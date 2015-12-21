// dumper interface
package registration

type Dumper interface {
	Dump(output chan<- byte) error
}
