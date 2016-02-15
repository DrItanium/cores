package iris16

import "testing"

func Test_TerminateCall(t *testing.T) {
	if core, err := New(); err != nil {
		t.Errorf("Couldn't create core: %s", err)
	} else if inst, err := NewDecodedInstruction(InstructionGroupMisc, MiscOpSystemCall, SystemCallTerminate, 0, 0); err != nil {
		t.Errorf("Couldn't create decoded instruction: %s", err)
	} else if err := core.Invoke(inst); err != nil {
		t.Errorf("Couldn't invoke system command: %s", err)
	} else if !core.terminateExecution {
		t.Errorf("Terminate system call didn't not tell core to terminate!")
	} else {
		t.Logf("Terminate system call did tell core to terminate!")
	}
}
