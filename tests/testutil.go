package main

import (
	"fmt"

	"github.com/platform9/cctl/statefileutil"
)

func testReadStateFile() {
	cs, _ := statefileutil.ReadStateFile()
	if cs != nil {
		fmt.Println("ClusterState object ", cs)
	} else {
		fmt.Println("Problem acquiring cluster state")
	}
}

func main() {
	testReadStateFile()
}
