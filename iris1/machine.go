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
	NumDataBytes             = 5
)
const (
	// reserved registers
	FalseRegister = iota
	TrueRegister
	InstructionPointer
	StackPointer
	PredicateRegister
	CountRegister
	CallPointer
	UserRegisterBegin
)

const (
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
const (
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

type HalfWord uint16
type Word uint32
type Instruction uint64

var masks = []struct {
	Shift, Mask Instruction
}{
	{0, 0x00000000000000FF},
	{8, 0x000000000000FF00},
	{16, 0x0000000000FF0000},
	{24, 0x00000000FF000000},
	{32, 0x000000FF00000000},
	{40, 0x0000FF0000000000},
}

func (this Instruction) group() byte {
	return byte((this & masks[0].Mask) & 0x7)
}
func (this Instruction) op() byte {
	return byte(((this & masks[0].Mask) & 0xF8) >> 3)
}
func (this Instruction) register(index int) (byte, error) {
	if index >= len(masks) {
		return 0, fmt.Errorf("Register index: %d is out of range!", index)
	} else {
		mask := masks[index]
		return byte((this & mask.Mask) >> mask.Shift), nil
	}
}

func (this *Instruction) setGroup(group byte) {
	*this = ((*this &^ 0x7) | Instruction(group))
}
func (this *Instruction) setOp(op byte) {
	*this = ((*this &^ 0xF8) | (Instruction(op) << 3))
}
func (this *Instruction) setByte(index int, value byte) error {
	if index >= len(masks) {
		return NewError(ErrorEncodeByteOutOfRange, uint(index))
	} else {
		mask := masks[index]
		*this = (*this &^ mask.Mask) | (Instruction(value) << mask.Shift)
		return nil
	}
}

type DecodedInstruction struct {
	Group byte
	Op    byte
	Data  [NumDataBytes]byte
}

func (this Instruction) Decode() (*DecodedInstruction, error) {
	var di DecodedInstruction
	di.Group = this.group()
	di.Op = this.op()
	for i := 1; i <= NumDataBytes; i++ {
		if value, err := this.register(i); err != nil {
			return nil, err
		} else {
			di.Data[i-1] = value
		}
	}
	return &di, nil
}

func (this *DecodedInstruction) SetImmediate(value Word) {
	this.Data[1] = byte(value)
	this.Data[2] = byte(value >> 8)
	this.Data[3] = byte(value >> 16)
	this.Data[4] = byte(value >> 24)
}
func (this *DecodedInstruction) Immediate() Word {
	return Word((Word(this.Data[4]) << 24) | (Word(this.Data[3]) << 16) | (Word(this.Data[2]) << 8) | Word(this.Data[1]))
}

func (this *DecodedInstruction) Encode() *Instruction {
	i := new(Instruction)
	// encode group
	i.setGroup(this.Group)
	i.setOp(this.Op)
	for index, val := range this.Data {
		i.setByte(index+1, val)
	}
	return i
}

type Error struct {
	value, code uint
}

func NewError(code, value uint) error {
	return &Error{code: code, value: value}
}

func (this Error) Error() string {
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
	gpr    [RegisterCount - UserRegisterBegin]Word
	memory *memController
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
type memControllerInput struct {
	Address Word
	Width   byte
}
type memControllerOutput struct {
	Err    error
	Result []byte
}
type memController struct {
	rawMemory []byte
	input     chan memControllerInput
	output    chan memControllerOutput
}

func newMemController(size uint32) *memController {
	var mc memController
	mc.rawMemory = make([]byte, size)
	mc.input = make(chan memControllerInput)
	mc.output = make(chan memControllerOutput)
	return &mc
}
func (this *memController) memory(address Word, width byte) ([]byte, error) {
	if address >= Word(len(this.rawMemory)) {
		return nil, fmt.Errorf("Attempted to access memory address %x outside of range!", address)
	} else if (address + Word(width)) >= Word(len(this.rawMemory)) {
		return nil, fmt.Errorf("Attempted to access %d cells starting at memory address %x! This will go outside range!", width, address)
	} else if width == 0 {
		return nil, fmt.Errorf("Attempted to read 0 bytes starting at address %x!", address)
	} else {
		return this.rawMemory[address:(Word(width-1) + address)], nil
	}
}
func (this *memController) setMemory(address Word, data []byte) error {
	if address >= Word(len(this.rawMemory)) {
		return fmt.Errorf("Memory address %x is outside of memory range!", address)
	} else if (address + Word(len(data))) >= Word(len(this.rawMemory)) {
		return fmt.Errorf("Writing %d cells starting at memory address %x will go out of range!", len(data), address)
	} else {
		for ind, val := range data {
			this.rawMemory[address+Word(ind)] = val
		}
		return nil
	}
}
func (this *memController) code(address Word) (Instruction, error) {
	if b, err := this.memory(address, 6); err != nil {
		return 0, err
	} else {
		var i Instruction
		for ind, val := range b {
			i.setByte(ind, val)
		}
		return i, nil
	}
}
func (this *memController) setCode(address Word, val Instruction) error {
	//decode an instruction
	contents := make([]byte, 6)
	for i := 0; i < len(contents); i++ {
		mask := masks[i]
		contents[i] = byte((val & mask.Mask) >> mask.Shift)
	}
	return this.setMemory(address, contents)
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

var defaultExecutionUnits = []struct {
	Group byte
	Unit  ExecutionUnit
}{
	{Group: InstructionGroupArithmetic, Unit: arithmetic},
	{Group: InstructionGroupMove, Unit: move},
	{Group: InstructionGroupJump, Unit: jump},
	{Group: InstructionGroupCompare, Unit: compare},
	{Group: InstructionGroupMisc, Unit: misc},
}

func New(memorySize) (*Core, error) {
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
	for _, unit := range defaultExecutionUnits {
		if err := c.InstallExecutionUnit(unit.Group, unit.Unit); err != nil {
			return nil, err
		}
	}
	for i := 0; i < SystemCallCount; i++ {
		if err := c.InstallSystemCall(byte(i), defaultSystemCall); err != nil {
			return nil, err
		}
	}
	c.InstallSystemCall(SystemCallTerminate, terminateSystemCall)
	c.InstallSystemCall(SystemCallPanic, panicSystemCall)
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
func (this *Core) Invoke(inst *DecodedInstruction) error {
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
		return this.Invoke(di)
	}
}
func terminateSystemCall(core *Core, inst *DecodedInstruction) error {
	core.terminateExecution = true
	return nil
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

func (this *Core) InstructionAddress() Word {
	return this.Register(InstructionPointer)
}
func (this *Core) NextInstructionAddress() Word {
	return this.Register(InstructionPointer) + 1
}
func (this *Core) PredicateValue(index byte) bool {
	return this.Register(index) != 0
}

func NewDecodedInstruction(group, op, data0, data1, data2 byte) (*DecodedInstruction, error) {
	if group >= MajorOperationGroupCount {
		return nil, fmt.Errorf("Provided group (%d) is out of range!", group)
	} else {
		var di DecodedInstruction
		di.Group = group
		di.Op = op
		di.Data[0] = data0
		di.Data[1] = data1
		di.Data[2] = data2
		return &di, nil
	}
}
func NewDecodedInstructionImmediate(group, op, data0 byte, imm Word) (*DecodedInstruction, error) {
	return NewDecodedInstruction(group, op, data0, byte(imm), byte(imm>>8))
}

const (
	// System commands
	SystemCallTerminate = iota
	SystemCallPanic
	NumberOfSystemCalls
)

func init() {
	if NumberOfSystemCalls > 256 {
		panic("Too many system commands defined!")
	}
}

func (this *Core) TerminateExecution() bool {
	return this.terminateExecution
}

func (this *Core) CurrentInstruction() Instruction {
	return this.code[this.Register(InstructionPointer)]
}

func (this *Core) AdvanceProgramCounter() error {
	if this.advancePc {
		if err := this.SetRegister(InstructionPointer, this.NextInstructionAddress()); err != nil {
			return err
		}
	} else {
		this.advancePc = true
	}
	return nil
}

func (this *Core) ExecuteCurrentInstruction() error {
	return this.Dispatch(this.CurrentInstruction())
}
