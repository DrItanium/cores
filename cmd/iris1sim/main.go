// iris1 simulator
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/DrItanium/cores/iris1"
)

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		fmt.Println("No input files provided!")
	} else {
		if core, err := iris1.New(); err != nil {
			fmt.Printf("ERROR: couldn't create a new iris1 core: %s\n", err)
		} else if inst, err := iris1.NewDecodedInstruction(iris1.InstructionGroupMisc, iris1.MiscOpSystemCall, iris1.SystemCallTerminate, 0, 0); err != nil {
			fmt.Printf("ERROR: couldn't create a new decoded instruction: %s\n", err)
		} else if err := core.Invoke(inst); err != nil {
			fmt.Printf("ERROR: execution failed of termination setup instruction: %s\n", err)
		} else {
			// install the program
			for !core.TerminateExecution() {
				// read the current instruction
				if err := core.ExecuteCurrentInstruction(); err != nil {
					fmt.Printf("ERROR during execution: %s\n", err)
				} else if err := core.AdvanceProgramCounter(); err != nil {
					fmt.Printf("ERROR during the advancement of the program counter: %s", err)
				}
			}
		}
	}
}

const (
	SixteenBitMemory      = 65536
	InstructionMemorySize = SixteenBitMemory * 4
	DataMemorySize        = SixteenBitMemory * 2
	ImageSize             = InstructionMemorySize + DataMemorySize
)

func installMemoryImage(core *iris1.Core, img []byte) error {
	if len(img) != ImageSize {
		return fmt.Errorf("provided image size is not %d", ImageSize)
	} else {
		code := img[:InstructionMemorySize]
		data := img[InstructionMemorySize:]
		if err := installCode(core, code); err != nil {
			return err
		} else if err := installData(core, data); err != nil {
			return err
		} else {
			return nil
		}
	}
}
func installData(core *iris1.Core, data []byte) error {
	if len(data) != DataMemorySize {
		return fmt.Errorf("provided data memory doesn't equal: %d", DataMemorySize)
	} else {
		slice := data
		for i := 0; i < SixteenBitMemory; i++ {
			section := slice[:2]
			if err := core.SetDataMemory(iris1.Word(i), iris1.Word(binary.LittleEndian.Uint16(section))); err != nil {
				return err
			}
			slice = slice[4:]
		}
		return nil
	}
}
func installCode(core *iris1.Core, data []byte) error {
	if len(data) != InstructionMemorySize {
		return fmt.Errorf("provided code memory doesn't equal: %d", InstructionMemorySize)
	} else {
		slice := data
		for i := 0; i < SixteenBitMemory; i++ {
			section := slice[:4]
			if err := core.SetCodeMemory(iris1.Word(i), iris1.Instruction(binary.LittleEndian.Uint32(section))); err != nil {
				return err
			}
			slice = slice[4:]
		}
		return nil
	}
}
