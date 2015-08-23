package iris1

import "testing"

func NewDecodedInstructionArithmetic(op, dest, src0, src1 byte) (*DecodedInstruction, error) {
	var di DecodedInstruction
	di.Group = InstructionGroupArithmetic
	di.Op = op
	di.Data[0] = dest
	di.Data[1] = src0
	di.Data[2] = src1
	return &di, nil
}
func Test_Add_1(t *testing.T) {
	if core, err := New(DefaultMemorySize); err != nil {
		t.Fatalf("Couldn't create core %s", err)
	} else if di, err := NewDecodedInstructionArithmetic(ArithmeticOpAdd, 32, TrueRegister, TrueRegister); err != nil {
		t.Errorf("Couldn't construct arithmetic instruction: %s", err)
	} else if err := core.Invoke(di); err != nil {
		t.Errorf("Execution failed: %s", err)
	} else if core.Register(di.Data[0]) != 2 {
		t.Errorf("Adding 1 and 1 did not yield 2!")
	}
}

func Test_Sub_1(t *testing.T) {
	if core, err := New(DefaultMemorySize); err != nil {
		t.Fatalf("Couldn't create core %s", err)
	} else if di, err := NewDecodedInstructionArithmetic(ArithmeticOpSub, 32, TrueRegister, TrueRegister); err != nil {
		t.Errorf("Couldn't construct arithmetic instruction: %s", err)
	} else if err := core.Invoke(di); err != nil {
		t.Errorf("Execution failed: %s", err)
	} else if core.Register(di.Data[0]) != 0 {
		t.Errorf("Subtracting 1 and 1 did not yield 0!")
	}
}

func Test_Mul_1(t *testing.T) {
	if core, err := New(DefaultMemorySize); err != nil {
		t.Fatalf("Couldn't create core %s", err)
	} else if di, err := NewDecodedInstructionArithmetic(ArithmeticOpMul, 32, UserRegisterBegin, UserRegisterBegin); err != nil {
		t.Errorf("Couldn't construct arithmetic instruction: %s", err)
	} else {
		if err := core.SetRegister(UserRegisterBegin, 2); err != nil {
			t.Errorf("Couldn't set register %d to 2: %s", UserRegisterBegin, err)
		} else if err := core.Invoke(di); err != nil {
			t.Errorf("Execution failed: %s", err)
		} else if core.Register(di.Data[0]) != 4 {
			t.Errorf("Multiplying 2 and 2 did not yield 4!")
		}
	}
}

func Test_Div_1(t *testing.T) {
	if core, err := New(DefaultMemorySize); err != nil {
		t.Fatalf("Couldn't create core %s", err)
	} else if di, err := NewDecodedInstructionArithmetic(ArithmeticOpDiv, UserRegisterBegin+1, UserRegisterBegin, UserRegisterBegin); err != nil {
		t.Errorf("Couldn't construct arithmetic instruction: %s", err)
	} else if err := core.SetRegister(UserRegisterBegin, 2); err != nil {
		t.Errorf("Couldn't set register %d to 2: %s", UserRegisterBegin, err)
	} else if err := core.SetRegister(di.Data[0], 44); err != nil {
		t.Errorf("Couldn't set register %d to 44: %s", di.Data[0], err)
	} else if err := core.Invoke(di); err != nil {
		t.Errorf("Execution failed: %s", err)
	} else if val := core.Register(di.Data[0]); val != 1 {
		t.Errorf("Dividing 2 and 2 did not yield 1, value is %d!", val)
	}
}

func Test_Div_2(t *testing.T) {
	if core, err := New(DefaultMemorySize); err != nil {
		t.Fatalf("Couldn't create core %s", err)
	} else if di, err := NewDecodedInstructionArithmetic(ArithmeticOpDiv, UserRegisterBegin+1, UserRegisterBegin, FalseRegister); err != nil {
		t.Errorf("Couldn't construct arithmetic instruction: %s", err)
	} else if err := core.SetRegister(UserRegisterBegin, 2); err != nil {
		t.Errorf("Couldn't set register %d to 2: %s", UserRegisterBegin, err)
	} else if err := core.SetRegister(di.Data[0], 44); err != nil {
		t.Errorf("Couldn't set register %d to 44: %s", di.Data[0], err)
	} else if err := core.Invoke(di); err == nil {
		t.Errorf("Dividing by zero did not cause execution to fail!")
	} else if err.Error() != DivideByZeroMessage {
		t.Errorf("Execution failed but divide by zero was not the culprit: %s", err)
	} else {
		t.Logf("Dividing by zero did cause execution to fail: %s", err)
	}
}

func Test_Rem_1(t *testing.T) {
	if core, err := New(DefaultMemorySize); err != nil {
		t.Fatalf("Couldn't create core %s", err)
	} else if di, err := NewDecodedInstructionArithmetic(ArithmeticOpRem, UserRegisterBegin+1, UserRegisterBegin, UserRegisterBegin); err != nil {
		t.Errorf("Couldn't construct arithmetic instruction: %s", err)
	} else if err := core.SetRegister(UserRegisterBegin, 2); err != nil {
		t.Errorf("Couldn't set register %d to 2: %s", UserRegisterBegin, err)
	} else if err := core.SetRegister(di.Data[0], 44); err != nil {
		t.Errorf("Couldn't set register %d to 44: %s", di.Data[0], err)
	} else if err := core.Invoke(di); err != nil {
		t.Errorf("Execution failed: %s", err)
	} else if val := core.Register(di.Data[0]); val != 0 {
		t.Errorf("Modulus 2 and 2 did not yield 1, value is %d!", val)
	}
}

func Test_Rem_2(t *testing.T) {
	if core, err := New(DefaultMemorySize); err != nil {
		t.Fatalf("Couldn't create core %s", err)
	} else if di, err := NewDecodedInstructionArithmetic(ArithmeticOpRem, UserRegisterBegin+1, UserRegisterBegin, FalseRegister); err != nil {
		t.Errorf("Couldn't construct arithmetic instruction: %s", err)
	} else if err := core.SetRegister(UserRegisterBegin, 2); err != nil {
		t.Errorf("Couldn't set register %d to 2: %s", UserRegisterBegin, err)
	} else if err := core.SetRegister(di.Data[0], 44); err != nil {
		t.Errorf("Couldn't set register %d to 44: %s", di.Data[0], err)
	} else if err := core.Invoke(di); err == nil {
		t.Errorf("Modulus by zero did not cause execution to fail!")
	} else if err.Error() != DivideByZeroMessage {
		t.Errorf("Execution failed but modulus by zero was not the culprit: %s", err)
	} else {
		t.Logf("Modulus by zero did cause execution to fail: %s", err)
	}
}
