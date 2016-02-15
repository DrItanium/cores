// instruction decoder module
package iris2

import (
	"fmt"
	"github.com/DrItanium/cores/iris16"
)

type DecoderUnit struct {
	core    *iris16.Core
	err     chan error
	out     chan *DecodedInstruction
	in      chan Word
	Error   <-chan error
	Input   chan<- Word
	Control <-chan Word
	Result  <-chan *DecodedInstruction
}

func NewDecoderUnit(control <-chan Word) (*DecoderUnit, error) {
	var dc DecoderUnit
	if core, err := iris16.New(); err != nil {
		return nil, err
	} else {
		dc.core = core
		// need to install the microcode somehow
	}

	dc.err = make(chan error)
	dc.out = make(chan *DecodedInstruction)
	dc.in = make(chan Word)
	dc.Error = dc.err
	dc.Input = dc.in
	dc.Result = dc.out
	dc.Control = control
	return &dc
}

func (this *DecoderUnit) Startup() error {
	if this.running {
		return fmt.Errorf("Can't startup an already running decoder unit")
	} else {
		this.running = true
		go this.body()
		go this.controlQuery()
		return nil
	}
}
func (this *DecoderUnit) controlQuery() {
	<-this.Control
	if err := this.shutdown(); err != nil {
		this.err <- err
	}
}
func (this *DecoderUnit) body() {
	for this.running {
		select {
		case raw, more := <-this.in:
			if more {
				// an instruction in iris2 is 64-bits long but
				// it can either be a packet or a single long
				// instruction, this is determined by the group
				// bits at the front of the instruction

			}
		}
	}
}
func (this *DecoderUnit) shutdown() error {
	if !this.running {
		return fmt.Errorf("Can't shutdown an already shutdown decoder unit")
	} else {
		this.running = false
		close(this.in)
		return nil
	}
}

type Instruction uint32

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

type DecodedInstruction struct {
	Group, Op                     byte
	Destination, Source0, Source1 byte
	Immediate                     Word
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
