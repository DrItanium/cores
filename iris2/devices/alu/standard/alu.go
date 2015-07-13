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
	nop = iota
	add
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

type aluOperation func(*Alu, *instruction) ([]byte, error)

func (fn aluOperation) Invoke(alu *Alu, inst *instruction) ([]byte, error) {
	return fn(alu, inst)
}

func (this *Alu) opset(inst *instruction) ([]byte, error) {
	if !inst.hasArgs() {
		return nil, fmt.Errorf("No arguments provided for the alu set operation!")
	}
	if len(inst.args) < 2 {
		return nil, fmt.Errorf("Too few arguments provided to the alu set operation!")
	} else if len(inst.args) > 2 {
		return nil, fmt.Errorf("Too many arguments provided to the alu set operation!")
	}
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
	default:
		return nil, fmt.Errorf("Got an illegal alu tag type during set: %d", reg.dataType)
	}
	return []byte{}, nil
}
func (this *Alu) opget(inst *instruction) ([]byte, error) {
	if !inst.hasArgs() {
		return nil, fmt.Errorf("No arguments provided for the alu get operation!")
	}
	if inst.args[0].immediate {
		return nil, fmt.Errorf("Getting an immediate does not make sense!")
	}

	if result, err1 := this.regs[inst.args[0].register].slice(); err1 != nil {
		return nil, err1
	} else {
		return result, nil
	}
}
func (this *Alu) opnop(inst *instruction) ([]byte, error) {
	if inst.hasArgs() {
		return nil, fmt.Errorf("alu nop should have no arguments!")
	} else {
		return []byte{}, nil
	}
}

//func (this *Alu) opadd(inst *instruction) ([]byte,

var lookupTable = []aluOperation{
	(*Alu).opnop,
	(*Alu).opnop,
	(*Alu).opnop,
	(*Alu).opnop,
	(*Alu).opnop,
	(*Alu).opnop,
	(*Alu).opnop,
	(*Alu).opnop,
	(*Alu).opnop,
	(*Alu).opnop,
	(*Alu).opnop,
	(*Alu).opnop,
	(*Alu).opset,
	(*Alu).opget,
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
					out.Value, out.Error = lookupTable[inst.op].Invoke(this, inst)
				}
			}
		}
		this.output <- out
	}
}
