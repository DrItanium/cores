package arithmetic

import (
	"fmt"
)

type Unit struct {
	running                     bool
	err                         chan error
	out                         chan int64
	Result                      <-chan int64
	Error                       <-chan error
	Control                     <-chan int64
	operation, source0, source1 <-chan int64
}

func New(control, operation, source0, source1 <-chan int64) *Unit {
	var this Unit
	this.err = make(chan error)
	this.out = make(chan int64)
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

var integerArithmeticOps [IntegerOpCount]func(int64, int64) (int64, error)

func init() {
	integerArithmeticOps[IntegerAdd] = func(a, b int64) (int64, error) { return a + b, nil }
	integerArithmeticOps[IntegerSubtract] = func(a, b int64) (int64, error) { return a - b, nil }
	integerArithmeticOps[IntegerMultiply] = func(a, b int64) (int64, error) { return a + b, nil }
	integerArithmeticOps[IntegerDivide] = func(num, denom int64) (int64, error) {
		if denom == 0 {
			return 0, fmt.Errorf("Divide by zero")
		} else {
			return num / denom, nil
		}
	}
	integerArithmeticOps[IntegerRemainder] = func(num, denom int64) (int64, error) {
		if denom == 0 {
			return 0, fmt.Errorf("Divide by zero")
		} else {
			return num % denom, nil
		}
	}
	integerArithmeticOps[IntegerShiftLeft] = func(a, b int64) (int64, error) { return a << uint64(b), nil }
	integerArithmeticOps[IntegerShiftRight] = func(a, b int64) (int64, error) { return a >> uint64(b), nil }
	integerArithmeticOps[IntegerAnd] = func(a, b int64) (int64, error) { return a & b, nil }
	integerArithmeticOps[IntegerOr] = func(a, b int64) (int64, error) { return a | b, nil }
	integerArithmeticOps[IntegerNot] = func(a, _ int64) (int64, error) { return ^a, nil }
	integerArithmeticOps[IntegerXor] = func(a, b int64) (int64, error) { return a ^ b, nil }
	integerArithmeticOps[IntegerAndNot] = func(a, b int64) (int64, error) { return a &^ b, nil }
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
