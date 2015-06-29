// classic iris1 core
package classic

import (
	"github.com/DrItanium/iris1"
)

type Backend struct {
	gpr   [iris1.RegisterCount - iris1.UserRegisterBegin]iris1.Word
	code  [iris1.MemorySize]iris1.Instruction
	data  [iris1.MemorySize]iris1.Word
	stack [iris1.MemorySize]iris1.Word
	// internal registers that should be easy to find
	instructionPointer iris1.Word
	stackPointer       iris1.Word
	link               iris1.Word
	count              iris1.Word
	predicate          iris1.Word
}

func (this *Backend) SetRegister(index byte, value iris1.Word) error {
	switch index {
	case iris1.FalseRegister:
		return iris1.NewError(iris1.ErrorWriteToFalseRegister, uint(value))
	case iris1.TrueRegister:
		return iris1.NewError(iris1.ErrorWriteToTrueRegister, uint(value))
	case iris1.InstructionPointer:
		this.instructionPointer = value
	case iris1.StackPointer:
		this.stackPointer = value
	case iris1.PredicateRegister:
		this.predicate = value
	case iris1.CountRegister:
		this.count = value
	case iris1.LinkRegister:
		this.link = value
	default:
		this.gpr[index-iris1.UserRegisterBegin] = value
	}
	return nil
}
func (this Backend) GetRegister(index byte) iris1.Word {
	switch index {
	case iris1.FalseRegister:
		return 0
	case iris1.TrueRegister:
		return 1
	case iris1.InstructionPointer:
		return this.instructionPointer
	case iris1.StackPointer:
		return this.stackPointer
	case iris1.PredicateRegister:
		return this.predicate
	case iris1.LinkRegister:
		return this.link
	case iris1.CountRegister:
		return this.count
	default:
		// do the offset calculation
		return this.gpr[index-iris1.UserRegisterBegin]
	}
}

func (this Backend) CodeMemory(address iris1.Word) iris1.Instruction {
	return this.code[address]
}
func (this *Backend) SetCodeMemory(address iris1.Word, value iris1.Instruction) error {
	this.code[address] = value
	return nil
}
func (this *Backend) Push(value iris1.Word) {
	this.stackPointer++
	this.stack[this.stackPointer] = value
}
func (this Backend) Peek() iris1.Word {
	return this.stack[this.stackPointer]
}
func (this *Backend) Pop() iris1.Word {
	value := this.stack[this.stackPointer]
	this.stackPointer--
	return value
}
func (this Backend) DataMemory(address iris1.Word) iris1.Word {
	return this.data[address]
}
func (this *Backend) SetDataMemory(address, value iris1.Word) error {
	this.data[address] = value
	return nil
}
