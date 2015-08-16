package main

import (
	"flag"
	"fmt"
	"github.com/DrItanium/cores/iris1"
	"github.com/DrItanium/cores/translation"
	"io"
	"os"
)

var target = flag.String("target", "", "Core to target")
var backends map[string]translation.Decoder

func supportedBackends() {
	fmt.Println("Supported backends:")
	for key, _ := range backends {
		fmt.Println("\t- ", key)
	}
}
func init() {
	backends = make(map[string]translation.Decoder)
	backends["iris1"] = iris1.GetDecoder()

	// this should always be the last part of this init function
	if len(backends) == 0 {
		panic("No backends specified!")
	}
}
func main() {
	var rawIn io.Reader
	flag.Parse()
	if *target == "" {
		flag.Usage()
		supportedBackends()
		return
	} else {
		rawIn = os.Stdin
		if dec, ok := backends[*target]; ok {
			if lst, err := dec.Decode(rawIn); err != nil {
				fmt.Println(err)
			} else {
				for _, val := range lst {
					fmt.Println(val)
				}
			}
		} else {
			fmt.Printf("ERROR: unknown target %s\n", *target)
			supportedBackends()
		}
	}
}
