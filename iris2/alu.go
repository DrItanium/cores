// declaration of the basic alu
package standard

import (
	"fmt"
	"github.com/DrItanium/cores/iris2"
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
)

type aluValue struct {
	dataType byte
	data     iris2.Word
}
type Alu struct {
	input            chan iris2.Packet
	output           chan iris2.Packet
	terminate        bool
	internalRegister aluValue
}

func New() *Alu {
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
	Nop = iota
	Add
	Sub
	Mul
	Div
	Mod
	Shl
	Shr
	And
	Or
	Not
	Xor

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

func aluAdd(x, y, ret *aluValue) error {
	return nil
}

var aluOperations = []aluOperation{
	func(_, _, _ *aluValue) error { return nil }, // nop
	aluAdd,
	aluSub,
	aluMul,
	aluDiv,
	aluMod,
	aluShl,
	aluShr,
	aluAdd,
	aluOr,
}

//func aluAdd8

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
