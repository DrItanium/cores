// machine description of iris1
package iris1

const (
	RegisterCount             = 256
	MemorySize                = 65536
	MajorOperationGroupCount  = 8
	PredicateRegisterIndex    = 255
	StackPointerRegisterIndex = 254
)

type uint16 Word
type uint32 DWord
type DWord Instruction

type Core struct {
	Gpr                [RegisterCount]Word
	Code               [MemorySize]Instruction
	Data               [MemorySize]Word
	Stack              [MemorySize]Word
	Pc                 Word
	AdvancePC          bool
	TerminateExecution bool
}
