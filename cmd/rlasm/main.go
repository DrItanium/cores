// assembler for iris1 which uses a lisp syntax
package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/DrItanium/cores/lisp"
	"io"
	"os"
)

type Encoder interface {
	Encode(lisp.List, io.Writer) error
}

var target = flag.String("target", "", "Core to target")
var output = flag.String("output", "", "Output encoded asm to, default is standard out")

var backends = make(map[string]Encoder)

func main() {
	var out io.Writer
	var rawIn io.Reader
	if len(backends) == 0 {
		panic("No backends specified!")
	} else {
		flag.Parse()
		if *target == "" {
			flag.Usage()
			fmt.Println("Supported backends:")
			for key, _ := range backends {
				fmt.Println("\t- ", key)
			}
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
				}
			}
		}
	}
}
