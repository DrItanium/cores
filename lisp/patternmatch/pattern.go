package patternmatch

import (
	"fmt"
	"github.com/DrItanium/cores/lisp"
	"reflect"
	"strings"
)

const (
	SinglefieldVariablePrefix = "?"
	MultifieldVariablePrefix  = "$?"
)

type SinglefieldVariable lisp.Atom

func (this SinglefieldVariable) String() string {
	return fmt.Sprintf("?%s", string(this))
}
func (this SinglefieldVariable) Type() reflect.Type {
	return reflect.TypeOf(this)
}

func (this SinglefieldVariable) Value() interface{} {
	return this
}

type MultifieldVariable lisp.Atom

func (this MultifieldVariable) String() string {
	return fmt.Sprintf("$?%s", string(this))
}
func (this MultifieldVariable) Type() reflect.Type {
	return reflect.TypeOf(this)
}

func (this MultifieldVariable) Value() interface{} {
	return this
}
func extract(l lisp.List, index int) (interface{}, error) {
	if index >= len(l) {
		return nil, fmt.Errorf("Index %d is out of range for the given list!", index)
	}
	return l[index], nil
}
func extractOfType(l lisp.List, index int, t reflect.Type) (interface{}, error) {
	if val, err := extract(l, index); err != nil {
		return nil, err
	} else if reflect.TypeOf(val).Name() != t.Name() {
		return nil, fmt.Errorf("Indexed type is not of type %t, instead it is of type %t", t, reflect.TypeOf(val))
	} else {
		return val, nil
	}
}

var stringType = reflect.TypeOf("")
var atomType = reflect.TypeOf(lisp.Atom(nil))

func extractString(l lisp.List, index int) (string, error) {
	if result, err := extractOfType(l, index, stringType); err != nil {
		return "", err
	} else {
		return result.(string), nil
	}
}

func extractStringAndCheckPrefix(l lisp.List, index int, prefix string) (string, error) {
	if str, err := extractString(l, index); err != nil {
		return str, err
	} else if !strings.HasPrefix(str, prefix) {
		return str, fmt.Errorf("Extracted value (%s) does not start with %s", str, prefix)
	} else {
		return str[len(prefix):], nil
	}
}

func extractSinglefieldArgument(l lisp.List, index int) (SinglefieldVariable, error) {
	if str, err := extractStringAndCheckPrefix(l, index, SinglefieldVariablePrefix); err != nil {
		return nil, err
	} else {
		return SinglefieldVariable([]byte(str)), nil
	}
}

func extractMultifieldArgument(l lisp.List, index int) (MultifieldVariable, error) {
	if str, err := extractStringAndCheckPrefix(l, index, MultifieldVariablePrefix); err != nil {
		return nil, err
	} else {
		return MultifieldVariable([]byte(str)), nil
	}
}

func extractAtom(l lisp.List, index int) (lisp.Atom, error) {
	if result, err := extractOfType(l, index, atomType); err != nil {
		return nil, err
	} else {
		return result.(lisp.Atom), nil
	}
}
