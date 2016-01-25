// register file execution unit
package iris2

import "fmt"

type registerFile struct {
	// internal registers that should be easy to find
	instructionPointer Word
	stackPointer       Word
	callPointer        Word
	predicate          Word
	gpr                [RegisterCount - UserRegisterBegin]Word

	err   chan error
	out   chan Word
	index chan byte
	value chan interface{}

	Error     <-chan error
	Control   <-chan Word
	Index     chan<- byte
	Value     chan<- interface{}
	Result    <-chan Word
	Operation <-chan byte

	running bool
	temp    Word // temporary storage for internal swap operations
}

const (
	// reserved registers
	FalseRegister = iota
	TrueRegister
	InstructionPointer
	StackPointer
	PredicateRegister
	CallPointer
	UserRegisterBegin
)
const (
	registerFileGet = iota
	registerFileSet
	registerFileSwap
	registerFileMove
	registerFileOpCount
)

func newRegisterFile(control <-chan Word, op <-chan byte) *registerFile {
	var this registerFile
	this.err = make(chan error)
	this.out = make(chan Word)
	this.index = make(chan byte)
	this.value = make(chan interface{})

	this.Error = this.err
	this.Control = control
	this.Operation = op
	this.Result = this.out
	this.Index = this.index
	this.Value = this.value

	return &this
}
func (this *registerFile) setRegister(index byte, value Word) error {
	switch index {
	case FalseRegister:
		return fmt.Errorf("Can't write to the zero/false register!")
	case TrueRegister:
		return fmt.Errorf("Can't write to the one/true register!")
	case InstructionPointer:
		this.instructionPointer = value
	case StackPointer:
		this.stackPointer = value
	case PredicateRegister:
		this.predicate = value
	case CallPointer:
		this.callPointer = value
	default:
		this.gpr[index-UserRegisterBegin] = value
	}
	return nil
}
func (this *registerFile) getRegister(index byte) Word {
	switch index {
	case FalseRegister:
		return 0
	case TrueRegister:
		return 1
	case InstructionPointer:
		return this.instructionPointer
	case StackPointer:
		return this.stackPointer
	case PredicateRegister:
		return this.predicate
	case CallPointer:
		return this.callPointer
	default:
		// do the offset calculation
		return this.gpr[index-UserRegisterBegin]
	}
}

func (this *registerFile) swapRegisters(a, b byte) error {
	if a != b {
		this.temp = this.getRegister(a)
		if err := this.setRegister(a, this.getRegister(b)); err != nil {
			return err
		} else if err := this.setRegister(b, this.temp); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		return nil
	}
}
func (this *registerFile) moveRegister(a, b byte) error {
	if a != b {
		return this.setRegister(a, this.getRegister(b))
	} else {
		return nil
	}
}
func (this *registerFile) Startup() error {
	if this.running {
		return fmt.Errorf("This register file is already running!")
	} else {
		this.running = true
		go this.body()
		go this.controlQuery()
		return nil
	}
}
func (this *registerFile) body() {
	for this.running {
		select {
		case op := <-this.Operation:
			if op >= registerFileOpCount {
			} else {
				arg0 := <-this.index
				arg1 := <-this.value
				switch op {
				case registerFileGet:
					this.out <- this.getRegister(arg0)
				case registerFileSet:
					switch arg1.(type) {
					case Word:
						if err := this.setRegister(arg1.(Word)); err != nil {
							this.err <- err
						}
					default:
						this.err <- fmt.Errorf("Only words are allowed in a register file set operation!")
					}
				case registerFileSwap:
					switch arg1.(type) {
					case byte:
						if err := this.swapRegisters(arg0, arg1.(byte)); err != nil {
							this.err <- err
						}
					default:
						this.err <- fmt.Errorf("Only bytes are allowed in a register file swap operation!")
					}
				case registerFileMove:
					switch arg1.(type) {
					case byte:
						if err := this.moveRegister(arg0, arg1.(byte)); err != nil {
							this.err <- err
						}
					default:
						this.err <- fmt.Errorf("Only bytes are allowed in a register file move operation!")
					}
				default:
					this.err <- fmt.Errorf("Illegal register file operation %d", op)
				}
			}
		}
	}
}

func (this *registerFile) controlQuery() {
	<-this.Control
	if err := this.shutdown(); err != nil {
		this.err <- err
	}
}

func (this *registerFile) shutdown() error {
	if !this.running {
		return fmt.Errorf("Can't shutdown a register file that isn't running!")
	} else {
		this.running = false
		return nil
	}
}
