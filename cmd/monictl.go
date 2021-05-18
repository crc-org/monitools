package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jsliacan/monitools/tools" // local tools package
)

func main() {

	// where to log
	t := time.Now()
	timestamp := t.Format("20060102150405")
	if err := os.MkdirAll("logs", 0766); err != nil {
		log.Fatal("Unable to create logs directory")
	}
	logFilePath := filepath.Join("logs", "monitools_"+timestamp+".log")
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Could not open a file for logging: %v", err)
	}
	log.SetOutput(logFile)

	// set up data folder
	dirName := fmt.Sprintf("data_%s", time.Now().Format("2006-01-02"))
	defaultDir := filepath.Join("data", dirName) // data/data_<date>

	// Command line flags
	var dirPath string
	flag.StringVar(&dirPath, "d", defaultDir, "destination directory")
	var numRepeats int
	flag.IntVar(&numRepeats, "n", 5, "number of checks of CPU load")
	var sleepLength int
	flag.IntVar(&sleepLength, "s", 1, "sleep between repeats [in seconds]")

	flag.Parse()

	// Local information
	//
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		log.Fatalf("Unable to create directory: %s", dirPath)
	}

	// Let the user know about the settings they're using
	fmt.Println("-------------")
	fmt.Println("Running monitoring tools with the following settings:")
	fmt.Printf("Data directory: %s\n", dirPath)
	fmt.Printf("Number of repeats: %d\n", numRepeats)
	fmt.Printf("Pauses between repeats: %ds\n", sleepLength)
	fmt.Printf("Logging into: %s\n", logFilePath)
	fmt.Println("-------------")

	cpuChan := make(chan error)
	trafficChan := make(chan error)
	crioChan := make(chan error)

	// ================
	// start collecting
	// ================

	// transmitted/received MiB on crc interface
	trafficFile := filepath.Join(dirPath, "traffic.json")
	go tools.RecordTraffic(trafficFile, numRepeats, sleepLength, trafficChan)
	log.Println("going to record traffic going in/out of the VM")

	// CPU usage by 'qemu' process
	cpuFile := filepath.Join(dirPath, "cpu.json")
	go tools.RecordHostCPUUsage(cpuFile, numRepeats, sleepLength, cpuChan)
	log.Println("going to record CPU usage percentage attributed to qemu")

	// CRI-O stats as reported by 'crictl'
	go tools.GetCRIStatsFromVM(dirPath, crioChan)
	log.Println("going to retrieve crictl stats from the CRC VM")

	// ================
	// done collecting
	// ================

	if err := <-trafficChan; err != nil {
		log.Fatalf("failed to record traffic flow %s", err)
	} else {
		log.Printf("recorded traffic (RX/TX) %d times at %d sec intervals", numRepeats, sleepLength)
	}

	if err := <-cpuChan; err != nil {
		log.Fatalf("failed to record CPU percentage %s", err)
	} else {
		log.Printf("recorded CPU usage percentage %d times at %d sec intervals", numRepeats, sleepLength)
	}

	if err := <-crioChan; err != nil {
		log.Fatalf("could not retrieve crictl stats: %s", err)
	} else {
		log.Println("crictl stats successfully retrieved")
	}
}
