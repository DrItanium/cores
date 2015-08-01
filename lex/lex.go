// generic lexing library for assembler syntax
package lex

import (
	"bufio"
	"fmt"
)

type Node struct {
	Value string
	Type  uint32
}

type Translator func(string) ([]Node, error)

func (this Translator) Translate(input *bufio.Scanner) ([]Node, error) {
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
