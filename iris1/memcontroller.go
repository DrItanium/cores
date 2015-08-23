package iris1

import "fmt"

type memControllerInput struct {
	Code    byte
	Address Word
	Value   interface{}
	Width   byte
}
type memController struct {
	rawMemory    []byte
	input        chan memControllerInput
	output       chan interface{}
	signalInput  chan string
	signalOutput chan error
	terminate    bool
	err          bool
	errMsg       string
}

func (this *memController) signalHandler() {
	for {
		result := <-this.signalInput
		if result == "shutdown" {
			this.terminate = true
			this.signalOutput <- nil
			return
		} else if result == "pause" {
			this.terminate = true
		} else if result == "resume" {
			this.terminate = false
			// restart the memory controller
			go this.handler()
		} else if result == "startup" {
			this.terminate = false
			go this.handler()
		} else if result == "recover" {
			this.err = false
			this.errMsg = ""
		} else if result == "status" {
			if this.err {
				this.signalOutput <- fmt.Errorf("Memory Controller error: %s", errMsg)
			} else {
				this.signalOutput <- nil
			}
		} else {
			this.signalOutput <- fmt.Errorf("ERROR: illegal signal %s", result)
		}
	}
}

const (
	memControllerOperationRead = iota
	memControllerOperationReadInstruction
	memControllerOperationWrite
)

var shiftTable = []Dword{0, 8, 16, 24, 32, 40, 48, 56}
var memControllerError = fmt.Errorf("MEMORY CONTROLLER ERROR!")

func (this *memController) handler() {
	toInst := func(slice []byte) Instruction {
		var i Instruction
		for ind, val := range slice {
			i.setByte(ind, val)
		}
		return i
	}
	toDword := func(slice []byte) Dword {
		var w Dword
		for ind, val := range slice {
			w = w | Dword(Dword(val)>>shiftTable[ind])
		}
		return w
	}
	for !this.terminate {
		// if we have an error then just sit and spin!
		if !this.err {
			// listen in on stuff
			select {
			case in := <-this.input:
				switch in.Code {
				case memControllerOperationRead:
					if r, e := this.memory(in.Address, in.Width); e != nil {
						this.errMsg = e.Error()
						this.err = true
						this.output <- memControllerError
					}
					switch in.Width {
					case 1:
						this.output <- byte(toDword(r))
					case 2:
						this.output <- QuarterWord(toDword(r))
					case 3, 4:
						this.output <- Word(toDword(r))
					case 6, 7, 8:
						this.output <- toDword(r)
					case 48:
						packet := make([]Instruction, 8)
						q := r
						for i := 0; i < 8; i++ {
							packet[i] = toInst(q[:6])
							q = q[6:]
						}
						this.output <- packet
					default:
						this.output <- r
					}
				case memControllerOperationReadInstruction:
					if r, e := this.memory(in.Address, 6); e != nil {
						this.output <- memControllerError
						this.err = true
						this.errMsg = e.Error()
					} else {
						this.output <- toInst(r)
					}
				case memControllerOperationWrite:
					switch t := in.Value.(type) {
					default:
					}
				default:
					this.errMsg = fmt.Sprintf("ERROR: unknown memory operation %d", in.Code)
					this.err = true
					this.output <- memControllerError
				}
			default:
				// do nothing, just wait around
			}
		}
	}
}
func newMemController(size uint32, signalOutput chan error) *memController {
	var mc memController
	mc.rawMemory = make([]byte, size)
	mc.input = make(chan memControllerInput)
	mc.output = make(chan memControllerOutput)
	mc.signalInput = make(chan string)
	mc.signalOutput = signalOutput
	go mc.signalHandler()
	mc.signalInput <- "startup"
	return &mc
}

func (this *memController) memory(address Word, width byte) ([]byte, error) {
	if address >= Word(len(this.rawMemory)) {
		return nil, fmt.Errorf("Attempted to access memory address %x outside of range!", address)
	} else if (address + Word(width)) >= Word(len(this.rawMemory)) {
		return nil, fmt.Errorf("Attempted to access %d cells starting at memory address %x! This will go outside range!", width, address)
	} else if width == 0 {
		return nil, fmt.Errorf("Attempted to read 0 bytes starting at address %x!", address)
	} else {
		return this.rawMemory[address:(Word(width-1) + address)], nil
	}
}
func (this *memController) setMemory(address Word, data []byte) error {
	if address >= Word(len(this.rawMemory)) {
		return fmt.Errorf("Memory address %x is outside of memory range!", address)
	} else if (address + Word(len(data))) >= Word(len(this.rawMemory)) {
		return fmt.Errorf("Writing %d cells starting at memory address %x will go out of range!", len(data), address)
	} else {
		for ind, val := range data {
			this.rawMemory[address+Word(ind)] = val
		}
		return nil
	}
}
