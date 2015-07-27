// iris1 simulator
package main

import (
	"flag"
	"fmt"
	"github.com/DrItanium/cores/iris1"
)

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		fmt.Println("No input files provided!")
	} else {
		if core, err := iris1.New(); err != nil {
			fmt.Printf("ERROR: couldn't create a new iris1 core: %s\n", err)
		} else if inst, err := iris1.NewDecodedInstruction(iris1.InstructionGroupMisc, iris1.MiscOpSystemCall, iris1.SystemCallTerminate, 0, 0); err != nil {
			fmt.Printf("ERROR: couldn't create a new decoded instruction: %s\n", err)
		} else if err := core.Invoke(inst); err != nil {
			fmt.Printf("ERROR: execution failed of termination setup instruction: %s\n", err)
		} else {
			for !core.TerminateExecution() {
			}
		}
	}

}
