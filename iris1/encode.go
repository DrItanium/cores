package iris1

import (
	//	"bufio"
	"fmt"
	"github.com/DrItanium/cores/lisp"
	"github.com/DrItanium/cores/parse/keyword"
	"github.com/DrItanium/cores/translation"
	"io"
	"strconv"
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
	{Symbol: "+", Src1: "?src1", Op: ArithmeticOpAdd},
	{Symbol: "-", Src1: "?src1", Op: ArithmeticOpSub},
	{Symbol: "*", Src1: "?src1", Op: ArithmeticOpMul},
	{Symbol: "/", Src1: "?src1", Op: ArithmeticOpDiv},
	{Symbol: "mod", Src1: "?src1", Op: ArithmeticOpRem},
	{Symbol: "<<", Src1: "?src1", Op: ArithmeticOpShiftLeft},
	{Symbol: ">>", Src1: "?src1", Op: ArithmeticOpShiftRight},
	{Symbol: "arithmetic-and", Src1: "?src1", Op: ArithmeticOpBinaryAnd},
	{Symbol: "arithmetic-or", Src1: "?src1", Op: ArithmeticOpBinaryOr},
	{Symbol: "arithmetic-not", Src1: "", Op: ArithmeticOpBinaryNot},
	{Symbol: "arithmetic-xor", Src1: "?src1", Op: ArithmeticOpBinaryXor},
	{Symbol: "1+", Src1: "?src1", Op: ArithmeticOpIncrement},
	{Symbol: "1-", Src1: "?src1", Op: ArithmeticOpDecrement},
	{Symbol: "2*", Src1: "?src1", Op: ArithmeticOpDouble},
	{Symbol: "2/", Src1: "?src1", Op: ArithmeticOpHalve},
	{Symbol: "+", Src1: "?upper", Op: ArithmeticOpAddImmediate},
	{Symbol: "-", Src1: "?upper", Op: ArithmeticOpSubImmediate},
	{Symbol: "*", Src1: "?upper", Op: ArithmeticOpMulImmediate},
	{Symbol: "/", Src1: "?upper", Op: ArithmeticOpDivImmediate},
	{Symbol: "mod", Src1: "?upper", Op: ArithmeticOpRemImmediate},
	{Symbol: "<<", Src1: "?upper", Op: ArithmeticOpShiftLeftImmediate},
	{Symbol: ">>", Src1: "?upper", Op: ArithmeticOpShiftRightImmediate},
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

func atomOfRegister(value interface{}) (byte, error) {
	if atom, err := asAtom(value); err != nil {
		return 0, err
	} else {
		str := atom.String()
		if !registers.IsKeyword(str) {
			return 0, fmt.Errorf("%s is not a register!", str)
		} else {
			return registers_StringToByte[str], nil
		}
	}
}
func immParse(atom lisp.Atom, width int) (uint64, error) {
	str := atom.String()
	// parse the immedate value
	if num, err := strconv.ParseUint(str, 10, width); err == nil {
		return num, nil
	} else if num, err := strconv.ParseUint(str, 16, width); err == nil {
		return num, nil
	} else if num, err := strconv.ParseUint(str, 2, width); err == nil {
		return num, nil
	} else {
		return 0, fmt.Errorf("Expected an %d bit immediate value in decimal, binary, or hex format. Got %s instead!", width, str)
	}
}
func applyToAtomNumeric(value interface{}, width int) (uint64, error) {
	if atom, err := asAtom(value); err != nil {
		return 0, err
	} else {
		return immParse(atom, width)
	}
}
func atomOfImm8(value interface{}) (byte, error) {
	val, err := applyToAtomNumeric(value, 8)
	return byte(val), err
}
func atomOfImm16(value interface{}) (Word, error) {
	val, err := applyToAtomNumeric(value, 16)
	return Word(val), err
}

type µcodeNetworkNode interface {
	Populate(lisp.List, instructionBuilder) error
}

type instructionBuilder map[string]interface{}

func (this instructionBuilder) set(name string, value interface{}) error {
	val, ok := this[name]
	if !ok {
		this[name] = value
	} else if val != value {
		return fmt.Errorf("Provided a different value for field %s, expected %s but got %s instead!", name, val, value)
	}
	return nil
}

func (this instructionBuilder) get(name string) (bool, interface{}) {
	val, result := this[name]
	return result, val
}

type µcodePopulator interface {
	Matches(interface{}, instructionBuilder) error
}
type µcodePopulatorFunc func(interface{}, instructionBuilder) error

func (this µcodePopulatorFunc) Matches(l interface{}, bld instructionBuilder) error {
	return this(l, bld)
}

func setImm8(l interface{}, bld instructionBuilder, field string) error {
	if word, err := atomOfImm8(l); err == nil {
		if ok, v := bld.get(field); ok {
			switch v.(type) {
			case byte:
				dat := v.(byte)
				if word != dat {
					return fmt.Errorf("Attempted to reassign %s from %d to %d", field, dat, word)
				}
			default:
				// it is a mismatch
				return fmt.Errorf("Attempted to reassign %s from %s to %d", field, v, word)
			}
		} else {
			bld.set(field, word)
		}
	} else if reg, err := atomOfRegister(l); err == nil {
		// we can't be a register!
		return fmt.Errorf("Attempted to assign %s to register r%d!", field, reg)
	} else if atom, err := asAtom(l); err != nil {
		return err
	} else {
		str := atom.String()
		if ok, v := bld.get(field); ok {
			switch v.(type) {
			case string:
				dat := v.(string)
				if str != dat {
					return fmt.Errorf("Attempted to reassign %s from %s to %s", field, dat, str)
				}
			default:
				return fmt.Errorf("Attempted to reassign %s from %s to %s", field, v, str)
			}
		} else {
			bld.set(field, str)
		}
	}
	return nil
}

func setRegister(l interface{}, bld instructionBuilder, field string) error {
	//TODO: add support for aliasing registers
	if reg, err := atomOfRegister(l); err != nil {
		return err
	} else if ok, content := bld.get(field); ok {
		dat := content.(byte)
		if dat != reg {
			return fmt.Errorf("Attempted to reassign %s from %d to %d. This is not allowed!", field, dat, reg)
		}
	} else {
		bld.set(field, reg)
	}
	return nil
}
func isSymbol(atom lisp.Atom, value string) bool {
	return atom.String() == value
}

func explicitSymbolMatch(symb string) µcodePopulator {
	return µcodePopulatorFunc(func(l interface{}, bld instructionBuilder) error {
		if atom, err := asAtom(l); err != nil {
			return err
		} else if !isSymbol(atom, symb) {
			return fmt.Errorf("Expected symbol %s but got %s instead.", symb, atom.String())
		} else {
			return nil
		}
	})
}

type µcodeListMatcher struct {
}

var variableTranslations = map[string]µcodePopulator{
	"?dest":   µcodePopulatorFunc(func(l interface{}, bld instructionBuilder) error { return setRegister(l, bld, "?dest") }),
	"?src0":   µcodePopulatorFunc(func(l interface{}, bld instructionBuilder) error { return setRegister(l, bld, "?src0") }),
	"?src1":   µcodePopulatorFunc(func(l interface{}, bld instructionBuilder) error { return setRegister(l, bld, "?src1") }),
	"?vector": µcodePopulatorFunc(func(l interface{}, bld instructionBuilder) error { return setImm8(l, bld, "?vector") }),
	"?upper":  µcodePopulatorFunc(func(l interface{}, bld instructionBuilder) error { return setImm8(l, bld, "?upper") }),
	"?lower":  µcodePopulatorFunc(func(l interface{}, bld instructionBuilder) error { return setImm8(l, bld, "?lower") }),
	"?word": µcodePopulatorFunc(func(l interface{}, bld instructionBuilder) error {
		if word, err := atomOfImm16(l); err == nil {
			if ok, v := bld.get("?word"); ok {
				switch v.(type) {
				case Word:
					dat := v.(Word)
					if word != dat {
						return fmt.Errorf("Attempted to reassign ?word from %d to %d", dat, word)
					}
				default:
					// it is a mismatch
					return fmt.Errorf("Attempted to reassign ?word from %s to %d", v, word)
				}
			} else {
				bld.set("?word", word)
			}
		} else if reg, err := atomOfRegister(l); err == nil {
			// we can't be a register!
			return fmt.Errorf("Attempted to assign ?word to register r%d!", reg)
		} else if atom, err := asAtom(l); err != nil {
			return err
		} else {
			str := atom.String()
			if ok, v := bld.get("?word"); ok {
				switch v.(type) {
				case string:
					dat := v.(string)
					if str != dat {
						return fmt.Errorf("Attempted to reassign ?word from %s to %s", dat, str)
					}
				default:
					return fmt.Errorf("Attempted to reassign ?word from %s to %s", v, str)
				}
			} else {
				bld.set("?word", str)
			}
		}
		return nil
	}),
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
		"(goto ?word)":                                  jumpCode(JumpOpUnconditionalImmediate),
		"(goto ?dest)":                                  jumpCode(JumpOpUnconditionalRegister),
		"(call ?word)":                                  jumpCode(JumpOpUnconditionalImmediateCall),
		"(call ?dest)":                                  jumpCode(JumpOpUnconditionalRegisterCall),
		"(if ?dest then (goto ?word))":                  jumpCode(JumpOpConditionalTrueImmediate),
		"(if (not ?dest) then (goto ?word))":            jumpCode(JumpOpConditionalFalseImmediate),
		"(if ?dest then (goto ?src0))":                  jumpCode(JumpOpConditionalTrueRegister),
		"(if (not ?dest) then (goto ?src0))":            jumpCode(JumpOpConditionalFalseRegister),
		"(if ?dest then (call ?word))":                  jumpCode(JumpOpConditionalTrueImmediateCall),
		"(if (not ?dest) then (call ?word))":            jumpCode(JumpOpConditionalFalseImmediateCall),
		"(if ?dest then (call ?src0))":                  jumpCode(JumpOpConditionalTrueRegisterCall),
		"(if (not ?dest) then (call ?src0))":            jumpCode(JumpOpConditionalFalseRegisterCall),
		"(goto (if ?dest then ?src0 else ?src1))":       jumpCode(JumpOpIfThenElseNormalPredTrue),
		"(goto (if (not ?dest) then ?src0 else ?src1))": jumpCode(JumpOpIfThenElseNormalPredFalse),
		"(call (if ?dest then ?src0 else ?src1))":       jumpCode(JumpOpIfThenElseCallPredTrue),
		"(call (if (not ?dest) then ?src0 else ?src1))": jumpCode(JumpOpIfThenElseCallPredFalse),
		// move forms
		"(set ?dest ?src0)":                      moveCode(MoveOpMove),
		"(set ?dest ?word)":                      moveCode(MoveOpSet),
		"(swap ?dest ?src0)":                     moveCode(MoveOpSwap),
		"(swap ?dest (data-at ?src0))":           moveCode(MoveOpSwapRegAddr),
		"(swap (data-at ?dest) (data-at ?src0))": moveCode(MoveOpSwapAddrAddr),
		"(swap ?dest (data-at ?word))":           moveCode(MoveOpSwapRegMem),
		"(swap (data-at ?dest) (data-at ?word))": moveCode(MoveOpSwapAddrMem),
		"(set ?dest (data-at ?src0))":            moveCode(MoveOpLoad),
		"(set ?dest (data-at ?word))":            moveCode(MoveOpLoadMem),
		"(set (data-at ?dest) ?src0)":            moveCode(MoveOpStore),
		"(set (data-at ?dest) (data-at ?src0))":  moveCode(MoveOpStoreAddr),
		"(set (data-at ?dest) (data-at ?word))":  moveCode(MoveOpStoreMem),
		"(set (data-at ?dest) ?word)":            moveCode(MoveOpStoreImm),
		"(push ?dest)":                           moveCode(MoveOpPush),
		"(push ?word)":                           moveCode(MoveOpPushImmediate),
		"(pop ?dest)":                            moveCode(MoveOpPop),
		"(peek ?dest)":                           moveCode(MoveOpPeek),
		// misc
		"(system ?vector ?src0 ?src1)": miscCode(MiscOpSystemCall),
	}
	arithmeticString := "(set ?dest (%s ?src0 %s))"
	for _, element := range µcodeArithmeticSymbols {
		µcode[fmt.Sprintf(arithmeticString, element.Symbol, element.Src1)] = arithmeticCode(element.Op)
	}
	combinatorialCompareString := "(set ?dest (%s ?dest (%s ?src0 ?src1)))"
	compareString := "(set ?dest (%s ?src0 ?src1))"
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

func parse(l lisp.List, out io.Writer) error {
	return nil
}
