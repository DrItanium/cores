// the description of the internal µcore which is a simple SISD core
package µcore

// floating point numbers are stored in words
type Word uint64
type Instruction Word
type Port chan Word
type Packet struct {
	Output []Port
	Input  []Port
}
type Link chan Packet

type Core interface {
	Output() Link
	Input() Link
}
