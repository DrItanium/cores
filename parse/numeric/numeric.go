// numeric parsing operations
package numeric

import (
	"strings"
)

func IsHexNumber(input string) bool {
	if !strings.HasPrefix(input, "0x") {
		return false
	} else {
		rest := input[2:]
		for _, r := range rest {
			switch r {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			case 'a', 'A', 'b', 'B', 'c', 'C', 'd', 'D', 'e', 'E', 'f', 'F':
			default:
				return false
			}
		}
		return true
	}
}

func IsBinaryNumber(input string) bool {
	if !strings.HasPrefix(input, "0b") {
		return false
	} else {
		rest := input[2:]
		for _, r := range rest {
			if r != '0' && r != '1' {
				return false
			}
		}
		return true
	}
}

func IsDecimalNumber(input string) bool {
	for _, r := range input {
		switch r {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
			return false
		}
	}
	return true
}
