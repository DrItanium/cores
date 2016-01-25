package cond

type SignedUnit struct {
	running                     bool
	out                         chan int64
	err                         chan error
	Error                       <-chan error
	Result, Control             <-chan int64
	operation, source0, source1 <-chan int64
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

func NewSignedUnit(control, operation, source0, source1 <-chan int64) *SignedUnit {
	var s SignedUnit
	s.out = make(chan int64)
	s.err = make(chan error)
	s.operation = operation
	s.source0 = source0
	s.source1 = source1
	s.Error = err
	s.Control = control
	s.Result = out
	return &s
}
func buildIntegerFunction(cond func(int64, int64) bool) func(int64, int64) int64 {
	return func(a, b int64) int64 {
		if cond(a, b) {
			return 1
		} else {
			return 0
		}
	}
}

var dispatchTable [NumberOfCondStates]func(int64, int64) int64

func init() {
	dispatchTable[Equal] = buildIntegerFunction(func(a, b int64) bool { return a == b })
	dispatchTable[NotEqual] = buildIntegerFunction(func(a, b int64) bool { return a != b })
	dispatchTable[LessThan] = buildIntegerFunction(func(a, b int64) bool { return a < b })
	dispatchTable[GreaterThan] = buildIntegerFunction(func(a, b int64) bool { return a > b })
	dispatchTable[LessThanOrEqual] = buildIntegerFunction(func(a, b int64) bool { return a <= b })
	dispatchTable[GreaterThanOrEqual] = buildIntegerFunction(func(a, b int64) bool { return a >= b })
	dispatchTable[SelectSource0] = func(a, _ int64) int64 { return a }
	dispatchTable[SelectSource1] = func(_, b int64) int64 { return b }
	dispatchTable[PassTrue] = func(_, _ int64) int64 { return 1 }
	dispatchTable[PassFalse] = func(_, _ int64) int64 { return 0 }

}
func (this *SignedUnit) body() {
	var result bool
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

func (this *SignedUnit) controlQuery() {
	<-this.Control
	this.shutdown()
}

func (this *SignedUnit) shutdown() {
	this.running = false
}

func (this *SignedUnit) Startup() error {
	if this.running {
		return fmt.Errorf("Given conditional unit is already running!")
	} else {
		this.running = true
		go this.body()
		go this.controlQuery()
		return nil
	}
}
