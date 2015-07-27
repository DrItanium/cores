package iris1

import "testing"

func Test_Add_1(t *testing.T) {
	if core, err := New(); err != nil {
		t.Fatalf("Couldn't create core %s", err.Error())
	} else {
		var di DecodedInstruction
		di.Op = ArithmeticOpAdd
		di.Group = InstructionGroupArithmetic
		di.Data[0] = 32
		di.Data[1] = TrueRegister
		di.Data[2] = TrueRegister
		if err0 := core.InvokeExecution(&di); err0 != nil {
			t.Errorf("Execution exception %d", err0.Error())
		} else if core.Register(di.Data[0]) != 2 {
			t.Errorf("Adding 1 and 1 did not yield 2!")
		}
	}
}
