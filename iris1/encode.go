package iris1

import (
	//	"bufio"
	"fmt"
	"github.com/DrItanium/cores/lisp"
	"github.com/DrItanium/cores/parse/keyword"
	"github.com/DrItanium/cores/translation"
	//	"github.com/DrItanium/cores/parse/numeric"
	"io"
	"log"
	"strconv"
	"strings"
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
func init() {
	genControlByte := func(op, group byte) byte {
		return (op << 3) | group
	}
	keywordTranslationTable = map[string]byte{
		// misc ops
		"system": genControlByte(MiscOpSystemCall, InstructionGroupMisc),
		// arithmetic ops
		"add":    genControlByte(ArithmeticOpAdd, InstructionGroupArithmetic),
		"addi":   genControlByte(ArithmeticOpAddImmediate, InstructionGroupArithmetic),
		"sub":    genControlByte(ArithmeticOpSub, InstructionGroupArithmetic),
		"subi":   genControlByte(ArithmeticOpSubImmediate, InstructionGroupArithmetic),
		"mul":    genControlByte(ArithmeticOpMul, InstructionGroupArithmetic),
		"muli":   genControlByte(ArithmeticOpMulImmediate, InstructionGroupArithmetic),
		"div":    genControlByte(ArithmeticOpDiv, InstructionGroupArithmetic),
		"divi":   genControlByte(ArithmeticOpDivImmediate, InstructionGroupArithmetic),
		"rem":    genControlByte(ArithmeticOpRem, InstructionGroupArithmetic),
		"remi":   genControlByte(ArithmeticOpRemImmediate, InstructionGroupArithmetic),
		"shl":    genControlByte(ArithmeticOpShiftLeft, InstructionGroupArithmetic),
		"shli":   genControlByte(ArithmeticOpShiftLeftImmediate, InstructionGroupArithmetic),
		"shr":    genControlByte(ArithmeticOpShiftRight, InstructionGroupArithmetic),
		"shri":   genControlByte(ArithmeticOpShiftRightImmediate, InstructionGroupArithmetic),
		"and":    genControlByte(ArithmeticOpBinaryAnd, InstructionGroupArithmetic),
		"or":     genControlByte(ArithmeticOpBinaryOr, InstructionGroupArithmetic),
		"not":    genControlByte(ArithmeticOpBinaryNot, InstructionGroupArithmetic),
		"xor":    genControlByte(ArithmeticOpBinaryXor, InstructionGroupArithmetic),
		"incr":   genControlByte(ArithmeticOpIncrement, InstructionGroupArithmetic),
		"decr":   genControlByte(ArithmeticOpDecrement, InstructionGroupArithmetic),
		"double": genControlByte(ArithmeticOpDouble, InstructionGroupArithmetic),
		"halve":  genControlByte(ArithmeticOpHalve, InstructionGroupArithmetic),
		// compare ops
		"eq":      genControlByte(CompareOpEq, InstructionGroupCompare),
		"eq-and":  genControlByte(CompareOpEqAnd, InstructionGroupCompare),
		"eq-or":   genControlByte(CompareOpEqOr, InstructionGroupCompare),
		"eq-xor":  genControlByte(CompareOpEqXor, InstructionGroupCompare),
		"neq":     genControlByte(CompareOpNeq, InstructionGroupCompare),
		"neq-and": genControlByte(CompareOpNeqAnd, InstructionGroupCompare),
		"neq-or":  genControlByte(CompareOpNeqOr, InstructionGroupCompare),
		"neq-xor": genControlByte(CompareOpNeqXor, InstructionGroupCompare),
		"lt":      genControlByte(CompareOpLessThan, InstructionGroupCompare),
		"lt-and":  genControlByte(CompareOpLessThanAnd, InstructionGroupCompare),
		"lt-or":   genControlByte(CompareOpLessThanOr, InstructionGroupCompare),
		"lt-xor":  genControlByte(CompareOpLessThanXor, InstructionGroupCompare),
		"le":      genControlByte(CompareOpLessThanOrEqualTo, InstructionGroupCompare),
		"le-and":  genControlByte(CompareOpLessThanOrEqualToAnd, InstructionGroupCompare),
		"le-or":   genControlByte(CompareOpLessThanOrEqualToOr, InstructionGroupCompare),
		"le-xor":  genControlByte(CompareOpLessThanOrEqualToXor, InstructionGroupCompare),
		"gt":      genControlByte(CompareOpGreaterThan, InstructionGroupCompare),
		"gt-and":  genControlByte(CompareOpGreaterThanAnd, InstructionGroupCompare),
		"gt-or":   genControlByte(CompareOpGreaterThanOr, InstructionGroupCompare),
		"gt-xor":  genControlByte(CompareOpGreaterThanXor, InstructionGroupCompare),
		"ge":      genControlByte(CompareOpGreaterThanOrEqualTo, InstructionGroupCompare),
		"ge-and":  genControlByte(CompareOpGreaterThanOrEqualToAnd, InstructionGroupCompare),
		"ge-or":   genControlByte(CompareOpGreaterThanOrEqualToOr, InstructionGroupCompare),
		"ge-xor":  genControlByte(CompareOpGreaterThanOrEqualToXor, InstructionGroupCompare),
		// Jump operations
		"goto-imm":        genControlByte(JumpOpUnconditionalImmediate, InstructionGroupJump),
		"call-imm":        genControlByte(JumpOpUnconditionalImmediateCall, InstructionGroupJump),
		"goto-reg":        genControlByte(JumpOpUnconditionalRegister, InstructionGroupJump),
		"call-reg":        genControlByte(JumpOpUnconditionalRegisterCall, InstructionGroupJump),
		"goto-imm-if1":    genControlByte(JumpOpConditionalTrueImmediate, InstructionGroupJump),
		"call-imm-if1":    genControlByte(JumpOpConditionalTrueImmediateCall, InstructionGroupJump),
		"goto-imm-if0":    genControlByte(JumpOpConditionalFalseImmediate, InstructionGroupJump),
		"call-imm-if0":    genControlByte(JumpOpConditionalFalseImmediateCall, InstructionGroupJump),
		"goto-reg-if1":    genControlByte(JumpOpConditionalTrueRegister, InstructionGroupJump),
		"call-reg-if1":    genControlByte(JumpOpConditionalTrueRegisterCall, InstructionGroupJump),
		"goto-reg-if0":    genControlByte(JumpOpConditionalFalseRegister, InstructionGroupJump),
		"call-reg-if0":    genControlByte(JumpOpConditionalFalseRegisterCall, InstructionGroupJump),
		"call-select-if1": genControlByte(JumpOpIfThenElseCallPredTrue, InstructionGroupJump),
		"call-select-if0": genControlByte(JumpOpIfThenElseCallPredFalse, InstructionGroupJump),
		"goto-select-if1": genControlByte(JumpOpIfThenElseNormalPredTrue, InstructionGroupJump),
		"goto-select-if0": genControlByte(JumpOpIfThenElseNormalPredFalse, InstructionGroupJump),
		// move operations
		"move":           genControlByte(MoveOpMove, InstructionGroupMove),
		"swap":           genControlByte(MoveOpSwap, InstructionGroupMove),
		"swap-reg-addr":  genControlByte(MoveOpSwapRegAddr, InstructionGroupMove),
		"swap-addr-addr": genControlByte(MoveOpSwapAddrAddr, InstructionGroupMove),
		"swap-reg-mem":   genControlByte(MoveOpSwapRegMem, InstructionGroupMove),
		"swap-addr-mem":  genControlByte(MoveOpSwapAddrMem, InstructionGroupMove),
		"set":            genControlByte(MoveOpSet, InstructionGroupMove),
		"load":           genControlByte(MoveOpLoad, InstructionGroupMove),
		"load-mem":       genControlByte(MoveOpLoadMem, InstructionGroupMove),
		"store":          genControlByte(MoveOpStore, InstructionGroupMove),
		"store-addr":     genControlByte(MoveOpStoreAddr, InstructionGroupMove),
		"store-mem":      genControlByte(MoveOpStoreMem, InstructionGroupMove),
		"store-imm":      genControlByte(MoveOpStoreImm, InstructionGroupMove),
		"push":           genControlByte(MoveOpPush, InstructionGroupMove),
		"push-imm":       genControlByte(MoveOpPushImmediate, InstructionGroupMove),
		"pop":            genControlByte(MoveOpPop, InstructionGroupMove),
		"peek":           genControlByte(MoveOpPeek, InstructionGroupMove),
	}
	// setup the keywords and register parsers
	registers = keyword.New()
	for i := 0; i < RegisterCount; i++ {
		registers.AddKeyword(fmt.Sprintf("r%d", i))
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
