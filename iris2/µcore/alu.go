package µcore

type Alu struct {
	vector [256]func(Port, Port, Port)
	in, out Link
}


func NewAlu(
