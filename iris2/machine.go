// next gen iris 64-bit implementation
package iris2

import (
	"fmt"
	"github.com/DrItanium/cores"
	"math"
	"reflect"
)

const (
	RegisterCount = 256

	RegisterKindWord = iota
	RegisterKindHalfWord
	RegisterKindQuarterWord
	RegisterKindByte
	RegisterKindBool
	RegisterKindFloat32
	RegisterKindFloat64
	RegisterKindLast

	MaxRegisterKinds      = 16
	NumberOfRegisterKinds = MaxRegisterKinds - RegisterKindLast
)

type Word uint64

type Register struct {
	readonly  bool
	unsigned  bool
	valueType byte
	value     Word
}

func (this *Register) Readonly() bool {
	return this.readonly
}
func (this *Register) Unsigned() bool {
	return this.unsigned
}
func (this *Register) Kind() byte {
	return this.valueType
}
func (this *Register) RawBits() Word {
	return this.value
}
func (this *Register) SetRawBits(value Word) {
	this.value = value
}
func (this *Register) SetValue(value interface{}) error {
	if this.readonly {
		return fmt.Errorf("Attempted to write to a readonly value")
	} else {
		switch t := value.(type) {
		case bool:
			this.unsigned = false
			this.valueType = RegisterKindBool
			if value.(bool) {
				this.value = 1
			} else {
				this.value = 0
			}
		case int8:
			this.value = Word(value.(int8))
			this.unsigned = false
			this.valueType = RegisterKindByte
		case uint8:
			this.value = Word(value.(uint8))
			this.unsigned = true
			this.valueType = RegisterKindByte
		case int16:
			this.unsigned = false
			this.value = Word(value.(int16))
			this.valueType = RegisterKindQuarterWord
		case uint16:
			this.unsigned = true
			this.value = Word(value.(uint16))
			this.valueType = RegisterKindQuarterWord
		case int32:
			this.unsigned = false
			this.value = Word(value.(int32))
			this.valueType = RegisterKindHalfWord
		case uint32:
			this.unsigned = true
			this.value = Word(value.(uint32))
			this.valueType = RegisterKindHalfWord
		case int64:
			this.unsigned = false
			this.value = Word(value.(int64))
			this.valueType = RegisterKindWord
		case uint64:
			this.unsigned = true
			this.value = Word(value.(uint64))
			this.valueType = RegisterKindWord
		case int:
			this.unsigned = false
			this.valueType = RegisterKindWord
			this.value = Word(value.(int))
		case uint:
			this.unsigned = true
			this.valueType = RegisterKindWord
			this.value = Word(value.(uint))
		case Word:
			this.unsigned = true
			this.valueType = RegisterKindWord
			this.value = value.(Word)
		case float32:
			this.unsigned = false
			this.valueType = RegisterKindFloat32
			this.value = Word(math.Float32bits(value.(float32)))
		case float64:
			this.unsigned = false
			this.valueType = RegisterKindFloat64
			this.value = Word(math.Float64bits(value.(float64)))
		default:
			return fmt.Errorf("Illegal value of type: %t", t)
		}
		return nil
	}
}

func (this *Register) GetValue() (interface{}, error) {
	switch this.valueType {
	case RegisterKindWord:
		if this.unsigned {
			return this.value, nil
		} else {
			return int64(this.value), nil
		}
	case RegisterKindHalfWord:
		if this.unsigned {
			return uint32(this.value), nil
		} else {
			return int32(this.value), nil
		}
	case RegisterKindQuarterWord:
		if this.unsigned {
			return uint16(this.value), nil
		} else {
			return int16(this.value), nil
		}
	case RegisterKindBool:
		return this.value != 0, nil
	case RegisterKindByte:
		if this.unsigned {
			return uint8(this.value), nil
		} else {
			return int8(this.value), nil
		}
	case RegisterKindFloat32:
		return math.Float32frombits(uint32(this.value)), nil
	case RegisterKindFloat64:
		return math.Float64frombits(uint64(this.value)), nil
	default:
		return nil, fmt.Errorf("Unknown type code: %d", this.valueType)
	}
}
func (this *Register) GetValueAs(valueType reflect.Type) (interface{}, error) {
	k := valueType.Kind()
	switch k {
	case reflect.Bool:
		return this.value != 0, nil
	case reflect.Int8:
		return int8(this.value), nil
	case reflect.Uint8:
		return uint8(this.value), nil
	case reflect.Int16:
		return int16(this.value), nil
	case reflect.Uint16:
		return uint16(this.value), nil
	case reflect.Int32:
		return int32(this.value), nil
	case reflect.Uint32:
		return uint32(this.value), nil
	case reflect.Int64:
		return int64(this.value), nil
	case reflect.Uint64:
		return uint64(this.value), nil
	case reflect.Float32:
		return math.Float32frombits(uint32(this.value)), nil
	case reflect.Float64:
		return math.Float64frombits(uint64(this.value)), nil
	default:
		return nil, fmt.Errorf("Illegal type: %s", k.String())
	}
}

const (
	RegisterZeroValue = iota
	RegisterOneValue
	RegisterFloat32Zero
	RegisterFloat32One
	RegisterFloat64Zero
	RegisterFloat64One
	RegisterIP
	RegisterDataStack
	RegisterCallStack
)

const (
	// DeviceTable layout
	_alu = iota
	_memoryController
	_compareUnit
	_branchUnit
	_numberOfDevices

	// register offsets
	ZeroRegister = iota
	OneRegister
	InstructionPointer
	StackPointer
)

type Core struct {
	Registers [RegisterCount]Register
	devices   [_numberOfDevices]cores.Device
}

func NewCore(alu, memController, compareUnit, branchUnit cores.Device) *Core {
	var c Core
	c.devices[_alu] = alu
	c.devices[_memoryController] = memController
	c.devices[_compareUnit] = compareUnit
	c.devices[_branchUnit] = branchUnit
	c.Registers[RegisterZeroValue].readonly = true
	c.Registers[RegisterOneValue].readonly = true
	c.Registers[RegisterOneValue].value = 1
	c.Registers[RegisterFloat32Zero].readonly = true
	c.Registers[RegisterFloat32Zero].value = Word(math.Float32bits(0.0))
	c.Registers[RegisterFloat32One].readonly = true
	c.Registers[RegisterFloat32One].value = Word(math.Float32bits(1.0))
	c.Registers[RegisterFloat64Zero].readonly = true
	c.Registers[RegisterFloat64Zero].value = Word(math.Float64bits(0.0))
	c.Registers[RegisterFloat64One].readonly = true
	c.Registers[RegisterFloat64One].value = Word(math.Float64bits(1.0))

	return &c
}
