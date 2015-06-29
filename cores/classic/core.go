package classic

import (
	"fmt"
	"github.com/DrItanium/iris1"
)

const (
	InstructionGroupArithmetic = iota
	InstructionGroupMove
	InstructionGroupJump
	InstructionGroupCompare
	InstructionGroupMisc
	// arithmetic operations
	ArithmeticOpAdd = iota
	ArithmeticOpSub
	ArithmeticOpMul
	ArithmeticOpDiv
	ArithmeticOpRem
	ArithmeticOpShiftLeft
	ArithmeticOpShiftRight
	ArithmeticOpBinaryAnd
	ArithmeticOpBinaryOr
	ArithmeticOpBinaryNot
	ArithmeticOpBinaryXor
	ArithmeticOpAddImmediate
	ArithmeticOpSubImmediate
	ArithmeticOpMulImmediate
	ArithmeticOpDivImmediate
	ArithmeticOpRemImmediate
	ArithmeticOpShiftLeftImmediate
	ArithmeticOpShiftRightImmediate
	// always last
	ArithmeticOpCount
	// Move Operations
	MoveOpMove = iota
	MoveOpSwap
	MoveOpSwapRegAddr
	MoveOpSwapAddrAddr
	MoveOpSwapRegMem
	MoveOpSwapAddrMem
	MoveOpSet
	MoveOpLoad
	MoveOpLoadMem
	MoveOpStore
	MoveOpStoreAddr
	MoveOpStoreMem
	MoveOpStoreImm
	MoveOpPush
	MoveOpPushImmediate
	MoveOpPop
	MoveOpPeek
	// always last
	MoveOpCount
	// Jump Operations
	JumpOpUnconditionalImmediate = iota
	JumpOpUnconditionalImmediateLink
	JumpOpUnconditionalRegister
	JumpOpUnconditionalRegisterLink
	JumpOpConditionalTrueImmediate
	JumpOpConditionalTrueImmediateLink
	JumpOpConditionalTrueRegister
	JumpOpConditionalTrueRegisterLink
	JumpOpConditionalFalseImmediate
	JumpOpConditionalFalseImmediateLink
	JumpOpConditionalFalseRegister
	JumpOpConditionalFalseRegisterLink
	JumpOpIfThenElseNormalPredTrue
	JumpOpIfThenElseNormalPredFalse
	JumpOpIfThenElseLinkPredTrue
	JumpOpIfThenElseLinkPredFalse
	// Compare operations
	CompareOpEq = iota
	CompareOpEqAnd
	CompareOpEqOr
	CompareOpEqXor
	CompareOpNeq
	CompareOpNeqAnd
	CompareOpNeqOr
	CompareOpNeqXor
	CompareOpLessThan
	CompareOpLessThanAnd
	CompareOpLessThanOr
	CompareOpLessThanXor
	CompareOpGreaterThan
	CompareOpGreaterThanAnd
	CompareOpGreaterThanOr
	CompareOpGreaterThanXor
	CompareOpLessThanOrEqualTo
	CompareOpLessThanOrEqualToAnd
	CompareOpLessThanOrEqualToOr
	CompareOpLessThanOrEqualToXor
	CompareOpGreaterThanOrEqualTo
	CompareOpGreaterThanOrEqualToAnd
	CompareOpGreaterThanOrEqualToOr
	CompareOpGreaterThanOrEqualToXor
	// Misc operations
	MiscOpSystemCall = iota
	// System commands
	SystemCommandTerminate = iota
	SystemCommandPanic     = 255
)

func NewCore() (*iris1.Core, error) {
	var b Backend
	core, err := iris1.New(&b)
	if err != nil {
		return nil, err
	}
	if err0 := core.InstallExecutionUnit(InstructionGroupArithmetic, arithmetic); err0 != nil {
		return nil, err0
	}
	if err0 := core.InstallExecutionUnit(InstructionGroupMove, move); err0 != nil {
		return nil, err0
	}
	return core, nil
}

func arithmetic(core *iris1.Core, inst *iris1.DecodedInstruction) error {
	if inst.Op >= ArithmeticOpCount {
		return fmt.Errorf("Op index %d is not a valid arithmetic operation", inst.Op)
	} else {
		dest := inst.Data[0]
		src0 := core.Register(inst.Data[1])
		src1 := core.Register(inst.Data[2])
		imm := iris1.Word(inst.Data[2])
		result := iris1.Word(0)
		switch inst.Op {
		case ArithmeticOpAdd:
			result = src0 + src1
		case ArithmeticOpSub:
			result = src0 - src1
		case ArithmeticOpMul:
			result = src0 * src1
		case ArithmeticOpDiv:
			if src1 == 0 {
				return fmt.Errorf("Divide by zero error!")
			} else if src1 == 1 {
				result = src0
			} else {
				result = src0 / src1
			}
		case ArithmeticOpRem:
			if src1 == 0 {
				return fmt.Errorf("Divide by zero error!")
			} else if src1 == 1 {
				result = 0
			} else {
				result = src0 % src1
			}
		case ArithmeticOpShiftLeft:
			result = src0 << src1
		case ArithmeticOpShiftRight:
			result = src0 >> src1
		case ArithmeticOpBinaryAnd:
			result = src0 & src1
		case ArithmeticOpBinaryOr:
			result = src0 | src1
		case ArithmeticOpBinaryNot:
			result = ^src0
		case ArithmeticOpBinaryXor:
			result = src0 ^ src1
		case ArithmeticOpAddImmediate:
			result = src0 + imm
		case ArithmeticOpSubImmediate:
			result = src0 - imm
		case ArithmeticOpMulImmediate:
			result = src0 * imm
		case ArithmeticOpDivImmediate:
			if imm == 0 {
				return fmt.Errorf("Divide by zero error!")
			} else if imm == 1 {
				result = src0
			} else {
				result = src0 / imm
			}
		case ArithmeticOpRemImmediate:
			if imm == 0 {
				return fmt.Errorf("Divide by zero error!")
			} else if imm == 1 {
				result = src0
			} else {
				result = src0 % imm
			}
		case ArithmeticOpShiftLeftImmediate:
			result = src0 << imm
		case ArithmeticOpShiftRightImmediate:
			result = src0 >> imm
		default:
			return fmt.Errorf("Programmer failure! Report it as such!")
		}
		return core.SetRegister(dest, result)
	}
}
func swapMemory(core *iris1.Core, addr0, data0, addr1, data1 iris1.Word) error {
	if err := core.SetDataMemory(addr0, data1); err != nil {
		return err
	} else if err := core.SetDataMemory(addr1, data0); err != nil {
		return err
	} else {
		return nil
	}
}
func swapMemoryAndRegister(core *iris1.Core, reg byte, data0, addr, data1 iris1.Word) error {
	if err := core.SetRegister(reg, data1); err != nil {
		return err
	} else if err := core.SetDataMemory(addr, data0); err != nil {
		return err
	} else {
		return nil
	}
}
func move(core *iris1.Core, inst *iris1.DecodedInstruction) error {
	if inst.Op >= MoveOpCount {
		return fmt.Errorf("Op index %d is not a valid move operation", inst.Op)
	} else {
		dest := inst.Data[0]
		src0 := inst.Data[1]
		switch inst.Op {
		case MoveOpMove:
			return core.SetRegister(dest, core.Register(src0))
		case MoveOpSwap:
			r0 := core.Register(dest)
			r1 := core.Register(src0)
			if err := core.SetRegister(src0, r0); err != nil {
				return err
			} else if err := core.SetRegister(dest, r1); err != nil {
				return err
			} else {
				return nil
			}
		case MoveOpSwapRegAddr:
			reg := core.Register(dest)
			memaddr := core.Register(src0)
			memcontents := core.DataMemory(memaddr)
			return swapMemoryAndRegister(core, dest, reg, memaddr, memcontents)
		case MoveOpSwapAddrAddr:
			addr0 := core.Register(dest)
			addr1 := core.Register(src0)
			mem0 := core.DataMemory(addr0)
			mem1 := core.DataMemory(addr1)
			return swapMemory(core, addr0, mem0, addr1, mem1)
		case MoveOpSwapRegMem:
			addr := inst.Immediate()
			return swapMemoryAndRegister(core, dest, core.Register(dest), addr, core.DataMemory(addr))
		case MoveOpSwapAddrMem:
			addr0 := core.Register(dest)
			addr1 := inst.Immediate()
			mem0 := core.DataMemory(addr0)
			mem1 := core.DataMemory(addr1)
			return swapMemory(core, addr0, mem0, addr1, mem1)
		case MoveOpSet:
			return core.SetRegister(dest, inst.Immediate())
		case MoveOpLoad:
			return core.SetRegister(dest, core.DataMemory(core.Register(src0)))
		case MoveOpLoadMem:
			return core.SetRegister(dest, core.DataMemory(inst.Immediate()))
		case MoveOpStore:
			return core.SetDataMemory(core.Register(dest), core.Register(src0))
		case MoveOpStoreAddr:
			return core.SetDataMemory(core.Register(dest), core.DataMemory(core.Register(src0)))
		case MoveOpStoreMem:
			return core.SetDataMemory(core.Register(dest), core.DataMemory(inst.Immediate()))
		case MoveOpStoreImm:
			return core.SetDataMemory(core.Register(dest), inst.Immediate())
		case MoveOpPush:
			core.Push(core.Register(dest))
			return nil
		case MoveOpPushImmediate:
			core.Push(inst.Immediate())
			return nil
		case MoveOpPop:
			return core.SetRegister(dest, core.Pop())
		case MoveOpPeek:
			return core.SetRegister(dest, core.Peek())
		default:
			return fmt.Errorf("Programmer failure! Report it as such!")
		}
	}
}
