package pattern

type Match interface {
	Priority() int
	Invoke(value []byte) bool
}

type MatchBody func([]byte) bool

type MatchFunction struct {
	fn       MatchBody
	priority int
}

func (this *MatchFunction) Invoke(value []byte) bool {
	return this.fn(value)
}

func (this *MatchFunction) Priority() int {
	return this.priority
}

func NewMatch(fn MatchBody, priority int) Match {
	return &MatchFunction{fn: fn, priority: priority}
}
