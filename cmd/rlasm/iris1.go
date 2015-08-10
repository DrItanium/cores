package main

import (
	"fmt"
	"github.com/DrItanium/cores/iris1"
	"github.com/DrItanium/cores/lisp"
	"github.com/DrItanium/cores/parse/keyword"
	"io"
	"log"
	//	"github.com/DrItanium/cores/parse/numeric"
)

const iris1BackendName = "iris1"

var iris1keywords *keyword.Parser
var iris1registers *keyword.Parser

type iris1Encoder func(lisp.List, io.Writer) error

func (this iris1Encoder) Encode(l lisp.List, out io.Writer) error {
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
	// arithmetic ops
	iris1keywords.AddKeywordList([]string{
		"add",
		"addi",
		"sub",
		"subi",
		"mul",
		"muli",
		"div",
		"divi",
		"rem",
		"remi",
		"shl",
		"shli",
		"shr",
		"shri",
		"and",
		"or",
		"not",
		"xor",
		"incr",
		"decr",
		"double",
		"halve",
	})
	// compare ops
	iris1keywords.AddKeywordList([]string{
		"eq",
		"eq-and",
		"eq-or",
		"eq-xor",
		"neq",
		"neq-and",
		"neq-or",
		"neq-xor",
		"lt",
		"lt-and",
		"lt-or",
		"lt-xor",
		"gt",
		"gt-and",
		"gt-or",
		"gt-xor",
		"le",
		"le-and",
		"le-or",
		"le-xor",
		"ge",
		"ge-and",
		"ge-or",
		"ge-xor",
	})
	// jump operations
	iris1keywords.AddKeywordList([]string{
		"goto-imm",
		"call-imm",
		"goto-reg",
		"call-reg",
		"goto-imm-if1",
		"call-imm-if1",
		"goto-reg-if1",
		"call-reg-if1",
		"call-select-if1",
		"goto-select-if1",
		"goto-imm-if0",
		"call-imm-if0",
		"goto-reg-if0",
		"call-reg-if0",
		"call-select-if0",
		"goto-select-if0",
	})
	// move operations
	iris1keywords.AddKeywordList([]string{
		"move",
		"swap",
		"swap-reg-addr",
		"swap-addr-addr",
		"swap-reg-mem",
		"swap-addr-mem",
		"set",
		"load",
		"load-mem",
		"store",
		"store-addr",
		"store-mem",
		"store-imm",
		"push",
		"push-imm",
		"pop",
		"peek",
	})
	// misc operations
	iris1keywords.AddKeyword("system")
	// directives
	iris1keywords.AddKeywordList([]string{
		"label",
		"org",
		"segment",
		"value",
		"string",
	})

}

type iris1ExtendedCore struct {
	Core   *iris1.Core
	Labels map[string]iris1.Word
}

func iris1Parse(l lisp.List, out io.Writer) error {
	// now iterate through all the set of lisp lists
	var core iris1ExtendedCore
	if c, err := iris1.New(); err != nil {
		return err
	} else {
		core.Core = c
		core.Labels = make(map[string]iris1.Word)
	}

	for _, element := range l {
		// if we encounter an atom at the top level then we should ignore it
		switch typ := element.(type) {
		case lisp.Atom:
			log.Printf("Ignoring atom %s", element)
		case lisp.List:
			nList := element.(lisp.List)
			if err := iris1_ParseList(&core, nList, out); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unknown type %t", typ)

		}
	}
	return nil
}

func iris1_ParseList(core *iris1ExtendedCore, l lisp.List, out io.Writer) error {
	// use the first arg as the op and the rest as arguments
	if len(l) == 0 {
		return nil
	}
	first := l[0]
	//rest := l[1:]
	switch t := first.(type) {
	case lisp.Atom:
		atom := first.(lisp.Atom)
		if iris1keywords.IsKeyword(atom.String()) {
			// determine what kind of operation we are looking at
			log.Printf("%s", atom)
		} else {
			return fmt.Errorf("First argument (%s) is not a keyword", atom)
		}
	default:
		return fmt.Errorf("ERROR: first argument (%s) of operation is not an atom (%t),", first, t)
	}
	return nil
}
