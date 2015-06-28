// machine description of iris1
package iris1

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

type Core struct {
	Gpr                [RegisterCount]Word
	Code               [MemorySize]Instruction
	Data               [MemorySize]Word
	Stack              [MemorySize]Word
	Pc                 Word
	AdvancePC          bool
	TerminateExecution bool
}

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
)
