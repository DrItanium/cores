package main

import (
	"fmt"
	"testing"
)

var inputs = []struct {
	Value      string
	ShouldFail bool
}{
	{"(+ 1 2 3 4) (", true},
	{"(- 1 2 3 4) )", true},
	{"(* 1 2 3 4)", false},
	{"(deffunction a (x y) (* (+ x y) 2))", false},
}

type hexExecution struct {
	LittleEndian bool
	Value        Atom
	Check        string
}

func (this hexExecution) Invoke() error {
	if result := this.Value.HexValue(this.LittleEndian); result != this.Check {
		return fmt.Errorf("hex generation expected %s but got %s instead", this.Check, result)
	} else {
		return nil
	}
}

var hexInputs = []hexExecution{
	hexExecution{LittleEndian: true, Value: Atom{0xED, 0xFD}, Check: "0xFDED"},
	hexExecution{LittleEndian: false, Value: Atom{0xFD, 0xED}, Check: "0xFDED"},
}

func parseLispString(t *testing.T, index int) {
	if index >= len(inputs) {
		t.Errorf("Invalid parsing input %d", index)
		t.Fail()
	}

	if targ, err := ParseString(inputs[index].Value); err != nil {
		if inputs[index].ShouldFail {
			t.Logf("Expected failure occurred! %s", err)
		} else {
			t.Error(err)
			t.Fail()
		}
	} else {
		t.Logf("Parsed as: %s", targ)
	}
}
func Test_ParseLispStrings(t *testing.T) {
	for i := 0; i < len(inputs); i++ {
		parseLispString(t, i)
	}
}

func Test_HexParsing(t *testing.T) {
	for _, hv := range hexInputs {
		if err := hv.Invoke(); err != nil {
			t.Error(err)
			t.Fail()
		}
	}
}
