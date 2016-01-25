package branch

import (
	"fmt"
)

// 16bit unit
type Unit16 struct {
	running                    bool
	condition, onTrue, onFalse chan uint16
	alive                      chan bool
	out                        chan uint16
	Result                     <-chan uint16
}

func New16() (*Unit16, error) {
	var unit Unit16
	unit.out = make(chan uint16)
	unit.Result = unit.out
	unit.alive = make(chan bool)
	unit.condition = make(chan uint16)
	unit.onTrue = make(chan uint16)
	unit.onFalse = make(chan uint16)
	if err := unit.Startup(); err != nil {
		return nil, err
	} else {
		return &unit, nil
	}
}
func (this *Unit16) Startup() error {
	if this.running {
		return fmt.Errorf("Branch unit is already running!")
	} else {
		this.running = true
		go this.body()
		return nil
	}
}
func (this *Unit16) body() {
	for <-this.alive {
		if cond, onTrue, onFalse := <-this.condition, <-this.onTrue, <-this.onFalse; cond != 0 {
			this.out <- onTrue
		} else {
			this.out <- onFalse
		}
	}
}
func (this *Unit16) Terminate() error {
	if this.running {
		this.alive <- false
		this.running = false
		return nil
	} else {
		return fmt.Errorf("Given unit is already shutdown")
	}
}
func (this *Unit16) Dispatch(condition, onTrue, onFalse uint16) error {
	if !this.running {
		return fmt.Errorf("Can't dispatch on a deactivated unit!")
	} else {
		this.alive <- true
		this.condition <- condition
		this.onTrue <- onTrue
		this.onFalse <- onFalse
		return nil
	}
}
