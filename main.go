package main

import (
	"fmt"
	"log"
	"monitools/tools" // local tools package
	"os"
	"path/filepath"
	"time"
)

// Example program using monitools functions
func main() {

	// Required to push to Github code-ready/crc-data
	githubRepo := "crc-data"
	githubTokenLocation := os.Getenv("GITHUB_TOKEN_LOCATION")
	if githubTokenLocation == "" {
		log.Println("Need to set GITHUB_TOKEN_LOCATION first")
	}
	githubOrg := "code-ready"
	githubUser := "Justin Case" // not sure if this is needed
	githubEmail := "jsliacan@redhat.com"

	// Local information
	dirName := fmt.Sprintf("%s%s", "data_", time.Now().Format("2006-01-02"))
	dirPath := filepath.Join("data", dirName) // data/data_<date>
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.Mkdir(dirPath, 0777)
	}
	// Set frequencies and reps
	napLength := 1 // in sec
	numRepeats := 5

	cpuChan := make(chan bool)
	crioChan := make(chan error)

	/*
		// setup & start
		err := tools.RunCRCCommand([]string{"setup"})
		if err != nil {
			os.Exit(1)
		}

		err = tools.RunCRCCommand([]string{"start"})
		if err != nil {
			os.Exit(1)
		}
	*/

	// ================
	// start collecting
	// ================

	// CPU usage by 'qemu' process
	cpuFile := filepath.Join(dirPath, "cpu.csv")
	go tools.RecordHostCPUUsage(cpuFile, numRepeats, napLength, cpuChan)
	log.Println("going to record CPU usage percentage attributed to qemu")

	// CRI-O stats as reported by 'crictl'
	go tools.GetCRIStatsFromVM(dirPath, crioChan)
	log.Println("going to retrieve crictl stats from the CRC VM")

	// ================
	// done collecting
	// ================

	if <-cpuChan != true {
		log.Fatalf("failed to record CPU percentage")
	} else {
		log.Printf("recorded CPU usage percentage %d times at %d sec intervals", numRepeats, napLength)
	}

	if err := <-crioChan; err != nil {
		log.Fatalf("could not retrieve crictl stats: %s", err)
	} else {
		log.Println("crictl stats successfully retrieved")
	}

	err :=	tools.PushTodaysData(dirPath, githubRepo, githubTokenLocation, githubOrg, githubUser, githubEmail)
	if err != nil {
		fmt.Println(err)
	}
	
	/*
		// stop & delete & clean up
		err = tools.RunCRCCommand([]string{"stop", "-f"})
		if err != nil {
			os.Exit(1)
		}

		err = tools.RunCRCCommand([]string{"delete", "-f"})
		if err != nil {
			os.Exit(1)
		}

		err = tools.RunCRCCommand([]string{"cleanup"})
		if err != nil {
			os.Exit(1)
		}
	*/
}
