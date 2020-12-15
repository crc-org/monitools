package main

import (
	"fmt"

	"monitools/tools" // local tools package
)

func main() {

	fmt.Println(tools.AvgCPUOverNSeconds(5))

}
