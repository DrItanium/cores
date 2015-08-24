package pattern

import (
	"strings"
)

type Match interface {
	Priority() int
	Invoke(value string) bool
}

type MatchBody func(string) bool

type MatchFunction struct {
	fn       MatchBody
	priority int
}

func (this *MatchFunction) Invoke(value string) bool {
	return this.fn(value)
}

func (this *MatchFunction) Priority() int {
	return this.priority
}

func NewMatch(fn MatchBody, priority int) Match {
	return &MatchFunction{fn: fn, priority: priority}
}

type literalMatchFunction struct {
	against  string
	priority int
}

func (this *literalMatchFunction) Priority() int {
	return this.priority
}
func (this *literalMatchFunction) Invoke(value string) bool {
	return this.against == value
}

func NewLiteralMatch(literal string, priority int) Match {
	return &literalMatchFunction{against: literal, priority: priority}
}

var SinglefieldMatch, MultifieldMatch, OptionalfieldMatch Match

func init() {
	SinglefieldMatch = NewMatch(func(v string) bool { return strings.HasPrefix(v, SinglefieldVariablePrefix) }, 0)
	MultifieldMatch = NewMatch(func(v string) bool { return strings.HasPrefix(v, MultifieldVariablePrefix) }, 0)
	OptionalfieldMatch = NewMatch(func(v string) bool { return strings.HasPrefix(v, OptionalVariablePrefix) }, 0)
}
