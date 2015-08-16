package iris1

import (
	"encoding/binary"
	"fmt"
	"github.com/DrItanium/cores/lisp"
	"github.com/DrItanium/cores/translation"
	"io"
)

func GetDecoder() translation.Decoder {
	return tDecoder(unparse)
}

type tDecoder func(io.Reader) (lisp.List, error)

func (this tDecoder) Decode(in io.Reader) (lisp.List, error) {
	return this(in)
}
func unparse(in io.Reader) (lisp.List, error) {
	i := make([]byte, 4)
	if count, err := in.Read(i); err != nil {
		return nil, err
	} else if count < 4 {
		return nil, fmt.Errorf("Input stream is not divisible by four evenly!")
	} else {
		inst := Instruction(binary.LittleEndian.Uint32(i))
		if di, err := inst.Decode(); err != nil {
			return nil, err
		} else {
			return unparseFuncs.invoke(di)
		}
	}

}

type instructionDecoder func(*DecodedInstruction) (lisp.List, error)

func (this instructionDecoder) invoke(inst *DecodedInstruction) (lisp.List, error) {
	return this(inst)
}

type unparseFuncList [MajorOperationGroupCount]instructionDecoder

func (this unparseFuncList) invoke(inst *DecodedInstruction) (lisp.List, error) {
	return this[inst.Group].invoke(inst)
}

var unparseFuncs unparseFuncList

func unimplementedUnparse(i *DecodedInstruction) (lisp.List, error) {
	return nil, fmt.Errorf("Unimplemented group %d!", i.Group)
}

func init() {
	for i := 0; i < MajorOperationGroupCount; i++ {
		unparseFuncs[i] = unimplementedUnparse
	}
	//unparseFuncs[InstructionGroupArithmetic] = unparseArithmetic
	unparseFuncs[InstructionGroupMove] = unparseMove
	unparseFuncs[InstructionGroupJump] = unparseJump
	unparseFuncs[InstructionGroupCompare] = unparseCompare
}

var dataAtSymbol = lisp.Atom([]byte("data-at"))
var pushSymbol = lisp.Atom([]byte("push"))
var popSymbol = lisp.Atom([]byte("pop"))
var peekSymbol = lisp.Atom([]byte("peek"))
var setSymbol = lisp.Atom([]byte("set"))
var swapSymbol = lisp.Atom([]byte("swap"))
var registerAtoms [RegisterCount]lisp.Atom

func init() {
	for i := 0; i < RegisterCount; i++ {
		registerAtoms[i] = lisp.Atom([]byte(fmt.Sprintf("r%d", i)))
	}
}
func immediateAtom(imm Word) lisp.Atom {
	return lisp.Atom([]byte(fmt.Sprintf("0x%X", imm)))
}
func registerAtom(reg byte) lisp.Atom {
	return registerAtoms[reg]
}
func dataAtRegister(reg byte) lisp.List {
	return lisp.List{
		dataAtSymbol,
		registerAtom(reg),
	}
}
func dataAtImmediate(imm Word) lisp.List {
	return lisp.List{
		dataAtSymbol,
		immediateAtom(imm),
	}
}
func unparseTwoArg(first, second interface{}) lisp.List {
	return lisp.List{first, second}
}
func unparseThreeArg(first, second, third interface{}) lisp.List {
	return lisp.List{first, second, third}
}
func unparseSet(first, second interface{}) lisp.List {
	return unparseThreeArg(setSymbol, first, second)
}
func unparseSwap(first, second interface{}) lisp.List {
	return unparseThreeArg(swapSymbol, first, second)
}

func unparseMove(inst *DecodedInstruction) (lisp.List, error) {
	dest := inst.Data[0]
	src0 := inst.Data[1]
	imm := inst.Immediate()
	switch inst.Op {
	case MoveOpPush:
		return unparseTwoArg(pushSymbol, registerAtom(dest)), nil
	case MoveOpPushImmediate:
		return unparseTwoArg(pushSymbol, immediateAtom(imm)), nil
	case MoveOpPop:
		return unparseTwoArg(popSymbol, registerAtom(dest)), nil
	case MoveOpPeek:
		return unparseTwoArg(peekSymbol, registerAtom(dest)), nil
	case MoveOpSet:
		return unparseSet(registerAtom(dest), immediateAtom(imm)), nil
	case MoveOpMove:
		return unparseSet(registerAtom(dest), registerAtom(src0)), nil
	case MoveOpSwap:
		return unparseSwap(registerAtom(dest), registerAtom(src0)), nil
	case MoveOpSwapRegAddr:
		return unparseSwap(registerAtom(dest), dataAtRegister(src0)), nil
	case MoveOpSwapAddrAddr:
		return unparseSwap(dataAtRegister(dest), dataAtRegister(src0)), nil
	case MoveOpSwapRegMem:
		return unparseSwap(registerAtom(dest), dataAtImmediate(imm)), nil
	case MoveOpSwapAddrMem:
		return unparseSwap(dataAtRegister(dest), dataAtImmediate(imm)), nil
	case MoveOpLoad:
		return unparseSet(registerAtom(dest), dataAtRegister(src0)), nil
	case MoveOpLoadMem:
		return unparseSet(registerAtom(dest), dataAtImmediate(imm)), nil
	case MoveOpStore:
		return unparseSet(dataAtRegister(dest), registerAtom(src0)), nil
	case MoveOpStoreAddr:
		return unparseSet(dataAtRegister(dest), dataAtRegister(src0)), nil
	case MoveOpStoreMem:
		return unparseSet(dataAtRegister(dest), dataAtImmediate(imm)), nil
	case MoveOpStoreImm:
		return unparseSet(dataAtRegister(dest), immediateAtom(imm)), nil
	default:
		return nil, fmt.Errorf("Illegal move op (id %d)!", inst.Op)
	}
}

// jump forms
var symbolIf = lisp.Atom([]byte("if"))
var symbolThen = lisp.Atom([]byte("then"))
var symbolElse = lisp.Atom([]byte("else"))
var symbolNot = lisp.Atom([]byte("not"))
var symbolGoto = lisp.Atom([]byte("goto"))
var symbolCall = lisp.Atom([]byte("call"))

func unparseGenericArgs(contents ...interface{}) lisp.List {
	// play and fast loose with copy over ops
	list := make(lisp.List, len(contents))
	copy(list, contents)
	return list
}
func unparseOnTrue(reg byte) lisp.Atom {
	return registerAtom(reg)
}
func unparseNot(reg byte) lisp.List {
	return unparseGenericArgs(symbolNot, registerAtom(reg))
}
func unparseSelect(cond interface{}, onTrue, onFalse byte) lisp.List {
	return unparseGenericArgs(symbolIf, cond, symbolThen, registerAtom(onTrue), symbolElse, registerAtom(onFalse))
}
func unparseIfThen(cond, onTrue interface{}) lisp.List {
	return unparseGenericArgs(symbolIf, cond, symbolThen, onTrue)
}
func unparseGoto(arg interface{}) lisp.List {
	return unparseGenericArgs(symbolGoto, arg)
}
func unparseGotoImmediate(imm Word) lisp.List {
	return unparseGoto(immediateAtom(imm))
}
func unparseGotoRegister(reg byte) lisp.List {
	return unparseGoto(registerAtom(reg))
}
func unparseCall(arg interface{}) lisp.List {
	return unparseGenericArgs(symbolCall, arg)
}
func unparseCallImmediate(imm Word) lisp.List {
	return unparseCall(immediateAtom(imm))
}
func unparseCallRegister(reg byte) lisp.List {
	return unparseCall(registerAtom(reg))
}
func unparseJump(inst *DecodedInstruction) (lisp.List, error) {
	dest, src0, src1, imm := inst.Data[0], inst.Data[1], inst.Data[2], inst.Immediate()
	switch inst.Op {
	case JumpOpUnconditionalImmediate:
		return unparseGoto(immediateAtom(imm)), nil
	case JumpOpUnconditionalImmediateCall:
		return unparseCall(immediateAtom(imm)), nil
	case JumpOpUnconditionalRegister:
		return unparseGoto(registerAtom(dest)), nil
	case JumpOpUnconditionalRegisterCall:
		return unparseCall(registerAtom(dest)), nil
	case JumpOpConditionalTrueImmediate:
		return unparseIfThen(unparseOnTrue(dest), unparseGotoImmediate(imm)), nil
	case JumpOpConditionalTrueImmediateCall:
		return unparseIfThen(unparseOnTrue(dest), unparseCallImmediate(imm)), nil
	case JumpOpConditionalTrueRegister:
		return unparseIfThen(unparseOnTrue(dest), unparseGotoRegister(src0)), nil
	case JumpOpConditionalTrueRegisterCall:
		return unparseIfThen(unparseOnTrue(dest), unparseCallRegister(src0)), nil
	case JumpOpConditionalFalseImmediate:
		return unparseIfThen(unparseNot(dest), unparseGotoImmediate(imm)), nil
	case JumpOpConditionalFalseImmediateCall:
		return unparseIfThen(unparseNot(dest), unparseCallImmediate(imm)), nil
	case JumpOpConditionalFalseRegister:
		return unparseIfThen(unparseNot(dest), unparseGotoRegister(src0)), nil
	case JumpOpConditionalFalseRegisterCall:
		return unparseIfThen(unparseNot(dest), unparseCallRegister(src0)), nil
	case JumpOpIfThenElseNormalPredTrue:
		return unparseGoto(unparseSelect(unparseOnTrue(dest), src0, src1)), nil
	case JumpOpIfThenElseNormalPredFalse:
		return unparseGoto(unparseSelect(unparseNot(dest), src0, src1)), nil
	case JumpOpIfThenElseCallPredTrue:
		return unparseCall(unparseSelect(unparseOnTrue(dest), src0, src1)), nil
	case JumpOpIfThenElseCallPredFalse:
		return unparseCall(unparseSelect(unparseNot(dest), src0, src1)), nil
	default:
		return nil, fmt.Errorf("Illegal jump op (id %d)!", inst.Op)
	}
}

const (
	symbolCompareEq = iota
	symbolCompareNeq
	symbolCompareLessThan
	symbolCompareGreaterThan
	symbolCompareLessThanOrEqualTo
	symbolCompareGreaterThanOrEqualTo
)

var symbolCompareOps = []lisp.Atom{
	lisp.Atom("="),
	lisp.Atom("<>"),
	lisp.Atom("<"),
	lisp.Atom(">"),
	lisp.Atom("<="),
	lisp.Atom(">="),
}

const (
	symbolCompareModifierNone = iota
	symbolCompareModifierAnd
	symbolCompareModifierOr
	symbolCompareModifierXor
)

var symbolCompareModifiers = []lisp.Atom{
	nil,
	lisp.Atom("&&"),
	lisp.Atom("||"),
	lisp.Atom("xor"),
}

func unparseCompareOp(compareModifier, compareOp int, dest, src0, src1 byte) lisp.List {
	destAtom := registerAtom(dest)
	baseOp := unparseGenericArgs(symbolCompareOps[compareOp], registerAtom(src0), registerAtom(src1))
	var lst lisp.List
	if modifier := symbolCompareModifiers[compareModifier]; modifier == nil {
		lst = baseOp
	} else {
		lst = unparseGenericArgs(modifier, destAtom, baseOp)
	}
	return unparseSet(destAtom, lst)
}

func unparseCompare(inst *DecodedInstruction) (lisp.List, error) {
	dest, src0, src1 := inst.Data[0], inst.Data[1], inst.Data[2]
	switch inst.Op {
	case CompareOpEq:
		return unparseCompareOp(symbolCompareModifierNone, symbolCompareEq, dest, src0, src1), nil
	case CompareOpEqAnd:
		return unparseCompareOp(symbolCompareModifierAnd, symbolCompareEq, dest, src0, src1), nil
	case CompareOpEqOr:
		return unparseCompareOp(symbolCompareModifierOr, symbolCompareEq, dest, src0, src1), nil
	case CompareOpEqXor:
		return unparseCompareOp(symbolCompareModifierXor, symbolCompareEq, dest, src0, src1), nil
	case CompareOpNeq:
		return unparseCompareOp(symbolCompareModifierNone, symbolCompareNeq, dest, src0, src1), nil
	case CompareOpNeqAnd:
		return unparseCompareOp(symbolCompareModifierAnd, symbolCompareNeq, dest, src0, src1), nil
	case CompareOpNeqOr:
		return unparseCompareOp(symbolCompareModifierOr, symbolCompareNeq, dest, src0, src1), nil
	case CompareOpNeqXor:
		return unparseCompareOp(symbolCompareModifierXor, symbolCompareNeq, dest, src0, src1), nil
	case CompareOpLessThan:
		return unparseCompareOp(symbolCompareModifierNone, symbolCompareLessThan, dest, src0, src1), nil
	case CompareOpLessThanAnd:
		return unparseCompareOp(symbolCompareModifierAnd, symbolCompareLessThan, dest, src0, src1), nil
	case CompareOpLessThanOr:
		return unparseCompareOp(symbolCompareModifierOr, symbolCompareLessThan, dest, src0, src1), nil
	case CompareOpLessThanXor:
		return unparseCompareOp(symbolCompareModifierXor, symbolCompareLessThan, dest, src0, src1), nil
	case CompareOpGreaterThan:
		return unparseCompareOp(symbolCompareModifierNone, symbolCompareGreaterThan, dest, src0, src1), nil
	case CompareOpGreaterThanAnd:
		return unparseCompareOp(symbolCompareModifierAnd, symbolCompareGreaterThan, dest, src0, src1), nil
	case CompareOpGreaterThanOr:
		return unparseCompareOp(symbolCompareModifierOr, symbolCompareGreaterThan, dest, src0, src1), nil
	case CompareOpGreaterThanXor:
		return unparseCompareOp(symbolCompareModifierXor, symbolCompareGreaterThan, dest, src0, src1), nil
	case CompareOpLessThanOrEqualTo:
		return unparseCompareOp(symbolCompareModifierNone, symbolCompareLessThanOrEqualTo, dest, src0, src1), nil
	case CompareOpLessThanOrEqualToAnd:
		return unparseCompareOp(symbolCompareModifierAnd, symbolCompareLessThanOrEqualTo, dest, src0, src1), nil
	case CompareOpLessThanOrEqualToOr:
		return unparseCompareOp(symbolCompareModifierOr, symbolCompareLessThanOrEqualTo, dest, src0, src1), nil
	case CompareOpLessThanOrEqualToXor:
		return unparseCompareOp(symbolCompareModifierXor, symbolCompareLessThanOrEqualTo, dest, src0, src1), nil
	case CompareOpGreaterThanOrEqualTo:
		return unparseCompareOp(symbolCompareModifierNone, symbolCompareGreaterThanOrEqualTo, dest, src0, src1), nil
	case CompareOpGreaterThanOrEqualToAnd:
		return unparseCompareOp(symbolCompareModifierAnd, symbolCompareGreaterThanOrEqualTo, dest, src0, src1), nil
	case CompareOpGreaterThanOrEqualToOr:
		return unparseCompareOp(symbolCompareModifierOr, symbolCompareGreaterThanOrEqualTo, dest, src0, src1), nil
	case CompareOpGreaterThanOrEqualToXor:
		return unparseCompareOp(symbolCompareModifierXor, symbolCompareGreaterThanOrEqualTo, dest, src0, src1), nil
	default:
		return nil, fmt.Errorf("Illegal compare op (id %d)!", inst.Op)
	}
}

const (
	symbolArithmeticOpAdd = iota
	symbolArithmeticOpSub
	symbolArithmeticOpMul
	symbolArithmeticOpDiv
	symbolArithmeticOpRem
	symbolArithmeticOpShiftLeft
	symbolArithmeticOpShiftRight
	symbolArithmeticOpArithmeticAnd
	symbolArithmeticOpArithmeticOr
	symbolArithmeticOpArithmeticNot
	symbolArithmeticOpArithmeticXor
	symbolArithmeticOpIncrement
	symbolArithmeticOpDecrement
	symbolArithmeticOpDouble
	symbolArithmeticOpHalve
)

var arithmeticSymbols = []lisp.Atom{
	lisp.Atom("+"),
	lisp.Atom("-"),
	lisp.Atom("*"),
	lisp.Atom("/"),
	lisp.Atom("rem"),
	lisp.Atom("shl"),
	lisp.Atom("shr"),
	lisp.Atom("arithmetic-and"),
	lisp.Atom("arithmetic-or"),
	lisp.Atom("arithmetic-not"),
	lisp.Atom("arithmetic-xor"),
	lisp.Atom("1+"),
	lisp.Atom("1-"),
	lisp.Atom("2*"),
	lisp.Atom("2/"),
}

const (
	arithmeticOpRegisterStyle = iota
	arithmeticOpImmediateStyle
	arithmeticOpUnaryStyle
)

var arithmeticTranslationTable = map[byte]struct {
	Index int
	Style int
}{
	ArithmeticOpAdd:                 {Index: symbolArithmeticOpAdd, Style: arithmeticOpRegisterStyle},
	ArithmeticOpSub:                 {Index: symbolArithmeticOpSub, Style: arithmeticOpRegisterStyle},
	ArithmeticOpMul:                 {Index: symbolArithmeticOpMul, Style: arithmeticOpRegisterStyle},
	ArithmeticOpDiv:                 {Index: symbolArithmeticOpDiv, Style: arithmeticOpRegisterStyle},
	ArithmeticOpRem:                 {Index: symbolArithmeticOpRem, Style: arithmeticOpRegisterStyle},
	ArithmeticOpShiftLeft:           {Index: symbolArithmeticOpShiftLeft, Style: arithmeticOpRegisterStyle},
	ArithmeticOpShiftRight:          {Index: symbolArithmeticOpShiftRight, Style: arithmeticOpRegisterStyle},
	ArithmeticOpBinaryAnd:           {Index: symbolArithmeticOpArithmeticAnd, Style: arithmeticOpRegisterStyle},
	ArithmeticOpBinaryOr:            {Index: symbolArithmeticOpArithmeticOr, Style: arithmeticOpRegisterStyle},
	ArithmeticOpBinaryXor:           {Index: symbolArithmeticOpArithmeticXor, Style: arithmeticOpRegisterStyle},
	ArithmeticOpBinaryNot:           {Index: symbolArithmeticOpArithmeticNot, Style: arithmeticOpUnaryStyle},
	ArithmeticOpIncrement:           {Index: symbolArithmeticOpIncrement, Style: arithmeticOpUnaryStyle},
	ArithmeticOpDecrement:           {Index: symbolArithmeticOpDecrement, Style: arithmeticOpUnaryStyle},
	ArithmeticOpDouble:              {Index: symbolArithmeticOpDouble, Style: arithmeticOpUnaryStyle},
	ArithmeticOpHalve:               {Index: symbolArithmeticOpHalve, Style: arithmeticOpUnaryStyle},
	ArithmeticOpAddImmediate:        {Index: symbolArithmeticOpAdd, Style: arithmeticOpImmediateStyle},
	ArithmeticOpSubImmediate:        {Index: symbolArithmeticOpSub, Style: arithmeticOpImmediateStyle},
	ArithmeticOpMulImmediate:        {Index: symbolArithmeticOpMul, Style: arithmeticOpImmediateStyle},
	ArithmeticOpDivImmediate:        {Index: symbolArithmeticOpDiv, Style: arithmeticOpImmediateStyle},
	ArithmeticOpRemImmediate:        {Index: symbolArithmeticOpRem, Style: arithmeticOpImmediateStyle},
	ArithmeticOpShiftLeftImmediate:  {Index: symbolArithmeticOpShiftLeft, Style: arithmeticOpImmediateStyle},
	ArithmeticOpShiftRightImmediate: {Index: symbolArithmeticOpShiftRight, Style: arithmeticOpImmediateStyle},
}

func unparseArithmeticOp(symbolIndex int, dest, src0 byte, src1 interface{}) lisp.List {
	var l lisp.List
	if src1 == nil {
		l = unparseGenericArgs(arithmeticSymbols[symbolIndex], registerAtom(src0))
	} else {
		l = unparseGenericArgs(arithmeticSymbols[symbolIndex], registerAtom(src0), src1)
	}
	return unparseSet(registerAtom(dest), l)
}
func unparseArithmeticOpRegister(symbolIndex int, dest, src0, src1 byte) lisp.List {
	return unparseArithmeticOp(symbolIndex, dest, src0, registerAtom(src1))
}
func unparseArithmeticOpImmediate(symbolIndex int, dest, src0, src1 byte) lisp.List {
	return unparseArithmeticOp(symbolIndex, dest, src0, immediateAtom(Word(src1)))
}
func unparseArithmeticOpUnary(symbolIndex int, dest, src0 byte) lisp.List {
	return unparseArithmeticOp(symbolIndex, dest, src0, nil)
}

func unparseArithmetic(inst *DecodedInstruction) (lisp.List, error) {
	if val, ok := arithmeticTranslationTable[inst.Op]; !ok {
		return nil, fmt.Errorf("Illegal arithmetic op (id %d)!", inst.Op)
	} else {
		dest, src0, src1 := inst.Data[0], inst.Data[1], inst.Data[2]
		switch val.Style {
		case arithmeticOpRegisterStyle:
			return unparseArithmeticOpRegister(val.Index, dest, src0, src1), nil
		case arithmeticOpImmediateStyle:
			return unparseArithmeticOpImmediate(val.Index, dest, src0, src1), nil
		case arithmeticOpUnaryStyle:
			return unparseArithmeticOpUnary(val.Index, dest, src0), nil
		default:
			return nil, fmt.Errorf("Unknown arithmetic op style %d!", val.Style)
		}
	}
}

// misc operations
var symbolSystem = lisp.Atom("system")

func unparseMisc(inst *DecodedInstruction) (lisp.List, error) {
	switch inst.Op {
	case MiscOpSystemCall:
		// A hack right now since system calls need to be fixed up as it right now isn't very well designed!
		return unparseGenericArgs(symbolSystem, immediateAtom(Word(inst.Data[0])), registerAtom(inst.Data[1]), registerAtom(inst.Data[2])), nil
	default:
		return nil, fmt.Errorf("Unknown misc op (id %d)!", inst.Op)
	}
}
