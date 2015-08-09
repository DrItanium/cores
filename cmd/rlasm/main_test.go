package main

import (
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
