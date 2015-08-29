package iris1

import (
	//	"bufio"
	"fmt"
	"github.com/DrItanium/cores/lisp"
	"github.com/DrItanium/cores/parse/keyword"
	"github.com/DrItanium/cores/translation"
	"io"
	"log"
	"strconv"
	"strings"
	"unicode"
)

var keywords *keyword.Parser
var registers *keyword.Parser
var directives *keyword.Parser

var keywordTranslationTable map[string]byte

type translator func(lisp.List, io.Writer) error

func (this translator) Encode(l lisp.List, out io.Writer) error {
	return this(l, out)
}
func GetEncoder() translation.Encoder {
	return translator(parse)
}

const (
	DataAtKeyword = "data-at"
)

var registers_StringToByte map[string]byte
var µcodeArithmeticSymbols = []struct {
	Symbol, Src1 string
	Op           byte
}{
	{Symbol: "+", Src1: "r:src1", Op: ArithmeticOpAdd},
	{Symbol: "-", Src1: "r:src1", Op: ArithmeticOpSub},
	{Symbol: "*", Src1: "r:src1", Op: ArithmeticOpMul},
	{Symbol: "/", Src1: "r:src1", Op: ArithmeticOpDiv},
	{Symbol: "mod", Src1: "r:src1", Op: ArithmeticOpRem},
	{Symbol: "<<", Src1: "r:src1", Op: ArithmeticOpShiftLeft},
	{Symbol: ">>", Src1: "r:src1", Op: ArithmeticOpShiftRight},
	{Symbol: "arithmetic-and", Src1: "r:src1", Op: ArithmeticOpBinaryAnd},
	{Symbol: "arithmetic-or", Src1: "r:src1", Op: ArithmeticOpBinaryOr},
	{Symbol: "arithmetic-not", Src1: "", Op: ArithmeticOpBinaryNot},
	{Symbol: "arithmetic-xor", Src1: "r:src1", Op: ArithmeticOpBinaryXor},
	{Symbol: "1+", Src1: "r:src1", Op: ArithmeticOpIncrement},
	{Symbol: "1-", Src1: "r:src1", Op: ArithmeticOpDecrement},
	{Symbol: "2*", Src1: "r:src1", Op: ArithmeticOpDouble},
	{Symbol: "2/", Src1: "r:src1", Op: ArithmeticOpHalve},
	{Symbol: "+", Src1: "i8:src1", Op: ArithmeticOpAddImmediate},
	{Symbol: "-", Src1: "i8:src1", Op: ArithmeticOpSubImmediate},
	{Symbol: "*", Src1: "i8:src1", Op: ArithmeticOpMulImmediate},
	{Symbol: "/", Src1: "i8:src1", Op: ArithmeticOpDivImmediate},
	{Symbol: "mod", Src1: "i8:src1", Op: ArithmeticOpRemImmediate},
	{Symbol: "<<", Src1: "i8:src1", Op: ArithmeticOpShiftLeftImmediate},
	{Symbol: ">>", Src1: "i8:src1", Op: ArithmeticOpShiftRightImmediate},
}

var µcodeCompareSymbols = []struct {
	Symbol        string
	Combinatorial bool
	CombineSymbol string
	Op            byte
}{
	{Symbol: "=", Op: CompareOpEq},
	{Symbol: "=", Combinatorial: true, CombineSymbol: "&&", Op: CompareOpEqAnd},
	{Symbol: "=", Combinatorial: true, CombineSymbol: "||", Op: CompareOpEqOr},
	{Symbol: "=", Combinatorial: true, CombineSymbol: "xor", Op: CompareOpEqXor},
	{Symbol: "<>", Op: CompareOpNeq},
	{Symbol: "<>", Combinatorial: true, CombineSymbol: "&&", Op: CompareOpNeqAnd},
	{Symbol: "<>", Combinatorial: true, CombineSymbol: "||", Op: CompareOpNeqOr},
	{Symbol: "<>", Combinatorial: true, CombineSymbol: "xor", Op: CompareOpNeqXor},
	{Symbol: "<", Op: CompareOpLessThan},
	{Symbol: "<", Combinatorial: true, CombineSymbol: "&&", Op: CompareOpLessThanAnd},
	{Symbol: "<", Combinatorial: true, CombineSymbol: "||", Op: CompareOpLessThanOr},
	{Symbol: "<", Combinatorial: true, CombineSymbol: "xor", Op: CompareOpLessThanXor},
	{Symbol: ">", Op: CompareOpGreaterThan},
	{Symbol: ">", Combinatorial: true, CombineSymbol: "&&", Op: CompareOpGreaterThanAnd},
	{Symbol: ">", Combinatorial: true, CombineSymbol: "||", Op: CompareOpGreaterThanOr},
	{Symbol: ">", Combinatorial: true, CombineSymbol: "xor", Op: CompareOpGreaterThanXor},
	{Symbol: "<=", Op: CompareOpLessThanOrEqualTo},
	{Symbol: "<=", Combinatorial: true, CombineSymbol: "&&", Op: CompareOpLessThanOrEqualToAnd},
	{Symbol: "<=", Combinatorial: true, CombineSymbol: "||", Op: CompareOpLessThanOrEqualToOr},
	{Symbol: "<=", Combinatorial: true, CombineSymbol: "xor", Op: CompareOpLessThanOrEqualToXor},
	{Symbol: ">=", Op: CompareOpGreaterThanOrEqualTo},
	{Symbol: ">=", Combinatorial: true, CombineSymbol: "&&", Op: CompareOpGreaterThanOrEqualToAnd},
	{Symbol: ">=", Combinatorial: true, CombineSymbol: "||", Op: CompareOpGreaterThanOrEqualToOr},
	{Symbol: ">=", Combinatorial: true, CombineSymbol: "xor", Op: CompareOpGreaterThanOrEqualToXor},
}

// generator strings
var µcode map[string]string

type networkMatcher func(interface{}) bool

/*
// the network used to perform matching
type networkNode struct {
	fn       networkMatcher
	collect  bool
	children []*networkNode
}

func (this *networkNode) Match(cell interface{}) bool {
	return this.fn(cell)
}
func newNetworkNode(list lisp.List) (*networkNode, error) {
	if len(list) == 0 {
		return nil, fmt.Errorf("empty list!")
	}
	first := list[0]
	switch t := first.(type) {
	case lisp.Atom:
		at := first.(lisp.Atom)
		str := at.String()
		if str == "r" {
			// it is a register so check for that
		} else if str == "i8" {

		} else if str == "i16" {

		} else {

		}
	case lisp.List:
	default:
		return nil, fmt.Errorf("unknown type %t found during µcode network creation!", t)
	}
	return nil, nil
}

var top *networkNode
*/
func asList(value interface{}) (lisp.List, error) {
	switch t := value.(type) {
	case lisp.List:
		return value.(lisp.List), nil
	default:
		return nil, fmt.Errorf("Value of type %t is not a lisp.List", t)
	}
}
func asAtom(value interface{}) (lisp.Atom, error) {
	switch t := value.(type) {
	case lisp.Atom:
		return value.(lisp.Atom), nil
	default:
		return nil, fmt.Errorf("Value of type %t is not a lisp.Atom", t)
	}
}
func isSymbol(atom lisp.Atom, value string) bool {
	return atom.String() == value
}
func isImmediate(atom lisp.Atom) bool {
	return isImmediate16(atom) || isImmediate8(atom)
}
func isImmediate16(atom lisp.Atom) bool {
	return isSymbol(atom, "i16")
}
func isImmediate8(atom lisp.Atom) bool {
	return isSymbol(atom, "i8")
}
func isDestinationRegister(atom lisp.Atom) bool {
	return isSymbol(atom, "r:dest")
}
func isSource0Register(atom lisp.Atom) bool {
	return isSymbol(atom, "r:src0")
}
func isSource1Register(atom lisp.Atom) bool {
	return isSymbol(atom, "r:src1")
}
func isSpecificSymbol(str string) func(lisp.Atom) bool {
	return func(at lisp.Atom) bool {
		return isSymbol(at, str)
	}
}

type instructionBuilder map[string]interface{}

func parseNumber(atom lisp.Atom, width int) (byte, error) {
	str := atom.String()
	if num, err := strconv.ParseUint(str, 10, width); err != nil {
		return 0, err
	} else {
		return byte(num), nil
	}
}
func listOfLength(value interface{}, count int) (lisp.List, error) {
	if l, err := asList(value); err != nil {
		return nil, err
	} else if len(l) != count {
		return nil, fmt.Errorf("expected list to be of length %d but it was actually %d", count, len(l))
	} else {
		return l, nil
	}
}
func atomOfSymbol(value interface{}, symb string) (lisp.Atom, error) {
	if atom, err := asAtom(value); err != nil {
		return nil, err
	} else if !isSymbol(atom, symb) {
		return nil, fmt.Errorf("Expected symbol %s but got %s instead.", symb, atom.String())
	} else {
		return atom, nil
	}
}
func parseTupleStringNumber(list interface{}, title string, width int) (byte, error) {
	if l, err := listOfLength(list, 2); err != nil {
		return 0, err
	} else if _, err := atomOfSymbol(l[0], title); err != nil {
		return 0, err
	} else if val, err := asAtom(l[1]); err != nil {
		return 0, err
	} else {
		return parseNumber(val, width)
	}
}
func parseGroupEncoding(list interface{}) (byte, error) {
	return parseTupleStringNumber(list, "group", 3)
}
func parseOpEncoding(list interface{}) (byte, error) {
	return parseTupleStringNumber(list, "op", 5)
}

func parseControlEncoding(q interface{}) (byte, error) {
	if lst, err := listOfLength(q, 3); err != nil {
		return 0, err
	} else if _, err := atomOfSymbol(lst[0], "control"); err != nil {
		return 0, err
	} else if group, err := parseGroupEncoding(lst[1]); err != nil {
		return 0, err
	} else if op, err := parseOpEncoding(lst[2]); err != nil {
		return 0, err
	} else {
		return byte(op<<3) | group, nil
	}
}

func init() {

	genControlCode := func(group, op byte) string {
		return fmt.Sprintf("(control (group %d) (op %d))", group, op)
	}
	jumpCode := func(op byte) string {
		return genControlCode(InstructionGroupJump, op)
	}
	moveCode := func(op byte) string {
		return genControlCode(InstructionGroupMove, op)
	}
	arithmeticCode := func(op byte) string {
		return genControlCode(InstructionGroupArithmetic, op)
	}
	compareCode := func(op byte) string {
		return genControlCode(InstructionGroupCompare, op)
	}
	miscCode := func(op byte) string {
		return genControlCode(InstructionGroupMisc, op)
	}
	µcode = map[string]string{
		// jump operations
		"(goto i16)":                                       jumpCode(JumpOpUnconditionalImmediate),
		"(goto r:dest)":                                    jumpCode(JumpOpUnconditionalRegister),
		"(call i16)":                                       jumpCode(JumpOpUnconditionalImmediateCall),
		"(call r:dest)":                                    jumpCode(JumpOpUnconditionalRegisterCall),
		"(if r:dest then (goto i16))":                      jumpCode(JumpOpConditionalTrueImmediate),
		"(if (not r:dest) then (goto i16))":                jumpCode(JumpOpConditionalFalseImmediate),
		"(if r:dest then (goto r:src0))":                   jumpCode(JumpOpConditionalTrueRegister),
		"(if (not r:dest) then (goto r:src0))":             jumpCode(JumpOpConditionalFalseRegister),
		"(if r:dest then (call i16))":                      jumpCode(JumpOpConditionalTrueImmediateCall),
		"(if (not r:dest) then (call i16))":                jumpCode(JumpOpConditionalFalseImmediateCall),
		"(if r:dest then (call r:src0))":                   jumpCode(JumpOpConditionalTrueRegisterCall),
		"(if (not r:dest) then (call r:src0))":             jumpCode(JumpOpConditionalFalseRegisterCall),
		"(goto (if r:dest then r:src0 else r:src1))":       jumpCode(JumpOpIfThenElseNormalPredTrue),
		"(goto (if (not r:dest) then r:src0 else r:src1))": jumpCode(JumpOpIfThenElseNormalPredFalse),
		"(call (if r:dest then r:src0 else r:src1))":       jumpCode(JumpOpIfThenElseCallPredTrue),
		"(call (if (not r:dest) then r:src0 else r:src1))": jumpCode(JumpOpIfThenElseCallPredFalse),
		// move forms
		"(set r:dest r:src0)":                      moveCode(MoveOpMove),
		"(set r:dest i16)":                         moveCode(MoveOpSet),
		"(swap r:dest r:src0)":                     moveCode(MoveOpSwap),
		"(swap r:dest (data-at r:src0))":           moveCode(MoveOpSwapRegAddr),
		"(swap (data-at r:dest) (data-at r:src0))": moveCode(MoveOpSwapAddrAddr),
		"(swap r:dest (data-at i16))":              moveCode(MoveOpSwapRegMem),
		"(swap (data-at r:dest) (data-at i16))":    moveCode(MoveOpSwapAddrMem),
		"(set r:dest (data-at r:src0))":            moveCode(MoveOpLoad),
		"(set r:dest (data-at i16))":               moveCode(MoveOpLoadMem),
		"(set (data-at r:dest) r:src0)":            moveCode(MoveOpStore),
		"(set (data-at r:dest) (data-at r:src0))":  moveCode(MoveOpStoreAddr),
		"(set (data-at r:dest) (data-at i16))":     moveCode(MoveOpStoreMem),
		"(set (data-at r:dest) i16)":               moveCode(MoveOpStoreImm),
		"(push r:dest)":                            moveCode(MoveOpPush),
		"(push i16)":                               moveCode(MoveOpPushImmediate),
		"(pop r:dest)":                             moveCode(MoveOpPop),
		"(peek r:dest)":                            moveCode(MoveOpPeek),
		// misc
		"(system i8 r:src0 r:src1)": miscCode(MiscOpSystemCall),
	}
	arithmeticString := "(set r:dest (%s r:src0 %s))"
	for _, element := range µcodeArithmeticSymbols {
		µcode[fmt.Sprintf(arithmeticString, element.Symbol, element.Src1)] = arithmeticCode(element.Op)
	}
	combinatorialCompareString := "(set r:dest (%s r:dest (%s r:src0 r:src1)))"
	compareString := "(set r:dest (%s r:src0 r:src1))"
	for _, element := range µcodeCompareSymbols {
		var msg string
		if element.Combinatorial {
			msg = fmt.Sprintf(combinatorialCompareString, element.CombineSymbol, element.Symbol)
		} else {
			msg = fmt.Sprintf(compareString, element.Symbol)
		}
		µcode[msg] = compareCode(element.Op)
	}
	// setup the keywords and register parsers
	registers = keyword.New()
	registers_StringToByte = make(map[string]byte)
	for i := 0; i < RegisterCount; i++ {
		str := fmt.Sprintf("r%d", i)
		registers.AddKeyword(str)
		registers_StringToByte[str] = byte(i)
	}
	// now just iterate through the table we built and store the keys in the set of keywords
	keywords = keyword.New()
	for key, _ := range keywordTranslationTable {
		keywords.AddKeyword(key)
	}
	directives := keyword.New()
	// directives
	directives.AddKeywordList([]string{
		"label",
		"org",
		"segment",
		"value",
		"string",
	})

	// now setup the match "network"
}

type segment struct {
	Address Word
	Labels  map[string]Word
}

func newSegment() *segment {
	var s segment
	s.Labels = make(map[string]Word)
	return &s
}

type extendedCore struct {
	Core        *Core
	segments    map[string]*segment
	currSegment *segment
	segName     string
}

// registers the given name (in the current segment
func (this *extendedCore) registerLabel(name string) error {
	if addr, ok := this.currSegment.Labels[name]; ok {
		return fmt.Errorf("Label %s is already defined for address %x in segment %s", name, addr, this.segName)
	} else {
		this.currSegment.Labels[name] = this.currSegment.Address
		return nil
	}
}

func (this *extendedCore) changeSegment(name string) error {
	if name == this.segName {
		return nil // we're already there so do nothing
	}
	if seg, ok := this.segments[name]; !ok {
		return fmt.Errorf("Illegal segment %s!", name)
	} else {
		this.currSegment = seg
		this.segName = name
		return nil
	}
}
func (this *extendedCore) getCurrentSegmentName() string {
	return this.segName
}

// set the address of the current segment
func (this *extendedCore) setAddress(addr Word) {
	this.currSegment.Address = addr
}

type coreTransformer func(*extendedCore, lisp.List) error

func handleLabel(core *extendedCore, contents lisp.List) error {
	// first argument check
	if len(contents) == 1 {
		return fmt.Errorf("No titles passed to label")
	} else if len(contents) > 2 {
		return fmt.Errorf("Too many arguments passed to label!")
	} else {
		// do a type check on the actual argument
		arg := contents[1]
		switch t := arg.(type) {
		case lisp.Atom:
			// check to see if we are looking at a keyword, directive, or register
			atom := arg.(lisp.Atom)
			str := atom.String()
			if keywords.IsKeyword(str) {
				return fmt.Errorf("Illegal label name '%s', is a operation name!", str)
			} else if directives.IsKeyword(str) {
				return fmt.Errorf("Illegal label name '%s', is a directive!", str)
			} else if registers.IsKeyword(str) {
				return fmt.Errorf("Illegal label name '%s', is a register!", str)
			} else {
				return core.registerLabel(str)
			}
		default:
			return fmt.Errorf("Label argument was not an atom, instead it was a %t", t)
		}
		return nil
	}
}

func handleSegment(core *extendedCore, contents lisp.List) error {
	if len(contents) == 1 {
		return fmt.Errorf("No segment name provided!")
	} else if len(contents) > 2 {
		return fmt.Errorf("Too many arguments passed to the segment directive!")
	} else {
		arg := contents[1]
		switch t := arg.(type) {
		case lisp.Atom:
			atom := arg.(lisp.Atom)
			return core.changeSegment(atom.String())
		default:
			return fmt.Errorf("Segment argument was not an atom, instead it was a %t", t)
		}
	}
}

func handleOrg(core *extendedCore, contents lisp.List) error {
	if len(contents) == 1 {
		return fmt.Errorf("Too few arguments provided to the org directive!")
	} else if len(contents) > 2 {
		return fmt.Errorf("Too many arguments provided to the org directive!")
	} else {
		arg := contents[1]
		switch t := arg.(type) {
		case lisp.Atom:
			// we need to now do numeric conversion
			atom := arg.(lisp.Atom)
			str := atom.String()
			var addr Word
			// check and see if we can parse this as any kind of number
			if hex, err := _parseHexNumber(str); err == nil {
				addr = hex
			} else if bits, err := _parseBinaryNumber(str); err == nil {
				addr = bits
			} else if number, err := _parseDecimalNumber(str); err == nil {
				addr = number
			} else {
				return fmt.Errorf("Provided \"number\" (%s) is not a parseable or legal number!", str)
			}
			// we found a conversion so update the core with the new address of the current segment
			core.setAddress(addr)
			return nil
		default:
			return fmt.Errorf("Org argument was not an atom, instead it was a %t", t)
		}
	}
}

var directiveTranslations = map[string]coreTransformer{
	"label":   handleLabel,
	"segment": handleSegment,
	"org":     handleOrg,
}

func parse(l lisp.List, out io.Writer) error {
	// now iterate through all the set of lisp lists
	var core extendedCore
	if c, err := New(); err != nil {
		return err
	} else {
		// setup the core
		core.Core = c
		code := newSegment()
		core.currSegment = code
		core.segName = "code"
		core.segments = map[string]*segment{
			"code": code,
			"data": newSegment(),
		}
	}
	// buildup the core
	for _, element := range l {
		// if we encounter an atom at the top level then we should ignore it
		switch typ := element.(type) {
		case lisp.Atom:
			log.Printf("Ignoring atom %s", element)
		case lisp.List:
			nList := element.(lisp.List)
			if err := _parseList(&core, nList); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unknown type %t", typ)

		}
	}
	// output the core to the io.Writer
	//bout := bufio.NewWriter(out)
	return nil
}
func _parseGenericNumber(base int, prefix, input string) (Word, error) {
	var num string
	if strings.HasPrefix(input, prefix) {
		num = input[len(prefix):]
	} else {
		num = input
	}
	val, err := strconv.ParseUint(num, base, 16)
	return Word(val), err
}
func _parseHexNumber(input string) (Word, error) {
	return _parseGenericNumber(16, "0x", input)
}
func _parseBinaryNumber(input string) (Word, error) {
	return _parseGenericNumber(2, "0b", input)
}
func _parseDecimalNumber(input string) (Word, error) {
	return _parseGenericNumber(0, "", input)
}
func _parseNumber(input string) (Word, error) {
	if val, err := _parseDecimalNumber(input); err == nil {
		return val, nil
	} else if val, err = _parseBinaryNumber(input); err == nil {
		return val, nil
	} else if val, err = _parseHexNumber(input); err == nil {
		return val, nil
	} else {
		return 0, fmt.Errorf("Couldn't parse given input: %s", err)
	}
}
func isNumber(input string) bool {
	_, err := _parseNumber(input)
	return err == nil
}

func _parseList(core *extendedCore, l lisp.List) error {
	// use the first arg as the op and the rest as arguments
	if len(l) == 0 {
		return fmt.Errorf("Empty list provided!")
	}
	first := l[0]
	//rest := l[1:]
	switch t := first.(type) {
	case lisp.Atom:
		atom := first.(lisp.Atom)
		if keywords.IsKeyword(atom.String()) {
			// determine what kind of operation we are looking at
			log.Printf("%s", atom)
		} else if directives.IsKeyword(atom.String()) {
			if fn, ok := directiveTranslations[atom.String()]; !ok {
				return fmt.Errorf("Unimplemented directive: %s", atom.String())
			} else {
				return fn(core, l)
			}
		} else {
			return fmt.Errorf("First argument (%s) is not a keyword nor directive!", atom)
		}
	default:
		return fmt.Errorf("ERROR: first argument (%s) of operation is not an atom (%t),", first, t)
	}
	return nil
}

func _parseDonuts(core *extendedCore, list lisp.List) error {
	switch len(list) {
	case 3:
		return parseThreeElementList(core, list)
	default:
		return fmt.Errorf("Provided list %s doesn't match up to anything!", list)
	}
}

type parseFunction func(*extendedCore, lisp.List) error

var threeElementFuncs = map[string]parseFunction{
	"set": parseSet,
}

func parseThreeElementList(core *extendedCore, list lisp.List) error {
	car := list[0]
	cdr := list[1:]
	switch t := car.(type) {
	case lisp.Atom:
		first := car.(lisp.Atom)
		if fn, ok := threeElementFuncs[string(first)]; ok {
			return fn(core, cdr)
		} else {
			return fmt.Errorf("Unknown op %s in %s", first, list)
		}
	default:
		return fmt.Errorf("First element of %s is not an atom, it is a %t!", list, t)
	}
}
func isKeywordOrNumber(atom lisp.Atom) bool {
	str := atom.String()
	return registers.IsKeyword(str) || isNumber(str)
}
func equalsLiteralValue(atom lisp.Atom, str string) bool {
	return str == atom.String()
}
func isQuoted(runes []rune) bool {
	return runes[0] == '"' && runes[len(runes)-1] == '"'
}
func isSingleWord(runes []rune) bool {
	if !isQuoted(runes) {
		for _, r := range runes {
			if !unicode.IsPrint(r) || unicode.IsSpace(r) || r == '"' || r == '\'' {
				return false
			}
		}
		return true
	}
	return false
}
func isLabel(atom lisp.Atom) bool {
	runes := []rune(atom.String())
	f := runes[0]
	return unicode.IsPrint(f) && !unicode.IsNumber(f) && f != 'r' && isSingleWord(runes)

}
func isDataAtPhrase(list lisp.List) bool {
	if len(list) == 2 {
		// we need to go through this
		first := list[0]
		switch first.(type) {
		case lisp.Atom:
			// now check to see if it is the data-at keyword
			if equalsLiteralValue(first.(lisp.Atom), DataAtKeyword) {
				// now check the second argument
				second := list[1]
				switch second.(type) {
				case lisp.Atom:
					return isKeywordOrNumber(second.(lisp.Atom))
				default:
				}
			}
		default:
		}
	}
	return false
}
func parseSet_FirstRegister(core *extendedCore, first byte, second interface{}) error {
	switch t := second.(type) {
	case lisp.Atom:
		// move
		at := second.(lisp.Atom)
		str := at.String()
		// can either be an immediate or a register
		if registers.IsKeyword(str) {
			// we found a register!
		} else if _, err := _parseNumber(str); err == nil {
			// immediate form

		} else if isLabel(at) {
			// it is a label so we just save it for now
		} else {
			return fmt.Errorf("Unknown atom %s passed as second argument to (set r%d %s)!", at, first, at)
		}
	case lisp.List:
		// arithmetic, compare, and some move
		// we now check the contents of the sublist
		sublist := second.(lisp.List)
		if isDataAtPhrase(sublist) {

		} else {
			return fmt.Errorf("Unknown sublist %s provided as second argument to set function", sublist)
		}
	default:
		return fmt.Errorf("Second argument of (set r%d %s) is of disallowed type %t!", first, second, t)
	}
	return nil
}
func parseSet_FirstList(core *extendedCore, first lisp.List, second interface{}) error {
	switch t := second.(type) {
	case lisp.Atom:
	case lisp.List:
		sublist := second.(lisp.List)
		if isDataAtPhrase(sublist) {

		} else {
			return fmt.Errorf("Unknown sublist %s provided as second argument to set function!", sublist)
		}
	default:
		return fmt.Errorf("Second argument of (set %s %s) is of disallowed type %t!", first, second, t)
	}
	return nil
}
func parseSet(core *extendedCore, args lisp.List) error {
	first := args[0]
	second := args[1]
	// check the type of the first argument
	switch tFirst := first.(type) {
	case lisp.Atom:
		// we are now dealing with either arithmetic, move, or compare ops
		str := first.(lisp.Atom).String()
		if registers.IsKeyword(str) {
			// now we need to checkout the second argument
			return parseSet_FirstRegister(core, registers_StringToByte[str], second)
		} else {
			return fmt.Errorf("First argument of (set %s %s) is not a register!", first, second)
		}
	case lisp.List:
		// this goes to the move operations
		return parseSet_FirstList(core, first.(lisp.List), second)
	default:
		return fmt.Errorf("First argument of (set %s %s) is of incorrect type %t!", first, second, tFirst)
	}
}
