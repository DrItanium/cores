package iris1

import "fmt"

const (
	branchBitCallForm = iota
	branchBitReturnForm
	branchBitConditionalForm
	branchBitIfThenElseForm
	branchBitImmediateForm
	branchBitCount
)
const (
	branchBitReturnFormMask      = 1 << branchBitReturnForm
	branchBitIfThenElseFormMask  = 1 << branchBitIfThenElseForm
	branchBitCallFormMask        = 1 << branchBitCallForm
	branchBitImmediateFormMask   = 1 << branchBitImmediateForm
	branchBitConditionalFormMask = 1 << branchBitConditionalForm
)

func init() {
	if branchBitCount > 5 {
		panic(fmt.Sprintf("too many branch bits defined!"))
	}
}

var branchExtractionFunctions = map[byte][2]byte{
	branchBitReturnForm:      [2]byte{branchBitReturnFormMask, branchBitReturnForm},
	branchBitIfThenElseForm:  [2]byte{branchBitIfThenElseFormMask, branchBitIfThenElseForm},
	branchBitCallForm:        [2]byte{branchBitCallFormMask, branchBitCallForm},
	branchBitImmediateForm:   [2]byte{branchBitImmediateFormMask, branchBitImmediateForm},
	branchBitConditionalForm: [2]byte{branchBitConditionalFormMask, branchBitConditionalForm},
}

type branchBits byte

func boolToByte(value bool) byte {
	if value {
		return 1
	} else {
		return 0
	}
}
func (this *branchBits) setBit(index byte, value bool) {
	result := branchExtractionFunctions[index]
	mask, shift := branchBits(result[0]), branchBits(result[1])
	v := branchBits(boolToByte(value))
	*this = (*this &^ mask) | (v << shift)
}
func (this branchBits) extractBit(index byte) bool {
	result := branchExtractionFunctions[index]
	mask, shift := branchBits(result[0]), branchBits(result[1])
	return ((this & mask) >> shift) == 1
}
func (this branchBits) callForm() bool {
	return this.extractBit(branchBitCallForm)
}
func (this branchBits) ifThenElseForm() bool {
	return this.extractBit(branchBitIfThenElseForm)
}
func (this branchBits) immediateForm() bool {
	return this.extractBit(branchBitImmediateForm)
}
func (this branchBits) conditionalForm() bool {
	return this.extractBit(branchBitConditionalForm)
}
func (this branchBits) returnForm() bool {
	return this.extractBit(branchBitReturnForm)
}

func (this *branchBits) setReturnForm(value bool) {
	this.setBit(branchBitReturnForm, value)
}
func (this *branchBits) setCallForm(value bool) {
	this.setBit(branchBitCallForm, value)
}
func (this *branchBits) setConditionalForm(value bool) {
	this.setBit(branchBitConditionalForm, value)
}
func (this *branchBits) setImmediateForm(value bool) {
	this.setBit(branchBitImmediateForm, value)
}
func (this *branchBits) setIfThenElseForm(value bool) {
	this.setBit(branchBitIfThenElseForm, value)
}

func branch(core *Core, addr Word, call bool) error {
	if call {
		return core.Call(addr)
	} else {
		return core.SetRegister(InstructionPointer, addr)
	}
}
func selectNextAddress(core *Core, cond bool, onTrue, onFalse Word, call bool) error {
	core.advancePc = false
	var next Word
	if cond {
		next = onTrue
	} else {
		next = onFalse
	}
	return branch(core, next, call)
}
func conditionalJump(core *Core, cond bool, onTrue Word, call bool) error {
	return selectNextAddress(core, cond, onTrue, core.Register(InstructionPointer)+1, call)
}
func unconditionalJump(core *Core, addr Word, call bool) error {
	return branch(core, addr, call)
}
func undefinedJumpFunction(_ *Core, _ *DecodedInstruction) error {
	return fmt.Errorf("Illegal jump operation!")
}
func uncondOp(core *Core, call, ret, imm bool, inst *DecodedInstruction) error {
	var addr Word
	if ret {
		if imm {
			return fmt.Errorf("A return instruction combined with an immediate makes no sense")
		} else {
			addr = core.Return()
		}
	} else {
		// in the case of call and branch the same behavior will occur
		if imm {
			addr = inst.Immediate()
		} else {
			addr = core.Register(inst.Data[0])
		}
	}
	return branch(core, addr, call)
}
func condOp(core *Core, call, ret, imm bool, inst *DecodedInstruction) error {
	addr := core.Register(InstructionPointer) + 1
	cond := core.Register(inst.Data[0]) == 1
	shouldCall := call
	if ret {
		if imm {
			return fmt.Errorf("A return instruction combined with an immediate makes no sense")
		} else if cond {
			addr = core.Return()
		}
		// we shouldn't even get here if call and ret are both true so no need to check again
	} else {
		if cond {
			if imm {
				addr = inst.Immediate()
			} else {
				addr = core.Register(inst.Data[1])
			}
		} else {
			// it may turn out that the cond is false but we're a call instruction so don't call in this case
			shouldCall = false
		}
	}
	return branch(core, addr, shouldCall)
}
func ifThenElseOp(core *Core, call, ret, imm bool, inst *DecodedInstruction) error {
	if imm {
		return fmt.Errorf("The immediate flag should never be set with an if then else form")
	} else if ret {
		return fmt.Errorf("Can't mix return instructions and the if then else form")
	} else {
		// extract the predicate condition
		var addr Word
		if core.Register(inst.Data[0]) == 1 {
			addr = core.Register(inst.Data[1])
		} else {
			addr = core.Register(inst.Data[2])
		}
		return branch(core, addr, call)
	}
	return nil
}
func jump(core *Core, inst *DecodedInstruction) error {
	bb := branchBits(inst.Op)
	ret, ifthenelse, call, imm, cond := bb.returnForm(), bb.ifThenElseForm(), bb.callForm(), bb.immediateForm(), bb.conditionalForm()
	if ret && call {
		return fmt.Errorf("Instruction can't call and return at the same time!")
	} else if ifthenelse && cond {
		return fmt.Errorf("Can't have both \"if then else\" form and \"conditional\" form set in the same instruction!")
	} else if ifthenelse {
		return ifThenElseOp(core, call, ret, imm, inst)
	} else if cond {
		return condOp(core, call, ret, imm, inst)
	} else {
		return uncondOp(core, call, ret, imm, inst)
	}
}
