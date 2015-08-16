package iris1

import (
	//	"bufio"
	"fmt"
	"github.com/DrItanium/cores/encoder"
	"github.com/DrItanium/cores/lisp"
	"github.com/DrItanium/cores/parse/keyword"
	"io"
	"log"
	//	"github.com/DrItanium/cores/parse/numeric"
)

const backendName = "iris1"

var keywords *keyword.Parser
var registers *keyword.Parser
var directives *keyword.Parser

var keywordTranslationTable map[string]byte

type translator func(lisp.List, io.Writer) error

func (this translator) Encode(l lisp.List, out io.Writer) error {
	return this(l, out)
}
func GetEncoder() encoder.Encoder {
	return translator(parse)
}
func init() {
	keywordTranslationTable = map[string]byte{
		// arithmetic ops
		"add":    byte((ArithmeticOpAdd << 3) | InstructionGroupArithmetic),
		"addi":   byte((ArithmeticOpAddImmediate << 3) | InstructionGroupArithmetic),
		"sub":    byte((ArithmeticOpSub << 3) | InstructionGroupArithmetic),
		"subi":   byte((ArithmeticOpSubImmediate << 3) | InstructionGroupArithmetic),
		"mul":    byte((ArithmeticOpMul << 3) | InstructionGroupArithmetic),
		"muli":   byte((ArithmeticOpMulImmediate << 3) | InstructionGroupArithmetic),
		"div":    byte((ArithmeticOpDiv << 3) | InstructionGroupArithmetic),
		"divi":   byte((ArithmeticOpDivImmediate << 3) | InstructionGroupArithmetic),
		"rem":    byte((ArithmeticOpRem << 3) | InstructionGroupArithmetic),
		"remi":   byte((ArithmeticOpRemImmediate << 3) | InstructionGroupArithmetic),
		"shl":    byte((ArithmeticOpShiftLeft << 3) | InstructionGroupArithmetic),
		"shli":   byte((ArithmeticOpShiftLeftImmediate << 3) | InstructionGroupArithmetic),
		"shr":    byte((ArithmeticOpShiftRight << 3) | InstructionGroupArithmetic),
		"shri":   byte((ArithmeticOpShiftRightImmediate << 3) | InstructionGroupArithmetic),
		"and":    byte((ArithmeticOpBinaryAnd << 3) | InstructionGroupArithmetic),
		"or":     byte((ArithmeticOpBinaryOr << 3) | InstructionGroupArithmetic),
		"not":    byte((ArithmeticOpBinaryNot << 3) | InstructionGroupArithmetic),
		"xor":    byte((ArithmeticOpBinaryXor << 3) | InstructionGroupArithmetic),
		"incr":   byte((ArithmeticOpIncrement << 3) | InstructionGroupArithmetic),
		"decr":   byte((ArithmeticOpDecrement << 3) | InstructionGroupArithmetic),
		"double": byte((ArithmeticOpDouble << 3) | InstructionGroupArithmetic),
		"halve":  byte((ArithmeticOpHalve << 3) | InstructionGroupArithmetic),
		// compare ops
		"eq":      byte((CompareOpEq << 3) | InstructionGroupCompare),
		"eq-and":  byte((CompareOpEqAnd << 3) | InstructionGroupCompare),
		"eq-or":   byte((CompareOpEqOr << 3) | InstructionGroupCompare),
		"eq-xor":  byte((CompareOpEqXor << 3) | InstructionGroupCompare),
		"neq":     byte((CompareOpNeq << 3) | InstructionGroupCompare),
		"neq-and": byte((CompareOpNeqAnd << 3) | InstructionGroupCompare),
		"neq-or":  byte((CompareOpNeqOr << 3) | InstructionGroupCompare),
		"neq-xor": byte((CompareOpNeqXor << 3) | InstructionGroupCompare),
		"lt":      byte((CompareOpLessThan << 3) | InstructionGroupCompare),
		"lt-and":  byte((CompareOpLessThanAnd << 3) | InstructionGroupCompare),
		"lt-or":   byte((CompareOpLessThanOr << 3) | InstructionGroupCompare),
		"lt-xor":  byte((CompareOpLessThanXor << 3) | InstructionGroupCompare),
		"le":      byte((CompareOpLessThanOrEqualTo << 3) | InstructionGroupCompare),
		"le-and":  byte((CompareOpLessThanOrEqualToAnd << 3) | InstructionGroupCompare),
		"le-or":   byte((CompareOpLessThanOrEqualToOr << 3) | InstructionGroupCompare),
		"le-xor":  byte((CompareOpLessThanOrEqualToXor << 3) | InstructionGroupCompare),
		"gt":      byte((CompareOpGreaterThan << 3) | InstructionGroupCompare),
		"gt-and":  byte((CompareOpGreaterThanAnd << 3) | InstructionGroupCompare),
		"gt-or":   byte((CompareOpGreaterThanOr << 3) | InstructionGroupCompare),
		"gt-xor":  byte((CompareOpGreaterThanXor << 3) | InstructionGroupCompare),
		"ge":      byte((CompareOpGreaterThanOrEqualTo << 3) | InstructionGroupCompare),
		"ge-and":  byte((CompareOpGreaterThanOrEqualToAnd << 3) | InstructionGroupCompare),
		"ge-or":   byte((CompareOpGreaterThanOrEqualToOr << 3) | InstructionGroupCompare),
		"ge-xor":  byte((CompareOpGreaterThanOrEqualToXor << 3) | InstructionGroupCompare),
		// Jump operations
		"goto-imm":        byte((JumpOpUnconditionalImmediate << 3) | InstructionGroupJump),
		"call-imm":        byte((JumpOpUnconditionalImmediateCall << 3) | InstructionGroupJump),
		"goto-reg":        byte((JumpOpUnconditionalRegister << 3) | InstructionGroupJump),
		"call-reg":        byte((JumpOpUnconditionalRegisterCall << 3) | InstructionGroupJump),
		"goto-imm-if1":    byte((JumpOpConditionalTrueImmediate << 3) | InstructionGroupJump),
		"call-imm-if1":    byte((JumpOpConditionalTrueImmediateCall << 3) | InstructionGroupJump),
		"goto-imm-if0":    byte((JumpOpConditionalFalseImmediate << 3) | InstructionGroupJump),
		"call-imm-if0":    byte((JumpOpConditionalFalseImmediateCall << 3) | InstructionGroupJump),
		"goto-reg-if1":    byte((JumpOpConditionalTrueRegister << 3) | InstructionGroupJump),
		"call-reg-if1":    byte((JumpOpConditionalTrueRegisterCall << 3) | InstructionGroupJump),
		"goto-reg-if0":    byte((JumpOpConditionalFalseRegister << 3) | InstructionGroupJump),
		"call-reg-if0":    byte((JumpOpConditionalFalseRegisterCall << 3) | InstructionGroupJump),
		"call-select-if1": byte((JumpOpIfThenElseCallPredTrue << 3) | InstructionGroupJump),
		"call-select-if0": byte((JumpOpIfThenElseCallPredFalse << 3) | InstructionGroupJump),
		"goto-select-if1": byte((JumpOpIfThenElseNormalPredTrue << 3) | InstructionGroupJump),
		"goto-select-if0": byte((JumpOpIfThenElseNormalPredFalse << 3) | InstructionGroupJump),
	}
	// setup the keywords and register parsers
	registers = keyword.New()
	for i := 0; i < RegisterCount; i++ {
		registers.AddKeyword(fmt.Sprintf("r%d", i))
	}
	keywords = keyword.New()
	// arithmetic ops
	keywords.AddKeywordList([]string{
		"add",
		"addi",
		"sub",
		"subi",
		"mul",
		"muli",
		"div",
		"divi",
		"rem",
		"remi",
		"shl",
		"shli",
		"shr",
		"shri",
		"and",
		"or",
		"not",
		"xor",
		"incr",
		"decr",
		"double",
		"halve",
	})
	// compare ops
	keywords.AddKeywordList([]string{
		"eq",
		"eq-and",
		"eq-or",
		"eq-xor",
		"neq",
		"neq-and",
		"neq-or",
		"neq-xor",
		"lt",
		"lt-and",
		"lt-or",
		"lt-xor",
		"gt",
		"gt-and",
		"gt-or",
		"gt-xor",
		"le",
		"le-and",
		"le-or",
		"le-xor",
		"ge",
		"ge-and",
		"ge-or",
		"ge-xor",
	})
	// jump operations
	keywords.AddKeywordList([]string{
		"goto-imm",
		"call-imm",
		"goto-reg",
		"call-reg",
		"goto-imm-if1",
		"call-imm-if1",
		"goto-reg-if1",
		"call-reg-if1",
		"call-select-if1",
		"goto-select-if1",
		"goto-imm-if0",
		"call-imm-if0",
		"goto-reg-if0",
		"call-reg-if0",
		"call-select-if0",
		"goto-select-if0",
	})
	// move operations
	keywords.AddKeywordList([]string{
		"move",
		"swap",
		"swap-reg-addr",
		"swap-addr-addr",
		"swap-reg-mem",
		"swap-addr-mem",
		"set",
		"load",
		"load-mem",
		"store",
		"store-addr",
		"store-mem",
		"store-imm",
		"push",
		"push-imm",
		"pop",
		"peek",
	})
	// misc operations
	keywords.AddKeyword("system")
	directives := keyword.New()
	// directives
	directives.AddKeywordList([]string{
		"label",
		"org",
		"segment",
		"value",
		"string",
	})

}

type extendedCore struct {
	Core     *Core
	Labels   map[string]Word
	segment  string
	instAddr Word
	dataAddr Word
}
type coreTransformer func(*extendedCore, lisp.List) error

var directiveTranslations = map[string]coreTransformer{
	"label": func(core *extendedCore, contents lisp.List) error {
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
					log.Printf("Constructed label: %s", str)
				}
			default:
				return fmt.Errorf("Label argument was not an atom, instead it was a %t", t)
			}
			//if _, ok := core.Labels[contents
			return nil
		}
	},
	"org": nil,
}

func parse(l lisp.List, out io.Writer) error {
	// now iterate through all the set of lisp lists
	var core extendedCore
	if c, err := New(); err != nil {
		return err
	} else {
		core.Core = c
		core.Labels = make(map[string]Word)
	}
	// buildup the core
	for _, element := range l {
		// if we encounter an atom at the top level then we should ignore it
		switch typ := element.(type) {
		case lisp.Atom:
			log.Printf("Ignoring atom %s", element)
		case lisp.List:
			nList := element.(lisp.List)
			if err := _ParseList(&core, nList); err != nil {
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

func _ParseList(core *extendedCore, l lisp.List) error {
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
