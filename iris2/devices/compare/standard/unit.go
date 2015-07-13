// standard compare unit
package standard

import (
	"fmt"
	"github.com/DrItanium/cores/manip"
	//"github.com/DrItanium/cores/iris2"
)

type CompareUnit struct {
	temporary bool
}
type binaryOperation func(iris2.Word, iris2.Word) iris2.Word

func New() *CompareUnit {
	var c CompareUnit

	return &c
}

const (
	get = iota
	set
	nop
	eq
	ne
	gt
	lt
	ge
	le
	compareOpCount
)

func init() {
	if compareOpCount > 16 {
		panic(fmt.Errorf("COMPARE UNIT INTERNAL ERROR: number of defined compare ops is %d when the max is 16", compareOpCount))
	}
}

type compareOperation struct {
	name               string
	hasCustomOperation bool
	customOp           func(*CompareUnit, []byte) ([]byte, error)
	op                 binaryOperation
}

const (
	flagMask      = 0xF0
	operationMask = 0x0F

	combineOperationOverwrite = iota
	combineOperationAnd
	combineOperationOr
	combineOperationXor
	combineOperationNor
	combineOperationNand
	combineOperationCount
)

func init() {
	if combineOperationCount > 16 {
		panic(fmt.Errorf("COMPARE UNIT INTENRAL ERROR: number of defined combine operations is %d when the max is 16", combineOperationCount))
	}
}

func getFlags(value byte) byte {
	return manip.Mask8(value, flagMask, 4)
}
func getOperation(value byte) byte {
	return manip.Mask8(value, operationMask, 0)
}

func (this *compareOperation) invoke(unit *CompareUnit, data []byte) ([]byte, error) {
	if this.hasCustomOperation {
		return this.customOp(unit, data)
	} else {
		if len(data) != 3 {
			return nil, fmt.Errorf("Expecting three bytes in the command stream, got %d", len(data))
		} else {
			header, src0, src1 := data[0], data[1] != 0, data[2] != 0
			q := getFlags(header)
			if fn, ok := combineOperations[q]; !ok {
				return nil, fmt.Errorf("Illegal compare combine operation %d", q)
			} else {
				unit.temporary = fn(this.op(src0, src1), unit.temporary) // perform the combining
				if unit.temporary {
					return []byte{1}, nil
				} else {
					return []byte{0}, nil
				}
			}
		}
	}
}

var combineOperations = map[byte]binaryOperation{
	combineOperationOverwrite: func(x, _ bool) bool { return x },
	combineOperationAnd:       func(x, y bool) bool { return x && y },
	combineOperationOr:        func(x, y bool) bool { return x || y },
	combineOperationXor:       func(x, y bool) bool { return x != y }, // golang doesn't have an logical xor operator but != is equivalent
	combineOperationNor:       func(x, y bool) bool { return !(x || y) },
	combineOperationNand:      func(x, y bool) bool { return !(x && y) },
}
var dispatchTable = map[int]*compareOperation{
	get: &compareOperation{"get", true, nil, nil},
	set: &compareOperation{"set", true, nil, nil},
	nop: &compareOperation{"nop", true, nil, nil},
	eq:  &compareOperation{"eq", false, nil, func(x, y iris2.Word) iris2.Word { return x == y }},
	ne:  &compareOperation{"ne", false, nil, func(x, y iris2.Word) iris2.Word { return x != y }},
	gt:  &compareOperation{"gt", false, nil, func(x, y iris2.Word) iris2.Word { return x > y }},
	lt:  &compareOperation{"lt", false, nil, func(x, y iris2.Word) iris2.Word { return x < y }},
	ge:  &compareOperation{"ge", false, nil, func(x, y iris2.Word) iris2.Word { return x >= y }},
	le:  &compareOperation{"le", false, nil, func(x, y iris2.Word) iris2.Word { return x <= y }},
}
