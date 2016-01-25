package cond

import (
	"fmt"
	"github.com/DrItanium/cores/iris2"
)

type Unit struct {
	running                     bool
	out                         chan iris2.Word
	err                         chan error
	Error                       <-chan error
	Result, Control             <-chan iris2.Word
	operation, source0, source1 <-chan iris2.Word
}

const (
	Equal = iota
	NotEqual
	LessThan
	GreaterThan
	LessThanOrEqual
	GreaterThanOrEqual
	SelectSource0
	SelectSource1
	PassTrue
	PassFalse
	NumberOfCondStates
)

func New(control, operation, source0, source1 <-chan iris2.Word) *Unit {
	var s Unit
	s.out = make(chan iris2.Word)
	s.err = make(chan error)
	s.operation = operation
	s.source0 = source0
	s.source1 = source1
	s.Error = s.err
	s.Control = control
	s.Result = s.out
	return &s
}
func buildIntegerFunction(cond func(iris2.Word, iris2.Word) bool) func(iris2.Word, iris2.Word) iris2.Word {
	return func(a, b iris2.Word) iris2.Word {
		if cond(a, b) {
			return 1
		} else {
			return 0
		}
	}
}

var dispatchTable [NumberOfCondStates]func(iris2.Word, iris2.Word) iris2.Word

func init() {
	dispatchTable[Equal] = buildIntegerFunction(func(a, b iris2.Word) bool { return a == b })
	dispatchTable[NotEqual] = buildIntegerFunction(func(a, b iris2.Word) bool { return a != b })
	dispatchTable[LessThan] = buildIntegerFunction(func(a, b iris2.Word) bool { return a < b })
	dispatchTable[GreaterThan] = buildIntegerFunction(func(a, b iris2.Word) bool { return a > b })
	dispatchTable[LessThanOrEqual] = buildIntegerFunction(func(a, b iris2.Word) bool { return a <= b })
	dispatchTable[GreaterThanOrEqual] = buildIntegerFunction(func(a, b iris2.Word) bool { return a >= b })
	dispatchTable[SelectSource0] = func(a, _ iris2.Word) iris2.Word { return a }
	dispatchTable[SelectSource1] = func(_, b iris2.Word) iris2.Word { return b }
	dispatchTable[PassTrue] = func(_, _ iris2.Word) iris2.Word { return 1 }
	dispatchTable[PassFalse] = func(_, _ iris2.Word) iris2.Word { return 0 }

}
func (this *Unit) body() {
	for this.running {
		select {
		case op := <-this.operation:
			if op >= NumberOfCondStates {
				this.err <- fmt.Errorf("operation index %d is an undefined instruction!", op)
			} else if op < 0 {
				this.err <- fmt.Errorf("Send an operation index %d which is less than zero", op)
			} else {
				this.out <- dispatchTable[op](<-this.source0, <-this.source1)
			}
		}
	}
}

func (this *Unit) controlQuery() {
	<-this.Control
	this.shutdown()
}

func (this *Unit) shutdown() {
	this.running = false
}

func (this *Unit) Startup() error {
	if this.running {
		return fmt.Errorf("Given conditional unit is already running!")
	} else {
		this.running = true
		go this.body()
		go this.controlQuery()
		return nil
	}
}
