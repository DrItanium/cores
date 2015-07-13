package standard

import (
	"testing"
)

func Test_Nop(t *testing.T) {
	alu := New()
	result := alu.Send([]byte{nop})
	<-result
	alu.Terminate()
}
