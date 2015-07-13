// common slice operations
package cores

import (
	"fmt"
	"reflect"
)

func IsContainer(slice interface{}) bool {
	typ := reflect.TypeOf(slice)
	k := typ.Kind()
	switch k {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return true
	default:
		return false
	}
}
func Empty(slice interface{}) (bool, error) {
	if IsContainer(slice) {
		return reflect.ValueOf(slice).Len() != 0, nil
	} else {
		return false, fmt.Errorf("Provided thing is not a container, it is a %t", reflect.TypeOf(slice))
	}
}

func First(slice interface{}) (interface{}, error) {
	switch reflect.TypeOf(slice).Kind() {
	case reflect.String, reflect.Array, reflect.Slice:
		val := reflect.ValueOf(slice)
		if val.Len() > 0 {
			return val.Index(0), nil
		} else {
			return nil, nil
		}
	default:
		return nil, fmt.Errorf("Provided thing is not a sliceable thing, it is a %t", reflect.TypeOf(slice))
	}
}
func Rest(slice interface{}) (interface{}, error) {
	switch reflect.TypeOf(slice).Kind() {
	case reflect.String, reflect.Array, reflect.Slice:
		val := reflect.ValueOf(slice)
		if val.Len() > 0 {
			return val.Slice(1, val.Len()), nil
		} else {
			return nil, nil
		}
	default:
		return nil, fmt.Errorf("Provided thing is not a sliceable thing, it is a %t", reflect.TypeOf(slice))
	}
}
