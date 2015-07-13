// declaration of the basic alu
package standard

import (
	"fmt"
	"github.com/DrItanium/cores"
	"github.com/DrItanium/cores/iris2"
)

const (
	tagInternal = iota
	tagByte
	tagSbyte
	tagInt16
	tagUint16
	tagInt32
	tagUint32
	tagInt64
	tagUint64

	registerCount = 8
)

type aluValue struct {
	dataType byte
	data     iris2.Word
}
type Alu struct {
	input            chan cores.Packet
	output           chan cores.Packet
	terminate        bool
	internalRegister [registerCount]aluValue
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
		dataType:  byte(cores.Mask(first, 0x03, 0)),
		unsigned:  cores.Mask(first, 0x04, 2) != 0,
		immediate: cores.Mask(first, 0x08, 3) != 0,
		register:  byte(cores.Mask(first, 0xE0, 5)),
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

	maskAluGroup = 0x0F
	maskAluFlags = 0xF0

	saveResultFlag = 0x1
	flag2          = 0x2
	flag3          = 0x4
	flag4          = 0x8
)

func groupMajor(value byte) byte {
	return value & maskAluGroup
}

type flags byte

func (this flags) saveResult() bool {
	return manip.BitsSet(this, saveResult, 0)
}
func (this flags) flag2() bool {
	return manip.BitsSet(this, flag2, 1)
}
func (this flags) flag3() bool {
	return manip.BitsSet(this, flag3, 2)
}
func (this flags) flag4() bool {
	return manip.BitsSet(this, flag4, 3)
}
func getAluFlags(value byte) flags {
	return flags(manip.Mask(value, maskAluFlags, 4))
}

type aluOperation func(a, b, ret *aluValue) error

//func aluAdd8

func (this *Alu) parseInput() {
	for !this.terminate {
		var out cores.Packet
		input := <-this.input
		if !input.HasData() {
			out.Error = fmt.Errorf("alu: Command stream is empty")
		} else {
			op := groupMajor(input.First())
			//flags := getAluFlags(input.First())
			switch op {
			case nop:
				// do nothing
			case add:
			case sub:
			case mul:
			case div:
			case mod:
			case shl:
			case shr:
			case and:
			case or:
			case not:
			case xor:
			default:
				out.Error = fmt.Errorf("Illegal operation %d", op)
			}
		}
		this.output <- out
	}
}
