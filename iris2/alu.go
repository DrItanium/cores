package iris2

import (
	"fmt"
)

type Alu struct {
	running                          bool
	err                              chan error
	out, operation, source0, source1 chan Word
	Result                           <-chan Word
	Error                            <-chan error
	Control                          <-chan Word
	Operation, Source0, Source1      chan<- Word
}

func NewAlu(control <-chan Word) *Alu {
	var this Alu
	this.err = make(chan error)
	this.out = make(chan Word)
	this.operation = make(chan Word)
	this.source0 = make(chan Word)
	this.source1 = make(chan Word)
	this.Result = this.out
	this.Error = this.err
	this.Control = control
	this.Operation = this.operation
	this.Source0 = this.source0
	this.Source1 = this.source1
	return &this
}

const (
	IntegerAdd = iota
	IntegerSubtract
	IntegerMultiply
	IntegerDivide
	IntegerRemainder
	IntegerShiftLeft
	IntegerShiftRight
	IntegerAnd
	IntegerOr
	IntegerNot
	IntegerXor
	IntegerAndNot
	IntegerOpCount
)

var integerArithmeticOps [IntegerOpCount]func(Word, Word) (Word, error)

func init() {
	integerArithmeticOps[IntegerAdd] = func(a, b Word) (Word, error) { return a + b, nil }
	integerArithmeticOps[IntegerSubtract] = func(a, b Word) (Word, error) { return a - b, nil }
	integerArithmeticOps[IntegerMultiply] = func(a, b Word) (Word, error) { return a * b, nil }
	integerArithmeticOps[IntegerDivide] = func(num, denom Word) (Word, error) {
		if denom == 0 {
			return 0, fmt.Errorf("Divide by zero")
		} else {
			return num / denom, nil
		}
	}
	integerArithmeticOps[IntegerRemainder] = func(num, denom Word) (Word, error) {
		if denom == 0 {
			return 0, fmt.Errorf("Divide by zero")
		} else {
			return num % denom, nil
		}
	}
	integerArithmeticOps[IntegerShiftLeft] = func(a, b Word) (Word, error) { return a << uint64(b), nil }
	integerArithmeticOps[IntegerShiftRight] = func(a, b Word) (Word, error) { return a >> uint64(b), nil }
	integerArithmeticOps[IntegerAnd] = func(a, b Word) (Word, error) { return a & b, nil }
	integerArithmeticOps[IntegerOr] = func(a, b Word) (Word, error) { return a | b, nil }
	integerArithmeticOps[IntegerNot] = func(a, _ Word) (Word, error) { return ^a, nil }
	integerArithmeticOps[IntegerXor] = func(a, b Word) (Word, error) { return a ^ b, nil }
	integerArithmeticOps[IntegerAndNot] = func(a, b Word) (Word, error) { return a &^ b, nil }
}

func (this *Alu) Startup() error {
	if this.running {
		return fmt.Errorf("Arithmetic unit is already running")
	} else {
		this.running = true
		go this.body()
		go this.controlQuery()
		return nil
	}
}
func (this *Alu) controlQuery() {
	<-this.Control
	if err := this.shutdown(); err != nil {
		this.err <- err
	}
}
func (this *Alu) body() {
	for this.running {
		select {
		case op, more := <-this.operation:
			if more {
				if op >= IntegerOpCount {
					this.err <- fmt.Errorf("Index %d is not a legal instruction index!", op)
				} else if op < 0 {
					this.err <- fmt.Errorf("Index %d is less than zero!", op)
				} else if result, err := integerArithmeticOps[op](<-this.source0, <-this.source1); err != nil {
					this.err <- err
				} else {
					this.out <- result
				}
			}
		}
	}
}

func (this *Alu) shutdown() error {
	if !this.running {
		return fmt.Errorf("this unit is not currently running!")
	} else {
		this.running = false
		close(this.operation)
		close(this.source0)
		close(this.source1)
		return nil
	}
}

type Fpu struct {
	running          bool
	err              chan error
	out              chan float64
	Result           <-chan float64
	Error            <-chan error
	Control          <-chan Word
	operation        <-chan Word
	source0, source1 <-chan float64
}

func NewFpu(control, operation <-chan Word, source0, source1 <-chan float64) *Fpu {
	var this Fpu
	this.err = make(chan error)
	this.out = make(chan float64)
	this.Result = this.out
	this.Error = this.err
	this.Control = control
	this.operation = operation
	this.source0 = source0
	this.source1 = source1
	return &this
}

const (
	FloatAdd = iota
	FloatSubtract
	FloatMultiply
	FloatDivide
	FloatOpCount
)

var floatArithmeticOps [FloatOpCount]func(float64, float64) float64

func init() {
	floatArithmeticOps[FloatAdd] = func(a, b float64) float64 { return a + b }
	floatArithmeticOps[FloatSubtract] = func(a, b float64) float64 { return a - b }
	floatArithmeticOps[FloatMultiply] = func(a, b float64) float64 { return a * b }
	floatArithmeticOps[FloatDivide] = func(a, b float64) float64 { return a / b }
}

func (this *Fpu) Startup() error {
	if this.running {
		return fmt.Errorf("Arithmetic unit is already running")
	} else {
		this.running = true
		go this.body()
		go this.controlQuery()
		return nil
	}
}

func (this *Fpu) controlQuery() {
	<-this.Control
	if err := this.shutdown(); err != nil {
		this.err <- err
	}
}

func (this *Fpu) body() {
	for this.running {
		select {
		case op := <-this.operation:
			if op >= FloatOpCount {
				this.err <- fmt.Errorf("Index %d is not a legal instruction index!", op)
			} else if op < 0 {
				this.err <- fmt.Errorf("Index %d is less than zero!", op)
			} else {
				this.out <- floatArithmeticOps[op](<-this.source0, <-this.source1)
			}
		}
	}
}

func (this *Fpu) shutdown() error {
	if !this.running {
		return fmt.Errorf("this unit is not currently running!")
	} else {
		this.running = false
		return nil
	}
}
