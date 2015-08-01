// assembler for iris1
package main

import (
	"bufio"
	"fmt"
	"github.com/DrItanium/cores/iris1"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for count := 0; scanner.Scan(); count++ {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		} else {
			stmt := carveLine(line)
			for _, str := range stmt {
				if err := str.Parse(); err != nil {
					fmt.Printf("Error: Line %d: %s\n", count+1, err)
					return
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

type NodeType int

const (
	TypeUnknown NodeType = iota
	TypeEquals
	TypeComma
	TypeLabel
	TypeRegister
	TypeImmediate
	TypeBinaryImmediate
	TypeHexImmediate
	TypeComment
	TypeSymbol
)

type Node struct {
	Value interface{}
	Type  NodeType
}

func parseHexImmediate(str string) (iris1.Word, error) {
	val, err := strconv.ParseUint(str, 16, 16)
	return iris1.Word(val), err
}
func parseBinaryImmediate(str string) (iris1.Word, error) {
	val, err := strconv.ParseUint(str, 2, 16)
	return iris1.Word(val), err
}
func parseDecimalImmediate(str string) (iris1.Word, error) {
	val, err := strconv.ParseUint(str, 10, 16)
	return iris1.Word(val), err
}
func parseRegisterValue(str string) (byte, error) {
	val, err := strconv.ParseUint(str, 10, 8)
	return byte(val), err
}

type InvalidRegisterError struct {
	Value string
}

func (this *InvalidRegisterError) Error() string {
	return fmt.Sprintf("Register %s is not a valid register!", this.Value)
}
func InvalidRegister(value string) error {
	return &InvalidRegisterError{Value: value}
}
func (this *Node) parseLabel(val string) error {
	nVal := strings.TrimSuffix(val, ":")
	q, _ := utf8.DecodeRuneInString(nVal)
	if !unicode.IsLetter(q) {
		return fmt.Errorf("Label %s starts with a non letter %s!", nVal, q)
	} else {
		this.Type = TypeLabel
		this.Value = nVal
		// now parse the label as a entirely new node and see if we get a register back
		nod := Node{Value: this.Value, Type: TypeUnknown}
		if err := nod.Parse(); err != nil {
			switch err.(type) {
			case *strconv.NumError:
				j := err.(*strconv.NumError)
				if j.Err == strconv.ErrRange {
					return fmt.Errorf("Label %s is interpreted as an out of range value! This is not allowed as it is ambiguous!", this.Value)
				} else {
					return err
				}
			case *InvalidRegisterError:
				j := err.(*InvalidRegisterError)
				return fmt.Errorf("Label %s is interpreted as an out of range register! This is not allowed as it is ambiguous!", j.Value)
			default:
				return fmt.Errorf("Unkown error occurred: %s! Programmer failure!", err)
			}
		} else if nod.Type == TypeRegister {
			return fmt.Errorf("Label %s has the same name as register %s. This is not allowed!", this.Value)
		}
	}
	return nil
}
func (this *Node) parseRegister(val string) error {
	// convert this to a byte
	if v, err := parseRegisterValue(val[1:]); err != nil {
		switch err.(type) {
		case *strconv.NumError:
			j := err.(*strconv.NumError)
			if j.Err == strconv.ErrSyntax {
				this.Type = TypeSymbol
			} else {
				return InvalidRegister(val)
			}
		default:
			return err
		}
	} else {
		this.Type = TypeRegister
		this.Value = v
	}
	return nil
}
func (this *Node) parseHexImmediate(val string) error {
	this.Type = TypeHexImmediate
	if v, err := parseHexImmediate(val[2:]); err != nil {
		return err
	} else {
		this.Value = v
	}
	return nil
}
func (this *Node) parseBinaryImmediate(val string) error {
	this.Type = TypeBinaryImmediate
	if v, err := parseBinaryImmediate(val[2:]); err != nil {
		return err
	} else {
		this.Value = v
	}
	return nil
}
func (this *Node) Parse() error {
	if this.Type == TypeUnknown {
		val := this.Value.(string)
		if val == "=" {
			this.Type = TypeEquals
		} else if val == "," {
			this.Type = TypeComma
		} else if strings.HasSuffix(val, ":") {
			return this.parseLabel(val)
		} else if strings.HasPrefix(val, ";") {
			this.Type = TypeComma
			this.Value = strings.TrimPrefix(val, ";")
		} else if strings.HasPrefix(val, "r") {
			return this.parseRegister(val)
		} else if strings.HasPrefix(val, "0x") {
			return this.parseHexImmediate(val)
		} else if strings.HasPrefix(val, "0b") {
			return this.parseBinaryImmediate(val)
		}
	}
	return nil
}

func (this *Node) IsComment() bool {
	return this.Type == TypeComment
}
func (this *Node) IsLabel() bool {
	return this.Type == TypeLabel
}

type Statement []Node

func (this *Statement) Add(value string, t NodeType) {
	// always trim before adding
	str := strings.TrimSpace(value)
	if len(str) > 0 {
		*this = append(*this, Node{Value: str, Type: t})
	}
}
func (this *Statement) AddUnknown(value string) {
	this.Add(value, TypeUnknown)
}

func carveLine(line string) Statement {
	// trim the damn line first
	data := strings.TrimSpace(line)
	var s Statement
	if len(data) == 0 {
		return s
	}
	oldStart := 0
	start := 0
	// skip the strings at the beginning
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRuneInString(data[start:])
		if unicode.IsSpace(r) {
			s.AddUnknown(data[oldStart:start])
			oldStart = start
		} else if r == '=' {
			s.AddUnknown(data[oldStart:start])
			s.Add("=", TypeEquals)
			oldStart = start + width
		} else if r == ',' {
			s.AddUnknown(data[oldStart:start])
			s.Add(",", TypeComma)
			oldStart = start + width
		} else if r == ';' {
			// consume the rest of the data
			s.AddUnknown(data[oldStart:start])
			// then capture the comment
			s.Add(data[start:], TypeComment)
			oldStart = start
			break
		}
	}
	if oldStart < start {
		s.AddUnknown(data[oldStart:])
	}
	return s
}
