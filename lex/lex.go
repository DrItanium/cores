// generic lexing library for assembler syntax
package lex

import (
	"fmt"
	"text/scanner"
)

type Node struct {
	Value string
	Type  uint32
}

type Translator func(string) ([]Node, error)

func (this Translator) Translate(input *scanner.Scanner) ([]Node, error) {
	var nodes []Node
	for input.Scan() {
		nodes := append(nodes, this(input.Text())...)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	} else {
		return nodes, nil
	}
}
