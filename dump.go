// dumper interface
package cores

type Dumper interface {
	Dump(output chan<- byte) error
}
