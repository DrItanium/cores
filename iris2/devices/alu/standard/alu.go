// declaration of the basic alu
package standard

import (
	"encoding/binary"
	"fmt"
	"github.com/DrItanium/cores"
	"github.com/DrItanium/cores/iris2"
	"github.com/DrItanium/cores/manip"
)

const (
	tagByte = iota
	tagInt16
	tagInt32
	tagInt64
	tagEnd

	registerCount = 8
)

func init() {
	if tagEnd > 4 {
		panic(fmt.Errorf("ALU INTERNAL ERROR: %d tags are defined when the max is 4", tagEnd))
	}
}

type aluValue struct {
	dataType byte
	data     iris2.Word
}

func (this *aluValue) slice() ([]byte, error) {
	var out []byte
	switch this.dataType {
	case tagByte:
		out = make([]byte, 1)
		out[0] = byte(this.data)
	case tagInt16:
		out = make([]byte, 2)
		binary.LittleEndian.PutUint16(out, uint16(this.data))
	case tagInt32:
		out = make([]byte, 4)
		binary.LittleEndian.PutUint32(out, uint32(this.data))
	case tagInt64:
		out = make([]byte, 8)
		binary.LittleEndian.PutUint64(out, uint64(this.data))
	default:
		return nil, fmt.Errorf("Illegal data type code in alu register!")
	}
	return out, nil
}

type Alu struct {
	input     chan cores.Packet
	output    chan cores.Packet
	terminate bool
	regs      [registerCount]aluValue
}

func (this *Alu) setRegister(index int, value iris2.Word) error {
	if index >= registerCount {
		return fmt.Errorf("Register at index %d is not a real register (alu has %d registers)!", index, registerCount)
	} else {
		switch this.regs[index].dataType {
		case tagByte:
			this.regs[index].data = iris2.Word(byte(value))
		case tagInt16:
			this.regs[index].data = iris2.Word(int16(value))
		case tagInt32:
			this.regs[index].data = iris2.Word(int32(value))
		case tagInt64:
			this.regs[index].data = value
		}
		return nil
	}
}
func (this *Alu) getRegister(index int) ([]byte, error) {
	if index >= registerCount {
		return nil, fmt.Errorf("Register at index %d is not a real register (alu has %d registers)!", index, registerCount)
	} else {
		return this.regs[index].slice()
	}
}

type argument struct {
	dataType  byte
	unsigned  bool
	immediate bool
	register  byte
	data      []byte
}

var argSizeTable = []byte{1, 2, 4, 8}

func newArgument(value []byte) (*argument, int, error) {
	if len(value) == 0 {
		return nil, 0, fmt.Errorf("No bytes to parse for argument!")
	}
	count := 1
	first := value[0]
	a := argument{
		dataType:  byte(manip.Mask8(first, 0x03, 0)),
		unsigned:  manip.BitsSet8(first, 0x04, 2),
		immediate: manip.BitsSet8(first, 0x08, 3),
		register:  byte(manip.Mask8(first, 0xE0, 5)),
	}

	if a.immediate {
		if int(a.dataType) < len(argSizeTable) {
			return nil, 0, fmt.Errorf("dataType set as %d", a.dataType)
		}
		num := argSizeTable[a.dataType]
		a.data = make([]byte, num)
		copy(a.data, value[1:1+num])
		count += int(num)
	}

	return &a, count, nil
}

type instruction struct {
	op   byte
	args []*argument
}

func (this *instruction) hasArgs() bool {
	return len(this.args) != 0
}
func newInstruction(value []byte) (*instruction, error) {
	if len(value) == 0 {
		return nil, fmt.Errorf("No bytes to parse!")
	} else {
		var i instruction
		i.op = value[0]
		rest := value[1:]
		for len(rest) > 0 {
			arg, count, err := newArgument(rest)
			if err != nil {
				return nil, err
			}
			i.args = append(i.args, arg)
			rest = rest[count:]
		}
		return &i, nil
	}
}

func New() *Alu {
	var a Alu
	a.output = make(chan cores.Packet)
	a.input = make(chan cores.Packet)
	go a.parseInput()
	return &a
}

func (this *Alu) Terminate() {
	if !this.terminate {
		close(this.input)
		close(this.output)
		this.terminate = true
	}
}

const (
	add = iota
	sub
	mul
	div
	mod
	shl
	shr
	and
	or
	not
	xor
	nop
	set
	get
	lastAluOp
)

func init() {
	// make sure that we aren't going to barf on too many operations!
	if lastAluOp > 256 {
		panic(fmt.Errorf("%d alu operations defined, max is 256", lastAluOp))
	}
}

func (this *Alu) opset(inst *instruction) ([]byte, error) {
	first := inst.args[0]
	if first.immediate {
		return nil, fmt.Errorf("The first argument provided to the alu set operation must be a register!")
	}
	second := inst.args[1]
	if !second.immediate {
		return nil, fmt.Errorf("The second argument provided to the alu set operation must be an immediate!")
	}
	reg := &(this.regs[first.register])
	reg.dataType = second.dataType
	switch reg.dataType {
	case tagByte:
		reg.data = iris2.Word(second.data[0])
	case tagInt16:
		reg.data = iris2.Word(binary.LittleEndian.Uint16(second.data))
	case tagInt32:
		reg.data = iris2.Word(binary.LittleEndian.Uint32(second.data))
	case tagInt64:
		reg.data = iris2.Word(binary.LittleEndian.Uint64(second.data))
	}
	return []byte{}, nil
}
func (this *Alu) opget(inst *instruction) ([]byte, error) {
	first := inst.args[0]
	if first.immediate {
		return nil, fmt.Errorf("Getting an immediate does not make sense!")
	}

	if result, err1 := this.getRegister(int(first.register)); err1 != nil {
		return nil, err1
	} else {
		return result, nil
	}
}
func (this *Alu) opnop(inst *instruction) ([]byte, error) {
	return []byte{}, nil
}

type aluOperation struct {
	name                string
	requiresArguments   bool
	numberOfArguments   int
	customOperation     bool
	secondArgCantBeZero bool
	fn                  func(*Alu, *instruction) ([]byte, error)
	binaryOp            binaryOperation
}

func (this *aluOperation) invoke(alu *Alu, inst *instruction) ([]byte, error) {
	if this.requiresArguments {
		if !inst.hasArgs() {
			return nil, fmt.Errorf("Alu op %s requires arguments but none are given!", this.name)
		} else if len(inst.args) != this.numberOfArguments {
			return nil, fmt.Errorf("Alu op %s requires exactly %d arguments but %d are given!", this.name, this.numberOfArguments, len(inst.args))
		}
	} else {
		if inst.hasArgs() {
			return nil, fmt.Errorf("Alu op %s does not accept arguments but %d are given!", this.name, len(inst.args))
		}
	}
	if this.customOperation {
		return this.fn(alu, inst)
	} else {
		if this.numberOfArguments == 3 {
			destination, src0, src1 := inst.args[0], inst.args[1], inst.args[2]
			if destination.immediate {
				return nil, fmt.Errorf("Destination is tagged as an immediate!")
			} else if src0.immediate {
				return nil, fmt.Errorf("Source0 is tagged as an immediate!")
			} else if src1.immediate {
				return nil, fmt.Errorf("Source1 is tagged as an immediate!")
			} else if destination.register >= registerCount {
				return nil, fmt.Errorf("destination's associated register index (%d) is greater than %d", destination.register, registerCount)
			} else if src0.register >= registerCount {
				return nil, fmt.Errorf("src0's associated register index (%d) is greater than %d", src0.register, registerCount)
			} else if src1.register >= registerCount {
				return nil, fmt.Errorf("src1's associated register index (%d) is greater than %d", src1.register, registerCount)
			} else {
				alu.regs[destination.register].dataType = destination.dataType
				if d, s0, s1 := int(destination.register), alu.regs[src0.register].data, alu.regs[src1.register].data; this.secondArgCantBeZero && s1 == 0 {
					return nil, fmt.Errorf("Attempted to divide by zero, op is %s", this.name)
				} else if err0 := alu.setRegister(d, this.binaryOp(s0, s1)); err0 != nil {
					return nil, err0
				} else {
					return []byte{}, nil
				}
			}
		} else {
			return nil, fmt.Errorf("non three argument operations must have a custom implementation!")
		}
	}
}

func (this *Alu) unpack2(inst *instruction) (int, iris2.Word, error) {
	destination, src0 := inst.args[0], inst.args[1]
	if destination.immediate {
		return 0, 0, fmt.Errorf("Destination is tagged as an immediate!")
	} else if src0.immediate {
		return 0, 0, fmt.Errorf("Source0 is tagged as an immediate!")
	} else if destination.register >= registerCount {
		return 0, 0, fmt.Errorf("destination's associated register index (%d) is greater than %d", destination.register, registerCount)
	} else if src0.register >= registerCount {
		return 0, 0, fmt.Errorf("src0's associated register index (%d) is greater than %d", src0.register, registerCount)
	} else {
		this.regs[destination.register].dataType = destination.dataType
		return int(destination.register), this.regs[src0.register].data, nil
	}

}

type binaryOperation func(iris2.Word, iris2.Word) iris2.Word

func (fn binaryOperation) invoke(x, y iris2.Word) iris2.Word {
	return fn(x, y)
}

func (this *Alu) opnot(inst *instruction) ([]byte, error) {
	if destination, src0, err := this.unpack2(inst); err != nil {
		return nil, err
	} else {
		if err0 := this.setRegister(destination, ^src0); err0 != nil {
			return nil, err0
		} else {
			return []byte{}, nil
		}
	}
}

var lookupTable = map[int]*aluOperation{
	add: &aluOperation{"add", true, 3, false, false, nil, func(x, y iris2.Word) iris2.Word { return x + y }},
	sub: &aluOperation{"sub", true, 3, false, false, nil, func(x, y iris2.Word) iris2.Word { return x - y }},
	mul: &aluOperation{"mul", true, 3, false, false, nil, func(x, y iris2.Word) iris2.Word { return x * y }},
	shl: &aluOperation{"shl", true, 3, false, false, nil, func(x, y iris2.Word) iris2.Word { return x << y }},
	shr: &aluOperation{"shr", true, 3, false, false, nil, func(x, y iris2.Word) iris2.Word { return x >> y }},
	and: &aluOperation{"and", true, 3, false, false, nil, func(x, y iris2.Word) iris2.Word { return x & y }},
	or:  &aluOperation{"or", true, 3, false, false, nil, func(x, y iris2.Word) iris2.Word { return x | y }},
	xor: &aluOperation{"xor", true, 3, false, false, nil, func(x, y iris2.Word) iris2.Word { return x ^ y }},
	div: &aluOperation{"div", true, 3, false, true, nil, func(x, y iris2.Word) iris2.Word { return x / y }},
	mod: &aluOperation{"mod", true, 3, false, true, nil, func(x, y iris2.Word) iris2.Word { return x % y }},
	nop: &aluOperation{"nop", false, 0, true, false, (*Alu).opnop, nil},
	not: &aluOperation{"not", true, 2, true, false, (*Alu).opnot, nil},
	set: &aluOperation{"set", true, 1, true, false, (*Alu).opset, nil},
	get: &aluOperation{"get", true, 2, true, false, (*Alu).opget, nil},
}

func init() {
	if len(lookupTable) != lastAluOp {
		panic(fmt.Errorf("Number of operations in dispatch table (%d) does not match number of defined operations (%d)", len(lookupTable), lastAluOp))
	}
}

func (this *Alu) parseInput() {
	for !this.terminate {
		var out cores.Packet
		input := <-this.input
		if !input.HasData() {
			out.Error = fmt.Errorf("alu: Command stream is empty")
		} else {
			if inst, err := newInstruction(input.Value); err != nil {
				out.Error = err
			} else {
				if inst.op >= lastAluOp {
					out.Error = fmt.Errorf("Illegal operation %d", inst.op)
				} else {
					out.Value, out.Error = lookupTable[int(inst.op)].invoke(this, inst)
				}
			}
		}
		this.output <- out
	}
}

func (this *Alu) Send(value []byte) chan cores.Packet {
	var a cores.Packet
	a.Value = value
	this.input <- a
	return this.output
}
