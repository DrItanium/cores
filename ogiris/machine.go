// straight port of the 16bit iris core from my C version
package ogiris

import (
	"fmt"
)

type Word uint16
type Dword uint32

const (
	RegisterCount             = 256
	MemorySize                = 65536
	MajorOperationMax         = 8
	MinorOperationMax         = 32
	PredicateRegisterIndex    = 255
	StackPointerRegisterIndex = 254
)

var registeredGroups = []struct {
	Name       string
	Count, Max int
}{
	{Name: "instruction groups", Count: GroupCount, Max: MajorOperationMax},
	{Name: "arithmetic operations", Count: ArithmeticOpCount, Max: MinorOperationMax},
	{Name: "move operations", Count: MoveOpCount, Max: MinorOperationMax},
	{Name: "jump operations", Count: JumpOpCount, Max: MinorOperationMax},
	{Name: "compare operations", Count: CompareOpCount, Max: MinorOperationMax},
	{Name: "misc operations", Count: MiscOpCount, Max: MinorOperationMax},
}

// instruction groups
const (
	GroupArithmetic = iota
	GroupMove
	GroupJump
	GroupCompare
	GroupMisc
	// add more here
	GroupCount
)

// arithmetic ops
const (
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
	ArithmeticOpCount
)

const (
	MoveOpMove         = iota // move r? r?
	MoveOpSwap                // swap r? r?
	MoveOpSwapRegAddr         // swap.reg.addr r? r?
	MoveOpSwapAddrAddr        // swap.addr.addr r? r?
	MoveOpSwapRegMem          // swap.reg.mem r? $imm
	MoveOpSwapAddrMem         // swap.addr.mem r? $imm
	MoveOpSet                 // set r? $imm
	MoveOpLoad                // load r? r?
	MoveOpLoadMem             // load.mem r? $imm
	MoveOpStore               // store r? r?
	MoveOpStoreAddr           // store.addr r? r?
	MoveOpStoreMem            // memcopy r? $imm
	MoveOpStoreImm            // memset r? $imm
	// uses an indirect register for the stack pointer
	MoveOpPush          // push r?
	MoveOpPushImmediate // push.imm $imm
	MoveOpPop           // pop r?
	MoveOpCount
)

const (
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
	JumpOpCount
)

const (
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
	CompareOpCount
)

const (
	MiscOpSystemCall = iota
	MiscOpCount
)

func init() {
	for _, value := range registeredGroups {
		if value.Count > value.Max {
			panic(fmt.Sprintf("Too many %s defined, %d allowed but %d defined!", value.Name, value.Max, value.Count))
		}
	}
}

type Instruction Dword

func (this Instruction) Group() byte {
	return byte(((this & 0x000000FF) & 0x7))
}
func (this Instruction) Op() byte {
	return byte(((this & 0x000000FF) & 0xF8) >> 3)
}

func (this Instruction) Byte(index int) (byte, error) {
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
func (this Instruction) Source1() byte {
	if value, err := this.Byte(3); err != nil {
		panic(err) // we should never ever get here!
	} else {
		return value
	}
}
func (this Instruction) Source0() byte {
	if value, err := this.Byte(2); err != nil {
		panic(err) // we should never ever get here!
	} else {
		return value
	}
}
func (this Instruction) Destination() byte {
	if value, err := this.Byte(1); err != nil {
		panic(err) // we should never ever get here!
	} else {
		return value
	}
}
func (this Instruction) Immediate() Word {
	return (Word(this.Source1()) << 8) | Word(this.Source0())
}

func (this *Instruction) SetGroup(group byte) {
	*this = ((*this &^ 0x7) | Instruction(group))
}
func (this *Instruction) SetOp(op byte) {
	*this = ((*this &^ 0xF8) | (Instruction(op) << 3))
}
func (this *Instruction) SetByte(index int, value byte) error {
	switch index {
	case 1:
		*this = ((*this &^ 0x0000FF00) | (Instruction(value) << 8))
	case 2:
		*this = ((*this &^ 0x00FF0000) | (Instruction(value) << 16))
	case 3:
		*this = ((*this &^ 0xFF000000) | (Instruction(value) << 24))
	default:
		return fmt.Errorf("Provided index %d is out of range!", index)
	}
	return nil
}

func (this *Instruction) SetDestination(value byte) {
	this.SetByte(1, value)
}
func (this *Instruction) SetSource0(value byte) {
	this.SetByte(2, value)
}
func (this *Instruction) SetSource1(value byte) {
	this.SetByte(3, value)
}

func NewInstruction(group, op, dest, src0, src1 byte) *Instruction {
	var i Instruction
	i.SetOp(op)
	i.SetGroup(group)
	i.SetDestination(dest)
	i.SetSource0(src0)
	i.SetSource1(src1)
	return &i
}

type Core struct {
	Gpr                           [RegisterCount]Word
	Code                          [MemorySize]Instruction
	Data, Stack                   [MemorySize]Word
	Pc                            Word
	AdvancePc, TerminateExecution bool
}
type executionUnit func(*Core, *Instruction) error

var dispatchTable = map[byte]executionUnit{
	GroupArithmetic: (*Core).arithmetic,
	GroupMove:       (*Core).move,
	//GroupJump:       jump,
	//GroupCompare:    compare,
	//GroupMisc:       misc,
}

func (this *Core) Dispatch(value *Instruction) error {
	if fn, ok := dispatchTable[value.Group()]; !ok {
		return fmt.Errorf("Instruction group %d isn't used!", value.Group())
	} else {
		this.AdvancePc = true
		return fn(this, value)
	}
}

// arithmetic operations
func legalDenominator(denominator Word) error {
	if denominator == 0 {
		return fmt.Errorf("denominator is zero!")
	} else {
		return nil
	}
}
func div(numerator, denominator Word) (Word, error) {
	if err := legalDenominator(denominator); err != nil {
		return 0, fmt.Errorf("Attempted to divide by zero!")
	} else {
		return numerator / denominator, nil
	}
}

func rem(numerator, denominator Word) (Word, error) {
	if err := legalDenominator(denominator); err != nil {
		return 0, fmt.Errorf("Attempted to divide by zero and take the remainder!")
	} else {
		return numerator % denominator, nil
	}
}
func add(a, b Word) (Word, error) {
	return a + b, nil
}
func sub(a, b Word) (Word, error) {
	return a - b, nil
}
func mul(a, b Word) (Word, error) {
	return a * b, nil
}
func shiftleft(a, b Word) (Word, error) {
	return a << b, nil
}
func shiftright(a, b Word) (Word, error) {
	return a >> b, nil
}

type arithmeticUnit func(Word, Word) (Word, error)

func reg_reg_ArithOp(fn arithmeticUnit) executionUnit {
	return func(c *Core, inst *Instruction) error {
		if out, err := fn(c.Gpr[inst.Source0()], c.Gpr[inst.Source1()]); err != nil {
			return err
		} else {
			c.Gpr[inst.Destination()] = out
			return nil
		}
	}
}
func reg_imm_ArithOp(fn arithmeticUnit) executionUnit {
	return func(c *Core, inst *Instruction) error {
		if out, err := fn(c.Gpr[inst.Source0()], Word(inst.Source1())); err != nil {
			return err
		} else {
			c.Gpr[inst.Destination()] = out
			return nil
		}
	}
}

var arithmeticDispatch = map[byte]executionUnit{
	ArithmeticOpAdd:                 reg_reg_ArithOp(add),
	ArithmeticOpSub:                 reg_reg_ArithOp(sub),
	ArithmeticOpMul:                 reg_reg_ArithOp(mul),
	ArithmeticOpDiv:                 reg_reg_ArithOp(div),
	ArithmeticOpRem:                 reg_reg_ArithOp(rem),
	ArithmeticOpShiftLeft:           reg_reg_ArithOp(shiftleft),
	ArithmeticOpShiftRight:          reg_reg_ArithOp(shiftright),
	ArithmeticOpBinaryAnd:           reg_reg_ArithOp(func(a, b Word) (Word, error) { return a & b, nil }),
	ArithmeticOpBinaryOr:            reg_reg_ArithOp(func(a, b Word) (Word, error) { return a | b, nil }),
	ArithmeticOpBinaryNot:           reg_reg_ArithOp(func(a, _ Word) (Word, error) { return ^a, nil }),
	ArithmeticOpBinaryXor:           reg_reg_ArithOp(func(a, b Word) (Word, error) { return a ^ b, nil }),
	ArithmeticOpAddImmediate:        reg_imm_ArithOp(add),
	ArithmeticOpSubImmediate:        reg_imm_ArithOp(sub),
	ArithmeticOpMulImmediate:        reg_imm_ArithOp(mul),
	ArithmeticOpDivImmediate:        reg_imm_ArithOp(div),
	ArithmeticOpRemImmediate:        reg_imm_ArithOp(rem),
	ArithmeticOpShiftLeftImmediate:  reg_imm_ArithOp(shiftleft),
	ArithmeticOpShiftRightImmediate: reg_imm_ArithOp(shiftright),
}

func (this *Core) arithmetic(value *Instruction) error {
	if op, ok := arithmeticDispatch[value.Op()]; !ok {
		return fmt.Errorf("Illegal arithmetic operation %d", value.Op())
	} else {
		return op(this, value)
	}
}

// move operations
type swapOpGet func(*Core, byte) Word
type swapOpSet func(*Core, byte, Word)

func (this *Core) registerValue(index byte) Word {
	return this.Gpr[index]
}
func (this *Core) registerAddress(index byte) Word {
	return this.Data[this.Gpr[index]]
}
func (this *Core) setRegisterValue(index byte, value Word) {
	this.Gpr[index] = value
}
func (this *Core) setRegisterAddress(index byte, value Word) {
	this.Data[this.Gpr[index]] = value
}
func (this *Core) setImmediateAddress(imm, value Word) {
	this.Data[imm] = value
}

func (this *Core) moveop_move(value *Instruction) error {
	this.Gpr[value.Destination()] = this.Gpr[value.Source0()]
	return nil
}
func (this *Core) moveop_set(value *Instruction) error {
	this.Gpr[value.Destination()] = value.Immediate()
	return nil
}
func (this *Core) moveop_swap_base(value *Instruction, destGet, src0Get swapOpGet, destSet, src0Set swapOpSet) error {
	dest := value.Destination()
	src0 := value.Source0()
	dVal := destGet(this, dest)
	sVal := src0Get(this, src0)
	destSet(this, dest, sVal)
	src0Set(this, src0, dVal)
	return nil
}
func (this *Core) moveop_swap(value *Instruction) error {
	return this.moveop_swap_base(value, (*Core).registerValue, (*Core).registerValue, (*Core).setRegisterValue, (*Core).setRegisterValue)
}
func (this *Core) moveop_swap_reg_addr(value *Instruction) error {
	return this.moveop_swap_base(value, (*Core).registerValue, (*Core).registerAddress, (*Core).setRegisterValue, (*Core).setRegisterAddress)
}
func (this *Core) moveop_swap_addr_addr(value *Instruction) error {
	return this.moveop_swap_base(value, (*Core).registerAddress, (*Core).registerAddress, (*Core).setRegisterAddress, (*Core).setRegisterAddress)
}

func (this *Core) moveop_swap_reg_mem(value *Instruction) error {
	dest := value.Destination()
	imm := value.Immediate()
	dVal := this.registerValue(dest)
	aVal := this.Data[imm]
	this.setRegisterValue(dest, aVal)
	this.setImmediateAddress(imm, dVal)
	return nil
}
func (this *Core) moveop_swap_addr_mem(value *Instruction) error {
	dest := value.Destination()
	imm := value.Immediate()
	dVal := this.registerAddress(dest)
	aVal := this.Data[imm]
	this.setRegisterAddress(dest, aVal)
	this.setImmediateAddress(imm, dVal)
	return nil
}
func (this *Core) moveop_load_base(dest byte, value Word) error {
	this.setRegisterValue(dest, value)
	return nil
}
func (this *Core) moveop_load(value *Instruction) error {
	return this.moveop_load_base(value.Destination(), this.registerAddress(value.Source0()))
}
func (this *Core) moveop_load_mem(value *Instruction) error {
	return this.moveop_load_base(value.Destination(), this.Data[value.Immediate()])
}
func (this *Core) moveop_store_base(dest byte, value Word) error {
	this.setRegisterAddress(dest, value)
	return nil
}
func (this *Core) moveop_store(value *Instruction) error {
	return this.moveop_store_base(value.Destination(), this.registerValue(value.Source0()))
}
func (this *Core) moveop_store_imm(value *Instruction) error {
	return this.moveop_store_base(value.Destination(), value.Immediate())
}
func (this *Core) moveop_store_mem(value *Instruction) error {
	return this.moveop_store_base(value.Destination(), this.Data[value.Immediate()])
}
func (this *Core) moveop_store_addr(value *Instruction) error {
	return this.moveop_store_base(value.Destination(), this.registerAddress(value.Source0()))
}
func (this *Core) PushOntoStack(value Word) {
	index := this.Gpr[StackPointerRegisterIndex]
	index++
	this.Stack[index] = value
	this.Gpr[StackPointerRegisterIndex] = index
}
func (this *Core) PopOffStack() Word {
	index := this.Gpr[StackPointerRegisterIndex]
	value := this.Stack[index]
	index--
	this.Gpr[StackPointerRegisterIndex] = index
	return value
}
func (this *Core) moveop_push(value *Instruction) error {
	this.PushOntoStack(this.registerValue(value.Destination()))
	return nil
}
func (this *Core) moveop_push_imm(value *Instruction) error {
	this.PushOntoStack(value.Immediate())
	return nil
}
func (this *Core) moveop_pop(value *Instruction) error {
	this.setRegisterValue(value.Destination(), this.PopOffStack())
	return nil
}

// move operations
var moveDispatch = map[byte]executionUnit{
	MoveOpMove:         (*Core).moveop_move,           // move r? r?
	MoveOpSwap:         (*Core).moveop_swap,           // swap r? r?
	MoveOpSwapRegAddr:  (*Core).moveop_swap_reg_addr,  // swap.reg.addr r? r?
	MoveOpSwapAddrAddr: (*Core).moveop_swap_addr_addr, // swap.addr.addr r? r?
	MoveOpSwapRegMem:   (*Core).moveop_swap_reg_mem,   // swap.reg.mem r? $imm
	MoveOpSwapAddrMem:  (*Core).moveop_swap_addr_mem,  // swap.addr.mem r? $imm
	MoveOpSet:          (*Core).moveop_set,            // set r? $imm
	MoveOpLoad:         (*Core).moveop_load,           // load r? r?
	MoveOpLoadMem:      (*Core).moveop_load_mem,       // load.mem r? $imm
	MoveOpStore:        (*Core).moveop_store,          // store r? r?
	MoveOpStoreAddr:    (*Core).moveop_store_addr,     // store.addr r? r?
	MoveOpStoreMem:     (*Core).moveop_store_mem,      // memcopy r? $imm
	MoveOpStoreImm:     (*Core).moveop_store_imm,      // memset r? $imm
	// uses an indirect register for the stack pointer
	MoveOpPush:          (*Core).moveop_push,     // push r?
	MoveOpPushImmediate: (*Core).moveop_push_imm, // push.imm $imm
	MoveOpPop:           (*Core).moveop_pop,      // pop r?
}

func (this *Core) move(value *Instruction) error {
	if op, ok := moveDispatch[value.Op()]; !ok {
		return fmt.Errorf("Illegal move operation %d", value.Op())
	} else {
		return op(this, value)
	}
}
