// iris1 core with unicornhat interface
package unicorn

import (
	"github.com/DrItanium/cores/iris1"
	"github.com/DrItanium/cores/registration/machine"
	"github.com/DrItanium/unicornhat"
)

func RegistrationName() string {
	return "iris1-unicorn"
}

func Register() {}

type MachineRegistrar func(...interface{}) (machine.Machine, error)

func (this MachineRegistrar) New(args ...interface{}) (machine.Machine, error) {
	return this(args)
}

func generateCore(a ...interface{}) (machine.Machine, error) {
	// install system handlers after the fact
	if c, err := iris1.New(); err != nil {
		return nil, err
	} else {

	}
}

func init() {
	machine.Register(RegistrationName(), MachineRegistrar(generateCore))
}
