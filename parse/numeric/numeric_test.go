package numeric

import (
	"testing"
)

func Test_HexNumber_0(t *testing.T) {
	if !IsHexNumber("0xFDED") {
		t.Error("0xFDED was not identified as a hex number!")
	} else {
		t.Log("0xFDED was identified as a hex number!")
	}
}

func Test_HexNumber_1(t *testing.T) {
	if IsHexNumber("0xFDEDJJ") {
		t.Error("0xFDEDJJ was identified as a hex number!")
	} else {
		t.Log("0xFDEDJJ was identified as not being a hex number!")
	}
}
