// machine description of iris2
package iris2

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/DrItanium/cores/registration/machine"
	"math"
)

var arg_dataMemorySize = flag.Uint64("iris2.dataSegmentSize", dataMemorySize, "The size of the data segment")
var arg_codeMemorySize = flag.Uint64("iris2.codeSegmentSize", codeMemorySize, "The size of the code segment")
var arg_microcodeMemorySize = flag.Uint64("iris2.microcodeSegmentSize", microcodeMemorySize, "The size of the microcode segment")
var arg_stackMemorySize = flag.Uint64("iris2.stackSegmentSize", stackMemorySize, "The size of the stack segment")
var arg_callMemorySize = flag.Uint64("iris2.callSegmentSize", callMemorySize, "The size of the call segment")

func RegistrationName() string {
	return "iris2"
}

// Dummy function to force inclusion
func Register() {}

type MachineRegistrar func(...interface{}) (machine.Machine, error)

func (this MachineRegistrar) New(args ...interface{}) (machine.Machine, error) {
	return this(args)
}

func generateCore(a ...interface{}) (machine.Machine, error) {
	return newCore()
}

func init() {
	machine.Register(RegistrationName(), MachineRegistrar(generateCore))
}

const (
	generalPurposeRegisterCount = 256
	floatRegisterCount          = 256
	predicateRegisterCount      = 16
	majorOperationGroupCount    = 8
	systemCallCount             = 256
	dataMemorySize              = 268435456 / 8 // 256mb
	stackMemorySize             = 268435456 / 8
	callMemorySize              = 268435456 / 8
	codeMemorySize              = 268435456 / 8
	microcodeMemorySize         = 268435456 / 8
)

type errorCode int

const (
	// Error codes
	errorNone errorCode = iota
	errorPanic
	errorGetRegisterOutOfRange
	errorPutRegisterOutOfRange
	errorInvalidInstructionGroupProvided
	errorInvalidArithmeticOperation
	errorInvalidMoveOperation
	errorInvalidJumpOperation
	errorInvalidCompareOperation
	errorInvalidMiscOperation
	errorInvalidSystemCommand
	errorWriteToFalseRegister
	errorWriteToTrueRegister
	errorEncodeByteOutOfRange
	errorGroupValueOutOfRange
	errorOpValueOutOfRange
)

type instructionGroup byte

const (
	// Instruction groups
	igArithmetic instructionGroup = iota
	igMove
	igJump
	igCompare
	igMisc
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

type predicate bool
type halfWord uint32
type word uint64
type dword [2]uint64
type floatWord float64
type instruction word
type instructionField byte

type dataAbstraction interface {
	Float() (floatWord, error)
	Int() (word, error)
	Bool() (predicate, error)
	Instruction() (instruction, error)
}

func (this word) Float() (floatWord, error) {
	return floatWord(math.Float64frombits(this)), nil
}
func (this word) Int() (word, error) {
	return this, nil
}
func (this word) Bool() (predicate, error) {
	return predicate(word != 0), nil
}

func (this word) Instruction() (instruction, error) {
	return instruction(this), nil
}

func (this floatWord) Float() (floatWord, error) {
	return this, nil
}
func (this floatWord) Int() (word, error) {
	return word(math.Float64bits(this)), nil
}
func (this floatWord) Bool() (predicate, error) {
	return 0.0, fmt.Errorf("This is a floatWord, not a predicate. Implicit conversion will not take place due to precision issues!")
}

func (this floatWord) Instruction() (instruction, error) {
	return instruction(math.Float64bits(this)), nil
}

func (this predicate) Float() (floatWord, error) {
	return 0.0, fmt.Errorf("This is a predicate, not a floatWord. Implicit conversion will not take place due to precision issues!")
}
func (this predicate) Int() (word, error) {
	if this {
		return word(1), nil
	} else {
		return word(0), nil
	}
}
func (this predicate) Bool() (predicate, error) {
	return this, nil
}
func (this predicate) Instruction() (instruction, error) {
	if this {
		return instruction(1), nil
	} else {
		return instruction(0), nil
	}
}

const (
	fieldPredicate instructionField = iota
	fieldControlPrimary
	fieldControlSecondary
	fieldDest
	fieldSrc0
	fieldSrc1
	fieldMiscLower
	fieldMiscUpper
	fieldCount // always last
)

func init() {
	if fieldCount > 8 {
		panic("iris2: Too many fields described in the raw instruction field listing")
	}
}

var iFields = []struct {
	mask  word
	shift byte
}{
	{mask: 0x00000000000000FF, shift: 0},
	{mask: 0x000000000000FF00, shift: 8},
	{mask: 0x0000000000FF0000, shift: 16},
	{mask: 0x00000000FF000000, shift: 24},
	{mask: 0x000000FF00000000, shift: 32},
	{mask: 0x0000FF0000000000, shift: 40},
	{mask: 0x00FF000000000000, shift: 48},
	{mask: 0xFF00000000000000, shift: 56},
}

func (this instruction) Int() (word, error) {
	return word(this), nil
}
func (this instruction) Float() (floatWord, error) {
	return word(this).Float()
}
func (this instruction) Bool() (predicate, error) {
	return word(this).Bool()
}
func (this instruction) getField(index instructionField) (byte, error) {
	if index >= len(iFields) {
		return 0, fmt.Errorf("Field index %d is not a legal field!", index)
	} else {
		return byte((this & iFields[index].mask) >> iFields[index].mask), nil
	}
}
func (this *instruction) setField(index instructionField, value byte) error {
	if index >= len(iFields) {
		return fmt.Errorf("Field index %d is not a legal field!", index)
	} else {
		*this = (*this &^ iFields[index].mask) | (word(value) << iFields[index].shift)
		return nil
	}
}
func (this instruction) group() byte {
	return byte(this.getField(fieldControlPrimary) & 0x7)
}
func (this instruction) op() byte {
	return byte((this.getField(fieldControlPrimary) & 0xF8) >> 3)
}

func (this *instruction) setGroup(group byte) error {
	if fld, err := this.getField(fieldControlPrimary); err != nil {
		return err
	} else {
		return this.setField(fieldControlPrimary, (fld&^0x7)|group)
	}
}

func (this *instruction) setOp(op byte) error {
	if fld, err := this.getField(fieldControlPrimary); err != nil {
		return err
	} else {
		return this.setField(fieldControlPrimary, (fld&^0xF8)|(op<<3))
	}
}

type decodedInstruction struct {
	predicate byte
	group, op byte // one byte encoded
	control2  byte
	data      [5]byte
}

func (this instruction) decode() (*decodedInstruction, error) {
	var di decodedInstruction
	di.group = this.group()
	di.op = this.op()
	if p, err = this.getField(fieldPredicate); err != nil {
		return nil, err
	} else if d0, err := this.getField(fieldControlSecondary); err != nil {
		return nil, err
	} else if d1, err := this.getField(fieldDest); err != nil {
		return nil, err
	} else if d2, err := this.getField(fieldSrc0); err != nil {
		return nil, err
	} else if d3, err := this.getField(fieldSrc1); err != nil {
		return nil, err
	} else if d4, err := this.getField(fieldMiscLower); err != nil {
		return nil, err
	} else if d5, err := this.getField(fieldMiscUpper); err != nil {
		return nil, err
	} else {
		di.control2 = d0
		di.data = [5]byte{d1, d2, d3, d4, d5}
		return &di, nil
	}
}

var dataTranslationTable = []instructionField{
	fieldDest,
	fieldSrc0,
	fieldSrc1,
	fieldMiscLower,
	fieldMiscUpper,
}

func (this *decodedInstruction) Encode() (*instruction, error) {
	var i instruction
	// encode group
	if err := i.setField(fieldPredicate, this.predicate); err != nil {
		return nil, err
	} else if err := i.setField(fieldControlSecondary, this.control2); err != nil {
		return nil, err
	} else {
		i.setGroup(this.group)
		i.setOp(this.op)
		for x, v := range this.data {
			if err := i.setField(dataTranslationTable[x], v); err != nil {
				return nil, err
			}
		}
		return &i, nil
	}
}

type machError struct {
	value, code uint
}

func newError(code, value uint) error {
	return &machError{code: code, value: value}
}

func (this machError) Error() string {
	if this.code == 0 {
		return fmt.Sprintf("No Error with value %d!!! This should never ever showup!", this.value)
	} else if this.code >= uint(len(errorLookup)) {
		return fmt.Sprintf("Unknown error %d with value %d! Something really bad happened!", this.code, this.value)
	} else {
		return fmt.Sprintf(errorLookup[this.code], this.value)
	}
}

type gprReservedRegister byte
type fprReservedRegister byte
type predicateReservedRegister byte

const (
	// reserved registers
	gprZeroRegister gprReservedRegister = iota
	gprOneRegister
	gprInstructionPointer
	gprStackPointer
	gprCallPointer
	userGPRRegisterBegin
)

const (
	predicateTrueRegister predicateReservedRegister = iota
	predicateFalseRegister

	userPredicateRegisterBegin
)

const (
	fprZeroRegister fprReservedRegister = iota
	fprOneRegister

	userFprRegisterBegin
)

type registerSegment byte

const (
	rsGPR registerSegment = iota
	rsFPR
	rsPredicate
	rsCount
)

type ExecutionUnit func(*core, *decodedInstruction) error
type SystemCall ExecutionUnit
type codeMemorySegment []instruction
type wordMemorySegment []word

type segmentAbstraction interface {
	Set(address, value dataAbstraction) error
	Get(address dataAbstraction) (dataAbstraction, error)
}

func (this codeMemorySegment) Set(a, v dataAbstraction) error {
	if addr, err := a.Int(); err != nil {
		return err
	} else if value, err := a.Instruction(); err != nil {
		return err
	} else if addr >= len(this) {
		return fmt.Errorf("Can't store %x in illegal address %x", value, addr)
	} else {
		this[addr] = value
		return nil
	}
}
func (this codeMemorySegment) Get(a dataAbstraction) (dataAbstraction, error) {
	if addr, err := a.Int(); err != nil {
		return 0, err
	} else if addr >= len(this) {
		return fmt.Errorf("Can't read from %x", addr)
	} else {
		return this[addr], nil
	}
}

func (this wordMemorySegment) Set(a, v dataAbstraction) error {
	if addr, err := a.Int(); err != nil {
		return err
	} else if value, err := a.Int(); err != nil {
		return err
	} else if addr >= len(this) {
		return fmt.Errorf("Can't store %x in illegal address %x", value, addr)
	} else {
		this[addr] = value
		return nil
	}
}
func (this wordMemorySegment) Get(a dataAbstraction) (dataAbstraction, error) {
	if addr, err := a.Int(); err != nil {
		return 0, err
	} else if addr >= len(this) {
		return fmt.Errorf("Can't read from %x", addr)
	} else {
		return this[addr], nil
	}
}

type core struct {
	gpr                      [generalPurposeRegisterCount - userGPRRegisterBegin]word
	fpr                      [floatRegisterCount - userFprRegisterBegin]floatWord
	predicates               [predicateRegisterCount - userPredicateRegisterBegin]predicate
	code                     codeMemorySegment
	data, ucode, stack, call wordMemorySegment
	// internal registers that should be easy to find
	instructionPointer            halfWord
	stackPointer                  word
	callPointer                   word
	advancePc, terminateExecution bool
	groups                        [MajorOperationGroupCount]ExecutionUnit
	systemCalls                   [SystemCallCount]SystemCall
}

func (this *core) setFloatRegister(index fprReservedRegister, value floatWord) error {
	if index >= floatRegisterCount {
		return fmt.Errorf("Index %d is out of range for the given set of floating point registers!", index)
	} else {
		switch index {
		case fprZeroRegister:
			return newError(errorWriteToFloatZeroRegister, value)
		case fprOneRegister:
			return newError(errorWriteToFloatOneRegister, value)
		default:
			this.fpr[index-userFprRegisterBegin] = value
			return nil
		}
	}
}
func (this *core) setPredicateRegister(index predicateReservedRegister, value predicate) error {
	if index >= predicateRegisterCount {
		return fmt.Errorf("Index %d is out of range for the given set of predicate registers!", index)
	} else {
		switch index {
		case trueRegister:
			return newError(errorWriteToTrueRegister, value)
		case falseRegister:
			return newError(errorWriteToFalseRegister, value)
		default:
			this.predicates[index-userPredicateRegisterBegin] = value
			return nil
		}
	}
}
func (this *core) setGPRRegister(index gprReservedRegister, value Word) error {
	if index >= generalPurposeRegisterCount {
		return fmt.Errorf("Index %d is out of range for the given set of GPRs!", index)
	} else {
		switch index {
		case gprZeroRegister:
			return newError(errorWriteToIntZeroRegister, uint(value))
		case gprOneRegister:
			return newError(errorWriteToIntOneRegister, uint(value))
		case gprInstructionPointer:
			this.instructionPointer = value
		case gprStackPointer:
			this.stackPointer = value
		case gprCallPointer:
			this.callPointer = value
		default:
			this.gpr[index-userGPRRegisterBegin] = value
		}
		return nil
	}
}

func (this *core) setRegister(seg registerSegment, index byte, value dataAbstraction) error {
	switch seg {
	case rsGPR:
		if val, err := value.Int(); err != nil {
			return err
		} else {
			return this.setGPRRegister(gprReservedRegister(index), val)
		}
	case rsFPR:
		if val, err := value.Float(); err != nil {
			return err
		} else {
			return this.setFloatRegister(fprReservedRegister, val)
		}
	case rsPredicate:
		if val, err := value.Bool(); err != nil {
			return err
		} else {
			return this.setPredicateRegister(predicateReservedRegister(index), val)
		}
	default:
		return fmt.Errorf("Attempted to access illegal registerSegment %d!", seg)
	}
}
func (this *core) register(seg registerSegment, index byte) (dataAbstraction, error) {
	switch seg {
	case rsGPR:
		return this.gprRegister(gprReservedRegister(index))
	case rsFPR:
		return this.fprRegister(fprReservedRegister(index))
	case rsPredicate:
		return this.predicateRegister(predicateReservedRegister(index))
	default:
		return word(0), fmt.Errorf("Attempted to access illegal registerSegment %d!", seg)
	}
}

func (this *core) gprRegister(index gprReservedRegister) (word, error) {
	if index >= generalPurposeRegisterCount {
		return 0, fmt.Errorf("%d is out of range for a GPR!", index)
	} else {
		switch index {
		case gprZeroRegister:
			return 0, nil
		case gprOneRegister:
			return 1, nil
		case gprInstructionPointer:
			return this.instructionPointer, nil
		case gprStackPointer:
			return this.stackPointer, nil
		case gprCallPointer:
			return this.callPointer, nil
		default:
			// do the offset calculation
			return this.gpr[index-userGPRRegisterBegin], nil
		}
	}
}
func (this *core) fprRegister(index fprReservedRegister) (floatWord, error) {
	if index >= floatRegisterCount {
		return 0, fmt.Errorf("%d is out of range for a FPR!", index)
	} else {
		switch index {
		case fprZeroRegister:
			return 0.0, nil
		case fprOneRegister:
			return 1.0, nil
		default:
			// do the offset calculation
			return this.fpr[index-userFprRegisterBegin], nil
		}
	}
}

func (this *core) predicateRegister(index predicateReservedRegister) (predicate, error) {
	if index >= predicateRegisterCount {
		return 0, fmt.Errorf("%d is out of range for a predicate!", index)
	} else {
		switch index {
		case predicateTrueRegister:
			return true, nil
		case predicateFalseRegister:
			return false, nil
		default:
			// do the offset calculation
			return this.predicates[index-userPredicateRegisterBegin], nil
		}
	}
}
func (this *core) wordSegment(seg segment) ([]word, error) {
	switch seg {
	case codeSegment:
		return nil, fmt.Errorf("The code segment is an instruction segment, not a word segment!")
	case dataSegment:
		return this.data, nil
	case microcodeSegment:
		return this.ucode, nil
	case callSegment:
		return this.call, nil
	case stackSegment:
		return this.stack, nil
	default:
		return nil, fmt.Errorf("Illegal code segment %d", seg)
	}
}
func setWordMemory(seg segment, addr, value word, mem []word) error {
	if addr >= seg.Size() {
		return fmt.Errorf("Attempted to write %x to address %x is outside legal memory bounds of the %s memory segment!", value, addr, seg.String())
	} else {
		mem[addr] = value
		return nil
	}
}
func getWordMemory(seg segment, addr word, mem []word) (word, error) {
	if addr >= seg.Size() {
		return fmt.Errorf("Attempted to read from address %x which is outside legal memory bounds of the %s memory segment!", addr, seg.String())
	} else {
		return mem[addr], nil
	}
}
func (this *core) setMemory(seg segment, addr, value word) error {
	switch seg {
	case codeSegment:
		if addr >= seg.Size() {
			return fmt.Errorf("Address %x is outside legal memory bounds of the code memory segment!", addr)
		} else {
			this.code[addr] = instruction(value)
			return nil
		}
	case dataSegment, microcodeSegment, callSegment, stackSegment:
		if mem, err := this.wordSegment(seg); err != nil {
			return err
		} else {
			return setWordMemory(seg, addr, value, mem)
		}
	default:
		return fmt.Errorf("Attempted to write %x to address %x in %s memory segment", value, addr, seg.String())
	}
}
func (this *core) memory(seg segment, addr word) error {
}
func (this *core) Call(addr Word) error {
	this.callPointer++
	this.call[this.callPointer] = this.NextInstructionAddress()
	return this.SetRegister(instructionPointer, addr)
}
func (this *core) Return() Word {
	value := this.call[this.callPointer]
	this.callPointer--
	return value
}
func (this *core) Push(value Word) {
	this.stackPointer++
	this.stack[this.stackPointer] = value
}
func (this *core) Peek() Word {
	return this.stack[this.stackPointer]
}
func (this *core) Pop() Word {
	value := this.stack[this.stackPointer]
	this.stackPointer--
	return value
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

var systemCalls = []struct {
	Op byte
	fn SystemCall
}{
	{Op: SystemCallTerminate, fn: terminateSystemCall},
	{Op: SystemCallPanic, fn: panicSystemCall},
	{Op: SystemCallPutc, fn: putcSystemCall},
}

func newCore() (*core, error) {
	var c core
	c.advancePc = true
	c.terminateExecution = false
	c.data = make(wordMemorySegment, *arg_dataMemorySize)
	c.code = make(codeMemorySegment, *arg_codeMemorySize)
	c.ucode = make(wordMemorySegment, *arg_microcodeMemorySize)
	c.stack = make(wordMemorySegment, *arg_stackMemorySize)
	c.call = make(wordMemorySegment, *arg_callMemorySize)
	if err := c.setRegister(gprInstructionPointer, 0); err != nil {
		return nil, err
	}
	// install default handlers
	for i := 0; i < MajorOperationGroupCount; i++ {
		if err := c.InstallExecutionUnit(byte(i), defaultExtendedUnit); err != nil {
			return nil, err
		}
	}
	// overwrite with real execution units
	for _, unit := range defaultExecutionUnits {
		if err := c.InstallExecutionUnit(unit.Group, unit.Unit); err != nil {
			return nil, err
		}
	}
	// install default system calls
	for i := 0; i < SystemCallCount; i++ {
		if err := c.InstallSystemCall(byte(i), defaultSystemCall); err != nil {
			return nil, err
		}
	}
	// overwrite
	for _, s := range systemCalls {
		if err := c.InstallSystemCall(s.Op, s.fn); err != nil {
			return nil, err
		}
	}
	return &c, nil
}

func (this *core) InstallExecutionUnit(group byte, fn ExecutionUnit) error {
	if group >= MajorOperationGroupCount {
		return newError(ErrorGroupValueOutOfRange, uint(group))
	} else {
		this.groups[group] = fn
		return nil
	}
}
func (this *core) invoke(inst *decodedInstruction) error {
	if p, err := this.predicateRegister(inst.predicate); err != nil {
		return err
	} else if p {
		return this.groups[inst.group](this, inst)
	} else {
		return nil
	}
}
func (this *core) InstallSystemCall(offset byte, fn SystemCall) error {
	this.systemCalls[offset] = fn
	return nil
}
func (this *core) SystemCall(inst *decodedInstruction) error {
	return this.systemCalls[inst.Data[0]](this, inst)
}

func defaultExtendedUnit(core *core, inst *decodedInstruction) error {
	return newError(ErrorInvalidInstructionGroupProvided, uint(inst.Group))
}

func (this *core) currentInstruction() (instruction, error) {
	ip := this.instructionPointer
	if ip >= codeSegment.size() {
		// should we chop it or error out....error out since it is safer
		// jumped outside legal memory
		return 0, fmt.Errorf("Attempted to load an instruction from address %x outside code segment!", ip)
	} else {
		return this.code[ip], nil
	}
}

func (this *core) advanceProgramCounter() error {
	if this.advancePc {
		value := this.instructionPointer + 1
		if value >= codeSegment.Size() {
			value = 0
		}
		this.instructionPointer = value
	} else {
		this.advancePc = true
	}
	return nil
}

func (this *core) Run() error {
	for !this.terminateExecution {
		// extract the current instruction
		this.advancePc = true
		if inst, err := this.currentInstruction(); err != nil {
			return fmt.Errorf("ERROR during extraction of current instruction: %s\n", err)
		} else if di, err := inst.Decode(); err != nil {
			return fmt.Errorf("ERROR during decode: %s\n", err)
		} else if err := this.invoke(di); err != nil {
			return fmt.Errorf("ERROR during execution: %s\n", err)
		} else if err := this.advanceProgramCounter(); err != nil {
			return fmt.Errorf("ERROR during the advancement of the program counter: %s", err)
		}
	}
	return nil
}

func (this *core) GetDebugStatus() bool {
	return false
}

func (this *core) SetDebug(_ bool) {

}

func readWord(input <-chan byte) (word, error) {
	if value, more := <-input; !more {
		// closed early it seems :(
		return 0, fmt.Errorf("Closed stream")
	} else if value2, more0 := <-input; !more0 {
		return 0, fmt.Errorf("Closed stream")
	} else if value3, more1 := <-input; !more1 {
		return 0, fmt.Errorf("Closed stream")
	} else if value4, more2 := <-input; !more2 {
		return 0, fmt.Errorf("Closed stream")
	} else if value5, more3 := <-input; !more3 {
		return 0, fmt.Errorf("Closed stream")
	} else if value6, more4 := <-input; !more4 {
		return 0, fmt.Errorf("Closed stream")
	} else if value7, more5 := <-input; !more5 {
		return 0, fmt.Errorf("Closed stream")
	} else if value8, more6 := <-input; !more6 {
		return 0, fmt.Errorf("Closed stream")
	} else {
		return binary.LittleEndian.Uint64([]byte{value, value2, value3, value4, value5, value6, value7, value8}), nil
	}

}
func (this *core) InstallProgram(input <-chan byte) error {
	for i := 0; i < codeMemorySize; i++ {
		if inst, err := readWord(input); err != nil {
			return err
		} else {
			this.code[i] = instruction(inst)
		}
	}
	for i := 0; i < dataMemorySize; i++ {
		if inst, err := readWord(input); err != nil {
			return err
		} else {
			this.data[i] = inst
		}
	}
	for i := 0; i < microcodeMemorySize; i++ {
		if inst, err := readWord(input); err != nil {
			return err
		} else {
			this.ucode[i] = inst
		}
	}
	for i := 0; i < stackMemorySize; i++ {
		if inst, err := readWord(input); err != nil {
			return err
		} else {
			this.stack[i] = inst
		}
	}
	for i := 0; i < callMemorySize; i++ {
		if inst, err := readWord(input); err != nil {
			return err
		} else {
			this.call[i] = inst
		}
	}
	return nil
}

func (this *core) Dump(output chan<- byte) error {
	w := make([]byte, 8)
	for _, dat := range this.code {
		binary.LittleEndian.PutUint32(w, uint64(dat))
		for _, v := range w {
			output <- v
		}
	}
	for _, dat := range this.data {
		binary.LittleEndian.PutUint16(w, uint64(dat))
		for _, v := range w {
			output <- v
		}
	}
	for _, dat := range this.ucode {
		binary.LittleEndian.PutUint16(w, uint64(dat))
		for _, v := range w {
			output <- v
		}
	}
	for _, val := range this.stack {
		binary.LittleEndian.PutUint16(w, uint64(val))
		for _, b := range w {
			output <- b
		}
	}
	for _, val := range this.call {
		binary.LittleEndian.PutUint16(w, uint64(val))
		for _, b := range w {
			output <- b
		}
	}
	return nil
}

func (this *core) Startup() error {
	return nil
}
func (this *core) Shutdown() error {
	return nil
}

type segment int

const (
	codeSegment segment = iota
	dataSegment
	microcodeSegment
	stackSegment
	callSegment
	numSegments
)

func init() {
	if numSegments > 255 {
		panic("Too many memory segments described!")
	}
}
func (this segment) acceptsWords() bool {
	return this == dataSegment || this == microcodeSegment || this == stackSegment || this == callSegment
}

func (this segment) String() string {
	switch this {
	case codeSegment:
		return "code"
	case stackSegment:
		return "stack"
	case dataSegment:
		return "data"
	case microcodeSegment:
		return "microcode"
	case callSegment:
		return "call"
	default:
		return fmt.Sprintf("unknown(%d)", this)
	}
}

func (this segment) Size() word {
	switch this {
	case codeSegment:
		return *arg_codeMemorySize
	case dataSegment:
		return *arg_dataMemorySize
	case microcodeSegment:
		return *arg_microcodeMemorySize
	case stackSegment:
		return *arg_stackMemorySize
	case callSegment:
		return *arg_callMemorySize
	default:
		return 0
	}
}
