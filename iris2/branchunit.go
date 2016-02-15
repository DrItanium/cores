package iris2

import (
	"fmt"
)

type BranchUnit struct {
	running                    bool
	out                        chan Word
	err                        chan error
	condition, onTrue, onFalse chan Word
	Control                    <-chan Word
	Result                     <-chan Word
	Error                      <-chan error
	Condition, OnTrue, OnFalse chan<- Word
}

func NewBranchUnit(control <-chan Word) *BranchUnit {
	var unit BranchUnit
	unit.out = make(chan Word)
	unit.err = make(chan error)
	unit.onTrue = make(chan Word)
	unit.onFalse = make(chan Word)
	unit.condition = make(chan Word)
	unit.Result = unit.out
	unit.Condition = unit.condition
	unit.Control = control
	unit.OnTrue = unit.onTrue
	unit.OnFalse = unit.onFalse
	return &unit
}
func (this *BranchUnit) Startup() error {
	if this.running {
		return fmt.Errorf("Branch unit is already running!")
	} else {
		this.running = true
		go this.body()
		go this.controlStream()
		return nil
	}
}
func (this *BranchUnit) controlStream() {
	<-this.Control
	if err := this.terminate(); err != nil {
		this.err <- err
	}
}
func (this *BranchUnit) body() {
	for this.running {
		select {
		case cond, more := <-this.condition:
			if more {
				if onTrue, onFalse := <-this.onTrue, <-this.onFalse; cond != 0 {
					this.out <- onTrue
				} else {
					this.out <- onFalse
				}
			}
		}
	}
}
func (this *BranchUnit) terminate() error {
	if this.running {
		this.running = false
		close(this.condition)
		close(this.onTrue)
		close(this.onFalse)
		return nil
	} else {
		return fmt.Errorf("Given unit is already shutdown")
	}
}
