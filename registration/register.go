// dummy package meant to catalog registrations for different machine backends
package registration

import (
	"github.com/DrItanium/cores/iris1"
	"github.com/DrItanium/cores/xand"
)

// does nothing
func Register() {}
func init() {
	iris1.Register()
	xand.Register()
}
