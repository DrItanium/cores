// next gen iris 64-bit implementation
package iris2

import (
	"fmt"
	"math"
	"reflect"
)

type Word uint64

type Register struct {
	value Word
}

func (this Register) Bool() bool {
	return this.value != 0
}
func (this Register) RawBits() Word {
	return this.value
}
func (this *Register) SetRawBits(value Word) {
	this.value = value
}
func (this *Register) SetValue(value interface{}) error {
	switch t := value.(type) {
	case bool:
		if value.(bool) {
			this.value = 1
		} else {
			this.value = 0
		}
	case int8:
		this.value = Word(value.(int8))
	case uint8:
		this.value = Word(value.(uint8))
	case int16:
		this.value = Word(value.(int16))
	case uint16:
		this.value = Word(value.(uint16))
	case int32:
		this.value = Word(value.(int32))
	case uint32:
		this.value = Word(value.(uint32))
	case int64:
		this.value = Word(value.(int64))
	case uint64:
		this.value = Word(value.(uint64))
	case int:
		this.value = Word(value.(int))
	case uint:
		this.value = Word(value.(uint))
	case Word:
		this.value = value.(Word)
	case float32:
		this.value = Word(math.Float32bits(value.(float32)))
	case float64:
		this.value = Word(math.Float64bits(value.(float64)))
	default:
		return fmt.Errorf("Illegal value of type: %t", t)
	}
	return nil
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
