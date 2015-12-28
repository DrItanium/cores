package iris2

import "fmt"

const (
	// Compare operations
	CompareOpEq = iota
	CompareOpEqAnd
	CompareOpEqOr
	CompareOpEqXor
	CompareOpNeq
	CompareOpNeqAnd
	CompareOpNeqOr
	CompareOpNeqXor
	CompareOpLessThan
	CompareOpLessThanAnd
	CompareOpLessThanOr
	CompareOpLessThanXor
	CompareOpGreaterThan
	CompareOpGreaterThanAnd
	CompareOpGreaterThanOr
	CompareOpGreaterThanXor
	CompareOpLessThanOrEqualTo
	CompareOpLessThanOrEqualToAnd
	CompareOpLessThanOrEqualToOr
	CompareOpLessThanOrEqualToXor
	CompareOpGreaterThanOrEqualTo
	CompareOpGreaterThanOrEqualToAnd
	CompareOpGreaterThanOrEqualToOr
	CompareOpGreaterThanOrEqualToXor

	CompareOpCount
)

func init() {
	if CompareOpCount > 32 {
		panic("Too many compare operations defined!")
	}
}

type compareOpCombine func(bool, bool) (bool, error)

const (
	CombineNone = iota
	CombineAnd
	CombineOr
	CombineXor
	CombineError
)

var combineOps = []compareOpCombine{
	func(_, n bool) (bool, error) { return n, nil },
	func(o, n bool) (bool, error) { return o && n, nil },
	func(o, n bool) (bool, error) { return o || n, nil },
	func(o, n bool) (bool, error) { return o != n, nil },
	func(_, _ bool) (bool, error) { return false, fmt.Errorf("invalid combine operation!") },
}

const (
	CompareBodyEq = iota
	CompareBodyNeq
	CompareBodyLt
	CompareBodyGt
	CompareBodyLe
	CompareBodyGe
	CompareBodyError
)

type compareOpBody func(word, word) (bool, error)

var bodyOps = []compareOpBody{
	func(a, b word) (bool, error) { return a == b, nil },
	func(a, b word) (bool, error) { return a != b, nil },
	func(a, b word) (bool, error) { return a < b, nil },
	func(a, b word) (bool, error) { return a > b, nil },
	func(a, b word) (bool, error) { return a <= b, nil },
	func(a, b word) (bool, error) { return a >= b, nil },
	func(a, b word) (bool, error) { return false, fmt.Errorf("Invalid compare body op!") },
}

type compareOp struct {
	Body, Combine int // record the index
}

var errorCompareOp = compareOp{Body: CompareBodyError, Combine: CombineError}

func (this *compareOp) Invoke(oldVal bool, new0, new1 word) (bool, error) {
	if result, err := bodyOps[this.Body](new0, new1); err != nil {
		return false, err
	} else {
		return combineOps[this.Combine](oldVal, result)
	}
}

var compareOps = [32]compareOp{
	{CompareBodyEq, CombineNone},
	{CompareBodyEq, CombineAnd},
	{CompareBodyEq, CombineOr},
	{CompareBodyEq, CombineXor},
	{CompareBodyNeq, CombineNone},
	{CompareBodyNeq, CombineAnd},
	{CompareBodyNeq, CombineOr},
	{CompareBodyNeq, CombineXor},
	{CompareBodyLt, CombineNone},
	{CompareBodyLt, CombineAnd},
	{CompareBodyLt, CombineOr},
	{CompareBodyLt, CombineXor},
	{CompareBodyGt, CombineNone},
	{CompareBodyGt, CombineAnd},
	{CompareBodyGt, CombineOr},
	{CompareBodyGt, CombineXor},
	{CompareBodyLe, CombineNone},
	{CompareBodyLe, CombineAnd},
	{CompareBodyLe, CombineOr},
	{CompareBodyLe, CombineXor},
	{CompareBodyGe, CombineNone},
	{CompareBodyGe, CombineAnd},
	{CompareBodyGe, CombineOr},
	{CompareBodyGe, CombineXor},
	errorCompareOp,
	errorCompareOp,
	errorCompareOp,
	errorCompareOp,
	errorCompareOp,
	errorCompareOp,
	errorCompareOp,
	errorCompareOp,
}

func boolToword(val bool) word {
	if val {
		return 1
	} else {
		return 0
	}
}
func compare(core *Core, inst *DecodedInstruction) error {
	if val, err := compareOps[inst.Op].Invoke(core.PredicateValue(inst.Data[0]), core.Register(inst.Data[1]), core.Register(inst.Data[2])); err != nil {
		return err
	} else {
		return core.SetRegister(inst.Data[0], boolToword(val))
	}
}
