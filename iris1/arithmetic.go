package iris1

import "fmt"

const (
	// arithmetic operations
	ArithmeticOpAdd = iota
	ArithmeticOpSub
	ArithmeticOpMul
	ArithmeticOpDiv
	ArithmeticOpRem
	ArithmeticOpShiftLeft
	ArithmeticOpShiftRight
	ArithmeticOpBinaryAnd
	ArithmeticOpBinaryOr
	ArithmeticOpBinaryNot
	ArithmeticOpBinaryXor
	ArithmeticOpIncrement
	ArithmeticOpDecrement
	ArithmeticOpDouble
	ArithmeticOpHalve
	ArithmeticOpAddImmediate
	ArithmeticOpSubImmediate
	ArithmeticOpMulImmediate
	ArithmeticOpDivImmediate
	ArithmeticOpRemImmediate
	ArithmeticOpShiftLeftImmediate
	ArithmeticOpShiftRightImmediate
	// always last
	ArithmeticOpCount
)

var unimplementedBinaryOp = func(a, b Word) (Word, error) { return 0, fmt.Errorf("Operation not implemented!") }

type ArithmeticOp struct {
	ImmediateForm bool
	fn            func(Word, Word) (Word, error)
}

func (this *ArithmeticOp) Invoke(first, second Word) (Word, error) {
	return this.fn(first, second)

}
func basicDivOperation(a, b Word, fn func(Word, Word) Word) (Word, error) {
	if b == 0 {
		return 0, fmt.Errorf("Divide by zero error!")
	} else {
		return fn(a, b), nil
	}
}
func div(a, b Word) (Word, error) {
	if b == 0 {
		return 0, fmt.Errorf("Divide by zero error!")
	} else if b == 1 {
		return a, nil
	} else {
		return a / b, nil
	}
}
func rem(a, b Word) (Word, error) {
	if b == 0 {
		return 0, fmt.Errorf("Divide by zero error!")
	} else if b == 1 {
		return 0, nil
	} else {
		return a % b, nil
	}
}

var arithmeticOps [32]ArithmeticOp

func init() {
	if ArithmeticOpCount > 32 {
		panic("Too many arithmetic operations defined! Programmer failure!")
	} else {
		for i := 0; i < 32; i++ {
			arithmeticOps[i].ImmediateForm = false
			arithmeticOps[i].fn = unimplementedBinaryOp
		}
		// setup the table itself
		arithmeticOps[ArithmeticOpAdd].fn = func(a, b Word) (Word, error) { return a + b, nil }                                                                          // add
		arithmeticOps[ArithmeticOpSub].fn = func(a, b Word) (Word, error) { return a - b, nil }                                                                          // sub
		arithmeticOps[ArithmeticOpMul].fn = func(a, b Word) (Word, error) { return a * b, nil }                                                                          // mul
		arithmeticOps[ArithmeticOpDiv].fn = div                                                                                                                          // divide
		arithmeticOps[ArithmeticOpRem].fn = rem                                                                                                                          // remainder
		arithmeticOps[ArithmeticOpShiftLeft].fn = func(a, b Word) (Word, error) { return a << b, nil }                                                                   // shift left
		arithmeticOps[ArithmeticOpShiftRight].fn = func(a, b Word) (Word, error) { return a >> b, nil }                                                                  // shift right
		arithmeticOps[ArithmeticOpBinaryAnd].fn = func(a, b Word) (Word, error) { return a & b, nil }                                                                    // binary and
		arithmeticOps[ArithmeticOpBinaryOr].fn = func(a, b Word) (Word, error) { return a | b, nil }                                                                     // binary or
		arithmeticOps[ArithmeticOpBinaryNot].fn = func(a, _ Word) (Word, error) { return ^a, nil }                                                                       // unary not
		arithmeticOps[ArithmeticOpBinaryXor].fn = func(a, b Word) (Word, error) { return a ^ b, nil }                                                                    // binary xor
		arithmeticOps[ArithmeticOpIncrement].fn = func(a, _ Word) (Word, error) { return arithmeticOps[ArithmeticOpAdd].Invoke(a, 1) }                                   // increment
		arithmeticOps[ArithmeticOpDecrement].fn = func(a, _ Word) (Word, error) { return arithmeticOps[ArithmeticOpSub].Invoke(a, 1) }                                   // decrement
		arithmeticOps[ArithmeticOpDouble].fn = func(a, _ Word) (Word, error) { return arithmeticOps[ArithmeticOpAdd].Invoke(a, a) }                                      // double
		arithmeticOps[ArithmeticOpHalve].fn = func(a, _ Word) (Word, error) { return arithmeticOps[ArithmeticOpDiv].Invoke(a, 2) }                                       // halve
		arithmeticOps[ArithmeticOpAddImmediate] = ArithmeticOp{true, func(a, b Word) (Word, error) { return arithmeticOps[ArithmeticOpAdd].Invoke(a, b) }}               // immediate form of add
		arithmeticOps[ArithmeticOpSubImmediate] = ArithmeticOp{true, func(a, b Word) (Word, error) { return arithmeticOps[ArithmeticOpSub].Invoke(a, b) }}               // immediate form of sub
		arithmeticOps[ArithmeticOpMulImmediate] = ArithmeticOp{true, func(a, b Word) (Word, error) { return arithmeticOps[ArithmeticOpMul].Invoke(a, b) }}               // immediate form of mul
		arithmeticOps[ArithmeticOpDivImmediate] = ArithmeticOp{true, func(a, b Word) (Word, error) { return arithmeticOps[ArithmeticOpDiv].Invoke(a, b) }}               // immediate form of div
		arithmeticOps[ArithmeticOpRemImmediate] = ArithmeticOp{true, func(a, b Word) (Word, error) { return arithmeticOps[ArithmeticOpRem].Invoke(a, b) }}               // immediate form of rem
		arithmeticOps[ArithmeticOpShiftLeftImmediate] = ArithmeticOp{true, func(a, b Word) (Word, error) { return arithmeticOps[ArithmeticOpShiftLeft].Invoke(a, b) }}   // immediate form of shift left
		arithmeticOps[ArithmeticOpShiftRightImmediate] = ArithmeticOp{true, func(a, b Word) (Word, error) { return arithmeticOps[ArithmeticOpShiftRight].Invoke(a, b) }} // immediate form of shift right
	}
}
func arithmetic(core *Core, inst *DecodedInstruction) error {
	var arg0, arg1 Word
	var err error
	dest := inst.Data[0]
	arg0 = core.Register(inst.Data[1])
	result := Word(0)
	invoke := arithmeticOps[inst.Op]
	if invoke.ImmediateForm {
		arg1 = inst.Immediate()
	} else {
		arg1 = core.Register(inst.Data[2])
	}
	if result, err = arithmeticOps[inst.Op].Invoke(arg0, arg1); err != nil {
		return err
	} else {
		return core.SetRegister(dest, result)
	}
}
