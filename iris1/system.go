package iris1

import "fmt"

const (
	// Misc operations
	MiscOpSystemCall = iota
	NumberOfMiscOperations
)
const (
	// System commands
	SystemCommandTerminate = iota
	SystemCommandPanic
	SystemCommandCount
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
	if SystemCommandCount > 256 {
		panic("Too many system commands defined!")
	}
	for i := 0; i < 32; i++ {
		miscOps[i] = badMiscOp
	}
	miscOps[MiscOpSystemCall] = systemCall
}

func misc(core *Core, inst *DecodedInstruction) error {
	return miscOps[inst.Op].Invoke(core, inst)
}
func systemCall(core *Core, inst *DecodedInstruction) error {
	switch inst.Data[0] {
	case SystemCommandTerminate:
		core.terminateExecution = true
	case SystemCommandPanic:
		// this is a special case that I haven't implemented yet
	default:
		return fmt.Errorf("Illegal signal %d", inst.Data[0])
	}
	return nil
}
