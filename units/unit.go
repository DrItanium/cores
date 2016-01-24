package unit

type Unit interface {
	SetOp(op byte)
	SetSource(index byte, value interface{}) error
	Exec() error
	Result() interface{}
}
