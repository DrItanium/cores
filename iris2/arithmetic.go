package iris2

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

var unimplementedBinaryOp = func(a, b word) (word, error) { return 0, fmt.Errorf("Operation not implemented!") }

type ArithmeticOp struct {
	ImmediateForm bool
	Fn            func(word, word) (word, error)
}

func (this *ArithmeticOp) Invoke(first, second word) (word, error) {
	return this.Fn(first, second)

}

const (
	DivideByZeroMessage = "Divide by zero error!"
)

type binOp func(word, word) (word, error)

func genericDivide(a, b word, denomIsOne, otherwise binOp) (word, error) {
	if b == 0 {
		return 0, fmt.Errorf(DivideByZeroMessage)
	} else if b == 1 {
		return denomIsOne(a, b)
	} else {
		return otherwise(a, b)
	}
}
func div(a, b word) (word, error) {
	return genericDivide(a, b, func(a, _ word) (word, error) { return a, nil }, func(a, b word) (word, error) { return a / b, nil })
}
func rem(a, b word) (word, error) {
	return genericDivide(a, b, func(_, _ word) (word, error) { return 0, nil }, func(a, b word) (word, error) { return a % b, nil })
}

var unimplementedArithmeticOp = ArithmeticOp{
	ImmediateForm: false,
	Fn:            unimplementedBinaryOp,
}
var arithmeticOps = [31]ArithmeticOp{
	ArithmeticOp{false, func(a, b word) (word, error) { return a + b, nil }},  // add
	ArithmeticOp{false, func(a, b word) (word, error) { return a - b, nil }},  // sub
	ArithmeticOp{false, func(a, b word) (word, error) { return a * b, nil }},  // mul
	ArithmeticOp{false, div},                                                  // divide
	ArithmeticOp{false, rem},                                                  // remainder
	ArithmeticOp{false, func(a, b word) (word, error) { return a << b, nil }}, // shift left
	ArithmeticOp{false, func(a, b word) (word, error) { return a >> b, nil }}, // shift right
	ArithmeticOp{false, func(a, b word) (word, error) { return a & b, nil }},  // binary and
	ArithmeticOp{false, func(a, b word) (word, error) { return a | b, nil }},  // binary or
	ArithmeticOp{false, func(a, _ word) (word, error) { return ^a, nil }},     // unary not
	ArithmeticOp{false, func(a, b word) (word, error) { return a ^ b, nil }},  // binary xor
	ArithmeticOp{false, func(a, _ word) (word, error) { return a + 1, nil }},  // increment
	ArithmeticOp{false, func(a, _ word) (word, error) { return a - 1, nil }},  // decrement
	ArithmeticOp{false, func(a, _ word) (word, error) { return a + a, nil }},  // double
	ArithmeticOp{false, func(a, _ word) (word, error) { return a / 2, nil }},  // halve
	ArithmeticOp{true, func(a, b word) (word, error) { return a + b, nil }},   // immediate form of add
	ArithmeticOp{true, func(a, b word) (word, error) { return a - b, nil }},   // immediate form of sub
	ArithmeticOp{true, func(a, b word) (word, error) { return a * b, nil }},   // immediate form of mul
	ArithmeticOp{true, div},                                                   // immediate form of div
	ArithmeticOp{true, rem},                                                   // immediate form of rem
	ArithmeticOp{true, func(a, b word) (word, error) { return a << b, nil }},  // immediate form of shift left
	ArithmeticOp{true, func(a, b word) (word, error) { return a >> b, nil }},  // immediate form of shift right
	unimplementedArithmeticOp,
	unimplementedArithmeticOp,
	unimplementedArithmeticOp,
	unimplementedArithmeticOp,
	unimplementedArithmeticOp,
	unimplementedArithmeticOp,
	unimplementedArithmeticOp,
	unimplementedArithmeticOp,
	unimplementedArithmeticOp,
}

func init() {
	if ArithmeticOpCount > 32 {
		panic("Too many arithmetic operations defined! Programmer failure!")
	}
}
func arithmetic(core *Core, inst *DecodedInstruction) error {
	var arg0, arg1 word
	var err error
	dest := inst.Data[0]
	arg0 = core.Register(inst.Data[1])
	result := word(0)
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
