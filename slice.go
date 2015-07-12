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
		return false, fmt.Errorf("Provided thing is not a slice, it is a %t", reflect.TypeOf(slice).Kind().String())
	}
}
