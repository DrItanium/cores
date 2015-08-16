package iris1

import (
	"encoding/binary"
	"fmt"
	"github.com/DrItanium/cores/lisp"
	"io"
)

func GetDecoder() translation.Decoder {
	return instructionDecoder(unparse)
}
func unparse(in io.Reader) (lisp.Lisp, error) {
	var i [4]byte
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
	switch inst.Op {
	case MoveOpPush:
		return unparseTwoArg(pushSymbol, registerAtom(inst.Data[0])), nil
	case MoveOpPushImmediate:
		return unparseTwoArg(pushSymbol, immediateAtom(inst.Immediate())), nil
	case MoveOpPop:
		return unparseTwoArg(popSymbol, registerAtom(inst.Data[0])), nil
	case MoveOpPeek:
		return unparseTwoArg(peekSymbol, registerAtom(inst.Data[0])), nil
	case MoveOpSet:
		return unparseSet(registerAtom(inst.Data[0]), immediateAtom(inst.Immediate())), nil
	case MoveOpMove:
		return unparseSet(registerAtom(inst.Data[0]), registerAtom(inst.Data[1])), nil
	case MoveOpSwap:
		return unparseSwap(registerAtom(inst.Data[0]), registerAtom(inst.Data[1])), nil
	case MoveOpSwapRegAddr:
		return unparseSwap(registerAtom(inst.Data[0]), dataAtRegister(inst.Data[1])), nil
	case MoveOpSwapAddrAddr:
		return unparseSwap(dataAtRegister(inst.Data[0]), dataAtRegister(inst.Data[1])), nil
	case MoveOpSwapRegMem:
		return unparseSwap(registerAtom(inst.Data[0]), dataAtImmediate(inst.Immediate())), nil
	case MoveOpSwapAddrMem:
		return unparseSwap(dataAtRegister(inst.Data[0]), dataAtImmediate(inst.Immediate())), nil
	case MoveOpLoad:
		return unparseSet(registerAtom(inst.Data[0]), dataAtRegister(inst.Data[1])), nil
	case MoveOpLoadMem:
		return unparseSet(registerAtom(inst.Data[0]), dataAtImmediate(inst.Immediate())), nil
	case MoveOpStore:
		return unparseSet(dataAtRegister(inst.Data[0]), registerAtom(inst.Data[1])), nil
	case MoveOpStoreAddr:
		return unparseSet(dataAtRegister(inst.Data[0]), dataAtRegister(inst.Data[1])), nil
	case MoveOpStoreMem:
		return unparseSet(dataAtRegister(inst.Data[0]), datatAtImmediate(inst.Immediate())), nil
	case MoveOpStoreImm:
		return unparseSet(dataAtRegister(inst.Data[0]), immediateAtom(inst.Immediate())), nil
	default:
		return nil, fmt.Errorf("Illegal operation move op id %d!", inst.Op)
	}
}
