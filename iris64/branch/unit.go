package branch

import (
	"fmt"
)

// 16bit unit
type Unit struct {
	running                    bool
	condition, onTrue, onFalse <-chan uint64
	Control                    <-chan uint64
	out                        chan uint64
	Result                     <-chan uint64
}

func New16(control, condition, onTrue, onFalse <-chan uint64) (*Unit, error) {
	var unit Unit
	unit.out = make(chan uint64)
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
	this.terminate()
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
