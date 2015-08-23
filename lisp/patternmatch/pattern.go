package patternmatch

import (
	"fmt"
	"github.com/DrItanium/cores/lisp"
	"reflect"
	"strings"
)

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

func extractString(l lisp.List, index int) (string, error) {
	if result, err := extractOfType(l, index, stringType); err != nil {
		return "", err
	} else {
		return result.(string), nil
	}
}

func extractStringAndCheckPrefix(l lisp.List, index int, prefix string) (string, error) {
	if str, err := extractString(l, index); err != nil {
		return "", err
	} else if !strings.HasPrefix(str, prefix) {
		return "", fmt.Errorf("Extracted value (%s) does not start with %s", str, prefix)
	} else {
		return str[len(prefix):], nil
	}
}

func extractSinglefieldArgument(l lisp.List, index int) (string, error) {
	return extractStringAndCheckPrefix(l, index, "?")
}

func extractMultifieldArgument(l lisp.List, index int) (string, error) {
	return extractStringAndCheckPrefix(l, index, "$?")
}

func extractAtom(l lisp.List, index int) (lisp.Atom, error) {
	val := l[index]
	switch t := val.(type) {
	case lisp.Atom:
		return val.(lisp.Atom), nil
	default:
		return nil, fmt.Errorf("ERROR: provided index doesn't refer to a lisp.Atom (%t instead)!", t)
	}
}
