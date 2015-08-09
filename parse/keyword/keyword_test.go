package keyword

import (
	"testing"
)

func Test_EmptyKeyword(t *testing.T) {
	p := New()
	if p.IsKeyword("") {
		t.Error("Empty string is never a keyword!")
	} else {
		t.Log("Empty string is a not a keyword!")
	}
}

func Test_SingleKeyword(t *testing.T) {
	p := New()
	p.AddKeyword("foo")
	if !p.IsKeyword("foo") {
		t.Error("Despite foo being a keyword, it wasn't believed to be such!")
	} else {
		t.Log("Foo is a keyword!")
	}
}
func Test_SingleKeyword_Case(t *testing.T) {
	p := New()
	p.AddKeyword("foo")
	if !p.IsKeyword("foo") {
		t.Error("Despite foo being a keyword, it wasn't believed to be such!")
	} else {
		t.Log("foo is a keyword!")
	}

	if p.IsKeyword("Foo") {
		t.Error("Foo (not foo) is marked as a keyword! This is wrong!")
	} else {
		t.Log("Foo is not a keyword and that is right!")
	}
}
