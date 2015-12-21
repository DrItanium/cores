// iris1 simulator
package main

import (
	"flag"
	"fmt"
	"github.com/DrItanium/cores/registration"
	"github.com/DrItanium/edgeworth"
	"io"
	"os"
)

var target = flag.String("target", "", "Target machine to simulate")
var listTargets = flag.Bool("list-targets", false, "List supported machines and exit")

func init() {
	// this does nothing but prevent compilation errors
	registration.Register()
}
func listRegisteredTargets() {
	fmt.Println("Supported targets: ")
	for _, val := range edgeworth.RegisteredMachines() {
		fmt.Println("\t - ", val)
	}
}
func main() {
	flag.Parse()
	if *listTargets {
		listRegisteredTargets()
		return
	}
	if *target == "" {
		fmt.Println("No target specified")
		return
	} else if !edgeworth.MachineExists(*target) {
		fmt.Printf("Specified target %s is not a supported target!\n", *target)
		listRegisteredTargets()
		return
	}
	if len(flag.Args()) != 1 {
		flag.Usage()
		fmt.Println("Only one input file accepted!")
	} else if mach, err0 := edgeworth.NewMachine(*target); err0 != nil {
		fmt.Println(err0)
	} else {
		// install the program
		done := make(chan error)
		done2 := make(chan error)
		data := make(chan byte)

		loadMemoryImage(flag.Args()[0], done, data)
		go func(mach edgeworth.Machine, err chan error, data chan byte) {
			err <- mach.InstallProgram(data)
		}(mach, done2, data)
		for i := 0; i < 2; i++ {
			select {
			case err := <-done:
				if err != nil {
					fmt.Printf("Error from rlasm: %s\n", err)
					return
				}
			case err := <-done2:
				if err != nil {
					fmt.Printf("Error from %s machine: %s\n", *target, err)
					return
				}
			}
		}
		if err := mach.Run(); err != nil {
			fmt.Printf("Something went wrong during machine execution: %s!", err)
			//TODO: dump image contents
		}
	}
}

func loadMemoryImage(path string, done chan error, data chan byte) {
	if file, err := os.Open(path); err != nil {
		go func(done chan error, err error) {
			done <- err
		}(done, err)
	} else {
		go func(file *os.File, done chan error, data chan byte) {
			defer file.Close()
			var err error
			storage := make([]byte, 1)
			for _, err0 := file.Read(storage); err0 != io.EOF; _, err0 = file.Read(storage) {
				if err0 != nil {
					err = err0
					break
				} else {
					data <- storage[0]
				}
			}
			close(data)
			done <- err
		}(file, done, data)
	}
}
