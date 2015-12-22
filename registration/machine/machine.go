// generic machine interface
package machine

import (
	"fmt"
	"github.com/DrItanium/cores"
)

var registrations map[string]Registration

func Register(name string, gen Registration) error {
	if registrations == nil {
		registrations = make(map[string]Registration)
	}
	if _, ok := registrations[name]; ok {
		return fmt.Errorf("Machine %s is already registered!", name)
	} else {
		registrations[name] = gen
		return nil
	}
}
func GetRegistered() []string {
	var names []string
	if registrations != nil {
		for name, _ := range registrations {
			names = append(names, name)
		}
	}
	return names
}

func New(name string, args ...interface{}) (Machine, error) {
	if registrations == nil {
		return nil, fmt.Errorf("No machines registered!")
	}
	if gen, ok := registrations[name]; ok {
		return gen.New(args)
	} else {
		return nil, fmt.Errorf("%s does not refer to a registered machine!", name)
	}
}
func IsRegistered(name string) bool {
	if registrations == nil {
		return false
	} else {
		_, ok := registrations[name]
		return ok
	}
}

type Registration interface {
	New(args ...interface{}) (Machine, error)
}
type Machine interface {
	cores.Dumper
	GetDebugStatus() bool
	SetDebug(value bool)
	InstallProgram(input <-chan byte) error
	Startup() error
	Shutdown() error
	Run() error
}

// Dummy function used to force inclusion of this library
func Activate() {}
