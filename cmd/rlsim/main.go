// iris1 simulator
package main

import (
	"flag"
	"fmt"
	_ "github.com/DrItanium/cores/registration"
	"github.com/DrItanium/cores/registration/machine"
	"io/ioutil"
	"os"
)

var target = flag.String("target", "", "Target machine to simulate")
var listTargets = flag.Bool("list-targets", false, "List supported machines and exit")
var input = flag.String("input", "", "input file to be processed (leave blank for stdin)")
var debug = flag.Bool("debug", false, "enable/disable debugging")

func listRegisteredTargets() {
	fmt.Fprintln(os.Stderr, "Supported targets: ")
	for _, val := range machine.GetRegistered() {
		fmt.Fprintln(os.Stderr, "\t - ", val)
	}
}
func main() {
	if listTargets, listUsage, err, code := body(); err != nil {
		if listUsage {
			flag.Usage()
		}
		if listTargets {
			listRegisteredTargets()
		}
		if str := err.Error(); len(str) > 0 {
			fmt.Fprintln(os.Stderr, str)
		}
		os.Exit(code)
	}
}
func body() (bool, bool, error, int) {
	flag.Parse()
	if *listTargets {
		return true, false, fmt.Errorf(""), 1
	}
	if *target == "" {
		return true, true, fmt.Errorf("No target backend specified"), 2
	} else if !machine.IsRegistered(*target) {
		return true, false, fmt.Errorf("Specified target %s is not a supported target!", *target), 3
	} else {
		var o *os.File
		if *input == "" {
			o = os.Stdin
		} else {
			if file, err := os.Open(*input); err != nil {
				return false, false, err, 4
			} else {
				defer file.Close()
				o = file
			}
		}
		if mach, err0 := machine.New(*target); err0 != nil {
			return false, false, err0, 5
		} else {
			// install the program
			done, done2 := make(chan error), make(chan error)
			data := make(chan byte, 1024)

			loadMemoryImage(o, done, data)
			go func(mach machine.Machine, err chan error, data chan byte) {
				err <- mach.InstallProgram(data)
			}(mach, done2, data)
			for i := 0; i < 2; i++ {
				select {
				case err := <-done:
					if err != nil {
						return false, false, fmt.Errorf("Error from rlasm: %s", err), 6
					}
				case err := <-done2:
					if err != nil {
						return false, false, fmt.Errorf("Error from %s machine: %s", *target, err), 7
					}
				}
			}
			if err := mach.Run(); err != nil {
				fmt.Printf("Something went wrong during machine execution: %s!", err)
				return false, false, fmt.Errorf("Something went wrong during machine execution: %s!", err), 8
			}
		}
		return false, false, nil, 0
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
