package ogiris

import (
	"testing"
)

func addTest(t *testing.T) {
	inst := NewInstruction(GroupArithmetic, ArithmeticOpAdd, 2, 0, 1)
	c := new(Core)
	c.Gpr[0] = 1
	c.Gpr[1] = 2
	if err := c.Dispatch(inst); err != nil {
		t.Error(err)
	} else if val := c.Gpr[2]; val != 3 {
		t.Errorf("Adding 1 and 2 did not yield 3, it yielded %d", val)
	} else {
		t.Logf("Adding 1 and 2 did yield 3!")
	}
}
func TestDecoding(t *testing.T) {
	addTest(t)
}

func TestArithmetic(t *testing.T) {
	addTest(t)
}
