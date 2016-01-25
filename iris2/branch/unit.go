package branch

import (
	"fmt"
	"github.com/DrItanium/cores/iris2"
)

type Unit struct {
	running                    bool
	out                        chan iris2.Word
	err                        chan error
	condition, onTrue, onFalse <-chan iris2.Word
	Control                    <-chan iris2.Word
	Result                     <-chan iris2.Word
	Error                      <-chan error
}

func New(control, condition, onTrue, onFalse <-chan iris2.Word) (*Unit, error) {
	var unit Unit
	unit.out = make(chan iris2.Word)
	unit.err = make(chan error)
	unit.Result = unit.out
	unit.condition = condition
	unit.Control = control
	unit.onTrue = onTrue
	unit.onFalse = onFalse
	if err := unit.Startup(); err != nil {
		return nil, err
	} else {
		return &unit, nil
	}
}
func (this *Unit) Startup() error {
	if this.running {
		return fmt.Errorf("Branch unit is already running!")
	} else {
		this.running = true
		go this.body()
		go this.controlStream()
		return nil
	}
}
func (this *Unit) controlStream() {
	<-this.Control
	if err := this.terminate(); err != nil {
		this.err <- err
	}
}
func (this *Unit) body() {
	for this.running {
		if cond, onTrue, onFalse := <-this.condition, <-this.onTrue, <-this.onFalse; cond != 0 {
			this.out <- onTrue
		} else {
			this.out <- onFalse
		}
	}
}
func (this *Unit) terminate() error {
	if this.running {
		this.running = false
		return nil
	} else {
		return fmt.Errorf("Given unit is already shutdown")
	}
}
