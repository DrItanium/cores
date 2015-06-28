// machine description of iris1
package iris1

import (
	"fmt"
)

const (
	RegisterCount            = 256
	MemorySize               = 65536
	MajorOperationGroupCount = 8
	SystemCallCount          = 256
	FalseRegister            = iota
	TrueRegister
	InstructionPointer
	StackPointer
	PredicateRegister
	CountRegister
	LinkRegister
	UserRegisterBegin
	// groups
	InstructionGroupArithmetic = iota
	InstructionGroupMove
	InstructionGroupJump
	InstructionGroupCompare
	InstructionGroupMisc
	InstructionGroupExtended0
	InstructionGroupExtended1
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
	SystemCommandPanic     = 255
	// Error codes
	ErrorNone = iota
	ErrorPanic
	ErrorGetRegisterOutOfRange
	ErrorPutRegisterOutOfRange
	ErrorInvalidInstructionGroupProvided
	ErrorInvalidArithmeticOperation
	ErrorInvalidMoveOperation
	ErrorInvalidJumpOperation
	ErrorInvalidCompareOperation
	ErrorInvalidMiscOperation
	ErrorInvalidSystemCommand
	ErrorWriteToFalseRegister
	ErrorWriteToTrueRegister
	ErrorEncodeByteOutOfRange
	ErrorGroupValueOutOfRange
	ErrorOpValueOutOfRange
)

var errorLookup = []string{
	"None",
	"The core was sent a panic signal with argument %d!",
	"Attempted to get the value of invalid register r%d",
	"Attempted to set the value of invalid register r%d",
	"Instruction group %d is not a valid instruction group!",
	"Illegal arithmetic operation %d",
	"Illegal move operation %d",
	"Illegal jump operation %d",
	"Illegal compare operation %d",
	"Illegal misc operation %d",
	"Invalid system command %d",
	"Attempted to write %d to false register",
	"Attempted to write %d to true register!",
	"Specified illegal byte offset %d to encode data into",
	"Provided group id %d is larger than the space allotted to specifying the group",
	"Provided op id %d is larger than the space allotted to specifying the op",
}

type Word uint16
type Dword uint32
type Instruction Dword

func (this Instruction) group() byte {
	return byte(((this & 0x000000FF) & 0x7))
}
func (this Instruction) op() byte {
	return byte(((this & 0x000000FF) & 0xF8) >> 3)
}
func (this Instruction) register(index int) (byte, error) {
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

func (this Instruction) immediate() Word {
	return Word((this & 0xFFFF0000) >> 16)
}

func (this *Instruction) setGroup(group byte) {
	*this = ((*this &^ 0x7) | Instruction(group))
}
func (this *Instruction) setOp(op byte) {
	*this = ((*this &^ 0xF8) | (Instruction(op) << 3))
}
func (this *Instruction) setByte(index int, value byte) error {
	switch index {
	case 1:
		*this = ((*this &^ 0x0000FF00) | (Instruction(value) << 8))
	case 2:
		*this = ((*this &^ 0x00FF0000) | (Instruction(value) << 16))
	case 3:
		*this = ((*this &^ 0xFF000000) | (Instruction(value) << 24))
	default:
		return newError(ErrorEncodeByteOutOfRange, uint(index))
	}
	return nil
}
func (this *Instruction) setImmediate(value Word) {
	*this = ((*this &^ 0xFFFF0000) | (Instruction(value) << 16))
}

type DecodedInstruction struct {
	Group byte
	Op    byte
	Data  [3]byte
}

func (this Instruction) Decode() (*DecodedInstruction, error) {
	var di DecodedInstruction
	di.Group = this.group()
	di.Op = this.op()
	if value, err := this.register(1); err != nil {
		return nil, err
	} else {
		di.Data[0] = value
	}
	if value, err := this.register(2); err != nil {
		return nil, err
	} else {
		di.Data[1] = value
	}
	if value, err := this.register(3); err != nil {
		return nil, err
	} else {
		di.Data[2] = value
	}
	return &di, nil
}

func (this *DecodedInstruction) SetImmediate(value Word) {
	this.Data[1] = byte(value)
	this.Data[2] = byte(value >> 8)
}
func (this *DecodedInstruction) Immediate() Word {
	return Word((Word(this.Data[2]) << 8) | Word(this.Data[1]))
}

func (this *DecodedInstruction) Encode() *Instruction {
	i := new(Instruction)
	// encode group
	i.setGroup(this.Group)
	i.setOp(this.Op)
	i.setByte(1, this.Data[0])
	i.setByte(2, this.Data[1])
	i.setByte(3, this.Data[2])
	return i
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

type ExecutionUnit func(*Core, *DecodedInstruction) error
type SystemCall ExecutionUnit
type Core interface {
	Register(index byte) Word
	SetRegister(index byte, value Word) error
	CodeMemory(address Word) Instruction
	SetCodeMemory(address Word, value Instruction) error
	StackMemory(address Word) Word
	SetStackMemory(address, value Word) error
	DataMemory(address Word) Word
	SetDataMemory(address, value Word) error
	// special registers
	InstructionPointer() Word
	SetInstructionPointer(value Word) error
	StackPointer() Word
	SetStackPointer(value Word) error
	Predicate() Word
	SetPredicate(value Word) error
	LinkRegister() Word
	SetLinkRegister(value Word) error
	CountRegister() Word
	SetCountRegister(value Word) error
	// execution unit related stuff
	InstallExecutionUnit(group byte, fn ExecutionUnit) error
	InvokeExecution(inst *DecodedInstruction) error
	// system call related stuff
	InstallSystemCall(index byte, fn SystemCall) error
	SystemCall(inst *DecodedInstruction) error
	HaltExecution()
	ResumeExecution()
	// how to dispatch from the front end to the backend
	Dispatch(inst Instruction) error
}
type Core struct {
	gpr   [RegisterCount - UserRegisterBegin]Word
	Code  [MemorySize]Instruction
	Data  [MemorySize]Word
	Stack [MemorySize]Word
	// internal registers that should be easy to find
	instructionPointer Word
	stackPointer       Word
	link               Word
	count              Word
	predicate          Word
	advancePc          bool
	terminateExecution bool
	Groups             [MajorOperationGroupCount]ExecutionUnit
	SystemCalls        [SystemCallCount]SystemCall
}

func defaultExtendedUnit(core *Core, inst *DecodedInstruction) error {
	return newError(ErrorInvalidInstructionGroupProvided, uint(inst.Group))
}

func (this *Core) SetRegister(index byte, value Word) error {
	switch index {
	case FalseRegister:
		return newError(ErrorWriteToFalseRegister, uint(value))
	case TrueRegister:
		return newError(ErrorWriteToTrueRegister, uint(value))
	case InstructionPointer:
		this.instructionPointer = value
	case StackPointer:
		this.stackPointer = value
	case PredicateRegister:
		this.predicate = value
	case CountRegister:
		this.count = value
	case LinkRegister:
		this.link = value
	default:
		this.gpr[index-UserRegisterBegin] = value
	}
	return nil
}
func (this *Core) GetRegister(index byte) Word {
	switch index {
	case FalseRegister:
		return 0
	case TrueRegister:
		return 1
	case InstructionPointer:
		return this.instructionPointer
	case StackPointer:
		return this.stackPointer
	case PredicateRegister:
		return this.predicate
	case LinkRegister:
		return this.link
	case CountRegister:
		return this.count
	default:
		// do the offset calculation
		return this.gpr[index-UserRegisterBegin]
	}
}
func (this *Core) InstructionPointer() Word {
	return this.instructionPointer
}
func (this *Core) SetInstructionPointer(value Word) {
	this.instructionPointer = value
}
func (this *Core) HaltExecution() {
	this.terminateExecution = true
}
func (this *Core) ResumeExecution() {
	this.terminateExecution = false
}
func New() *Core {
	var c Core
	c.instructionPointer = 0
	c.advancePc = true
	c.terminateExecution = false
	c.predicate = 0
	c.stackPointer = 0xFFFF
	c.link = 0
	c.count = 0
	for i := 0; i < MajorOperationGroupCount; i++ {
		c.Xunits[i] = defaultExtendedUnit
	}
	for i := 0; i < SystemCallCount; i++ {
		c.SystemCalls[i] = defaultSystemCall
	}
	c.SystemCalls[SystemCommandTerminate] = terminateSystemCall
	c.SystemCalls[SystemCommandPanic] = panicSystemCall
	return &c
}
func (this *Core) Dispatch(inst Instruction) error {
	this.advancePc = true
	if di, err := inst.Decode(); err != nil {
		return err
	} else {
		return
		switch di.Group {

		case InstructionGroupArithmetic:
			return this.arithmetic(di)
		case InstructionGroupMove:
			return this.move(di)
		case InstructionGroupJump:
			return this.jump(di)
		case InstructionGroupCompare:
			return this.compare(di)
		case InstructionGroupMisc:
			return this.misc(di)
		case InstructionGroupExtended0: // expansion group0
			return this.extended0(di)
		case InstructionGroupExtended1: // expansion group1
			return this.extended1(di)
		default:
			return newError(ErrorInvalidInstructionGroupProvided, uint(di.Group))
		}
	}
}
func (this *Core) extended0(inst *DecodedInstruction) error {
	return this.Xunits[0](this, inst)
}
func (this *Core) extended1(inst *DecodedInstruction) error {
	return this.Xunits[1](this, inst)
}

func (this *Core) misc(inst *DecodedInstruction) error {
	switch inst.Op {
	case MiscOpSystemCall:
		return this.SystemCall(inst)
	default:
		return newError(ErrorInvalidMiscOperation, uint(inst.Op))
	}
}
func panicSystemCall(core *Core, inst *DecodedInstruction) error {
	// we don't want to panic the program itself but generate a new error
	// look at the data attached to the panic and encode it
	return newError(ErrorPanic, uint(inst.Immediate()))
}
func terminateSystemCall(core *Core, inst *DecodedInstruction) error {
	core.HaltExecution()
	return nil
}
func defaultSystemCall(core *Core, inst *DecodedInstruction) error {
	return newError(ErrorInvalidSystemCommand, uint(inst.Data[0]))
}
func (this *Core) SystemCall(inst *DecodedInstruction) error {
	return this.SystemCalls[inst.Data[0]](this, inst)
}

func (this *Core) arithmetic(inst *DecodedInstruction) error {
	switch inst.Op {
	default:
		return newError(ErrorInvalidArithmeticOperation, uint(inst.Op))
	}
}
func (this *Core) move(inst *DecodedInstruction) error {
	switch inst.Op {
	default:
		return newError(ErrorInvalidMoveOperation, uint(inst.Op))
	}
}
func (this *Core) jump(inst *DecodedInstruction) error {
	switch inst.Op {
	default:
		return newError(ErrorInvalidJumpOperation, uint(inst.Op))
	}
}
func (this *Core) compare(inst *DecodedInstruction) error {
	switch inst.Op {
	default:
		return newError(ErrorInvalidCompareOperation, uint(inst.Op))
	}
}
