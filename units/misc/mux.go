// Multiplxer circuit
package misc

import (
	"fmt"
)

type Multiplexer struct {
	running      bool
	destinations []chan<- interface{}
	selector     <-chan int64
	Control      <-chan int64
	source       <-chan interface{}
	err          chan error
	Error        <-chan error
}

func NewMultiplexer(control, selector <-chan int64, source <-chan interface{}, dest0 chan<- interface{}, dest ...chan<- interface{}) *Multiplexer {
	var mux Multiplexer
	mux.err = make(chan error)
	mux.Error = mux.err
	mux.destinations = []chan<- interface{}{dest0}
	mux.destinations = append(mux.destinations, dest...)
	mux.selector = selector
	mux.source = source
	mux.Control = control
	return &mux
}

func (this *Multiplexer) body() {
	for this.running {
		select {
		case index := <-this.selector:
			if index >= int64(len(this.destinations)) {
				this.err <- fmt.Errorf("Selected non existent destination: %d", index)
			} else if index < 0 {
				this.err <- fmt.Errorf("Select destination %d is less than zero!", index)
			} else {
				this.destinations[index] <- <-this.source
			}
		}
	}
}

func (this *Multiplexer) queryControl() {
	<-this.Control
	if err := this.shutdown(); err != nil {
		this.err <- err
	}
}

func (this *Multiplexer) shutdown() error {
	if !this.running {
		return fmt.Errorf("Attempted to shutdown a multiplexer which isn't running")
	} else {
		this.running = false
		return nil
	}
}

func (this *Multiplexer) Startup() error {
	if this.running {
		return fmt.Errorf("Given multiplexer is already running!")
	} else {
		this.running = true
		go this.body()
		go this.queryControl()
		return nil
	}
}
