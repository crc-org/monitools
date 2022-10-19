package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// RecordHostCPUUsage returns a list of n cpu usage stats
// (in %) taken with nap breaks in between each poll
// filename : relative location of file to write into
// reps     : number of times to record CPU usage
// nap      : sleep between the reps
// c        : channel used to report back to the main process
func RecordHostCPUUsage(filename string, reps int, nap int, c chan error) {
	c <- recordHostCPUUsage(filename, reps, nap)
}

func recordHostCPUUsage(filename string, reps int, nap int) error {
	napLength := time.Duration(nap)

	qemuPidByte, err := exec.Command("ps", "-g", "qemu", "-o", "pid=").Output()
	if err != nil {
		return err
	}
	// Remove new line from the output
	qemuPid := strings.TrimSpace(string(qemuPidByte))
	if qemuPid != "" {
		// collect data
		cpuData := make([]float64, reps) // can't initialize array of length calculated at runtime :(
		for i := 0; i < reps; i++ {
			// get qemu's line, static output, only once
			out, err := exec.Command("top", "-bn1", "-p", qemuPid).Output()
			if err != nil {
				return err
			}
			outTail := strings.Split(string(out), "qemu")[1]
			cpu, _ := strconv.ParseFloat(strings.Fields(outTail)[6], 64)
			cpuData[i] = cpu
			time.Sleep(napLength * time.Second)
		}

		// create CSV file and write data to it
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		jsonCPU, _ := json.MarshalIndent(cpuData, "", " ")
		err = ioutil.WriteFile(filename, jsonCPU, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

// RecordTraffic returns a list of n cpu usage stats
// (in %) taken with nap breaks in between each poll
// filename : relative location of file to write into
// reps     : number of times to record CPU usage
// nap      : sleep between the reps
// c        : channel used to report back to the main process
func RecordTraffic(filename string, reps int, nap int, c chan error) {
	c <- recordTraffic(filename, reps, nap)
}

func recordTraffic(filename string, reps int, nap int) error {
	napLength := time.Duration(nap)

	// collect data
	var rxtxData [][]string
	ifFace := "crc"
	for i := 0; i < reps; i++ {
		// get qemu's line, static output, only once
		rxFileName := fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", ifFace)
		rx, err := ioutil.ReadFile(rxFileName)
		if err != nil {
			fmt.Errorf("Not able to read %s", rxFileName)
		}

		txFileName := fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", ifFace)
		tx, err := ioutil.ReadFile(txFileName)
		if err != nil {
			fmt.Errorf("Not able to read %s", txFileName)
		}

		rxtxData = append(rxtxData, []string{strings.TrimSpace(string(rx)), strings.TrimSpace(string(tx))})
		time.Sleep(napLength * time.Second)
	}

	// create CSV file and write data to it
	f, err := os.Create(filename)
	if err != nil {
		fmt.Errorf("could not create %s err: %s", filename, err)
	}
	defer f.Close()

	jsonRxTx, _ := json.MarshalIndent(rxtxData, "", " ")
	err = ioutil.WriteFile(filename, jsonRxTx, 0644)
	if err != nil {
		fmt.Errorf("Could not write data to %s", filename)
	}
	return nil
}

// GetCRIStatsFromVM returns the output of `sudo crictl stats -o json`
// from inside the CRC VM
// destinationDir : location where dump JSON file will be saved
// c              : channel to report routines completion/error
func GetCRIStatsFromVM(destinationDir string, c chan error) {
	c <- getCRIStatsFromVM(destinationDir)
}

func getCRIStatsFromVM(destinationDir string) error {
	cmdCrictl := exec.Command("ssh", "-i", "~/.crc/machines/crc/id_ecdsa",
		"-o StrictHostKeyChecking=no", "-o UserKnownHostsFile=/dev/null", "core@192.168.130.11",
		"sudo", "crictl", "stats", "-o", "json")
	out, err := cmdCrictl.Output() // out is []byte
	if err != nil {
		log.Printf("could not capture output of the command: %s", cmdCrictl)
	}

	t := time.Now()
	timestamp := t.Format("20060102150405")
	filename := filepath.Join(destinationDir, "crictl-stats-"+timestamp+".json")

	f, err := os.Create(filename)
	if err != nil {
		fmt.Errorf("could not create %s err: %s", filename, err)
	}
	defer f.Close()

	_, err = f.Write(out)
	if err != nil {
		fmt.Errorf("could not write to %s err: %s", filename, err)
	}
	return f.Sync()
}

func IsCRCRunning() bool {
	out, err := exec.Command("virsh", "-c", "qemu:///system", "domstate", "crc").Output()
	if err != nil {
		return false
	}
	if strings.TrimSpace(string(out)) != "running" {
		return false
	}
	return true
}

func GetNodeResource(path string, c chan error) {
	c <- parseNodeDescribeToJSON(path)
}

// declaring a struct
type NodeDescribe struct {
	// defining struct variables
	Capacity           corev1.ResourceList   `json:"capacity"`
	Allocatable        corev1.ResourceList   `json:"allocatable"`
	NodeInfo           corev1.NodeSystemInfo `json:"nodeInfo"`
	NonTerminatedPods  []PodInfo             `json:"nonTerminatedPods"`
	AllocatedResources []Resource            `json:"allocatedResources"`
}

type PodInfo struct {
	Namespace      string `json:"namespace"`
	Name           string `json:"name"`
	CPURequests    string `json:"cpuRequests"`
	MemoryRequests string `json:"memoryRequests"`
}

type Resource struct {
	Name     string `json:"name"`
	Requests string `json:"requests"`
}

func parseNodeDescribeToJSON(path string) error {
	var nodeDescribe NodeDescribe
	kubeconfig := filepath.Join(homedir.HomeDir(), ".crc", "machines", "crc", "kubeconfig")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	var totalCPU, totalMem resource.Quantity
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			var cpu, mem resource.Quantity
			for _, container := range pod.Spec.Containers {
				cpu.Add(*container.Resources.Requests.Cpu())
				mem.Add(*container.Resources.Requests.Memory())
			}
			p := PodInfo{Namespace: pod.Namespace,
				Name:           pod.Name,
				CPURequests:    cpu.String(),
				MemoryRequests: mem.String(),
			}
			nodeDescribe.NonTerminatedPods = append(nodeDescribe.NonTerminatedPods, p)
			totalCPU.Add(cpu)
			totalMem.Add(mem)
		}
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, node := range nodes.Items {
		nodeDescribe.Capacity = node.Status.Capacity
		nodeDescribe.Allocatable = node.Status.Allocatable
		nodeDescribe.NodeInfo = node.Status.NodeInfo
		r := []Resource{
			{
				Name:     "cpu",
				Requests: totalCPU.String(),
			},
			{
				Name:     "memory",
				Requests: totalMem.String(),
			},
		}
		nodeDescribe.AllocatedResources = r
	}

	file, err := json.MarshalIndent(&nodeDescribe, "", " ")
	if err != nil {
		fmt.Printf("error marshalling json: %v", err)
	}
	_ = ioutil.WriteFile(path, file, 0644)

	return nil
}

func timedClusterStart(pullSecretPath string, bundlePath string) (string, error) {

	cmd := exec.Command("crc", "start", "-b", bundlePath, "-p", pullSecretPath)
	s := time.Now()
	out, err := cmd.Output()
	log.Printf("%s", string(out))
	startDuration := time.Since(s)
	if err != nil {
		return "0h0m0.0s", err
	}

	return startDuration.String(), nil
}

func clusterDelete() error {

	cmd := exec.Command("crc", "delete", "-f")
	err := cmd.Run()

	return err
}

func clusterCleanup() error {
	cmd := exec.Command("crc", "cleanup")
	out, err := cmd.Output()
	if err != nil {
		log.Printf("%s", string(out))
	}

	return err
}

func clusterSetup(bundlePath string) error {
	cmd := exec.Command("crc", "setup", "-b", bundlePath)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("%s", string(out))
	}

	return err
}

func recordStartTimes(filename string, reps int, pullSecretPath string, bundlePath string) error {

	// crc cleanup
	err := clusterCleanup()
	if err != nil {
		log.Printf("Could not clean up the host: %s", err)
		return err
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Could not retrieve user's home directory: %s", err)
		return err
	}
	// remove config file
	err = os.Remove(filepath.Join(homeDir, ".crc", "crc.json"))
	if err != nil {
		log.Printf("Could not remove crc config: %s", err)
		return err
	}
	// crc setup
	err = clusterSetup(bundlePath)
	if err != nil {
		log.Printf("Could not set up the host: %s", err)
		return err
	}

	// collect data
	var startTimes []string
	for i := 0; i < reps; i++ {

		startTime, _ := timedClusterStart(pullSecretPath, bundlePath)
		startTimes = append(startTimes, startTime)

		err1 := clusterDelete()
		if err1 != nil {
			log.Printf("Failed to delete cluster: %s", err1)
			os.Exit(1)
		}
	}

	// create CSV file and write data to it
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("could not create %s err: %s", filename, err)
	}
	defer f.Close()

	jsonStartTimes, _ := json.MarshalIndent(startTimes, "", " ")
	err = ioutil.WriteFile(filename, jsonStartTimes, 0644)
	if err != nil {
		log.Printf("Could not write data to %s", filename)
	}

	return nil
}

// RecordStartTimes returns a list of n cpu usage stats
// (in %) taken with nap breaks in between each poll
// filename : relative location of file to write into
// reps     : number of times to record CPU usage
// c        : channel used to report back to the main process
func RecordStartTimes(filename string, reps int, c chan error, pullSecretPath string, bundlePath string) {
	c <- recordStartTimes(filename, reps, pullSecretPath, bundlePath)
}
