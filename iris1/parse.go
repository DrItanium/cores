package iris1

import (
	//"encoding/binary"
	"fmt"
	"github.com/DrItanium/cores/registration/parser"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type segment int

const (
	codeSegment segment = iota
	dataSegment
)

type ParsingRegistrar func(...interface{}) (parser.Parser, error)

func (this ParsingRegistrar) New(args ...interface{}) (parser.Parser, error) {
	return this(args)
}

type labelEntry struct {
	seg  segment
	addr Word
}

type labelMap map[string]labelEntry

func generateParser(a ...interface{}) (parser.Parser, error) {
	var p _parser
	if core, err := New(); err != nil {
		return nil, err
	} else {
		p.core = core
		p.labels = make(labelMap)
		return &p, nil
	}
}

func init() {
	parser.Register(RegistrationName(), ParsingRegistrar(generateParser))
}

type _parser struct {
	core               *Core
	labels             labelMap
	dataAddr, codeAddr Word
	currSegment        segment
}

func (this *_parser) Dump(pipe chan<- byte) error {
	return nil
}
func (this *_parser) Process() error {
	return nil
}

func (this *_parser) Parse(lines <-chan parser.Entry) error {
	for line := range lines {
		stmt := carveLine(line.Line)
		for _, str := range stmt {
			if err := str.Parse(); err != nil {
				return fmt.Errorf("Error: Line %d: %s\n", line.Index, err)
			}
		}
	}
	return nil
}

type nodeType int

const (
	TypeUnknown nodeType = iota
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

type node struct {
	Value interface{}
	Type  nodeType
}

func parseHexImmediate(str string) (Word, error) {
	val, err := strconv.ParseUint(str, 16, 16)
	return Word(val), err
}
func parseBinaryImmediate(str string) (Word, error) {
	val, err := strconv.ParseUint(str, 2, 16)
	return Word(val), err
}
func parseDecimalImmediate(str string) (Word, error) {
	val, err := strconv.ParseUint(str, 10, 16)
	return Word(val), err
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
func (this *node) parseLabel(val string) error {
	nVal := strings.TrimSuffix(val, ":")
	q, _ := utf8.DecodeRuneInString(nVal)
	if !unicode.IsLetter(q) {
		return fmt.Errorf("Label %s starts with a non letter %s!", nVal, q)
	} else {
		this.Type = TypeLabel
		this.Value = nVal
		// now parse the label as a entirely new node and see if we get a register back
		nod := node{Value: this.Value, Type: TypeUnknown}
		if err := nod.Parse(); err != nil {
			switch err.(type) {
			case *strconv.NumError:
				j := err.(*strconv.NumError)
				if j.Err == strconv.ErrRange {
					return fmt.Errorf("Label %s is interpreted as an out of range value! This is not allowed as it is ambiguous!", this.Value)
				} else if j.Err == strconv.ErrSyntax {
					// probably legal, will require an extra pass most likely
					return nil
				} else {
					return err
				}
			case *InvalidRegisterError:
				j := err.(*InvalidRegisterError)
				return fmt.Errorf("Label %s is interpreted as an out of range register! This is not allowed as it is ambiguous!", j.Value)
			default:
				return fmt.Errorf("Unkown error occurred: %s! Programmer failure!", err)
			}
		} else {
			return nil
		}
	}
}
func (this *node) parseHexImmediate(val string) error {
	this.Type = TypeHexImmediate
	if v, err := parseHexImmediate(val[2:]); err != nil {
		return err
	} else {
		this.Value = v
		return nil
	}
}
func (this *node) parseBinaryImmediate(val string) error {
	this.Type = TypeBinaryImmediate
	if v, err := parseBinaryImmediate(val[2:]); err != nil {
		return err
	} else {
		this.Value = v
		return nil
	}
}
func (this *node) parseImmediate(val string) error {
	this.Type = TypeImmediate
	if v, err := parseDecimalImmediate(val[1:]); err != nil {
		return err
	} else {
		this.Value = v
		return nil
	}
}
func (this *node) Parse() error {
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
		} else if strings.HasPrefix(val, "#x") {
			return this.parseHexImmediate(val)
		} else if strings.HasPrefix(val, "#b") {
			return this.parseBinaryImmediate(val)
		} else if strings.HasPrefix(val, "#") {
			return this.parseImmediate(val)
		} else {

		}
	}
	return nil
}

func (this *node) IsComment() bool {
	return this.Type == TypeComment
}
func (this *node) IsLabel() bool {
	return this.Type == TypeLabel
}

type statement []node

func (this *statement) Add(value string, t nodeType) {
	// always trim before adding
	str := strings.TrimSpace(value)
	if len(str) > 0 {
		*this = append(*this, node{Value: str, Type: t})
	}
}
func (this *statement) AddUnknown(value string) {
	this.Add(value, TypeUnknown)
}

func carveLine(line string) statement {
	// trim the damn line first
	data := strings.TrimSpace(line)
	var s statement
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
