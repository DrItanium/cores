// Multiplxer circuit
package mux

import (
	"fmt"
	"github.com/DrItanium/cores/iris2"
)

type Unit struct {
	running     bool
	sources     []<-chan interface{}
	selector    <-chan iris2.Word
	Control     <-chan iris2.Word
	destination chan<- interface{}
	err         chan error
	Error       <-chan error
}

func New(control, selector <-chan iris2.Word, dest chan<- interface{}, src0 <-chan interface{}, srcs ...<-chan interface{}) *Unit {
	var mux Unit
	mux.err = make(chan error)
	mux.Error = mux.err
	mux.sources = []<-chan interface{}{src0}
	mux.sources = append(mux.sources, srcs...)
	mux.selector = selector
	mux.destination = dest
	mux.Control = control
	return &mux
}

func (this *Unit) body() {
	for this.running {
		select {
		case index := <-this.selector:
			if index >= iris2.Word(len(this.sources)) {
				this.err <- fmt.Errorf("Selected non existent source: %d", index)
			} else if index < 0 {
				this.err <- fmt.Errorf("Select source %d is less than zero!", index)
			} else {
				this.destination <- <-this.sources[index]
			}
		}
	}
}

func (this *Unit) queryControl() {
	<-this.Control
	if err := this.shutdown(); err != nil {
		this.err <- err
	}
}

func (this *Unit) shutdown() error {
	if !this.running {
		return fmt.Errorf("Attempted to shutdown a multiplexer which isn't running")
	} else {
		this.running = false
		return nil
	}
}

func (this *Unit) Startup() error {
	if this.running {
		return fmt.Errorf("Given multiplexer is already running!")
	} else {
		this.running = true
		go this.body()
		go this.queryControl()
		return nil
	}
}
