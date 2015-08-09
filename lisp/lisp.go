// lisp parsing code
package lisp

import (
	"bufio"
	"container/list"
	"encoding/binary"
	"fmt"
	"strings"
)

type Atom []byte

func (this Atom) String() string {
	return strings.TrimSpace(string(this))
}
func (this Atom) HexRepresentation(littleEndian bool) string {
	// this is an expensive operation!
	var str string
	var fn func(string, byte) string
	if littleEndian {
		// prefix
		fn = func(s string, v byte) string {
			return fmt.Sprintf("%X", v) + s
		}
	} else {
		// postfix
		fn = func(s string, v byte) string {
			return s + fmt.Sprintf("%X", v)
		}
	}
	for i := 0; i < len(this); i++ {
		str = fn(str, this[i])
	}
	return "0x" + str
}

func (this Atom) Uint16(bo binary.ByteOrder) uint16 {
	return bo.Uint16([]byte(this))
}

func (this Atom) Int16(bo binary.ByteOrder) int16 {
	return int16(this.Uint16(bo))
}

func (this Atom) Len() int {
	return len(this)
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

type AtomReconstructor func(Atom) string

func (this List) Reconstruct(onAtom AtomReconstructor) string {
	strs := make([]string, len(this))
	for ind, val := range this {
		switch val.(type) {
		case Atom:
			a := val.(Atom)
			strs[ind] = onAtom(a)
		case List:
			l := val.(List)
			strs[ind] = l.Reconstruct(onAtom)
		default:
			strs[ind] = fmt.Sprintf("%s", val)
		}
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
	updateList := func(lst List, c Atom) List {
		if len(c) > 0 {
			nC := make(Atom, len(c))
			copy(nC, c)
			return append(lst, nC)
		} else {
			return lst
		}
	}
	var l List
	var container Atom
	for c, err := buf.ReadByte(); err == nil; c, err = buf.ReadByte() {
		switch c {
		case '(':
			// save what we currently have to the current list
			l = updateList(l, container)
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
			} else {
				return updateList(l, container), nil
			}
		case ' ', '\n', '\t':
			l = updateList(l, container)
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
