// implementation of ajvondrak's xand core
package xand

import (
	"fmt"
	"github.com/DrItanium/cores/registration/machine"
)

func RegistrationName() string {
	return "xand"
}

func Register() {}

type MachineRegistrar func(...interface{}) (machine.Machine, error)

func (this MachineRegistrar) New(args ...interface{}) (machine.Machine, error) {
	return this(args)
}

func generateCore(a ...interface{}) (machine.Machine, error) {
	return New()
}

func init() {
	machine.Register(RegistrationName(), MachineRegistrar(generateCore))
}

type Core struct {
	pc     int8
	ir     [3]int8
	memory [128]int8
}

func (this *Core) fetch() bool {
	if (this.pc < 0) || (int(this.pc+2) >= len(this.memory)) {
		return false
	} else {
		this.ir[0] = this.memory[this.pc]
		this.ir[1] = this.memory[this.pc+1]
		this.ir[2] = this.memory[this.pc+2]
		return this.ir[0] >= 0 && this.ir[1] >= 0 && this.ir[2] >= 0
	}
}

func (this *Core) Run() error {
	for this.fetch() {
		// the xand operation it self
		this.memory[this.ir[0]] = this.memory[this.ir[0]] - this.memory[this.ir[1]]
		if this.memory[this.ir[0]] <= 0 {
			this.pc = this.ir[2]
		} else {
			this.pc += 3
		}
	}
	return nil
}

func (this *Core) Startup() error {
	return nil
}

func (this *Core) Shutdown() error {
	return nil
}

func (this *Core) GetDebugStatus() bool {
	return false
}

func (this *Core) SetDebug(_ bool) {

}

func (this *Core) InstallProgram(input <-chan byte) error {
	// read 128 bytes
	for i := 0; i < 128; i++ {
		if value, more := <-input; !more {
			return fmt.Errorf("Not a complete xand memory image")
		} else {
			this.memory[i] = int8(value)
		}
	}
	return nil
}

func (this *Core) Dump(output chan<- byte) error {
	for _, value := range this.memory {
		output <- byte(value)
	}
	return nil
}

func New() (*Core, error) {
	return &Core{}, nil
}
