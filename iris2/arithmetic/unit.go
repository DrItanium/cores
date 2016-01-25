package arithmetic

import (
	"fmt"
	"github.com/DrItanium/cores/iris2"
)

type Unit struct {
	running                     bool
	err                         chan error
	out                         chan iris2.Word
	Result                      <-chan iris2.Word
	Error                       <-chan error
	Control                     <-chan iris2.Word
	operation, source0, source1 <-chan iris2.Word
}

func NewIntegerUnit(control, operation, source0, source1 <-chan iris2.Word) *Unit {
	var this Unit
	this.err = make(chan error)
	this.out = make(chan iris2.Word)
	this.Result = this.out
	this.Error = this.err
	this.Control = control
	this.operation = operation
	this.source0 = source0
	this.source1 = source1
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

var integerArithmeticOps [IntegerOpCount]func(iris2.Word, iris2.Word) (iris2.Word, error)

func init() {
	integerArithmeticOps[IntegerAdd] = func(a, b iris2.Word) (iris2.Word, error) { return a + b, nil }
	integerArithmeticOps[IntegerSubtract] = func(a, b iris2.Word) (iris2.Word, error) { return a - b, nil }
	integerArithmeticOps[IntegerMultiply] = func(a, b iris2.Word) (iris2.Word, error) { return a * b, nil }
	integerArithmeticOps[IntegerDivide] = func(num, denom iris2.Word) (iris2.Word, error) {
		if denom == 0 {
			return 0, fmt.Errorf("Divide by zero")
		} else {
			return num / denom, nil
		}
	}
	integerArithmeticOps[IntegerRemainder] = func(num, denom iris2.Word) (iris2.Word, error) {
		if denom == 0 {
			return 0, fmt.Errorf("Divide by zero")
		} else {
			return num % denom, nil
		}
	}
	integerArithmeticOps[IntegerShiftLeft] = func(a, b iris2.Word) (iris2.Word, error) { return a << uiris2.Word(b), nil }
	integerArithmeticOps[IntegerShiftRight] = func(a, b iris2.Word) (iris2.Word, error) { return a >> uiris2.Word(b), nil }
	integerArithmeticOps[IntegerAnd] = func(a, b iris2.Word) (iris2.Word, error) { return a & b, nil }
	integerArithmeticOps[IntegerOr] = func(a, b iris2.Word) (iris2.Word, error) { return a | b, nil }
	integerArithmeticOps[IntegerNot] = func(a, _ iris2.Word) (iris2.Word, error) { return ^a, nil }
	integerArithmeticOps[IntegerXor] = func(a, b iris2.Word) (iris2.Word, error) { return a ^ b, nil }
	integerArithmeticOps[IntegerAndNot] = func(a, b iris2.Word) (iris2.Word, error) { return a &^ b, nil }
}

func (this *Unit) Startup() error {
	if this.running {
		return fmt.Errorf("Arithmetic unit is already running")
	} else {
		this.running = true
		go this.body()
		go this.controlQuery()
		return nil
	}
}
func (this *Unit) controlQuery() {
	<-this.Control
	if err := this.shutdown(); err != nil {
		this.err <- err
	}
}
func (this *Unit) body() {
	for this.running {
		select {
		case op := <-this.operation:
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

func (this *Unit) shutdown() error {
	if !this.running {
		return fmt.Errorf("this unit is not currently running!")
	} else {
		this.running = false
		return nil
	}
}

type FloatUnit struct {
	running          bool
	err              chan error
	out              chan float64
	Result           <-chan float64
	Error            <-chan error
	Control          <-chan iris2.Word
	operation        <-chan iris2.Word
	source0, source1 <-chan float64
}

func NewFloatUnit(control, operation <-chan iris2.Word, source0, source1 <-chan float64) *FloatUnit {
	var this FloatUnit
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

func (this *FloatUnit) Startup() error {
	if this.running {
		return fmt.Errorf("Arithmetic unit is already running")
	} else {
		this.running = true
		go this.body()
		go this.controlQuery()
		return nil
	}
}

func (this *FloatUnit) controlQuery() {
	<-this.Control
	if err := this.shutdown(); err != nil {
		this.err <- err
	}
}

func (this *FloatUnit) body() {
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

func (this *FloatUnit) shutdown() error {
	if !this.running {
		return fmt.Errorf("this unit is not currently running!")
	} else {
		this.running = false
		return nil
	}
}
