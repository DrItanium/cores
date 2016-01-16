// implementation of ajvondrak's xand core
package xand

import (
	"fmt"
	"github.com/DrItanium/cores/registration/machine"
	"github.com/DrItanium/cores/registration/parser"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func RegistrationName() string {
	return "xand"
}

func Register() {}

type MachineRegistrar func(...interface{}) (machine.Machine, error)

func (this MachineRegistrar) New(args ...interface{}) (machine.Machine, error) {
	return this(args)
}

func generateCore(a ...interface{}) (machine.Machine, error) {
	return New()
}

func init() {
	machine.Register(RegistrationName(), MachineRegistrar(generateCore))
}

type Core struct {
	pc     int8
	ir     [3]int8
	memory [128]int8
}

func (this *Core) fetch() bool {
	if (this.pc < 0) || (int(this.pc+2) >= len(this.memory)) {
		return false
	} else {
		this.ir[0] = this.memory[this.pc]
		this.ir[1] = this.memory[this.pc+1]
		this.ir[2] = this.memory[this.pc+2]
		return this.ir[0] >= 0 && this.ir[1] >= 0 && this.ir[2] >= 0
	}
}

func (this *Core) Run() error {
	for this.fetch() {
		// the xand operation it self
		this.memory[this.ir[0]] = this.memory[this.ir[0]] - this.memory[this.ir[1]]
		if this.memory[this.ir[0]] <= 0 {
			this.pc = this.ir[2]
		} else {
			this.pc += 3
		}
	}
	return nil
}

func (this *Core) Startup() error {
	return nil
}

func (this *Core) Shutdown() error {
	return nil
}

func (this *Core) GetDebugStatus() bool {
	return false
}

func (this *Core) SetDebug(_ bool) {

}

func (this *Core) InstallProgram(input <-chan byte) error {
	// read 128 bytes
	for i := 0; i < 128; i++ {
		if value, more := <-input; !more {
			return fmt.Errorf("Not a complete xand memory image")
		} else {
			this.memory[i] = int8(value)
		}
	}
	return nil
}

func (this *Core) Dump(output chan<- byte) error {
	for _, value := range this.memory {
		output <- byte(value)
	}
	return nil
}

func New() (*Core, error) {
	return &Core{}, nil
}

type ParsingRegistrar func(...interface{}) (parser.Parser, error)

func (this ParsingRegistrar) New(args ...interface{}) (parser.Parser, error) {
	return this(args)
}

func generateParser(a ...interface{}) (parser.Parser, error) {
	var p _parser
	if core, err := New(); err != nil {
		return nil, err
	} else {
		p.core = core
		p.labels = make(map[string]int8)
		return &p, nil
	}
}

func init() {
	parser.Register(RegistrationName(), ParsingRegistrar(generateParser))
}

type _parser struct {
	core       *Core
	labels     map[string]int8
	statements []statement
}

func (this *_parser) Dump(pipe chan<- byte) error {
	return this.core.Dump(pipe)
}

func (this *_parser) Parse(lines <-chan parser.Entry) error {
	for line := range lines {
		stmt := carveLine(line.Line)
		stmt.index = line.Index
		this.statements = append(this.statements, stmt)
	}
	return nil
}

type nodeType int

func (this nodeType) immediate() bool {
	return this == typeImmediate
}
func (this nodeType) comment() bool {
	return this == typeComment
}

const (
	typeId nodeType = iota
	typeImmediate
	typeLabel
	typeComment
	keywordXand
	keywordDotDotDot
)

type node struct {
	Value interface{}
	Type  nodeType
}

func parseDecimalImmediate(str string) (int8, error) {
	val, err := strconv.ParseInt(str, 10, 8)
	return int8(val), err
}

func (this *node) parseLabel(val string) error {
	nVal := strings.TrimSuffix(val, ":")
	q, _ := utf8.DecodeRuneInString(nVal)
	if !unicode.IsLetter(q) {
		return fmt.Errorf("Label %s starts with a non letter %s!", nVal, q)
	} else {
		this.Type = typeLabel
		this.Value = nVal
		// now parse the label as a entirely new node and see if we get a register back
		if nVal == "xand" {
			return fmt.Errorf("Can't name a label xand")
		} else {
			return nil
		}
	}
}

func (this *node) parseImmediate(val string) error {
	this.Type = typeImmediate
	if v, err := parseDecimalImmediate(val[1:]); err != nil {
		return err
	} else {
		this.Value = v
		return nil
	}
}

var keywords = map[string]nodeType{
	"xand": keywordXand,
	"...":  keywordDotDotDot,
}

func (this *node) parseGeneric(val string) error {
	if v, ok := keywords[val]; ok {
		this.Type = v
		return nil
	} else {
		return fmt.Errorf("Unknown statement %s", val)
	}
}

func (this *node) Parse() error {
	if this.Type == typeId {
		val := this.Value.(string)
		if this.parseGeneric(val) == nil {
		} else if strings.HasSuffix(val, ":") {
			return this.parseLabel(val)
		} else if strings.HasPrefix(val, ";") {
			this.Type = typeComment
			this.Value = strings.TrimPrefix(val, ";")
		} else if strings.HasPrefix(val, "#") {
			return this.parseImmediate(val)
		}
	}
	return nil
}

func (this *node) isComment() bool {
	return this.Type == typeComment
}

func (this *node) isLabel() bool {
	return this.Type == typeLabel
}

type statement struct {
	contents []*node
	index    int
}

func (this *statement) Add(value string, t nodeType) {
	// always trim before adding
	str := strings.TrimSpace(value)
	if len(str) > 0 {
		this.contents = append(this.contents, &node{Value: str, Type: t})
	}
}
func (this *statement) AddUnknown(value string) {
	this.Add(value, typeId)
}
func (this *statement) String() string {
	str := fmt.Sprintf("%d: ", this.index)
	for _, n := range this.contents {
		str += fmt.Sprintf(" %T: %s ", n, *n)
	}
	return str
}
func (this *statement) First() (*node, error) {
	if len(this.contents) == 0 {
		return nil, fmt.Errorf("Empty statement!")
	} else {
		return this.contents[0], nil
	}
}
func (this *statement) Rest() []*node {
	return this.contents[1:]
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
		next := data[start:]
		r, width = utf8.DecodeRuneInString(next)
		if unicode.IsSpace(r) {
			s.AddUnknown(data[oldStart:start])
			oldStart = start
		} else if r == ';' {
			// consume the rest of the data
			s.AddUnknown(data[oldStart:start])
			// then capture the comment
			s.Add(data[start:], typeComment)
			oldStart = start
			break
		}
	}
	if oldStart < start {
		s.AddUnknown(data[oldStart:])
	}
	return s
}

func (this *_parser) Process() error {
	for _, stmt := range this.statements {
		if err := this.parseStatement(&stmt); err != nil {
			return fmt.Errorf("Error: line %d: msg: %s", stmt.index, err)
		}
	}
	for _, d := range this.deferredInstructions {
		str := d.trouble.Value.(string)
		if v, err := this.resolveLabel(str); err != nil {
			return err
		} else {
			q := d.inst
			q.SetImmediate(v)
			z := q.Encode()
			this.core.code[d.addr] = *z
		}
	}
	return nil
}

func (this *_parser) parseStatement(stmt *statement) error {
	first, err := stmt.First()
	if err != nil {
		return err
	}
	rest := stmt.Rest()
	switch first.Type {
	case typeComment:
		if len(rest) > 0 {
			panic("Programmer Failure! Found something following a comment node in a statement. This is impossible!!!!")
		} else {
			return nil
		}
	case typeLabel:
		if err := this.newLabel(first); err != nil {
			return err
		} else if len(rest) > 0 {
			// if there are more entries on the line then check them out
			var s statement
			s.index = stmt.index
			s.contents = rest
			return this.parseStatement(&s)
		}
	case keywordXand:
		if len(rest) == 3 {
			var s statement
			s.index = stmt.index
			s.contents = rest
			return this.parseStatement(&s)
		} else {
			return fmt.Errorf("xand requires three arguments")
		}
	case keywordDotDotDot:
		this.core.memory[this.core.pc] = this.core.pc + 1
		this.core.pc++
		// hmmm should we allow this to continue on?...nope
		if len(rest) > 0 {
			return fmt.Errorf("... has to terminate a statement")
		}
	case typeImmediate:
		// just install the value to the current address
		this.core.memory[this.core.pc] = first.Value.(int8)
		this.core.pc++
		if len(rest) > 0 {
			var s statement
			s.index = stmt.index
			s.contents = rest
			return this.parseStatement(&s)
		}
	case typeId:
		return fmt.Errorf("Unknown node %s", first.Value)
	default:
		return fmt.Errorf("Unhandled nodeType %d: %s", first.Type, first.Value)
	}
	return nil
}
