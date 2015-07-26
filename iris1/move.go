// move related operations
package iris1

import (
	"fmt"
)

const (
	// Move Operations
	MoveOpMove = iota
	MoveOpSwap
	MoveOpSwapRegAddr
	MoveOpSwapAddrAddr
	MoveOpSwapRegMem
	MoveOpSwapAddrMem
	MoveOpSet
	MoveOpLoad
	MoveOpLoadMem
	MoveOpStore
	MoveOpStoreAddr
	MoveOpStoreMem
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
	swapRegAddr,
	swapAddrAddr,
	swapRegMem,
	swapAddrMem,
	moveOpSet,
	load,
	loadImm,
	store,
	storeAddr,
	storeMem,
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
func swapRegAddr(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	src0 := inst.Data[1]
	reg := core.Register(dest)
	memaddr := core.Register(src0)
	memcontents := core.DataMemory(memaddr)
	return swapMemoryAndRegister(core, dest, reg, memaddr, memcontents)
}
func swapAddrAddr(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	src0 := inst.Data[1]
	addr0 := core.Register(dest)
	addr1 := core.Register(src0)
	mem0 := core.DataMemory(addr0)
	mem1 := core.DataMemory(addr1)
	return swapMemory(core, addr0, mem0, addr1, mem1)

}
func swapRegMem(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	addr := inst.Immediate()
	return swapMemoryAndRegister(core, dest, core.Register(dest), addr, core.DataMemory(addr))
}
func moveOpSet(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	return core.SetRegister(dest, inst.Immediate())
}
func swapAddrMem(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	addr0 := core.Register(dest)
	addr1 := inst.Immediate()
	mem0 := core.DataMemory(addr0)
	mem1 := core.DataMemory(addr1)
	return swapMemory(core, addr0, mem0, addr1, mem1)

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
func storeAddr(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	src0 := inst.Data[1]
	return core.SetDataMemory(core.Register(dest), core.DataMemory(core.Register(src0)))
}
func storeMem(core *Core, inst *DecodedInstruction) error {
	dest := inst.Data[0]
	return core.SetDataMemory(core.Register(dest), core.DataMemory(inst.Immediate()))
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
