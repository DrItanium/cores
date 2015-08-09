// assembler for iris1 which uses a lisp syntax
package main

import (
	"bufio"
	"container/list"
	"fmt"
	"os"
	"strings"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	if list, err := Parse(in); err != nil {
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
	strs := make([]string, len(this))
	for ind, val := range this {
		strs[ind] = fmt.Sprintf("%s", val)
	}
	str := strings.Join(strs, " ")
	return fmt.Sprintf("(%s)", str)
}
func ParseString(input string) (List, error) {
	return Parse(bufio.NewReader(strings.NewReader(input)))
}
func Parse(buf *bufio.Reader) (List, error) {
	q := newListParser()
	if result, err := q.parse(buf); err != nil {
		return nil, err
	} else if q.hasUnfinishedLists() {
		return nil, fmt.Errorf("Opening paren found without corresponding closing paren!")
	} else {
		return result, nil
	}
}

type listParser struct {
	depthStack *list.List
	depth      uint64
}

func newListParser() *listParser {
	return &listParser{depthStack: list.New(), depth: 0}
}
func (this *listParser) hasUnfinishedLists() bool {
	return this.depth != 0
}
func (this *listParser) openList() error {
	oldDepth := this.depth
	this.depthStack.PushFront(this.depth)
	this.depth++
	if this.depth < oldDepth {
		return fmt.Errorf("parsing stack overflow!")
	} else {
		return nil
	}
}
func (this *listParser) closeList() error {
	if this.depthStack.Len() == 0 {
		return fmt.Errorf("parsing stack underflow!")
	} else {
		this.depthStack.Remove(this.depthStack.Front())
		this.depth--
		return nil
	}
}
func (this *listParser) parse(buf *bufio.Reader) (List, error) {
	var l List
	var container Atom
	for c, err := buf.ReadByte(); err == nil; c, err = buf.ReadByte() {
		switch c {
		case '(':
			// push the paren to the stack
			if err := this.openList(); err != nil {
				return nil, err
			} else if val, err0 := this.parse(buf); err0 != nil {
				return nil, err0
			} else {
				l = append(l, val)
			}
		case ')':
			if err := this.closeList(); err != nil {
				return nil, err
			} else if len(container) > 0 {
				nContainer := make(Atom, len(container))
				copy(nContainer, container)
				l = append(l, nContainer)
			}
			return l, nil
		case ' ', '\n', '\t':
			if len(container) > 0 {
				nContainer := make(Atom, len(container))
				copy(nContainer, container)
				l = append(l, nContainer)
			}
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
