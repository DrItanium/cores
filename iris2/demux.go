// Multiplexer circuit
package iris2

import (
	"fmt"
)

type Demux struct {
	running      bool
	destinations []chan<- interface{}
	selector     <-chan Word
	Control      <-chan Word
	source       <-chan interface{}
	err          chan error
	Error        <-chan error
}

func NewDemux(control, selector <-chan Word, src <-chan interface{}) *Demux {
	var mux Demux
	mux.err = make(chan error)
	mux.Error = mux.err
	mux.selector = selector
	mux.source = src
	mux.Control = control
	return &mux
}

func (this *Demux) AddDestination(dest chan<- interface{}) {
	this.destinations = append(this.destinations, dest)
}

func (this *Demux) body() {
	for this.running {
		select {
		case index := <-this.selector:
			if index >= Word(len(this.destinations)) {
				this.err <- fmt.Errorf("Selected non existent source: %d", index)
			} else if index < 0 {
				this.err <- fmt.Errorf("Select source %d is less than zero!", index)
			} else {
				this.destinations[index] <- <-this.source
			}
		}
	}
}

func (this *Demux) queryControl() {
	<-this.Control
	if err := this.shutdown(); err != nil {
		this.err <- err
	}
}

func (this *Demux) shutdown() error {
	if !this.running {
		return fmt.Errorf("Attempted to shutdown a multiplexer which isn't running")
	} else {
		this.running = false
		return nil
	}
}

func (this *Demux) Startup() error {
	if this.running {
		return fmt.Errorf("Given multiplexer is already running!")
	} else {
		this.running = true
		go this.body()
		go this.queryControl()
		return nil
	}
}
