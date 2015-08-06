// assembler for iris1 which uses a lisp syntax
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	if list, err := New(in); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(list)
	}
}

type Atom []byte

func (this Atom) String() string {
	return strings.TrimSpace(string(this))
}

type List []interface{}

func (this List) String() string {
	out := "("
	for _, val := range this {
		switch val.(type) {
		case List:
			l := val.(List)
			out += l.String()
		default:
			out += fmt.Sprintf("%s ", val)
		}
	}
	out = strings.TrimSpace(out)
	out += ")"
	return out
}
func New(buf *bufio.Reader) (List, error) {
	var l List
	var container Atom
	for c, err := buf.ReadByte(); err == nil; c, err = buf.ReadByte() {
		switch c {
		case '(':
			if val, err0 := New(buf); err0 != nil {
				return nil, err0
			} else {
				l = append(l, val)
			}
		case ')':
			nContainer := make(Atom, len(container))
			copy(nContainer, container)
			l = append(l, nContainer)
			return l, nil
		case ' ', '\n', '\t':
			nContainer := make(Atom, len(container))
			copy(nContainer, container)
			l = append(l, nContainer)
			container = make(Atom, 0)
		case ';':
			// read the rest of the line
			if _, err := buf.ReadString('\n'); err != nil {
				return nil, err
			}
		default:
			container = append(container, c)
		}
	}
	return l, nil
}
