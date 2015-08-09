// assembler for iris1 which uses a lisp syntax
package main

import (
	"bufio"
	"fmt"
	"github.com/DrItanium/cores/iris1"
	"github.com/DrItanium/cores/lisp"
	"github.com/DrItanium/cores/parse/keyword"
	"os"
)

var keywords *keyword.Parser

func init() {
	keywords = keyword.New()
	keywords.AddKeyword("add")
	keywords.AddKeyword("addi")
	keywords.AddKeyword("label")
	keywords.AddKeyword("sub")
	for i := 0; i < iris1.RegisterCount; i++ {
		keywords.AddKeyword(fmt.Sprintf("r%d", i))
	}
}
func main() {
	in := bufio.NewReader(os.Stdin)
	if list, err := lisp.Parse(in); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(list)
	}
}

func IsKeyword(atom lisp.Atom) bool {
	return keywords.IsKeyword(atom.String())
}
