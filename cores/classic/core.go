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
