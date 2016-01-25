package branch

import (
	"fmt"
)

// 64bit unit
type Unit64 struct {
	running                    bool
	condition, onTrue, onFalse chan uint64
	alive                      chan bool
	out                        chan uint64
	Result                     <-chan uint64
}

func New64() (*Unit64, error) {
	var unit Unit64
	unit.out = make(chan uint64)
	unit.Result = unit.out
	unit.alive = make(chan bool)
	unit.condition = make(chan uint64)
	unit.onTrue = make(chan uint64)
	unit.onFalse = make(chan uint64)
	if err := unit.Startup(); err != nil {
		return nil, err
	} else {
		return &unit, nil
	}
}
func (this *Unit64) Startup() error {
	if this.running {
		return fmt.Errorf("Branch unit is already running!")
	} else {
		this.running = true
		go this.body()
		return nil
	}
}
func (this *Unit64) body() {
	for <-this.alive {
		if cond, onTrue, onFalse := <-this.condition, <-this.onTrue, <-this.onFalse; cond != 0 {
			this.out <- onTrue
		} else {
			this.out <- onFalse
		}
	}
}
func (this *Unit64) Terminate() error {
	if this.running {
		this.alive <- false
		this.running = false
		return nil
	} else {
		return fmt.Errorf("Given unit is already shutdown")
	}
}
func (this *Unit64) Dispatch(condition, onTrue, onFalse uint64) error {
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
