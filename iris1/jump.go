package iris1

import "fmt"

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

var jumpFunctions = [32]func(core *Core, inst *DecodedInstruction) error{
	func(core *Core, inst *DecodedInstruction) error { // branch immediate
		return unconditionalJump(core, inst.Immediate(), false)
	},
	func(core *Core, inst *DecodedInstruction) error { // call immediate
		return unconditionalJump(core, inst.Immediate(), true)
	},
	func(core *Core, inst *DecodedInstruction) error { // branch register
		return unconditionalJump(core, core.Register(inst.Data[0]), false)
	},
	func(core *Core, inst *DecodedInstruction) error { // call register
		return unconditionalJump(core, core.Register(inst.Data[0]), true)
	},
	func(core *Core, inst *DecodedInstruction) error { // conditional branch immediate
		return conditionalJump(core, core.PredicateValue(inst.Data[0]), inst.Immediate(), false)
	},
	func(core *Core, inst *DecodedInstruction) error { // conditional call immediate
		return conditionalJump(core, core.PredicateValue(inst.Data[0]), inst.Immediate(), true)
	},
	func(core *Core, inst *DecodedInstruction) error { // conditional branch register
		return conditionalJump(core, core.PredicateValue(inst.Data[0]), core.Register(inst.Data[1]), false)
	},
	func(core *Core, inst *DecodedInstruction) error { // conditional call register
		return conditionalJump(core, core.PredicateValue(inst.Data[0]), core.Register(inst.Data[1]), true)
	},
	func(core *Core, inst *DecodedInstruction) error { // conditional branch immediate (false)
		return conditionalJump(core, !core.PredicateValue(inst.Data[0]), inst.Immediate(), false)
	},
	func(core *Core, inst *DecodedInstruction) error { // conditional call immediate (false)
		return conditionalJump(core, !core.PredicateValue(inst.Data[0]), inst.Immediate(), true)
	},
	func(core *Core, inst *DecodedInstruction) error { // conditional branch register (false)
		return conditionalJump(core, !core.PredicateValue(inst.Data[0]), core.Register(inst.Data[1]), false)
	},
	func(core *Core, inst *DecodedInstruction) error { // conditional call register (false)
		return conditionalJump(core, !core.PredicateValue(inst.Data[0]), core.Register(inst.Data[1]), true)
	},
	func(core *Core, inst *DecodedInstruction) error { // if then else branch pred true
		return selectNextAddress(core, core.PredicateValue(inst.Data[0]), core.Register(inst.Data[1]), core.Register(inst.Data[2]), false)
	},
	func(core *Core, inst *DecodedInstruction) error { // if then else branch pred false
		return selectNextAddress(core, !core.PredicateValue(inst.Data[0]), core.Register(inst.Data[1]), core.Register(inst.Data[2]), false)
	},
	func(core *Core, inst *DecodedInstruction) error { // if then else call pred true
		return selectNextAddress(core, core.PredicateValue(inst.Data[0]), core.Register(inst.Data[1]), core.Register(inst.Data[2]), true)
	},
	func(core *Core, inst *DecodedInstruction) error { // if then else call pred false
		return selectNextAddress(core, !core.PredicateValue(inst.Data[0]), core.Register(inst.Data[1]), core.Register(inst.Data[2]), true)
	},
	func(core *Core, inst *DecodedInstruction) error { // return
		return branch(core, core.Return(), false)
	},
	undefinedJumpFunction,
	undefinedJumpFunction,
	undefinedJumpFunction,
	undefinedJumpFunction,
	undefinedJumpFunction,
	undefinedJumpFunction,
	undefinedJumpFunction,
	undefinedJumpFunction,
	undefinedJumpFunction,
	undefinedJumpFunction,
	undefinedJumpFunction,
	undefinedJumpFunction,
	undefinedJumpFunction,
	undefinedJumpFunction,
	undefinedJumpFunction,
}

func init() {
	if JumpOpCount > 32 {
		panic("Too many jump operations defined!")
	}
}

func jump(core *Core, inst *DecodedInstruction) error {
	return jumpFunctions[inst.Op](core, inst)
}
