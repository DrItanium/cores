package iris32

import (
	"fmt"
)

const (
	// System commands
	SystemCallTerminate = iota
	SystemCallPanic
	SystemCallPutc
	NumberOfSystemCalls
)

func init() {
	if NumberOfSystemCalls > 256 {
		panic("Too many system commands defined!")
	}
}
func defaultSystemCall(core *Core, inst *DecodedInstruction) error {
	return NewError(ErrorInvalidSystemCommand, uint(inst.Data[0]))
}

func putcSystemCall(core *Core, inst *DecodedInstruction) error {
	// extract the two registers we need
	var r rune
	lower, upper := inst.Data[1], inst.Data[2]
	if lower == upper {
		// it is an ascii value so only use the lower half
		r = rune(byte(core.Register(lower)))
	} else {
		// make a rune out of it
		r = rune((0x0000FFFF & uint32(core.Register(lower))) | (0xFFFF0000 & (uint32(core.Register(upper)) << 16)))
	}
	fmt.Printf("%c", r)
	return nil
}
func terminateSystemCall(core *Core, inst *DecodedInstruction) error {
	core.terminateExecution = true
	return nil
}
func panicSystemCall(core *Core, inst *DecodedInstruction) error {
	// we don't want to panic the program itself but generate a new error
	// look at the data attached to the panic and encode it
	return NewError(ErrorPanic, uint(inst.Immediate()))
}
