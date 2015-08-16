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
