package tools

import (
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
)

// RecordHostCPUUsage returns a list of n cpu usage stats
// (in %) taken with nap breaks inbetween each poll
// filename : relative location of file to write into
// reps     : number of times to record CPU usage
// nap      : sleep between the reps
// c        : channel used to report back to the main process
func RecordHostCPUUsage(filename string, reps int, nap int, c chan bool) {

	napLength := time.Duration(nap)
	success := true

	// collect data
	cpuData := make([]float64, reps) // can't initialize array of length calculated at runtime :(
	for i := 0; i < reps; i++ {
		// get qemu's line, static output, only once
		cmdTop := exec.Command("top", "-bn1", "-u", "qemu")
		out, err := cmdTop.Output()
		strout := string(out)
		if err != nil {
			log.Printf("could not capture output of the `top` command")
			success = false
			cpuData[i] = -1.0
		} else if !strings.Contains(strout, "qemu") {
			log.Printf("there is no `qemu` process")
			cpuData[i] = -2.0
		} else {
			outTail := strings.Split(string(out), "qemu")[1]
			cpu, _ := strconv.ParseFloat(strings.Fields(outTail)[6], 64)
			cpuData[i] = cpu
		}
		time.Sleep(napLength * time.Second)
	}

	// create CSV file and write data to it
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("could not create %s err: %s", filename, err)
		success = false
	}
	defer f.Close()

	jsonCPU, _ := json.MarshalIndent(cpuData, "", " ")
	err = ioutil.WriteFile(filename, jsonCPU, 0644)
	if err != nil {
		log.Printf("Could not write data to %s", filename)
		success = false
	}

	c <- success
}

// RecordTraffic returns a list of n cpu usage stats
// (in %) taken with nap breaks inbetween each poll
// filename : relative location of file to write into
// reps     : number of times to record CPU usage
// nap      : sleep between the reps
// c        : channel used to report back to the main process
func RecordTraffic(filename string, reps int, nap int, c chan bool) {
	napLength := time.Duration(nap)
	success := true

	// collect data
	var rxtxData [][]string
	ifFace := "crc"
	for i := 0; i < reps; i++ {
		// get qemu's line, static output, only once
		rxFileName :=  fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", ifFace)
		rx, err := ioutil.ReadFile(rxFileName)
		if err != nil {
			log.Printf("Not able to read %s", rxFileName)
			success = false
		}

		txFileName :=  fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", ifFace)
		tx, err := ioutil.ReadFile(txFileName)
		if err != nil {
			log.Printf("Not able to read %s", txFileName)
			success = false
		}

		rxtxData = append(rxtxData, []string{strings.TrimSpace(string(rx)), strings.TrimSpace(string(tx))})
		time.Sleep(napLength * time.Second)
	}

	// create CSV file and write data to it
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("could not create %s err: %s", filename, err)
		success = false
	}
	defer f.Close()

	jsonRxTx, _ := json.MarshalIndent(rxtxData, "", " ")
	err = ioutil.WriteFile(filename, jsonRxTx, 0644)
	if err != nil {
		log.Printf("Could not write data to %s", filename)
		success = false
	}

	c <- success
}

// GetCRIStatsFromVM returns the output of `sudo crictl stats -o json`
// from inside the CRC VM
// destinationDir : location where dump JSON file will be saved
// c              : channel to report routines completion/error
func GetCRIStatsFromVM(destinationDir string, c chan error) {

	cmdCrictl := exec.Command("ssh", "-i", "~/.crc/machines/crc/id_ecdsa",
		"core@192.168.130.11",
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
		log.Printf("could not create %s err: %s", filename, err)
	}
	defer f.Close()

	_, err = f.Write(out)
	if err != nil {
		log.Printf("could not write to %s err: %s", filename, err)
	}
	f.Sync()

	c <- err
}

// RunCRCCommand runs a CRC command with args
func RunCRCCommand(cmdArgs []string) error {

	completeCommand := exec.Command("crc", cmdArgs...)
	_, err := completeCommand.Output()
	if err != nil {
		log.Printf("could not successfully run the command: %s\n err: %s", completeCommand, err)
		return err
	}

	return err
}
