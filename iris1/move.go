// move related operations
package iris1

import "fmt"

func swapMemory(core *Core, addr0, data0, addr1, data1 Word) error {
	if err := core.SetDataMemory(addr0, data1); err != nil {
		return err
	} else if err := core.SetDataMemory(addr1, data0); err != nil {
		return err
	} else {
		return nil
	}
}
func swapMemoryAndRegister(core *Core, reg byte, data0, addr, data1 Word) error {
	if err := core.SetRegister(reg, data1); err != nil {
		return err
	} else if err := core.SetDataMemory(addr, data0); err != nil {
		return err
	} else {
		return nil
	}
}

const (
	// Move Operations
	MoveOpMove = iota
	MoveOpSwap
	MoveOpSet
	MoveOpLoad
	MoveOpLoadMem
	MoveOpStore
	MoveOpStoreImm
	MoveOpPush
	MoveOpPushImmediate
	MoveOpPop
	MoveOpPeek
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
	loadImm,
	store,
	storeImm,
	moveOpPush,
	moveOpPushImm,
	moveOpPop,
	moveOpPeek,
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
	unimplementedMoveOp,
	unimplementedMoveOp,
}

func init() {
	if MoveOpCount > 32 {
		panic("Too many move operations registered! Programmer Failure!")
	}
}
func moveRegister(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	src0 := inst.Data[1]
	return core.SetRegister(dest, core.Register(src0))
}
func swapRegisters(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	src0 := inst.Data[1]
	r0 := core.Register(dest)
	r1 := core.Register(src0)
	if err := core.SetRegister(src0, r0); err != nil {
		return err
	}
	if err := core.SetRegister(dest, r1); err != nil {
		return err
	}
	return nil
}
func moveOpSet(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	return core.SetRegister(dest, inst.Immediate())
}
func load(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	src0 := inst.Data[1]
	return core.SetRegister(dest, core.DataMemory(core.Register(src0)))
}
func loadImm(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	return core.SetRegister(dest, core.DataMemory(inst.Immediate()))
}

func store(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	src0 := inst.Data[1]
	return core.SetDataMemory(core.Register(dest), core.Register(src0))
}

func storeImm(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	return core.SetDataMemory(core.Register(dest), inst.Immediate())
}
func moveOpPush(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	core.Push(core.Register(dest))
	return nil
}
func moveOpPushImm(core *Core, inst *DecodedInstruction) error {
	core.Push(inst.Immediate())
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
