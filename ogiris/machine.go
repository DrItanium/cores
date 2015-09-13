// straight port of the 16bit iris core from my C version
package ogiris

import (
	"fmt"
)

type Word uint16
type Dword uint32

type Instruction Dword

const (
	RegisterCount             = 256
	MemorySize                = 65536
	MajorOperationMax         = 8
	MinorOperationMax         = 32
	PredicateRegisterIndex    = 255
	StackPointerRegisterIndex = 254
)

type Core struct {
	Gpr                           [RegisterCount]Word
	Code                          [MemorySize]Instruction
	Data, Stack                   [MemorySize]Word
	Pc                            Word
	AdvancePc, TerminateExecution bool
}

var registeredGroups = []struct {
	Name       string
	Count, Max int
}{
	{Name: "instruction groups", Count: GroupCount, Max: MajorOperationMax},
	{Name: "arithmetic operations", Count: ArithmeticOpCount, Max: MinorOperationMax},
	{Name: "move operations", Count: MoveOpCount, Max: MinorOperationMax},
}

// instruction groups
const (
	GroupArithmetic = iota
	GroupMove
	GroupJump
	GroupCompare
	GroupMisc
	// add more here
	GroupCount
)

// arithmetic ops
const (
	ArithmeticOpAdd = iota
	ArithmeticOpSub
	ArithmeticOpMul
	ArithmeticOpDiv
	ArithmeticOpRem
	ArithmeticOpShiftLeft
	ArithmeticOpShiftRight
	ArithmeticOpBinaryAnd
	ArithmeticOpBinaryOr
	ArithmeticOpBinaryNot
	ArithmeticOpBinaryXor
	ArithmeticOpAddImmediate
	ArithmeticOpSubImmediate
	ArithmeticOpMulImmediate
	ArithmeticOpDivImmediate
	ArithmeticOpRemImmediate
	ArithmeticOpShiftLeftImmediate
	ArithmeticOpShiftRightImmediate
	ArithmeticOpCount
)

const (
	MoveOpMove         = iota // move r? r?
	MoveOpSwap                // swap r? r?
	MoveOpSwapRegAddr         // swap.reg.addr r? r?
	MoveOpSwapAddrAddr        // swap.addr.addr r? r?
	MoveOpSwapRegMem          // swap.reg.mem r? $imm
	MoveOpSwapAddrMem         // swap.addr.mem r? $imm
	MoveOpSet                 // set r? $imm
	MoveOpLoad                // load r? r?
	MoveOpLoadMem             // load.mem r? $imm
	MoveOpStore               // store r? r?
	MoveOpStoreAddr           // store.addr r? r?
	MoveOpStoreMem            // memcopy r? $imm
	MoveOpStoreImm            // memset r? $imm
	// uses an indirect register for the stack pointer
	MoveOpPush          // push r?
	MoveOpPushImmediate // push.imm $imm
	MoveOpPop           // pop r?
	MoveOpCount
)

func init() {
	for _, value := range registeredGroups {
		if value.Count > value.Max {
			panic(fmt.Sprintf("Too many %s defined, %d allowed but %d defined!", value.Name, value.Max, value.Count))
		}
	}
}
