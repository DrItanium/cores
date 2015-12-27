// dummy package meant to catalog registrations for different machine backends
package registration

import (
	"github.com/DrItanium/cores/iris1"
	"github.com/DrItanium/cores/iris2"
)

// does nothing
func Register() {}
func init() {
	iris1.Register()
	iris2.Register()
}
