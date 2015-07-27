package iris1

import "fmt"

const (
	// Misc operations
	MiscOpSystemCall = iota
	NumberOfMiscOperations
)
const (
	// System commands
	SystemCallTerminate = iota
	SystemCallPanic
	NumberOfSystemCalls
)

type miscOpFunc func(*Core, *DecodedInstruction) error

func (this miscOpFunc) Invoke(core *Core, inst *DecodedInstruction) error {
	return this(core, inst)
}
func badMiscOp(_ *Core, _ *DecodedInstruction) error {
	return fmt.Errorf("Invalid misc operation!")
}

var miscOps [32]miscOpFunc

func init() {
	if NumberOfMiscOperations > 32 {
		panic("Too many misc operations defined!")
	}
	if NumberOfSystemCalls > 256 {
		panic("Too many system commands defined!")
	}
	for i := 0; i < 32; i++ {
		miscOps[i] = badMiscOp
	}
	miscOps[MiscOpSystemCall] = SystemCall
}

func misc(core *Core, inst *DecodedInstruction) error {
	return miscOps[inst.Op].Invoke(core, inst)
}
