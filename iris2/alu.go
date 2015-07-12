// declaration of the basic alu
package iris2

import (
	"fmt"
)

const (
	aluTagInternal = iota
	aluTagByte
	aluTagSbyte
	aluTagInt16
	aluTagUint16
	aluTagInt32
	aluTagUint32
	aluTagInt64
	aluTagUint64
	aluTagFloat32
	aluTagFloat64
)

type aluValue struct {
	dataType byte
	data     Word
}
type Alu struct {
	input            chan Packet
	output           chan Packet
	terminate        bool
	internalRegister aluValue
}

func NewAlu() *Alu {
	var a Alu
	a.output = make(chan Packet)
	a.input = make(chan Packet)
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
	aluOpNop = iota
	aluOpAdd
	aluOpSub
	aluOpMul
	aluOpDiv
	aluOpMod
	aluOpShl
	aluOpShr
	aluOpAnd
	aluOpOr
	aluOpNot
	aluOpXor

	maskAluGroup = 0x0F
	maskAluFlags = 0xF0

	aluSaveResultFlag = 0x1
	aluFlag2          = 0x2
	aluFlag3          = 0x4
	aluFlag4          = 0x8
)

func aluGroupMajor(value byte) byte {
	return value & maskAluGroup
}

type aluFlags byte

func (this aluFlags) saveResult() bool {
	return (this & aluSaveResultFlag) != 0
}
func (this aluFlags) flag2() bool {
	return ((this & aluFlag2) >> 1) != 0
}
func (this aluFlags) flag3() bool {
	return ((this & aluFlag3) >> 2) != 0
}
func (this aluFlags) flag4() bool {
	return ((this & aluFlag4) >> 3) != 0
}
func getAluFlags(value byte) aluFlags {
	return aluFlags((value & maskAluFlags) >> 4)
}

type aluOperation func(a, b, ret *aluValue) error

var aluOperations = []aluOperation{
	func(_, _, _ *aluValue) error { return nil }, // nop
	func(a, b, ret *aluValue) error {
		if a.dataType != b.dataType {

		}
		return nil
	}, // add
}

func (this *Alu) parseInput() {
	for !this.terminate {
		var out Packet
		input := <-this.input
		if !input.HasData() {
			out.Error = fmt.Errorf("alu: Command stream is empty")
		} else {
			op := aluGroupMajor(input.First())
			//flags := getAluFlags(input.First())
			switch op {
			case aluOpNop:
				// do nothing
			case aluOpAdd:

			case aluOpSub:
			case aluOpMul:
			case aluOpDiv:
			case aluOpMod:
			case aluOpShl:
			case aluOpShr:
			case aluOpAnd:
			case aluOpOr:
			case aluOpNot:
			case aluOpXor:
			default:
				out.Error = fmt.Errorf("Illegal operation %d", op)
			}
		}
		this.output <- out
	}
}
