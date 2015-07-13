// Device generic code
package cores

type Packet struct {
	Value []byte
	Error error
}

func (this *Packet) HasData() bool {
	return len(this.Value) > 0
}
func (this *Packet) HasError() bool {
	return this.Error != nil
}
func (this *Packet) First() byte {
	return this.Value[0]
}

func (this *Packet) Rest() []byte {
	return this.Value[1:]
}

type Device interface {
	Send(value []byte) chan Packet
	Terminate()
}
