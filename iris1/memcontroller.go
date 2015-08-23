package iris1

import "fmt"

type DeviceInput interface {
	Code() byte
	Address() Word
	Value() interface{}
	Width() byte
}
type memControllerInput struct {
	code, width byte
	address     Word
	value       interface{}
}

func (this *memControllerInput) Code() byte {
	return this.code
}
func (this *memControllerInput) Width() byte {
	return this.width
}
func (this *memControllerInput) Address() Word {
	return this.address
}
func (this *memControllerInput) Value() interface{} {
	return this.value
}
func newMemControllerInput(code, width byte, address Word, value interface{}) DeviceInput {
	var mc memControllerInput
	mc.code = code
	mc.width = width
	mc.address = address
	mc.value = value
	return &mc
}

type DeviceInputChannel chan DeviceInput
type memController struct {
	rawMemory    []byte
	Input        DeviceInputChannel
	Output       chan interface{}
	SignalInput  chan string
	SignalOutput chan error
	terminate    bool
	err          bool
	errMsg       string
}

func (this *memController) signalHandler() {
	for {
		result := <-this.SignalInput
		if result == "shutdown" {
			this.terminate = true
			this.SignalOutput <- nil
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
				this.SignalOutput <- fmt.Errorf("Memory Controller error: %s", this.errMsg)
			} else {
				this.SignalOutput <- nil
			}
		} else {
			this.SignalOutput <- fmt.Errorf("ERROR: illegal signal %s", result)
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
			case in := <-this.Input:
				switch in.Code() {
				case memControllerOperationRead:
					r, e := this.memory(in.Address(), in.Width())
					if e != nil {
						this.errMsg = e.Error()
						this.err = true
						this.Output <- memControllerError
					}
					switch in.Width() {
					case 1:
						this.Output <- byte(toDword(r))
					case 2:
						this.Output <- Halfword(toDword(r))
					case 3, 4:
						this.Output <- Word(toDword(r))
					case 6, 7, 8:
						this.Output <- toDword(r)
					case 48:
						packet := make([]Instruction, 8)
						q := r
						for i := 0; i < 8; i++ {
							packet[i] = toInst(q[:6])
							q = q[6:]
						}
						this.Output <- packet
					default:
						this.Output <- r
					}
				case memControllerOperationReadInstruction:
					if r, e := this.memory(in.Address(), in.Width()); e != nil {
						this.Output <- memControllerError
						this.err = true
						this.errMsg = e.Error()
					} else {
						this.Output <- toInst(r)
					}
				case memControllerOperationWrite:
					var e error
					val := in.Value()
					switch t := val.(type) {
					case []byte:
						e = this.setMemory(in.Address(), val.([]byte))
					case byte:
						e = this.setMemory(in.Address(), []byte{val.(byte)})
					case Halfword:
						hw := val.(Halfword)
						e = this.setMemory(in.Address(), []byte{byte(hw), byte(hw >> 8)})
					case Word:
						w := val.(Word)
						e = this.setMemory(in.Address(), []byte{byte(w), byte(w >> 8), byte(w >> 16), byte(w >> 24)})
					case Dword:
						w := val.(Dword)
						e = this.setMemory(in.Address(), []byte{
							byte(w),
							byte(w >> 8),
							byte(w >> 16),
							byte(w >> 24),
							byte(w >> 32),
							byte(w >> 40),
							byte(w >> 48),
							byte(w >> 56)})
					case Instruction:
						w := val.(Instruction)
						contents := make([]byte, in.Width())
						for i := 0; i < int(in.Width()); i++ {
							if val, err := w.register(i); err != nil {
								e = err
								break
							} else {
								contents[i] = val
							}
						}
						e = this.setMemory(in.Address(), contents)
					default:
						e = fmt.Errorf("ERROR: attempted to write unsupported type %t to memory!", t)
					}
					if e != nil {
						this.errMsg = e.Error()
						this.err = true
						this.Output <- memControllerError
					}
				default:
					this.errMsg = fmt.Sprintf("ERROR: unknown memory operation %d", in.Code())
					this.err = true
					this.Output <- memControllerError
				}
			default:
				// do nothing, just wait around
			}
		}
	}
}
func newMemController(size uint32) *memController {
	var mc memController
	mc.rawMemory = make([]byte, size)
	mc.Input = make(DeviceInputChannel)
	mc.Output = make(chan interface{})
	mc.SignalInput = make(chan string)
	mc.SignalOutput = make(chan error)
	go mc.signalHandler()
	mc.SignalInput <- "startup"
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
