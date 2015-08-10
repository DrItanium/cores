package main

import (
	"bufio"
	"fmt"
	"github.com/DrItanium/cores/iris1"
	"github.com/DrItanium/cores/lisp"
	"github.com/DrItanium/cores/parse/keyword"
	//	"github.com/DrItanium/cores/parse/numeric"
)

const iris1BackendName = "iris1"

var iris1keywords *keyword.Parser
var iris1registers *keyword.Parser

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
	iris1registers = keyword.New()
	for i := 0; i < iris1.RegisterCount; i++ {
		iris1registers.AddKeyword(fmt.Sprintf("r%d", i))
	}
	iris1keywords = keyword.New()
}

func iris1Parse(l lisp.List, out *bufio.Writer) error {
	// now iterate through all the set of lisp lists
	for _, element := range l {
		// if we encounter an atom at the top level then we should ignore it
		switch element.(type) {

		}
	}
	return nil
}
