package tools

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// AvgCPUOverNSeconds returns average %CPU usage
// taken over n times with 1s intervals
func AvgCPUOverNSeconds(n int) (float64, error) {

	sumCPUPercent := 0.0

	for i := 0; i < n; i++ {
		cmdTop := exec.Command("top", "-bn1", "-u", "qemu")
		out, err := cmdTop.Output()
		if err != nil {
			return 0, err
		}

		outTail := strings.Split(string(out), "qemu")[1]
		cpuPercent, err := strconv.ParseFloat(strings.Fields(outTail)[6], 32)
		if err != nil {
			return 0, err
		}
		fmt.Printf("%f ", cpuPercent)
		sumCPUPercent += cpuPercent
		time.Sleep(1 * time.Second)
	}
	return sumCPUPercent / float64(n), nil
}
