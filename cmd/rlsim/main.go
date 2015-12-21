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

func init() {
	// this does nothing but prevent compilation errors
	registration.Register()
}
func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		fmt.Println("Only one input file accepted!")
		fmt.Println("Supported targets: ")
		for _, val := range edgeworth.RegisteredMachines() {
			fmt.Println("\t - ", val)
		}
	} else if mach, err0 := edgeworth.NewMachine(flag.Args()[0]); err0 != nil {
		fmt.Println(err0)
	} else {
		// install the program
		done := make(chan error)
		done2 := make(chan error)
		data := make(chan byte)

		loadMemoryImage(flag.Args()[1], done, data)
		go func(mach edgeworth.Machine, err chan error, data chan byte) {
			err <- mach.InstallProgram(data)
		}(mach, done2, data)
		if err, err2 := <-done, <-done2; err != nil || err2 != nil {
			if err != nil {
				fmt.Printf("ERROR during memory image load (tool side): %s\n", err)
			}
			if err2 != nil {
				fmt.Printf("ERROR during memory image load (machine side): %s\n", err2)
			}
		} else {
			if err := mach.Run(); err != nil {
				fmt.Printf("Something wen't wrong during machine execution: %s!", err)
				//TODO: dump image contents
			}
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
