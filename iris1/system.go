package iris1

const (
	// Misc operations
	MiscOpSystemCall = iota
	// System commands
	SystemCommandTerminate = iota
	SystemCommandPanic
	SystemCommandCount
)

func init() {
	if SystemCommandCount > 256 {
		panic("Too many system commands defined!")
	}
}
