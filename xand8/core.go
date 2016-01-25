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

type Word int8
type BranchUnit struct {
	running                             bool
	control, cond, onTrue, onFalse      chan Word
	out                                 chan Word
	Control, Condition, OnTrue, OnFalse chan<- Word
	Result                              <-chan Word
}

func NewBranchUnit() *BranchUnit {
	var b BranchUnit
	b.control = make(chan Word)
	b.cond = make(chan Word, 4)
	b.onTrue = make(chan Word, 4)
	b.onFalse = make(chan Word, 4)
	b.out = make(chan Word, 4)
	b.Control = b.control
	b.Condition = b.cond
	b.OnTrue = b.onTrue
	b.OnFalse = b.onFalse
	b.Result = b.out
	go b.controlInvoke()
	return &b
}

func (this *BranchUnit) controlInvoke() {
	// wait until we get the init signal
	<-this.control
	this.running = true
	go this.body()
	// wait until we get another signal to terminate
	<-this.control
	// shut the machine down
	this.running = false
	close(this.cond)
	close(this.onTrue)
	close(this.onFalse)
	// wait until body finishes
	<-this.control
	close(this.control)
}

func (this *BranchUnit) body() {
	for this.running {
		select {
		case cond := <-this.cond:
			if t, f := <-this.onTrue, <-this.onFalse; cond <= 0 {
				this.out <- t
			} else {
				this.out <- f
			}
		default:
		}
	}
	// send the I'm done signal
	this.control <- 1
}

type MemoryUnit struct {
	running                  bool
	memory                   [MemorySize]Word
	control, op, addr, value chan Word
	out                      chan Word
	err                      chan error
	Control, Op, Addr, Value chan<- Word
	Result                   <-chan Word
	Error                    <-chan error
}

func NewMemoryUnit() *MemoryUnit {
	var b MemoryUnit
	b.err = make(chan error, 4)
	b.control = make(chan Word)
	b.op = make(chan Word, MemorySize)
	b.addr = make(chan Word, MemorySize)
	b.value = make(chan Word, MemorySize)
	b.out = make(chan Word, MemorySize)
	b.Control = b.control
	b.Error = b.err
	b.Op = b.op
	b.Addr = b.addr
	b.Value = b.value
	b.Result = b.out
	go b.controlInvoke()
	return &b
}

func (this *MemoryUnit) controlInvoke() {
	// wait until we get the init signal
	<-this.control
	this.running = true
	go this.body()
	// wait until we get another signal to terminate
	<-this.control
	// shut the machine down
	this.running = false
	close(this.op)
	close(this.addr)
	close(this.value)
	// wait until body finishes
	<-this.control
	close(this.control)
	close(this.err)
}

const (
	MemoryLoad = iota
	MemoryStore
)

func (this *MemoryUnit) body() {
	for this.running {
		select {
		case op := <-this.op:
			addr := <-this.addr
			if addr < 0 {
				this.err <- fmt.Errorf("Out of range!")
			} else {
				switch op {
				case MemoryLoad: // load
					this.out <- this.memory[addr]
					this.err <- nil
				case MemoryStore: // store
					this.memory[addr] = <-this.value
					this.err <- nil
				default:
					this.err <- fmt.Errorf("Illegal signal")
				}
			}
		default:
		}
	}
	// send the I'm done signal
	this.control <- 1
}

type Alu struct {
	running                bool
	control, op, a, b, out chan Word
	Control                chan<- Word
	First                  chan<- Word
	Second                 chan<- Word
	Op                     chan<- Word
	Result                 <-chan Word
}

func NewAlu() *Alu {
	var a Alu
	a.control = make(chan Word)
	a.op = make(chan Word, 4)
	a.a = make(chan Word, 4)
	a.b = make(chan Word, 4)
	a.out = make(chan Word, 4)
	a.Control = a.control
	a.Op = a.op
	a.First = a.a
	a.Second = a.b
	a.Result = a.out
	go a.controlInvoke()
	return &a
}
func (this *Alu) controlInvoke() {
	// wait until we get the init signal
	<-this.control
	this.running = true
	go this.body()
	// wait until we get another signal to terminate
	<-this.control
	// shut the machine down
	this.running = false
	close(this.op)
	close(this.a)
	close(this.b)
	close(this.out)
	// wait until body finishes
	<-this.control
	close(this.control)
}

const (
	AluSubtract = iota
	AluLessThanZero
	AluLessThanOrEqualToZero
)

func (this *Alu) body() {
	for this.running {
		select {
		case op := <-this.op:
			a := <-this.a
			switch op {
			case AluSubtract:
				result := (a - <-this.b)
				this.out <- result
				this.out <- result
			case AluLessThanZero:
				if a < 0 {
					this.out <- 1
				} else {
					this.out <- 0
				}
			case AluLessThanOrEqualToZero:
				if a <= 0 {
					this.out <- 1
				} else {
					this.out <- 0
				}
			}
		default:
		}
	}
}

const MemorySize = 128

func RegistrationName() string {
	return "xand8"
}

func generateCore(a ...interface{}) (machine.Machine, error) {
	return New()
}

func init() {
	machine.Register(RegistrationName(), machine.Registrar(generateCore))
}

type Core struct {
	pc     Word
	ir     [3]Word
	branch *BranchUnit
	memory *MemoryUnit
	alu    *Alu
	debug  bool
}

func New() (*Core, error) {
	var c Core
	c.branch = NewBranchUnit()
	c.memory = NewMemoryUnit()
	c.alu = NewAlu()
	c.branch.Control <- 0
	c.memory.Control <- 0
	c.alu.Control <- 0
	return &c, nil
}

func (this *Core) Run() error {
	for {
		this.memory.Addr <- this.pc
		this.memory.Addr <- this.pc + 1
		this.memory.Addr <- this.pc + 2
		this.memory.Op <- MemoryLoad
		this.memory.Op <- MemoryLoad
		this.memory.Op <- MemoryLoad
		if <-this.memory.Error != nil || <-this.memory.Error != nil || <-this.memory.Error != nil {
			return nil
		}
		a, b, c := <-this.memory.Result, <-this.memory.Result, <-this.memory.Result
		this.alu.First <- a
		this.alu.First <- b
		this.alu.First <- c
		this.alu.Op <- AluLessThanZero
		this.alu.Op <- AluLessThanZero
		this.alu.Op <- AluLessThanZero
		// Check and see if a, b, or c are less than zero. Halt if it
		// is the case
		ra, rb, rc := <-this.alu.Result, <-this.alu.Result, <-this.alu.Result
		if ra == 1 || rb == 1 || rc == 1 {
			return nil
		}
		// setup the conditional check ahead of time since we are
		// dependent on the resulting condition, nothing more
		this.branch.OnTrue <- c                    // if memory[a] <= 0 then c
		this.branch.OnFalse <- this.pc + 3         // else this.pc + 3
		this.memory.Addr <- a                      // denote that we want to load memory[a]
		this.memory.Addr <- b                      // denote that we want to load memory[b]
		this.memory.Addr <- a                      // put a into the address queue ahead of time here for the memory[a] store later on
		this.memory.Op <- MemoryLoad               // command the memory unit to load the contents of memory[a]
		this.memory.Op <- MemoryLoad               // command the memory unit to load the contents of memory[b]
		this.alu.First <- <-this.memory.Result     // load memory[a] into the alu first "register"
		this.alu.Second <- <-this.memory.Result    // load memory[b] into the alu second "register"
		this.alu.Op <- AluSubtract                 // tell the alu to perform the subtraction and load two copies of the result into the Result channel
		this.branch.Condition <- <-this.alu.Result // Use the first copy of the subtraction result as the condition to the branch unit
		this.pc = <-this.branch.Result             // get the selected value out of the branch unit
		this.memory.Value <- <-this.alu.Result     // store the second copy of the subtraction result into memory[a]
		this.memory.Op <- MemoryStore              // tell the memory unit to perform a storeA
		// clear out the errors from the memory unit before ending the
		// cycle
		if err := <-this.memory.Error; err != nil {
			return err
		} else if err := <-this.memory.Error; err != nil {
			return err
		} else if err := <-this.memory.Error; err != nil {
			return err
		}
		// and start the process again
	}
}

func (this *Core) Startup() error {
	return nil
}

func (this *Core) Shutdown() error {
	this.branch.Control <- 0
	this.memory.Control <- 0
	this.alu.Control <- 0
	return nil
}

func (this *Core) GetDebugStatus() bool {
	return false
}

func (this *Core) SetDebug(_ bool) {

}
func (this *Core) wait0(done chan<- error) {
	for i := 0; i < MemorySize; i++ {
		if err := <-this.memory.Error; err != nil {
			done <- err
		}
	}
	done <- nil
}
func (this *Core) InstallProgram(input <-chan byte) error {
	err := make(chan error)
	go this.wait0(err)
	// read 128 bytes
	for i := 0; i < MemorySize; i++ {
		if value, more := <-input; !more {
			return fmt.Errorf("Not a complete xand memory image")
		} else {
			this.memory.Addr <- Word(i)
			this.memory.Value <- Word(value)
			this.memory.Op <- 1
		}
	}
	return <-err
}

func (this *Core) dump0(output chan<- byte, done chan<- bool) {
	for i := 0; i < MemorySize; i++ {
		output <- byte(<-this.memory.Result)
	}
	done <- true
}
func (this *Core) Dump(output chan<- byte) error {
	done := make(chan bool)
	go this.dump0(output, done)
	for i := 0; i < MemorySize; i++ {

	}
	<-done
	return nil
}

func generateParser(a ...interface{}) (parser.Parser, error) {
	var p _parser
	if core, err := New(); err != nil {
		return nil, err
	} else {
		p.core = core
		p.labels = make(map[string]Word)
		return &p, nil
	}
}

func init() {
	parser.Register(RegistrationName(), parser.Registrar(generateParser))
}

type deferredAddress struct {
	addr  Word
	title string
}

type _parser struct {
	core       *Core
	labels     map[string]Word
	statements []*statement
	deferred   []deferredAddress
}

func (this *_parser) Dump(pipe chan<- byte) error {
	return this.core.Dump(pipe)
}

func (this *_parser) Parse(lines <-chan parser.Entry) error {
	for line := range lines {
		stmt := carveLine(line.Line)
		stmt.index = line.Index
		this.statements = append(this.statements, stmt)
		for _, str := range stmt.contents {
			if err := str.Parse(); err != nil {
				return fmt.Errorf("Error: line: %d : %s\n", line.Index, err)
			}
		}
	}
	return nil
}

type nodeType int

func (this nodeType) String() string {
	switch this {
	case typeId:
		return "id"
	case typeImmediate:
		return "immediate"
	case typeLabel:
		return "label"
	case typeComment:
		return "comment"
	case keywordXand:
		return "xand"
	case keywordDotDotDot:
		return "..."
	default:
		return fmt.Sprintf("%d", this)
	}
}
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

func parseDecimalImmediate(str string) (Word, error) {
	val, err := strconv.ParseInt(str, 10, 8)
	return Word(val), err
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

func carveLine(line string) *statement {
	// trim the damn line first
	data := strings.TrimSpace(line)
	var s statement
	if len(data) == 0 {
		return &s
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
	return &s
}

func (this *_parser) Process() error {
	for _, stmt := range this.statements {
		if err := this.parseStatement(stmt); err != nil {
			return fmt.Errorf("Error: line %d: msg: %s", stmt.index, err)
		}
	}
	for _, d := range this.deferred {
		if entry, ok := this.labels[d.title]; !ok {
			return fmt.Errorf("Label %s not defined!", d.title)
		} else {
			this.core.memory.Addr <- d.addr
			this.core.memory.Value <- entry
			this.core.memory.Op <- 1
			if err := <-this.core.memory.Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *_parser) newLabel(n *node) error {
	name := n.Value.(string)
	if _, ok := this.labels[name]; ok {
		return fmt.Errorf("Label %s is already defined!", name)
	} else {
		this.labels[name] = this.core.pc
		return nil
	}
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
			if this.core.pc < 0 {
				return fmt.Errorf("Too many instructions defined!")
			}
			// if there are more entries on the line then check them out
			var s statement
			s.index = stmt.index
			s.contents = rest
			return this.parseStatement(&s)
		}
	case keywordXand:
		if len(rest) == 3 {
			if this.core.pc < 0 {
				return fmt.Errorf("Too many instructions defined!")
			}
			var s statement
			s.index = stmt.index
			s.contents = rest
			return this.parseStatement(&s)
		} else {
			return fmt.Errorf("xand requires three arguments")
		}
	case keywordDotDotDot:
		this.core.memory.Addr <- this.core.pc
		this.core.memory.Value <- this.core.pc + 1
		this.core.memory.Op <- 1
		if err := <-this.core.memory.Error; err != nil {
			return err
		}
		this.core.pc++
		// hmmm should we allow this to continue on?...nope
		if len(rest) > 0 {
			return fmt.Errorf("... has to terminate a statement")
		}
	case typeImmediate:
		// just install the value to the current address
		this.core.memory.Addr <- this.core.pc
		this.core.memory.Value <- first.Value.(Word)
		this.core.memory.Op <- 1
		if err := <-this.core.memory.Error; err != nil {
			return err
		}
		this.core.pc++
		if len(rest) > 0 {
			if this.core.pc < 0 {
				return fmt.Errorf("Too many instructions defined!")
			}
			var s statement
			s.index = stmt.index
			s.contents = rest
			return this.parseStatement(&s)
		}
	case typeId:
		// defer statement for the time being
		if addr, ok := this.labels[first.Value.(string)]; !ok {
			this.deferred = append(this.deferred, deferredAddress{addr: this.core.pc, title: first.Value.(string)})
		} else {
			this.core.memory.Addr <- this.core.pc
			this.core.memory.Value <- addr
			this.core.memory.Op <- 1
			if err := <-this.core.memory.Error; err != nil {
				return err
			}
		}
		this.core.pc++
		if len(rest) > 0 {
			if this.core.pc < 0 {
				return fmt.Errorf("Too many instructions defined!")
			}
			var s statement
			s.index = stmt.index
			s.contents = rest
			return this.parseStatement(&s)
		}
	default:
		return fmt.Errorf("Unhandled nodeType %d: %s", first.Type, first.Value)
	}
	return nil
}
