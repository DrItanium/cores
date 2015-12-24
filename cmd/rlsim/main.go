// iris1 simulator
package main

import (
	"flag"
	"fmt"
	"github.com/DrItanium/cores/registration"
	"github.com/DrItanium/cores/registration/machine"
	"io/ioutil"
	"os"
)

var target = flag.String("target", "", "Target machine to simulate")
var listTargets = flag.Bool("list-targets", false, "List supported machines and exit")
var input = flag.String("input", "", "input file to be processed (leave blank for stdin)")
var debug = flag.Bool("debug", false, "enable/disable debugging")

func init() {
	// this does nothing but prevent compilation errors
	registration.Register()
}
func listRegisteredTargets() {
	fmt.Println("Supported targets: ")
	for _, val := range machine.GetRegistered() {
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
		fmt.Println("No target backend specified")
		flag.Usage()
		listRegisteredTargets()
		return
	} else if !machine.IsRegistered(*target) {
		fmt.Printf("Specified target %s is not a supported target!\n", *target)
		listRegisteredTargets()
		return
	} else {
		var o *os.File
		if *input == "" {
			o = os.Stdin
		} else {
			if file, err := os.Open(*input); err != nil {
				fmt.Println(err)
				return
			} else {
				defer file.Close()
				o = file
			}
		}
		if mach, err0 := machine.New(*target); err0 != nil {
			fmt.Println(err0)
		} else {
			// install the program
			done := make(chan error)
			done2 := make(chan error)
			data := make(chan byte, 1024)

			loadMemoryImage(o, done, data)
			go func(mach machine.Machine, err chan error, data chan byte) {
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
}

func loadMemoryImage(file *os.File, done chan error, data chan byte) {
	go func(file *os.File, done chan error, data chan byte) {
		var err error
		if b, e := ioutil.ReadAll(file); e != nil {
			err = e
		} else {
			for _, by := range b {
				data <- by
			}
		}
		close(data)
		done <- err
	}(file, done, data)
}
