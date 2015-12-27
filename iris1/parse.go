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
		p.labels = make(labelMap)
		p.aliases = make(map[string]byte)
		return &p, nil
	}
}

func init() {
	parser.Register(RegistrationName(), ParsingRegistrar(generateParser))
}

type segment int

const (
	codeSegment segment = iota
	dataSegment
	numSegments
)

type nodeType int

func (this nodeType) isComma() bool {
	return this == typeComma
}
func (this nodeType) registerOrAlias() bool {
	return this == typeRegister || this == typeAlias
}
func (this nodeType) immediate() bool {
	return this == typeHexImmediate || this == typeBinaryImmediate || this == typeImmediate
}
func (this nodeType) comment() bool {
	return this == typeComment
}
func (this nodeType) compareOperation() bool {
	return this == typeEquals || this == typeOr || this == typeAnd || this == typeXor
}

const (
	typeId nodeType = iota
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
	typeAnd
	typeOr
	typeXor
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
	keywordPush
	keywordPop
	keywordPeek
	// branch words
	keywordBranch
	keywordCall
	keywordReturn
	keywordIf
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
	keywordDouble
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
		nod := node{Value: this.Value, Type: typeId}
		if err := nod.Parse(); err != nil {
			switch err.(type) {
			case *strconv.NumError:
				j := err.(*strconv.NumError)
				if j.Err == strconv.ErrRange {
					return fmt.Errorf("Label %s is interpreted as an out of range value! This is not allowed as it is ambiguous!", this.Value)
				} else if j.Err == strconv.ErrSyntax {
					// probably legal, will require an extra pass most likely
					return fmt.Errorf("Syntax error from strconv (most likely not an error!): %s", j.Err)
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
	"call":       keywordCall,
	"return":     keywordReturn,
	"if":         keywordIf,
	"then":       keywordThen,
	"else":       keywordElse,
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
	"double":     keywordDouble,
	"eq":         keywordEqual,
	"ne":         keywordNotEqual,
	"lt":         keywordLessThan,
	"gt":         keywordGreaterThan,
	"le":         keywordLessThanOrEqualTo,
	"ge":         keywordGreaterThanOrEqualTo,
	"push":       keywordPush,
	"pop":        keywordPop,
}

func (this *node) parseGeneric(str string) error {
	if v, ok := keywords[str]; ok {
		this.Type = v
		return nil
	} else {
		return fmt.Errorf("Unknown statment %s", str)
	}
}
func (this *node) Parse() error {
	if this.Type == typeId {
		val := this.Value.(string)
		if this.parseGeneric(val) == nil {
			// parse keywords first and then other types of symbols
			// successful match....kinda a hack but it will work
			return nil
		} else if val == "=" {
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
		}
		// leave it typeId since that is legal as well
	}
	return nil
}

func (this *node) IsComment() bool {
	return this.Type == typeComment
}
func (this *node) IsLabel() bool {
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
		} else if r == '&' {
			s.AddUnknown(data[oldStart:start])
			s.Add("&", typeAnd)
			oldStart = start + width
		} else if r == '|' {
			s.AddUnknown(data[oldStart:start])
			s.Add("|", typeOr)
			oldStart = start + width
		} else if r == '^' {
			s.AddUnknown(data[oldStart:start])
			s.Add("^", typeXor)
			oldStart = start + width
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

type labelEntry struct {
	seg  segment
	addr Word
}

type labelMap map[string]labelEntry

type indirectAddress struct {
	seg     segment
	label   string
	address Word
}
type deferredInstruction struct {
	addr    Word
	inst    *DecodedInstruction
	trouble *node
}

type _parser struct {
	core                 *Core
	statements           []statement
	labels               labelMap
	addrs                [numSegments]Word
	currSegment          segment
	aliases              map[string]byte
	indirectAddresses    []indirectAddress
	deferredInstructions []deferredInstruction
}

func (this *_parser) Defer(inst *DecodedInstruction, trouble *node) {
	this.deferredInstructions = append(this.deferredInstructions, deferredInstruction{addr: this.addrs[codeSegment], inst: inst, trouble: trouble})
	this.addrs[codeSegment]++
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
		title, value := nodes[0], nodes[2]
		if title.Type != typeSymbol && title.Type != typeId {
			return fmt.Errorf("Name of an alias must be a symbol!")
		} else if nodes[1].Type != typeEquals {
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
func (this *_parser) resolveAlias(alias string) (byte, error) {
	if v, ok := this.aliases[alias]; !ok {
		return 0, fmt.Errorf("Alias %s does not exist!", alias)
	} else {
		return v, nil
	}
}
func (this *_parser) resolveRegister(n *node) (byte, error) {
	switch n.Type {
	case typeRegister:
		return n.Value.(byte), nil
	case typeAlias:
		return this.resolveAlias(n.Value.(string))
	default:
		return 0, fmt.Errorf("Illegal node type provided for register resolution!")
	}
}
func (this *_parser) installInstruction(inst *Instruction) error {
	if this.currSegment != codeSegment {
		return fmt.Errorf("Must install instructions to the code segment")
	} else {
		this.core.code[this.addrs[this.currSegment]] = *inst
		this.addrs[this.currSegment]++
		return nil
	}
}
func (this *_parser) installData(data Word) error {
	if this.currSegment != dataSegment {
		return fmt.Errorf("Must install words to the data segment!")
	} else {
		this.core.data[this.addrs[this.currSegment]] = data
		this.addrs[this.currSegment]++
		return nil
	}
}
func (this *_parser) resolveSingleArgMove(i *DecodedInstruction, arg *node) error {
	if dv, err := this.resolveRegister(arg); err != nil {
		return err
	} else {
		i.Data[0] = dv
		return nil
	}
}
func (this *_parser) isLabelReference(n *node) bool {
	if n.Type == typeId {
		_, ok := this.labels[n.Value.(string)]
		return ok
	} else {
		return false
	}
}
func (this *_parser) resolveLabel(name string) (Word, error) {
	if v, ok := this.labels[name]; !ok {
		return 0, fmt.Errorf("Label %s does not exist!", name)
	} else {
		return v.addr, nil
	}
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

func (this *_parser) Process() error {
	// build up labels and alias listings
	for _, stmt := range this.statements {
		if err := this.parseStatement(&stmt); err != nil {
			return fmt.Errorf("Error: line %d: msg: %s", stmt.index, err)
		}
	}
	// check the deferred labels now that we are done processing the whole file
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
	// get the first element and perform a correct dispatch
	first, err := stmt.First()
	if err != nil {
		return err
	}
	rest := stmt.Rest()
	switch first.Type {
	case typeComment: // do nothing, just continue
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
		} else {
			return nil
		}
	case keywordAdd, keywordSub, keywordMul, keywordDiv, keywordRem, keywordShiftLeft, keywordShiftRight, keywordAnd, keywordOr, keywordNot, keywordXor, keywordIncrement, keywordDecrement, keywordHalve:
		return this.parseArithmetic(first, rest)
	case keywordMove, keywordSet, keywordSwap, keywordLoad, keywordStore, keywordPop, keywordPush:
		return this.parseMove(first, rest)
	case keywordEqual, keywordNotEqual, keywordLessThan, keywordGreaterThan, keywordLessThanOrEqualTo, keywordGreaterThanOrEqualTo:
		return this.parseCompare(first, rest)
	case keywordSystem:
		return this.parseMisc(first, rest)
	case keywordBranch, keywordCall, keywordReturn:
		return this.parseJump(first, rest)
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
	case typeComma:
		return fmt.Errorf("Can't start a line with a comma")
	case typeEquals:
		return fmt.Errorf("Can't start a line with a equals sign")
	case typeId:
		return fmt.Errorf("Unknown node %s", first.Value)
	default:
		return fmt.Errorf("Unhandled nodeType %d: %s", first.Type, first.Value)
	}
	return nil
}

func (this *_parser) parseMove(first *node, rest []*node) error {
	deferred := false
	var d DecodedInstruction
	d.Group = InstructionGroupMove
	switch len(rest) {
	case 2:
		if !rest[1].Type.comment() {
			return fmt.Errorf("Too many arguments provided to a single argument move operation")
		}
		fallthrough
	case 1:
		dv := rest[0]
		switch first.Type {
		case keywordPop:
			d.Op = MoveOpPop
			if err := this.resolveSingleArgMove(&d, dv); err != nil {
				return err
			}
		case keywordPeek:
			d.Op = MoveOpPeek
			if err := this.resolveSingleArgMove(&d, dv); err != nil {
				return err
			}
		case keywordPush:
			if dv.Type.registerOrAlias() {
				d.Op = MoveOpPush
				if err := this.resolveSingleArgMove(&d, dv); err != nil {
					return err
				}
			} else if dv.Type.immediate() {
				d.Op = MoveOpPushImmediate
				d.SetImmediate(dv.Value.(Word))
			} else {
				return fmt.Errorf("Illegal operand type for push operation!")
			}
		}
	case 4:
		if !rest[3].Type.comment() {
			return fmt.Errorf("Too many arguments provided to a normal move operation")
		}
		fallthrough
	case 3:
		if dv, err := this.resolveRegister(rest[0]); err != nil {
			return err
		} else if rest[1].Type != typeEquals {
			return fmt.Errorf("An = is necessary to separate the destination from source of a move operation")
		} else if src := rest[2]; !src.Type.registerOrAlias() && !src.Type.immediate() && src.Type != typeId {
			return fmt.Errorf("the source of a move operation can be either a register, alias, immediate, or label")
		} else {
			d.Data[0] = dv
			if src.Type.registerOrAlias() {
				if sv, err := this.resolveRegister(src); err != nil {
					return err
				} else {
					d.Data[1], d.Data[2] = sv, 0
				}
				switch first.Type {
				case keywordMove:
					d.Op = MoveOpMove
				case keywordSwap:
					d.Op = MoveOpSwap
				case keywordLoad:
					d.Op = MoveOpLoad
				case keywordStore:
					d.Op = MoveOpStore
				default:
					return fmt.Errorf("Illegal move operation %d", first.Type)
				}
			} else {
				if src.Type.immediate() {
					d.SetImmediate(src.Value.(Word))
				} else if src.Type == typeId {
					if v, err := this.resolveLabel(src.Value.(string)); err != nil {
						// defer it for now
						deferred = true
						this.Defer(&d, src)
					} else {
						d.SetImmediate(v)
					}
				} else {
					panic(fmt.Sprintf("Programmer Failure: accepted but unhandled node type (%s) in move operation!", src.Value))
				}
				switch first.Type {
				case keywordSet:
					d.Op = MoveOpSet
				case keywordLoad:
					d.Op = MoveOpLoadMem
				case keywordStore:
					d.Op = MoveOpStoreImm
				default:
					return fmt.Errorf("move op %s doesn't have an immediate form", first.Value.(string))
				}
			}
		}
	default:
		return fmt.Errorf("Too few or too many arguments provided to a move operation!")
	}
	if !deferred {
		return this.installInstruction(d.Encode())
	} else {
		return nil
	}
}

var compareTable = map[nodeType]map[nodeType]byte{
	keywordEqual: map[nodeType]byte{
		typeEquals: CompareOpEq,
		typeAnd:    CompareOpEqAnd,
		typeOr:     CompareOpEqOr,
		typeXor:    CompareOpEqXor,
	},
	keywordNotEqual: map[nodeType]byte{
		typeEquals: CompareOpNeq,
		typeAnd:    CompareOpNeqAnd,
		typeOr:     CompareOpNeqOr,
		typeXor:    CompareOpNeqXor,
	},
	keywordLessThan: map[nodeType]byte{
		typeEquals: CompareOpLessThan,
		typeAnd:    CompareOpLessThanAnd,
		typeOr:     CompareOpLessThanOr,
		typeXor:    CompareOpLessThanXor,
	},
	keywordGreaterThan: map[nodeType]byte{
		typeEquals: CompareOpGreaterThan,
		typeAnd:    CompareOpGreaterThanAnd,
		typeOr:     CompareOpGreaterThanOr,
		typeXor:    CompareOpGreaterThanXor,
	},
	keywordLessThanOrEqualTo: map[nodeType]byte{
		typeEquals: CompareOpLessThanOrEqualTo,
		typeAnd:    CompareOpLessThanOrEqualToAnd,
		typeOr:     CompareOpLessThanOrEqualToOr,
		typeXor:    CompareOpLessThanOrEqualToXor,
	},
	keywordGreaterThanOrEqualTo: map[nodeType]byte{
		typeEquals: CompareOpGreaterThanOrEqualTo,
		typeAnd:    CompareOpGreaterThanOrEqualToAnd,
		typeOr:     CompareOpGreaterThanOrEqualToOr,
		typeXor:    CompareOpGreaterThanOrEqualToXor,
	},
}

func (this *_parser) parseCompare(first *node, rest []*node) error {
	if this.currSegment != codeSegment {
		return fmt.Errorf("Not in a code segment for this instruction")
	}
	var d DecodedInstruction
	d.Group = InstructionGroupCompare
	switch len(rest) {
	case 0, 1, 2, 3, 4:
		return fmt.Errorf("Too few arguments provided for the given compare operation")
	case 6:
		if rest[5].Type != typeComment {
			return fmt.Errorf("Too many arguments for the given compare operation")
		}
		fallthrough
	case 5:
		if dest := rest[0]; !dest.Type.registerOrAlias() {
			return fmt.Errorf("The destination of a compare operation must be a register or alias")
		} else if update := rest[1]; !update.Type.compareOperation() {
			return fmt.Errorf("Illegal assignment symbol %s for compare operation", update.Value)
		} else if sv0 := rest[2]; !sv0.Type.registerOrAlias() {
			return fmt.Errorf("The first source argument of a compare operation must be a register or alias")
		} else if rest[3].Type != typeComma {
			return fmt.Errorf("Source arguments in a compare operation must be separated by a comma!")
		} else if sv1 := rest[4]; !sv1.Type.registerOrAlias() {
			return fmt.Errorf("Second source argument in a compare operation must be a register or alias")
		} else {
			// determine the corresponding op
			d.Op = compareTable[first.Type][update.Type]
			if dv, err := this.resolveRegister(dest); err != nil {
				return err
			} else if s0, err := this.resolveRegister(sv0); err != nil {
				return err
			} else if s1, err := this.resolveRegister(sv1); err != nil {
				return err
			} else {
				d.Data = [3]byte{dv, s0, s1}
			}
		}
	default:
		return fmt.Errorf("Too many arguments provided to given compare operation")
	}
	return this.installInstruction(d.Encode())
}
func (this *_parser) parseSystem(d *DecodedInstruction, rest []*node) error {
	d.Op = MiscOpSystemCall
	switch len(rest) {
	case 0, 1, 2, 3, 4:
		return fmt.Errorf("Too few arguments passed to the given system operation")
	case 6:
		if rest[5].Type != typeComment {
			return fmt.Errorf("Too many arguments passed to the system operation")
		}
		fallthrough
	case 5:
		if id := rest[0]; !id.Type.immediate() {
			return fmt.Errorf("First argument of a system must be an 8-bit immediate")
		} else if idx := id.Value.(Word); idx > 255 {
			return fmt.Errorf("Provided system operation immediate is larger than 7-bits!")
		} else if !rest[1].Type.isComma() {
			return fmt.Errorf("Comma is required after immediate in system operation")
		} else if sv0 := rest[2]; !sv0.Type.registerOrAlias() {
			return fmt.Errorf("Second argument in system operation must be a register or alias")
		} else if s0, err := this.resolveRegister(sv0); err != nil {
			return err
		} else if !rest[3].Type.isComma() {
			return fmt.Errorf("second and third arguments in a system operation must be separated by a comma")
		} else if sv1 := rest[2]; !sv1.Type.registerOrAlias() {
			return fmt.Errorf("Third argument in system operation must be a register or alias")
		} else if s1, err := this.resolveRegister(sv1); err != nil {
			return err
		} else {
			d.Data = [3]byte{byte(idx), s0, s1}
		}
	default:
		return fmt.Errorf("too many arguments passed to the given system operation")
	}
	return nil
}
func (this *_parser) parseMisc(first *node, rest []*node) error {
	if this.currSegment != codeSegment {
		return fmt.Errorf("Currently not in code segment, can't insert instruction")
	}
	var d DecodedInstruction
	d.Group = InstructionGroupMisc
	switch first.Type {
	case keywordSystem:
		if err := this.parseSystem(&d, rest); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Illegal misc operation %s", first.Value)
	}
	return this.installInstruction(d.Encode())
}

func (this *_parser) parseArithmetic(t *node, nodes []*node) error {
	var inst DecodedInstruction
	inst.Group = InstructionGroupArithmetic
	switch len(nodes) {
	case 0, 1, 2:
		return fmt.Errorf("Too few arguments provided for an arithmetic instruction")
	case 4:
		if nodes[3].Type != typeComment {
			return fmt.Errorf("Found an extra argument for an arithmetic instruction")
		}
		fallthrough
	case 3:
		// increment, decrement, double, and halve forms
		if dest := nodes[0]; !dest.Type.registerOrAlias() {
			return fmt.Errorf("The destination operand must be a register or alias")
		} else if nodes[1].Type != typeEquals {
			return fmt.Errorf("The destination register of a arithmetic instruction must be separated from the source register with an =")
		} else if src0 := nodes[2]; !dest.Type.registerOrAlias() {
			return fmt.Errorf("The source operand must be a register or alias")
		} else {
			if dv, err := this.resolveRegister(dest); err != nil {
				return err
			} else if sv0, err := this.resolveRegister(src0); err != nil {
				return err
			} else {
				inst.Data = [3]byte{dv, sv0, 0}
			}
			switch t.Type {
			case keywordHalve:
				inst.Op = ArithmeticOpHalve
			case keywordDouble:
				inst.Op = ArithmeticOpDouble
			case keywordIncrement:
				inst.Op = ArithmeticOpIncrement
			case keywordDecrement:
				inst.Op = ArithmeticOpDecrement
			default:
				return fmt.Errorf("Illegal arithmetic operation %s", t.Value)
			}
		}
	case 6:
		if nodes[5].Type != typeComment {
			return fmt.Errorf("Found an extra argument for an arithmetic instruction")
		}
		fallthrough
	case 5:
		// check all of the arguments
		if dest := nodes[0]; !dest.Type.registerOrAlias() {
			return fmt.Errorf("The destination operand must be a register or alias")
		} else if nodes[1].Type != typeEquals {
			return fmt.Errorf("The destination register of a arithmetic instruction must be separated from the source registers with an =")
		} else if src0 := nodes[2]; !src0.Type.registerOrAlias() {
			return fmt.Errorf("The first source operand must be a register or alias")
		} else if nodes[3].Type != typeComma {
			return fmt.Errorf("The source operands of an arithmetic instruction must be separated by a comma")
		} else if src1 := nodes[4]; !src1.Type.registerOrAlias() && !src1.Type.immediate() {
			return fmt.Errorf("The second source operand must be a register, alias, or 8-bit immediate")
		} else {
			if dv, err := this.resolveRegister(dest); err != nil {
				return err
			} else if sv0, err := this.resolveRegister(src0); err != nil {
				return err
			} else {
				inst.Data[0], inst.Data[1] = dv, sv0
			}
			if src1.Type.immediate() {
				// immediate form
				if immediate := src1.Value.(Word); immediate > 255 {
					return fmt.Errorf("Immediate value for arithmetic operation is too large: %d > 255", immediate)
				} else {
					inst.Data[2] = byte(immediate)
				}
				switch t.Type {
				case keywordAdd:
					inst.Op = ArithmeticOpAddImmediate
				case keywordSub:
					inst.Op = ArithmeticOpSubImmediate
				case keywordMul:
					inst.Op = ArithmeticOpMulImmediate
				case keywordDiv:
					inst.Op = ArithmeticOpDivImmediate
				case keywordRem:
					inst.Op = ArithmeticOpRemImmediate
				case keywordShiftLeft:
					inst.Op = ArithmeticOpShiftLeftImmediate
				case keywordShiftRight:
					inst.Op = ArithmeticOpShiftRightImmediate
				default:
					return fmt.Errorf("Arithmetic operation %s does not have an immediate form!", t.Value)
				}
			} else {
				switch t.Type {
				case keywordAdd:
					inst.Op = ArithmeticOpAdd
				case keywordSub:
					inst.Op = ArithmeticOpSub
				case keywordMul:
					inst.Op = ArithmeticOpMul
				case keywordDiv:
					inst.Op = ArithmeticOpDiv
				case keywordRem:
					inst.Op = ArithmeticOpRem
				case keywordShiftLeft:
					inst.Op = ArithmeticOpShiftLeft
				case keywordShiftRight:
					inst.Op = ArithmeticOpShiftRight
				case keywordAnd:
					inst.Op = ArithmeticOpBinaryAnd
				case keywordOr:
					inst.Op = ArithmeticOpBinaryOr
				case keywordNot:
					inst.Op = ArithmeticOpBinaryNot
				case keywordXor:
					inst.Op = ArithmeticOpBinaryXor
				default:
					return fmt.Errorf("Illegal arithmetic operation %s", t.Value)
				}
				// parse it like the other registers at this point
				if sv1, err := this.resolveRegister(src1); err != nil {
					return err
				} else {
					inst.Data[2] = sv1
				}
			}
		}
	default:
		return fmt.Errorf("Too many arguments provided for an arithmetic instruction")
	}
	// now setup the code section
	return this.installInstruction(inst.Encode())
}

func (this *_parser) parseJump(first *node, rest []*node) error {
	if this.currSegment != codeSegment {
		return fmt.Errorf("Currently not in code segment, can't insert instruction")
	}
	var bb branchBits

	var deferred bool
	var d DecodedInstruction
	d.Group = InstructionGroupJump
	// The bit set calls are fully described here as a description of the encoding itself
	// check the first operating instruction to set some of the branch bits
	bb.setReturnForm(first.Type == keywordReturn)
	bb.setCallForm(first.Type == keywordCall)
	switch first.Type {
	case keywordReturn:
		bb.setIfThenElseForm(false)
		bb.setImmediateForm(false)
		switch len(rest) {
		case 1:
			if !rest[0].Type.comment() {
				return fmt.Errorf("Too many arguments passed to return")
			}
			fallthrough
		case 0:
			bb.setConditionalForm(false)
		case 3:
			if !rest[2].Type.comment() {
				return fmt.Errorf("Too many arguments passed to conditional return")
			}
			fallthrough
		case 2:
			if rest[0].Type != keywordIf {
				return fmt.Errorf("Expected if condition as second argument")
			} else if dest := rest[1]; !dest.Type.registerOrAlias() {
				return fmt.Errorf("The predicate register for the conditional return must be a register or alias")
			} else if dv, err := this.resolveRegister(dest); err != nil {
				return err
			} else {
				bb.setConditionalForm(true)
				d.Data = [3]byte{dv, 0, 0}
			}
		default:
			return fmt.Errorf("Too many arguments provided to a return instruction")
		}
	case keywordCall, keywordBranch:
		switch len(rest) {
		case 7:
			if !rest[6].Type.comment() {
				return fmt.Errorf("Too many arguments provided to the given jump op")
			}
			fallthrough
		case 6:
			bb.setImmediateForm(false)
			bb.setIfThenElseForm(true)
			bb.setConditionalForm(false)
			if rest[0].Type != keywordIf {
				return fmt.Errorf("The if then else form of branch or call requires that the if immediately follows the branch/call statement")
			} else if pred := rest[1]; !pred.Type.registerOrAlias() {
				return fmt.Errorf("The predicate in an if then else form branch/call instruction must be a register or alias")
			} else if p, err := this.resolveRegister(pred); err != nil {
				return err
			} else if rest[2].Type != keywordThen {
				return fmt.Errorf("A \"then\" must follow the predicate in an if then else form branch/call")
			} else if ont := rest[3]; !ont.Type.registerOrAlias() {
				return fmt.Errorf("The \"on predicate true\" field must refer to a register or alias")
			} else if ot, err := this.resolveRegister(ont); err != nil {
				return err
			} else if rest[4].Type != keywordElse {
				return fmt.Errorf("An \"else\" statement must follow the \"on predicate true\" field in a if then else form branch/call")
			} else if onf := rest[5]; !onf.Type.registerOrAlias() {
				return fmt.Errorf("The \"on predicate false\" field must refer to a register or alias")
			} else if of, err := this.resolveRegister(onf); err != nil {
				return err
			} else {
				d.Data = [3]byte{p, ot, of}
			}
		case 4:
			if !rest[3].Type.comment() {
				return fmt.Errorf("Too many arguments provided to the given jump op")
			}
			fallthrough
		case 3:
			bb.setIfThenElseForm(false)
			bb.setConditionalForm(true)
			if dest := rest[0]; !dest.Type.registerOrAlias() && !dest.Type.immediate() && dest.Type != typeId {
				return fmt.Errorf("Expected a register, alias, immediate, or label as the first argument to the given branch/call")
			} else if rest[1].Type != keywordIf {
				return fmt.Errorf("Expected an \"if\" statement following the destination parameter in the given branch/call")
			} else if pred := rest[2]; !pred.Type.registerOrAlias() {
				return fmt.Errorf("The predicate of the given branch/call must be a register or alias!")
			} else if p, err := this.resolveRegister(pred); err != nil {
				return err
			} else {
				bb.setImmediateForm(dest.Type.immediate() || dest.Type == typeId)
				d.Data[0] = p
				// now check out the destination and see if it needs to be deferred or not
				if dest.Type.registerOrAlias() {
					if q, err := this.resolveRegister(dest); err != nil {
						return err
					} else {
						d.Data[1] = q
					}
				} else if dest.Type.immediate() {
					d.SetImmediate(dest.Value.(Word))
				} else if dest.Type == typeId {
					// check for label and defer if necessary
					if v, err := this.resolveLabel(dest.Value.(string)); err != nil {
						// defer it for now
						deferred = true
						this.Defer(&d, dest)
					} else {
						d.SetImmediate(v)
					}
				} else {
					panic(fmt.Sprintf("Programmer Failure: accepted but unhandled node type (%s) in jump operation!", dest.Value))
				}
			}
		case 2:
			if !rest[1].Type.comment() {
				return fmt.Errorf("Too many arguments provided to the given jump op")
			}
			fallthrough
		case 1:
			bb.setIfThenElseForm(false)
			bb.setConditionalForm(false)
			if dest := rest[0]; !dest.Type.registerOrAlias() && !dest.Type.immediate() && dest.Type != typeId {
				return fmt.Errorf("Expected a register, alias, immediate, or label as the first argument to the given unconditional branch/call")
			} else {
				bb.setImmediateForm(dest.Type.immediate() || dest.Type == typeId)
				if dest.Type.registerOrAlias() {
					if q, err := this.resolveRegister(dest); err != nil {
						return err
					} else {
						d.Data[0] = q
					}
				} else if dest.Type.immediate() {
					d.SetImmediate(dest.Value.(Word))
				} else {
					if v, err := this.resolveLabel(dest.Value.(string)); err != nil {
						// defer it for now
						deferred = true
						this.Defer(&d, dest)
					} else {
						d.SetImmediate(v)
					}
				}
			}
		case 0:
			return fmt.Errorf("No arguments provided to given branch/call")
		default:
			return fmt.Errorf("Too many arguments provided to this jump op")
		}
	default:
		return fmt.Errorf("Illegal jump operation %s", first.Value)
	}
	d.Op = byte(bb)
	if deferred {
		return nil
	} else {
		return this.installInstruction(d.Encode())
	}
}
