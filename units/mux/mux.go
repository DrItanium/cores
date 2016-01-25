// Multiplxer circuit
package mux

import (
	"fmt"
)

type Multiplexer struct {
	running     bool
	sources     []<-chan interface{}
	selector    <-chan int64
	Control     <-chan int64
	destination chan<- interface{}
	err         chan error
	Error       <-chan error
}

func NewMultiplexer(control, selector <-chan int64, dest chan<- interface{}, src0 <-chan interface{}, srcs ...<-chan interface{}) *Multiplexer {
	var mux Multiplexer
	mux.err = make(chan error)
	mux.Error = mux.err
	mux.sources = []chan<- interface{}{src0}
	mux.sources = append(mux.sources, srcs...)
	mux.selector = selector
	mux.destination = dest
	mux.Control = control
	return &mux
}

func (this *Multiplexer) body() {
	for this.running {
		select {
		case index := <-this.selector:
			if index >= int64(len(this.sources0)) {
				this.err <- fmt.Errorf("Selected non existent source: %d", index)
			} else if index < 0 {
				this.err <- fmt.Errorf("Select source %d is less than zero!", index)
			} else {
				this.destination <- <-this.sources[index]
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
