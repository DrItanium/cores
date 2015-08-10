// assembler for iris1 which uses a lisp syntax
package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/DrItanium/cores/lisp"
	"os"
)

type Encoder interface {
	Encode(lisp.List, *bufio.Writer) error
}

var target = flag.String("target", "", "Core to target")

var backends = make(map[string]Encoder)

func main() {
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
			in := bufio.NewReader(os.Stdin)
			if list, err := lisp.Parse(in); err != nil {
				fmt.Println(err)
			} else {
				if enc, ok := backends[*target]; ok {
					if err := enc.Encode(list, nil); err != nil {
						fmt.Println(err)
					}
				} else {
					fmt.Printf("ERROR: unknown target %s\n", *target)
				}
			}
		}
	}
}
