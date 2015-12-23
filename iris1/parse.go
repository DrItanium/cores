package iris1

import (
	"encoding/binary"
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
	numSegments
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
		p.aliases = make(map[string]byte)
		return &p, nil
	}
}

func init() {
	parser.Register(RegistrationName(), ParsingRegistrar(generateParser))
}

type indirectAddress struct {
	seg     segment
	label   string
	address Word
}
type _parser struct {
	core              *Core
	statements        []statement
	labels            labelMap
	addrs             [numSegments]Word
	currSegment       segment
	aliases           map[string]byte
	indirectAddresses []indirectAddress
}

func (this *_parser) Dump(pipe chan<- byte) error {
	c, d := make([]byte, 4), make([]byte, 2)
	for _, val := range this.core.code {
		binary.LittleEndian.PutUint32(c, uint32(val))
		for _, b := range c {
			pipe <- b
		}
	}
	for _, val := range this.core.data {
		binary.LittleEndian.PutUint16(d, uint16(val))
		for _, b := range d {
			pipe <- b
		}
	}
	return nil
}
func (this *_parser) parseAlias(nodes []*node) error {
	if len(nodes) == 0 {
		return fmt.Errorf("alias directives require arguments!")
	} else if len(nodes) < 3 {
		return fmt.Errorf("too few arguments provided to the alias directive")
	} else if len(nodes) > 4 {
		return fmt.Errorf("too many arguments provided to the alias directive")
	} else {
		if len(nodes) == 4 && nodes[3].Type != typeComment {
			return fmt.Errorf("last argument to an alias declaration is not a comment!")
		}
		title, eq, value := nodes[0], nodes[1], nodes[2]
		if title.Type != typeSymbol && title.Type != typeUnknown {
			return fmt.Errorf("Name of an alias must be a symbol!")
		} else if eq.Type != typeEquals {
			return fmt.Errorf("= symbol must be between name and target register in an alias declaration")
		} else if value.Type != typeRegister {
			return fmt.Errorf("an alias can only refer to a register or another alias!")
		}
		name := title.Value.(string)
		if _, ok := this.aliases[name]; ok {
			return fmt.Errorf("Already registered alias %s!", name)
		}
		this.aliases[name] = value.Value.(byte)
		return nil
	}
}
func (this *_parser) setSegment(nodes []*node, seg segment, name string) error {
	if len(nodes) != 0 {
		return fmt.Errorf("%s directive takes in no arguments!", name)
	} else {
		this.currSegment = seg
		return nil
	}
}
func (this *_parser) setPosition(nodes []*node) error {
	if len(nodes) == 0 {
		return fmt.Errorf("No arguments provided to org directive")
	} else if len(nodes) > 1 {
		return fmt.Errorf("Too many arguments provided to an org directive")
	} else {
		addr := nodes[0]
		switch addr.Type {
		case typeHexImmediate, typeBinaryImmediate, typeImmediate:
			this.addrs[this.currSegment] = addr.Value.(Word)
			return nil
		default:
			return fmt.Errorf("Org directive requires an immediate value")
		}
	}
}

func (this *_parser) setData(nodes []*node) error {
	if len(nodes) == 0 {
		return fmt.Errorf("Word directive requires a value")
	} else if len(nodes) > 1 {
		return fmt.Errorf("Too many arguments provided to the word directive!")
	} else {
		switch this.currSegment {
		case codeSegment:
			return fmt.Errorf("Word directives can't be in the code segment!")
		case dataSegment:
			addr := nodes[0]
			switch addr.Type {
			case typeHexImmediate, typeBinaryImmediate, typeImmediate:
				this.core.data[this.addrs[this.currSegment]] = addr.Value.(Word)
			case typeLabel:
				this.indirectAddresses = append(this.indirectAddresses, indirectAddress{label: addr.Value.(string), seg: dataSegment, address: this.addrs[this.currSegment]})
			default:
				return fmt.Errorf("word directives can only accept immediates or labels right now")
			}
			this.addrs[this.currSegment]++
			return nil
		default:
			panic("Programmer Failure! Found self in a illegal segment!")
		}
	}
}
func (this *_parser) newLabel(n *node) error {
	name := n.Value.(string)
	if _, ok := this.labels[name]; ok {
		return fmt.Errorf("Label %s is already defined!", name)
	} else {
		this.labels[name] = labelEntry{
			seg:  this.currSegment,
			addr: this.addrs[this.currSegment],
		}
		return nil
	}
}

func (this *_parser) Process() error {
	// build up labels and alias listings
	for _, stmt := range this.statements {
		// get the first element and perform a correct dispatch
		first := stmt[0]
		rest := stmt[1:]
		switch first.Type {
		case typeDirectiveAlias:
			// go through the rest of the nodes
			return this.parseAlias(rest)
		case typeDirectiveCode:
			if err := this.setSegment(rest, codeSegment, "code"); err != nil {
				return err
			}
		case typeDirectiveData:
			if err := this.setSegment(rest, dataSegment, "data"); err != nil {
				return err
			}
		case typeDirectiveOrg:
			return this.setPosition(rest)
		case typeDirectiveWord:
			return this.setData(rest)
		case typeComment:
			// do nothing, just continue
		case typeUnknown:
			fallthrough
		default:
			return fmt.Errorf("Unknown node %s", first.Value)
		}
	}
	return nil
}

func (this *_parser) Parse(lines <-chan parser.Entry) error {
	for line := range lines {
		stmt := carveLine(line.Line)
		this.statements = append(this.statements, stmt)
		for _, str := range stmt {
			if err := str.Parse(); err != nil {
				return fmt.Errorf("Error: line: %d : %s\n", line.Index, err)
			}
			fmt.Print("\t", str, " ")
		}
		fmt.Println()
	}
	return nil
}

type nodeType int

const (
	typeUnknown nodeType = iota
	typeEquals
	typeComma
	typeLabel
	typeRegister
	typeImmediate
	typeBinaryImmediate
	typeHexImmediate
	typeComment
	typeSymbol
	typeAlias
	// directives
	typeDirective
	typeDirectiveData
	typeDirectiveCode
	typeDirectiveOrg
	typeDirectiveWord
	typeDirectiveAlias
	// keywords
	// memory words
	keywordSet
	keywordMove
	keywordLoad
	keywordStore
	keywordSwap
	// branch words
	keywordBranch
	keywordIf0
	keywordIf1
	keywordLink
	keywordThen
	keywordElse
	// arithmetic words
	keywordAdd
	keywordSub
	keywordMul
	keywordDiv
	keywordRem
	keywordShiftLeft
	keywordShiftRight
	keywordAnd
	keywordOr
	keywordNot
	keywordXor
	keywordIncrement
	keywordDecrement
	keywordHalve
	// compare words
	keywordEqual
	keywordNotEqual
	keywordLessThan
	keywordGreaterThan
	keywordLessThanOrEqualTo
	keywordGreaterThanOrEqualTo
	// misc words
	keywordSystem
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

type invalidRegisterError struct {
	Value string
}

func (this *invalidRegisterError) Error() string {
	return fmt.Sprintf("Register %s is not a valid register!", this.Value)
}
func InvalidRegister(value string) error {
	return &invalidRegisterError{Value: value}
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
		nod := node{Value: this.Value, Type: typeUnknown}
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
			case *invalidRegisterError:
				j := err.(*invalidRegisterError)
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
	this.Type = typeHexImmediate
	if v, err := parseHexImmediate(val[2:]); err != nil {
		return err
	} else {
		this.Value = v
		return nil
	}
}
func (this *node) parseBinaryImmediate(val string) error {
	this.Type = typeBinaryImmediate
	if v, err := parseBinaryImmediate(val[2:]); err != nil {
		return err
	} else {
		this.Value = v
		return nil
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
func (this *node) parseRegister(val string) error {
	this.Type = typeRegister
	if v, err := parseRegisterValue(val[1:]); err != nil {
		return err
	} else {
		this.Value = v
		return nil
	}
}

var directives = map[string]nodeType{
	"data":  typeDirectiveData,
	"code":  typeDirectiveCode,
	"org":   typeDirectiveOrg,
	"word":  typeDirectiveWord,
	"alias": typeDirectiveAlias,
}

func (this *node) parseDirective(val string) error {
	str := val[1:]
	this.Value = str
	if q, ok := directives[str]; ok {
		this.Type = q
	} else {
		this.Type = typeDirective
	}
	return nil
}

func (this *node) parseAlias(str string) error {
	this.Value = str[1:]
	this.Type = typeAlias
	return nil
}

var keywords = map[string]nodeType{
	"branch":     keywordBranch,
	"if0":        keywordIf0,
	"if1":        keywordIf1,
	"then":       keywordThen,
	"else":       keywordElse,
	"link":       keywordLink,
	"add":        keywordAdd,
	"sub":        keywordSub,
	"mul":        keywordMul,
	"div":        keywordDiv,
	"rem":        keywordRem,
	"shiftleft":  keywordShiftLeft,
	"shiftright": keywordShiftRight,
	"and":        keywordAnd,
	"or":         keywordOr,
	"not":        keywordNot,
	"xor":        keywordXor,
	"halve":      keywordHalve,
	"incr":       keywordIncrement,
	"decr":       keywordDecrement,
	"system":     keywordSystem,
	"set":        keywordSet,
	"move":       keywordMove,
	"swap":       keywordSwap,
	"load":       keywordLoad,
	"store":      keywordStore,
}

func (this *node) parseGeneric(str string) error {
	if v, ok := keywords[str]; ok {
		this.Type = v
	}
	return nil
}
func (this *node) Parse() error {
	if this.Type == typeUnknown {
		val := this.Value.(string)
		if val == "=" {
			this.Type = typeEquals
		} else if val == "," {
			this.Type = typeComma
		} else if strings.HasSuffix(val, ":") {
			return this.parseLabel(val)
		} else if strings.HasPrefix(val, ";") {
			this.Type = typeComment
			this.Value = strings.TrimPrefix(val, ";")
		} else if strings.HasPrefix(val, "#x") {
			return this.parseHexImmediate(val)
		} else if strings.HasPrefix(val, "#b") {
			return this.parseBinaryImmediate(val)
		} else if strings.HasPrefix(val, "#") {
			return this.parseImmediate(val)
		} else if strings.HasPrefix(val, "r") {
			return this.parseRegister(val)
		} else if strings.HasPrefix(val, ".") {
			return this.parseDirective(val)
		} else if strings.HasPrefix(val, "?") {
			return this.parseAlias(val)
		} else {
			return this.parseGeneric(val)
		}
	}
	return nil
}

func (this *node) IsComment() bool {
	return this.Type == typeComment
}
func (this *node) IsLabel() bool {
	return this.Type == typeLabel
}

type statement []*node

func (this *statement) Add(value string, t nodeType) {
	// always trim before adding
	str := strings.TrimSpace(value)
	if len(str) > 0 {
		*this = append(*this, &node{Value: str, Type: t})
	}
}
func (this *statement) AddUnknown(value string) {
	this.Add(value, typeUnknown)
}
func (this *statement) String() string {
	var str string
	for _, n := range *this {
		str += fmt.Sprintf(" %T: %s ", n, *n)
	}
	return str
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
			s.Add("=", typeEquals)
			oldStart = start + width
		} else if r == ',' {
			s.AddUnknown(data[oldStart:start])
			s.Add(",", typeComma)
			oldStart = start + width
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
