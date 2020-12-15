package tools

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// AvgCPUOverNSeconds returns average %CPU usage
// taken over n times with 1s intervals
func AvgCPUOverNSeconds(n int, c chan float64) {

	sumCPUPercent := 0.0

	for i := 0; i < n; i++ {
		cmdTop := exec.Command("top", "-bn1", "-u", "qemu")
		out, err := cmdTop.Output()
		if err != nil {
			log.Fatalf("could not capture output of the `top` command")
		}

		outTail := strings.Split(string(out), "qemu")[1]
		cpuPercent, err := strconv.ParseFloat(strings.Fields(outTail)[6], 32)
		if err != nil {
			log.Fatalf("could not parse CPU percentage from the output of the `top` command")
		}
		sumCPUPercent += cpuPercent
		time.Sleep(1 * time.Second)
	}

	c <- sumCPUPercent / float64(n)

}

// RecordHostCPUUsage returns a list of n cpu usage stats
// (in %) taken with nap breaks inbetween each poll
func RecordHostCPUUsage(reps int, nap int, c chan bool) {

	napLength := time.Duration(nap)

	filename := filepath.Join("data", "cpu.csv")
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("could not create %s err: %s", filename, err)
	}
	defer f.Close()

	for i := 0; i < reps; i++ {
		// get qemu's line, static output, only once
		cmdTop := exec.Command("top", "-bn1", "-u", "qemu")
		out, err := cmdTop.Output()
		if err != nil {
			log.Fatalf("could not capture output of the `top` command")
		}

		outTail := strings.Split(string(out), "qemu")[1]
		cpuPercent := strings.Fields(outTail)[6]

		// append to file
		_, err = f.WriteString(cpuPercent + "\n")
		if err != nil {
			log.Fatalf("could not write to %s err: %s", filename, err)
		}
		f.Sync()

		time.Sleep(napLength * time.Second)
	}

	c <- true
}

// GetCRIStatsFromVM returns the output of `sudo crictl stats -o yaml`
// from inside the CRC VM
func GetCRIStatsFromVM(c chan error) {

	cmdCrictl := exec.Command("ssh", "-i", "~/.crc/machines/crc/id_ecdsa", "core@192.168.130.11", "sudo", "crictl", "stats", "-o", "yaml")
	out, err := cmdCrictl.Output() // out is []byte
	if err != nil {
		log.Fatalf("could not capture output of the command: %s", cmdCrictl)
	}

	t := time.Now()
	timestamp := t.Format("20060102150405")
	filename := filepath.Join("data", "crictl-stats-"+timestamp+".yaml")

	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("could not create %s err: %s", filename, err)
	}
	defer f.Close()

	_, err = f.Write(out)
	if err != nil {
		log.Fatalf("could not write to %s err: %s", filename, err)
	}
	f.Sync()

	c <- err
}

// RunCRCCommand takes a list of string argument sto `crc` command
// and returns exitcode
func RunCRCCommand(cmdArgs []string) error {

	completeCommand := exec.Command("crc", cmdArgs...)
	_, err := completeCommand.Output()
	if err != nil {
		log.Fatalf("could not successfully run the command: %s\n err: %s", completeCommand, err)
	}

	return err
}
