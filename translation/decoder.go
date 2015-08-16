// decode instructions into lisp representation
package translation

import (
	"github.com/DrItanium/cores/lisp"
	"io"
)

type Decoder interface {
	Decode(io.Reader) (lisp.List, error)
}
