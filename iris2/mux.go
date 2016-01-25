// Multiplexer circuit
package iris2

import (
	"fmt"
)

type Mux struct {
	running     bool
	sources     []<-chan interface{}
	selector    chan Word
	destination chan interface{}
	err         chan error

	Control     <-chan Word
	Select      chan<- Word
	Destination <-chan interface{}
	Error       <-chan error
}

func NewMux(control chan<- Word) *Mux {
	var mux Mux
	mux.err = make(chan error)
	mux.destination = make(chan interface{})
	mux.selector = make(chan Word)
	mux.Error = mux.err
	mux.Destination = mux.destination
	mux.Control = control
	mux.Select = mux.selector
	return &mux
}
func (this *Mux) AddSource(src <-chan interface{}) {
	this.sources = append(this.sources, src)
}
func (this *Mux) body() {
	for this.running {
		select {
		case index, more := <-this.selector:
			if more {
				if index >= Word(len(this.sources)) {
					this.err <- fmt.Errorf("Selected non existent source: %d", index)
				} else if index < 0 {
					this.err <- fmt.Errorf("Select source %d is less than zero!", index)
				} else {
					this.destination <- <-this.sources[index]
				}
			}
		}
	}
}

func (this *Mux) queryControl() {
	<-this.Control
	if err := this.shutdown(); err != nil {
		this.err <- err
	}
}

func (this *Mux) shutdown() error {
	if !this.running {
		return fmt.Errorf("Attempted to shutdown a multiplexer which isn't running")
	} else {
		this.running = false
		close(this.selector)
		return nil
	}
}

func (this *Mux) Startup() error {
	if this.running {
		return fmt.Errorf("Given multiplexer is already running!")
	} else {
		this.running = true
		go this.body()
		go this.queryControl()
		return nil
	}
}
