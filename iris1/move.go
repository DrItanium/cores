// move related operations
package iris1

import "fmt"

const (
	// Move Operations
	MoveOpMove = iota
	MoveOpSwap
	MoveOpSet
	MoveOpLoad
	MoveOpStore
	MoveOpPush
	MoveOpPop
	MoveOpPeek
	MoveOpStoreCode
	MoveOpLoadCode
	// always last
	MoveOpCount
)

type MoveOp func(*Core, *DecodedInstruction) error

func (fn MoveOp) Invoke(core *Core, inst *DecodedInstruction) error {
	return fn(core, inst)

}

var unimplementedMoveOp = func(_ *Core, _ *DecodedInstruction) error { return fmt.Errorf("Unimplemented move operation!") }
var moveTable = [32]MoveOp{
	moveRegister,
	swapRegisters,
	moveOpSet,
	load,
	store,
	moveOpPush,
	moveOpPop,
	moveOpPeek,
	moveOpStoreCode,
	moveOpLoadCode,
	unimplementedMoveOp,
	unimplementedMoveOp,
	unimplementedMoveOp,
	unimplementedMoveOp,
	unimplementedMoveOp,
	unimplementedMoveOp,
	unimplementedMoveOp,
	unimplementedMoveOp,
	unimplementedMoveOp,
	unimplementedMoveOp,
	unimplementedMoveOp,
	unimplementedMoveOp,
	unimplementedMoveOp,
	unimplementedMoveOp,
	unimplementedMoveOp,
	unimplementedMoveOp,
	unimplementedMoveOp,
}

func init() {
	if MoveOpCount > 32 {
		panic("Too many move operations registered! Programmer Failure!")
	}
}
func moveOpStoreCode(core *Core, inst *DecodedInstruction) error {
	// this one is a little odd since we have to use the contents of two registers
	// build an instruction from lower and upper
	return core.SetCodeMemory(core.Register(inst.Data[0]), (Instruction(core.Register(inst.Data[1])) | (Instruction(core.Register(inst.Data[2])) << 16)))
}
func moveOpLoadCode(core *Core, inst *DecodedInstruction) error {
	// in this case we need to load an Instruction from memory and store it into the upper and lower
	// registers. The syntax is dest, src0 are lower and upper with src1 being dest
	idat := core.CodeMemory(core.Register(inst.Data[2]))
	// set the lower and upper halves
	lowerHalf, upperHalf := Word(idat), Word(idat>>16)
	if err := core.SetRegister(inst.Data[0], lowerHalf); err != nil {
		return err
	} else if err := core.SetRegister(inst.Data[1], upperHalf); err != nil {
		return err
	} else {
		return nil
	}
}

func moveRegister(core *Core, inst *DecodedInstruction) error {
	dest, src := inst.Data[0], core.Register(inst.Data[1])
	return core.SetRegister(dest, src)
}
func swapRegisters(core *Core, inst *DecodedInstruction) error {
	dest, src0 := inst.Data[0], inst.Data[1]
	r0, r1 := core.Register(dest), core.Register(src0)
	if err := core.SetRegister(src0, r0); err != nil {
		return err
	} else if err := core.SetRegister(dest, r1); err != nil {
		return err
	} else {
		return nil
	}
}
func moveOpSet(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	return core.SetRegister(dest, inst.Immediate())
}
func load(core *Core, inst *DecodedInstruction) error {
	var val Word
	dest, addr, seg := inst.Data[0], core.Register(inst.Data[1]), segment(inst.Data[2])
	switch seg {
	case dataSegment:
		val = core.DataMemory(addr)
	case microcodeSegment:
		val = core.MicrocodeMemory(addr)
	case stackSegment:
		val = core.StackMemory(addr)
	case callSegment:
		val = core.CallMemory(addr)
	case ioSegment:
		if q, err := core.IoMemory(addr); err != nil {
			return err
		} else {
			val = q
		}
	case codeSegment:
		return fmt.Errorf("Can't load from the code segment!")
	default:
		return fmt.Errorf("Attempted to load from illegal segment %d", seg)
	}
	return core.SetRegister(dest, val)
}

func store(core *Core, inst *DecodedInstruction) error {
	dest, src, seg := core.Register(inst.Data[0]), core.Register(inst.Data[1]), segment(inst.Data[2])
	switch segment(seg) {
	case dataSegment:
		return core.SetDataMemory(dest, src)
	case microcodeSegment:
		return core.SetMicrocodeMemory(dest, src)
	case stackSegment:
		return core.SetStackMemory(dest, src)
	case callSegment:
		return core.SetCallMemory(dest, src)
	case codeSegment:
		return fmt.Errorf("Can't write to code memory!")
	case ioSegment:
		return core.SetIoMemory(dest, src)
	default:
		return fmt.Errorf("Attempted to write to illegal segment %d", seg)
	}
}

func moveOpPush(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	core.Push(core.Register(dest))
	return nil
}

func moveOpPop(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	return core.SetRegister(dest, core.Pop())
}

func moveOpPeek(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	return core.SetRegister(dest, core.Peek())
}

func move(core *Core, inst *DecodedInstruction) error {
	return moveTable[inst.Op].Invoke(core, inst)
}
