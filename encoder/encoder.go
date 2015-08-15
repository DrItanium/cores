package encoder

import (
	"github.com/DrItanium/cores/lisp"
	"io"
)

type Encoder interface {
	Encode(lisp.List, io.Writer) error
}
