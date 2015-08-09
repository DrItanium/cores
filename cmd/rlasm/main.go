// assembler for iris1 which uses a lisp syntax
package main

import (
	"bufio"
	"container/list"
	"encoding/binary"
	"fmt"
	"github.com/DrItanium/cores/lisp"
	"os"
	"strings"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	if list, err := lisp.Parse(in); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(list)
	}
}
