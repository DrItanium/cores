// assembler for iris1 which uses a lisp syntax
package main

import (
	"bufio"
	"fmt"
	"github.com/DrItanium/cores/iris1"
	"github.com/DrItanium/cores/lisp"
	"github.com/DrItanium/cores/parse/keyword"
	"github.com/DrItanium/cores/parse/numeric"
	"os"
)

var keywords *keyword.Parser
var registers *keyword.Parser

func init() {
	keywords = keyword.New()
	registers = keyword.New()
	keywords.AddKeyword("add")
	keywords.AddKeyword("addi")
	keywords.AddKeyword("label")
	keywords.AddKeyword("sub")
	keywords.AddKeyword("hex")
	keywords.AddKeyword("binary")
	keywords.AddKeyword("decimal")
	keywords.AddKeyword("register")

	for i := 0; i < iris1.RegisterCount; i++ {
		registers.AddKeyword(fmt.Sprintf("r%d", i))
	}
}
func main() {
	in := bufio.NewReader(os.Stdin)
	if list, err := lisp.Parse(in); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(list.Reconstruct(TagReconstruct))
	}
}
func TagReconstruct(atom lisp.Atom) string {
	if IsKeyword(atom) {
		return fmt.Sprintf("(keyword %s)", atom)
	} else if numeric.IsHexNumber(atom.String()) {
		return fmt.Sprintf("(hex %s)", atom)
	} else if numeric.IsBinaryNumber(atom.String()) {
		return fmt.Sprintf("(binary %s)", atom)
	} else if numeric.IsDecimalNumber(atom.String()) {
		return fmt.Sprintf("(decimal %s)", atom)
	} else if IsRegister(atom) {
		return fmt.Sprintf("(register %s)", atom)
	} else {
		return atom.String()
	}
}
func IsRegister(atom lisp.Atom) bool {
	return registers.IsKeyword(atom.String())
}

func IsKeyword(atom lisp.Atom) bool {
	return keywords.IsKeyword(atom.String())
}
