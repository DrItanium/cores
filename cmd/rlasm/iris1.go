package main

import (
	"bufio"
	"fmt"
	"github.com/DrItanium/cores/iris1"
	"github.com/DrItanium/cores/lisp"
	"github.com/DrItanium/cores/parse/keyword"
	"github.com/DrItanium/cores/parse/numeric"
)

const iris1BackendName = "iris1"

var keywords *keyword.Parser
var registers *keyword.Parser

type iris1Encoder func(lisp.List, *bufio.Writer) error

func (this iris1Encoder) Encode(l lisp.List, out *bufio.Writer) error {
	return this(l, out)
}

func init() {
	if _, ok := backends[iris1BackendName]; ok {
		// it shouldn't exist already!
		panic("The iris1 backend is already defined! Impossible!")
	} else {
		backends[iris1BackendName] = iris1Encoder(iris1Parse)
	}
	// setup the keywords and register parsers
	registers = keyword.New()
	for i := 0; i < iris1.RegisterCount; i++ {
		registers.AddKeyword(fmt.Sprintf("r%d", i))
	}
}

func iris1Parse(l lisp.List, out *bufio.Writer) error {
	return nil
}

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
