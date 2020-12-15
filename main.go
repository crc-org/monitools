package main

import (
	"log"
	"monitools/tools" // local tools package
)

// Assume CRC cluster is Running
func main() {

	napLength := 1 // in sec
	numRepeats := 5

	bchan := make(chan bool)
	cchan := make(chan error)

	// SEND COLLECTORS COLLECTING ============

	// CPU usage by 'qemu' process
	go tools.RecordHostCPUUsage(numRepeats, napLength, bchan)
	log.Println("going to record CPU usage percentage attributed to qemu")

	// CRI-O stats as reported by 'crictl'
	go tools.GetCRIStatsFromVM(cchan)
	log.Println("going to retrieve crictl stats from the CRC VM")

	// DONE (report back) =============
	if <-bchan != true {
		log.Fatalf("failed to record CPU percentage")
	} else {
		log.Printf("recorded CPU usage percentage %d times at %d sec intervals", numRepeats, napLength)
	}

	if err := <-cchan; err != nil {
		log.Fatalf("could not retrieve crictl stats: %s", err)
	} else {
		log.Println("crictl stats successfully retrieved")
	}

	//err := tools.RunCRCCommand([]string{"stop", "-f"})

}
