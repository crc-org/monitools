package main

import (
	"log"

	"github.com/code-ready/monitools/tools"
)

// Run a function from monitools/tools and observe
// a new file 'cpu.csv' in 'examples' dir after
// 5 seconds
func main() {

	cpuChan := make(chan error)

	go tools.RecordHostCPUUsage("cpu.csv", 5, 1, cpuChan)

	if <-cpuChan != nil {
		log.Fatalf("failed to record CPU percentage")
	}
}
