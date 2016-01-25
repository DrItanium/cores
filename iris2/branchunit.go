package iris2

import (
	"fmt"
)

type BranchUnit struct {
	running                    bool
	out                        chan Word
	err                        chan error
	condition, onTrue, onFalse <-chan Word
	Control                    <-chan Word
	Result                     <-chan Word
	Error                      <-chan error
}

func NewBranchUnit(control, condition, onTrue, onFalse <-chan Word) (*BranchUnit, error) {
	var unit BranchUnit
	unit.out = make(chan Word)
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
		if cond, onTrue, onFalse := <-this.condition, <-this.onTrue, <-this.onFalse; cond != 0 {
			this.out <- onTrue
		} else {
			this.out <- onFalse
		}
	}
}
func (this *BranchUnit) terminate() error {
	if this.running {
		this.running = false
		return nil
	} else {
		return fmt.Errorf("Given unit is already shutdown")
	}
}
