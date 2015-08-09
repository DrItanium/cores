// assembler for iris1 which uses a lisp syntax
package main

import (
	"bufio"
	"fmt"
	"github.com/DrItanium/cores/lisp"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	if list, err := lisp.Parse(in); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(list)
	}
}
