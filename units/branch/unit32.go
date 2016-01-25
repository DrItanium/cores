package branch

import (
	"fmt"
)

// 32bit unit
type Unit32 struct {
	running                    bool
	condition, onTrue, onFalse chan uint32
	alive                      chan bool
	out                        chan uint32
	Result                     <-chan uint32
}

func New32() (*Unit32, error) {
	var unit Unit32
	unit.out = make(chan uint32)
	unit.Result = unit.out
	unit.alive = make(chan bool)
	unit.condition = make(chan uint32)
	unit.onTrue = make(chan uint32)
	unit.onFalse = make(chan uint32)
	if err := unit.Startup(); err != nil {
		return nil, err
	} else {
		return &unit, nil
	}
}
func (this *Unit32) Startup() error {
	if this.running {
		return fmt.Errorf("Branch unit is already running!")
	} else {
		this.running = true
		go this.body()
		return nil
	}
}
func (this *Unit32) body() {
	for <-this.alive {
		if cond, onTrue, onFalse := <-this.condition, <-this.onTrue, <-this.onFalse; cond != 0 {
			this.out <- onTrue
		} else {
			this.out <- onFalse
		}
	}
}
func (this *Unit32) Terminate() error {
	if this.running {
		this.alive <- false
		this.running = false
		return nil
	} else {
		return fmt.Errorf("Given unit is already shutdown")
	}
}
func (this *Unit32) Dispatch(condition, onTrue, onFalse uint32) error {
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
