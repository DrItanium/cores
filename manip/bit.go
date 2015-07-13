// bit manipulation operations
package manip

func Mask(value, mask, shift uint64) uint64 {
	return (value & mask) >> shift
}

func BitsSet(value, mask, shift uint64) bool {
	return Mask(value, mask, shift) != 0
}
