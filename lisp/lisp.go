// lisp parsing code
package lisp

import (
	"bufio"
	"container/list"
	"fmt"
	"io"
	"reflect"
	"strings"
)

type Lexeme interface {
	fmt.Stringer
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
	var c byte
	var err error
	for c, err = buf.ReadByte(); err == nil; c, err = buf.ReadByte() {
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
	if err == io.EOF {
		return l, nil
	} else {
		return l, err
	}
}

type CoreNumericType int

type Number interface {
	Integer() int64
	Uinteger() uint64
	Float64() float64
	Float32() float32
	Type() reflect.Type
}
type Integer int64

func (this Integer) Integer() int64 {
	return int64(this)
}
func (this Integer) Uinteger() uint64 {
	return uint64(this)
}
func (this Integer) Float64() float64 {
	return float64(this)
}
func (this Integer) Float32() float32 {
	return float32(this)
}

func (this Integer) Type() reflect.Type {
	return reflect.TypeOf(this)
}

func (this Integer) String() string {
	return fmt.Sprintf("%d", this)
}

type Uinteger uint64

func (this Uinteger) Integer() int64 {
	return int64(this)
}
func (this Uinteger) Uinteger() uint64 {
	return uint64(this)
}
func (this Uinteger) Float64() float64 {
	return float64(this)
}
func (this Uinteger) Float32() float32 {
	return float32(this)
}

func (this Uinteger) Type() reflect.Type {
	return reflect.TypeOf(this)
}

func (this Uinteger) String() string {
	return fmt.Sprintf("%d", this)
}

type Float64 float64

func (this Float64) Integer() int64 {
	return int64(this)
}
func (this Float64) Uinteger() uint64 {
	return uint64(this)
}
func (this Float64) Float64() float64 {
	return float64(this)
}
func (this Float64) Float32() float32 {
	return float32(this)
}

func (this Float64) Type() reflect.Type {
	return reflect.TypeOf(this)
}

func (this Float64) String() string {
	return fmt.Sprintf("%f", this)
}

type Float32 float32

func (this Float32) Integer() int64 {
	return int64(this)
}
func (this Float32) Uinteger() uint64 {
	return uint64(this)
}
func (this Float32) Float64() float64 {
	return float64(this)
}
func (this Float32) Float32() float32 {
	return float32(this)
}

func (this Float32) Type() reflect.Type {
	return reflect.TypeOf(this)
}

func (this Float32) String() string {
	return fmt.Sprintf("%f", this)
}
