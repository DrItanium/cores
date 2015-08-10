// assembler for iris1 which uses a lisp syntax
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

type Encoder interface {
	Encode(lisp.List, io.Reader, io.Writer) error
}

var target = flag.String("target", "", "Core to target")

var backends map[string]Encoder

func main() {
	if len(backends) == 0 {
		panic("No backends specified!")
	} else {
		flag.Parse()
		if *target == "" {
			flag.Usage()
			return
		} else {
			in := bufio.NewReader(os.Stdin)
			if list, err := lisp.Parse(in); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(list.Reconstruct(TagReconstruct))
			}
		}
	}
}

/*
func TagReconstruct(atom lisp.Atom) string {
	if IsKeyword(atom) {
		return fmt.Sprintf("(keyword %s)\n", atom)
	} else if numeric.IsHexNumber(atom.String()) {
		return fmt.Sprintf("(hex %s)\n", atom)
	} else if numeric.IsBinaryNumber(atom.String()) {
		return fmt.Sprintf("(binary %s)\n", atom)
	} else if numeric.IsDecimalNumber(atom.String()) {
		return fmt.Sprintf("(decimal %s)\n", atom)
	} else if IsRegister(atom) {
		return fmt.Sprintf("(register %s)\n", atom)
	} else {
		return fmt.Sprintf("(lexeme %s)\n", atom)
	}
}
func IsHexNumber(atom lisp.Atom) bool {
	return numeric.IsHexNumber(atom.String())
}
func IsBinaryNumber(atom lisp.Atom) bool {
	return numeric.IsBinaryNumber(atom.String())
}
func IsDecimalNumber(atom lisp.Atom) bool {
	return numeric.IsDecimalNumber(atom.String())
}
func IsRegister(atom lisp.Atom) bool {
	return registers.IsKeyword(atom.String())
}

func IsKeyword(atom lisp.Atom) bool {
	return keywords.IsKeyword(atom.String())
}
*/
