// machine description of iris1
package iris1

import (
	"fmt"
)

const (
	RegisterCount             = 256
	MemorySize                = 65536
	MajorOperationGroupCount  = 8
	PredicateRegisterIndex    = 255
	StackPointerRegisterIndex = 254
)

type Word uint16
type Dword uint32
type Instruction Dword

const (
	InstructionGroupArithmetic = iota
	InstructionGroupMove
	InstructionGroupJump
	InstructionGroupCompare
	InstructionGroupMisc
)

const (
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
	SystemCommandGetC
	SystemCommandPutC
	SystemCommandPanic
	// Error codes
	ErrorNone = iota
	ErrorGetRegisterOutOfRange
	ErrorPutRegisterOutOfRange
	ErrorInvalidInstructionGroupProvided
	ErrorInvalidArithmeticOperation
	ErrorInvalidMoveOperation
	ErrorInvalidJumpOperation
	ErrorInvalidCompareOperation
	ErrorInvalidMiscOperation
	ErrorInvalidSystemCommand
)

var errorLookup = []string{
	"None",
	"Attempted to get the value of invalid register r%d",
	"Attempted to set the value of invalid register r%d",
	"Instruction group %d is not a valid instruction group!",
	"Illegal arithmetic operation %d",
	"Illegal move operation %d",
	"Illegal jump operation %d",
	"Illegal compare operation %d",
	"Illegal misc operation %d",
	"Invalid system command %d",
}

type IrisError struct {
	value, code uint
}

func newError(code, value uint) error {
	return &IrisError{code: code, value: value}
}

func (this IrisError) Error() string {
	if this.code == 0 {
		return fmt.Sprintf("No Error with value %d!!! This should never ever showup!", this.value)
	} else if this.code >= uint(len(errorLookup)) {
		return fmt.Sprintf("Unknown error %d with value %d! Something really bad happened!", this.code, this.value)
	} else {
		return fmt.Sprintf(errorLookup[this.code], this.value)
	}
}

func (this Instruction) Group() byte {
	return byte(((this & 0x000000FF) & 0x7))
}
func (this Instruction) Op() byte {
	return byte(((this & 0x000000FF) & 0xF8) >> 3)
}
func (this Instruction) Register(index int) (byte, error) {
	switch index {
	case 0:
		return byte(this), nil
	case 1:
		return byte((this & 0x0000FF00) >> 8), nil
	case 2:
		return byte((this & 0x00FF0000) >> 16), nil
	case 3:
		return byte((this & 0xFF000000) >> 24), nil
	default:
		return 0, fmt.Errorf("Register index: %d is out of range!", index)
	}
}

func (this Instruction) Immediate() Word {
	return Word((this & 0xFFFF0000) >> 16)
}

type CoreInit func(*Core) error
type Core struct {
	Gpr                [RegisterCount]Word
	Code               [MemorySize]Instruction
	Data               [MemorySize]Word
	Stack              [MemorySize]Word
	Pc                 Word
	advancePc          bool
	terminateExecution bool
}

func (this *Core) SetRegister(index byte, value Word) {
	this.Gpr[index] = value
}
func (this *Core) GetRegister(index byte) Word {
	return this.Gpr[index]
}
func New(init CoreInit) (*Core, error) {
	var c Core
	c.Pc = 0
	c.advancePc = true
	c.terminateExecution = false
	if err := init(&c); err != nil {
		return nil, err
	} else {
		return &c, nil
	}
}
func (this *Core) arithmetic(inst Instruction) error {
	op := inst.Op()
	switch op {
	default:
		return newError(ErrorInvalidArithmeticOperation, uint(op))
	}
}
func (this *Core) move(inst Instruction) error {
	op := inst.Op()
	switch op {
	default:
		return newError(ErrorInvalidMoveOperation, uint(op))
	}
}
func (this *Core) jump(inst Instruction) error {
	op := inst.Op()
	switch op {
	default:
		return newError(ErrorInvalidJumpOperation, uint(op))
	}
}
func (this *Core) compare(inst Instruction) error {
	op := inst.Op()
	switch op {
	default:
		return newError(ErrorInvalidCompareOperation, uint(op))
	}
}
func (this *Core) misc(inst Instruction) error {
	op := inst.Op()
	switch op {
	default:
		return newError(ErrorInvalidMiscOperation, uint(op))
	}
}
func (this *Core) Dispatch(inst Instruction) error {
	this.advancePc = true
	group := inst.Group()
	switch group {
	case InstructionGroupArithmetic:
		return this.arithmetic(inst)
	case InstructionGroupMove:
		return this.move(inst)
	case InstructionGroupJump:
		return this.jump(inst)
	case InstructionGroupCompare:
		return this.compare(inst)
	case InstructionGroupMisc:
		return this.misc(inst)
	default:
		return newError(ErrorInvalidInstructionGroupProvided, uint(group))
	}
}
