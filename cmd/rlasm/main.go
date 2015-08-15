// assembler for iris1 which uses a lisp syntax
package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/DrItanium/cores/encoder"
	"github.com/DrItanium/cores/iris1"
	"github.com/DrItanium/cores/lisp"
	"io"
	"os"
)

var target = flag.String("target", "", "Core to target")
var output = flag.String("output", "", "Output encoded asm to, default is standard out")

var backends map[string]encoder.Encoder

func supportedBackends() {
	fmt.Println("Supported backends:")
	for key, _ := range backends {
		fmt.Println("\t- ", key)
	}
}
func main() {
	var out io.Writer
	var rawIn io.Reader
	flag.Parse()
	if *target == "" {
		flag.Usage()
		supportedBackends()
		return
	} else {
		rawIn = os.Stdin
		out = os.Stdout
		in := bufio.NewReader(rawIn)
		if list, err := lisp.Parse(in); err != nil {
			fmt.Println(err)
		} else {
			if enc, ok := backends[*target]; ok {
				if err := enc.Encode(list, out); err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Printf("ERROR: unknown target %s\n", *target)
				supportedBackends()
			}
		}
	}
}

func init() {
	backends = make(map[string]encoder.Encoder)
	backends["iris1"] = iris1.GetEncoder()

	// this should always be the last part of this init function
	if len(backends) == 0 {
		panic("No backends specified!")
	}
}
