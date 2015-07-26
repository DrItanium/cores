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

	FalseRegister = iota
	TrueRegister
	InstructionPointer
	StackPointer
	PredicateRegister
	CountRegister
	CallPointer
	UserRegisterBegin
	// groups
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

	// Instruction groups
	InstructionGroupArithmetic = iota
	InstructionGroupMove
	InstructionGroupJump
	InstructionGroupCompare
	InstructionGroupMisc
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
		return NewError(ErrorEncodeByteOutOfRange, uint(index))
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

func NewError(code, value uint) error {
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

type Core struct {
	gpr   [RegisterCount - UserRegisterBegin]Word
	code  [MemorySize]Instruction
	data  [MemorySize]Word
	stack [MemorySize]Word
	call  [MemorySize]Word
	// internal registers that should be easy to find
	instructionPointer Word
	stackPointer       Word
	callPointer        Word
	count              Word
	predicate          Word
	advancePc          bool
	terminateExecution bool
	groups             [MajorOperationGroupCount]ExecutionUnit
	systemCalls        [SystemCallCount]SystemCall
}

func (this *Core) SetRegister(index byte, value Word) error {
	switch index {
	case FalseRegister:
		return NewError(ErrorWriteToFalseRegister, uint(value))
	case TrueRegister:
		return NewError(ErrorWriteToTrueRegister, uint(value))
	case InstructionPointer:
		this.instructionPointer = value
	case StackPointer:
		this.stackPointer = value
	case PredicateRegister:
		this.predicate = value
	case CountRegister:
		this.count = value
	case CallPointer:
		this.callPointer = value
	default:
		this.gpr[index-UserRegisterBegin] = value
	}
	return nil
}
func (this *Core) Register(index byte) Word {
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
	case CountRegister:
		return this.count
	case CallPointer:
		return this.callPointer
	default:
		// do the offset calculation
		return this.gpr[index-UserRegisterBegin]
	}
}

func (this *Core) CodeMemory(address Word) Instruction {
	return this.code[address]
}
func (this *Core) SetCodeMemory(address Word, value Instruction) error {
	this.code[address] = value
	return nil
}
func (this *Core) Call(addr Word) error {
	this.callPointer++
	this.call[this.callPointer] = this.NextInstructionAddress()
	return this.SetRegister(InstructionPointer, addr)
}
func (this *Core) Return() Word {
	value := this.call[this.callPointer]
	this.callPointer--
	return value
}
func (this *Core) Push(value Word) {
	this.stackPointer++
	this.stack[this.stackPointer] = value
}
func (this *Core) Peek() Word {
	return this.stack[this.stackPointer]
}
func (this *Core) Pop() Word {
	value := this.stack[this.stackPointer]
	this.stackPointer--
	return value
}
func (this *Core) DataMemory(address Word) Word {
	return this.data[address]
}
func (this *Core) SetDataMemory(address, value Word) error {
	this.data[address] = value
	return nil
}

func New() (*Core, error) {
	var c Core
	c.advancePc = true
	c.terminateExecution = false
	if err := c.SetRegister(InstructionPointer, 0); err != nil {
		return nil, err
	}
	if err := c.SetRegister(PredicateRegister, 0); err != nil {
		return nil, err
	}
	if err := c.SetRegister(StackPointer, 0xFFFF); err != nil {
		return nil, err
	}
	if err := c.SetRegister(CallPointer, 0xFFFF); err != nil {
		return nil, err
	}
	if err := c.SetRegister(CountRegister, 0); err != nil {
		return nil, err
	}
	for i := 0; i < MajorOperationGroupCount; i++ {
		if err := c.InstallExecutionUnit(byte(i), defaultExtendedUnit); err != nil {
			return nil, err
		}
	}
	for i := 0; i < SystemCallCount; i++ {
		if err := c.InstallSystemCall(byte(i), defaultSystemCall); err != nil {
			return nil, err
		}
	}
	return &c, nil
}

func (this *Core) InstallExecutionUnit(group byte, fn ExecutionUnit) error {
	if group >= MajorOperationGroupCount {
		return NewError(ErrorGroupValueOutOfRange, uint(group))
	} else {
		this.groups[group] = fn
		return nil
	}
}
func (this *Core) InvokeExecution(inst *DecodedInstruction) error {
	return this.groups[inst.Group](this, inst)
}
func (this *Core) InstallSystemCall(offset byte, fn SystemCall) error {
	this.systemCalls[offset] = fn
	return nil
}
func (this *Core) SystemCall(inst *DecodedInstruction) error {
	return this.systemCalls[inst.Data[0]](this, inst)
}

func (this *Core) Dispatch(inst Instruction) error {
	this.advancePc = true
	if di, err := inst.Decode(); err != nil {
		return err
	} else {
		return this.InvokeExecution(di)
	}
}
func panicSystemCall(core *Core, inst *DecodedInstruction) error {
	// we don't want to panic the program itself but generate a new error
	// look at the data attached to the panic and encode it
	return NewError(ErrorPanic, uint(inst.Immediate()))
}
func defaultSystemCall(core *Core, inst *DecodedInstruction) error {
	return NewError(ErrorInvalidSystemCommand, uint(inst.Data[0]))
}

func (this *Core) ShouldExecute() bool {
	return this.terminateExecution
}
func (this *Core) HaltExecution() {
	this.terminateExecution = true
}
func (this *Core) ResumeExecution() {
	this.terminateExecution = false
}
func defaultExtendedUnit(core *Core, inst *DecodedInstruction) error {
	return NewError(ErrorInvalidInstructionGroupProvided, uint(inst.Group))
}

const (
	// Misc operations
	MiscOpSystemCall = iota
	// System commands
	SystemCommandTerminate = iota
	SystemCommandPanic     = 255
)

func (this *Core) InstructionAddress() Word {
	return this.Register(InstructionPointer)
}
func (this *Core) NextInstructionAddress() Word {
	return this.Register(InstructionPointer) + 1
}
func (this *Core) PredicateValue(index byte) bool {
	return this.Register(index) != 0
}
