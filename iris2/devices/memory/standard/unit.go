// definition of a memory controller
package standard

import (
	"encoding/binary"
	"fmt"
	"github.com/DrItanium/cores/iris2"
)

const (
	iPMask         = 0xFFFFFFFFF000000F // last four bits offset for instruction pointer
	iPMaskEnd      = 0x000000000000000F
	memoryMask     = 0xFFFFFFFFF0000000
	memoryCapacity = 0x000000000FFFFFFF

	stackMask = 0xFFFFFFFFFF000000
	stackSize = ^stackMask
)

func asInstructionPointerAddress(this iris2.Word) iris2.Word {
	return this &^ iPMask
}

func asMemoryAddress(this iris2.Word) iris2.Word {
	return this &^ memoryMask
}
func asStackAddress(this iris2.Word) iris2.Word {
	return this &^ stackMask
}

type MemoryController struct {
	input      chan iris2.Packet
	output     chan iris2.Packet
	terminated bool
	memory     [memoryCapacity]byte
}

func (this *MemoryController) Capacity() iris2.Word {
	return iris2.Word(len(this.memory))
}
func NewMemoryController() (*MemoryController, error) {
	var mc MemoryController
	mc.output = make(chan iris2.Packet)
	mc.input = make(chan iris2.Packet)
	mc.terminated = false
	go mc.parseInput()
	return &mc, nil
}

func (this *MemoryController) Terminate() {
	if !this.terminated {
		close(this.input)
		close(this.output)
		this.terminated = true
	}
}

const (
	memoryGroupMajorMask = 0x03
	memoryGroupMinorMask = 0x0C

	memoryMajorNop = iota
	memoryMajorLoad
	memoryMajorStore
	memoryMajorGetInstructionPacket

	memoryMinorByte = iota
	memoryMinorInt16
	memoryMinorInt32
	memoryMinorInt64
)

func groupMajor(this byte) byte {
	return byte(this & memoryGroupMajorMask)
}
func groupMinor(this byte) byte {
	return byte((this & memoryGroupMinorMask) >> 2)
}
func (this *MemoryController) loadInstruction(address iris2.Word) []byte {
	return this.rawLoad(asInstructionPointerAddress(address), iPMaskEnd)
}
func (this *MemoryController) loadData(address, count iris2.Word) []byte {
	return this.rawLoad(asMemoryAddress(address), count)
}
func (this *MemoryController) rawLoad(address, count iris2.Word) []byte {
	return this.memory[address : address+count]
}

func (this *MemoryController) storeValue(address iris2.Word, data []byte) {
	for ind, value := range data {
		this.memory[address+iris2.Word(ind)] = value
	}
}

func toWord(this []byte) iris2.Word {
	return iris2.Word(binary.LittleEndian.Uint64(this))
}

var sizeTranslationTable = []int{1, 2, 4, 8}

func (this *MemoryController) parseInput() {
	for !this.terminated {
		p := <-this.input
		var outPacket iris2.Packet
		outPacket.Error = nil
		if !p.HasData() {
			outPacket.Error = fmt.Errorf("Memory controller: provided command input stream is empty!")
		} else {
			streamLen := len(p.Value)
			cell := p.Value[0]
			if streamLen == 1 {
				switch groupMajor(cell) {
				case memoryMajorNop:
				case memoryMajorLoad:
					outPacket.Error = fmt.Errorf("Requested a memory address load with no address to load from")
				case memoryMajorStore:
					outPacket.Error = fmt.Errorf("Requested a memory address store with no address to store from nor the value to save")
				case memoryMajorGetInstructionPacket:
					outPacket.Error = fmt.Errorf("Custom operations are not supported!")
				}
			} else {
				args := p.Value[1:]
				if len(args) < 8 {
					outPacket.Error = fmt.Errorf("args is less than 8 bytes in length!")
				} else {
					address := toWord(args[:8])
					rest := args[8:]

					switch len(rest) {
					case 0:
						switch groupMajor(cell) {
						case memoryMajorLoad:
							outPacket.Value = this.loadData(address, iris2.Word(sizeTranslationTable[groupMinor(cell)]))
						case memoryMajorGetInstructionPacket:
							outPacket.Value = this.loadInstruction(address)
						default:
							outPacket.Error = fmt.Errorf("Only memory controller load instructions can be 9 bytes long")
						}
					case 1, 2, 4, 8:
						switch groupMajor(cell) {
						case memoryMajorStore:
							switch groupMinor(cell) {
							case memoryMinorByte, memoryMinorInt16, memoryMinorInt32, memoryMinorInt64:
								this.storeValue(address, rest)
							default:
								outPacket.Error = fmt.Errorf("Illegal minor operation %d", groupMinor(cell))
							}
						default:
							outPacket.Error = fmt.Errorf("Illegal operation given the instruction length major (%d)", groupMajor(cell))
						}
					default:
						outPacket.Error = fmt.Errorf("Command stream is of invalid length %d", len(rest))
					}
				}
			}
		}
		this.output <- outPacket
	}
}

func (this *MemoryController) Send(value []byte) chan iris2.Packet {
	var p iris2.Packet
	p.Error = nil
	p.Value = value
	this.input <- p
	return this.output
}
