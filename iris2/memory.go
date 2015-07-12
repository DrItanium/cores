// definition of a memory controller
package iris2

import (
	"encoding/binary"
	"fmt"
)

const (
	iPMask         = 0xFFFFFFFFF000000F // last four bits offset for instruction pointer
	iPMaskEnd      = 0x000000000000000F
	memoryMask     = 0xFFFFFFFFF0000000
	memoryCapacity = 0x000000000FFFFFFF

	stackMask = 0xFFFFFFFFFF000000
	stackSize = 0x0000000000FFFFFF
)

func (this Word) asInstructionPointerAddress() Word {
	return this &^ iPMask
}

func (this Word) asMemoryAddress() Word {
	return this &^ memoryMask
}
func (this Word) asStackAddress() Word {
	return this &^ stackMask
}

type MemoryController struct {
	input      chan Packet
	output     chan Packet
	terminated bool
	memory     [memoryCapacity]byte
}

func NewMemoryController() (*MemoryController, error) {
	var mc MemoryController
	mc.output = make(chan Packet)
	mc.input = make(chan Packet)
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
	return byte(this &^ memoryGroupMajorMask)
}
func groupMinor(this byte) byte {
	return byte((this &^ memoryGroupMinorMask) >> 2)
}
func (this *MemoryController) loadInstruction(address Word) []byte {
	return this.rawLoad(address.asInstructionPointerAddress(), iPMaskEnd)
}
func (this *MemoryController) loadData(address, count Word) []byte {
	return this.rawLoad(address.asMemoryAddress(), count)
}
func (this *MemoryController) rawLoad(address, count Word) []byte {
	return this.memory[address : address+count]
}

func (this *MemoryController) storeValue(address Word, data []byte) {
	for ind, value := range data {
		this.memory[address+Word(ind)] = value
	}
}

func mctoByte(this []byte) byte {
	return byte(this[0])
}
func mctoInt16(this []byte) uint16 {
	return binary.LittleEndian.Uint16(this)
}
func mctoInt32(this []byte) uint32 {
	return binary.LittleEndian.Uint32(this)
}
func mctoInt64(this []byte) uint64 {
	return binary.LittleEndian.Uint64(this)
}
func mctoWord(this []byte) Word {
	return Word(binary.LittleEndian.Uint64(this))
}

var sizeTranslationTable = []int{1, 2, 4, 8}

func (this *MemoryController) parseInput() {
	for !this.terminated {
		p := <-this.input
		var outPacket Packet
		outPacket.Error = nil
		streamLen := len(p.Value)
		if streamLen == 0 {
			outPacket.Error = fmt.Errorf("Memory controller: provided command input stream is empty!")
		} else {
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
					address := mctoWord(args[:8])
					rest := args[8:]

					switch len(rest) {
					case 0:
						switch groupMajor(cell) {
						case memoryMajorLoad:
							outPacket.Value = this.loadData(address, Word(sizeTranslationTable[groupMinor(cell)]))
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

func (this *MemoryController) Send(value []byte) chan Packet {
	var p Packet
	p.Error = nil
	p.Value = value
	this.input <- p
	return this.output
}
