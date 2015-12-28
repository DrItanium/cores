// instruction description and functions
package iris2

type instruction [32]byte

func (this instruction) Instruction() (instruction, error) {
	return this, nil
}
func (this instruction) Int() (word, error) {
	return nil, fmt.Errorf("Can't convert an instruction to an word")
}
func (this instruction) Float() (floatWord, error) {
	return nil, fmt.Errorf("Can't convert an instruction to an floatWord")
}
func (this instruction) Bool() (predicate, error) {
	return nil, fmt.Errorf("Can't convert an instruction to a bool")
}
