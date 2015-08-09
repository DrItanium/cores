package keyword

import (
	"testing"
)

func expectIsNotKeyword(psr *Parser, kw string, t *testing.T) {
	if psr.IsKeyword(kw) {
		t.Errorf("Expected %s to not be keyword!", kw)
		t.Fail()
	} else {
		t.Logf("%s is not keyword!", kw)
	}
}
func expectIsKeyword(psr *Parser, kw string, t *testing.T) {
	if !psr.IsKeyword(kw) {
		t.Errorf("Expected %s to be a keyword!", kw)
		t.Fail()
	} else {
		t.Logf("%s is a keyword!", kw)
	}
}

func Test_EmptyKeyword(t *testing.T) {
	p := New()
	expectIsNotKeyword(p, "", t)
}

func Test_SingleKeyword(t *testing.T) {
	p := New()
	p.AddKeyword("foo")
	expectIsKeyword(p, "foo", t)
}
func Test_SingleKeyword_Case(t *testing.T) {
	p := New()
	p.AddKeyword("foo")
	expectIsKeyword(p, "foo", t)
	expectIsNotKeyword(p, "Foo", t)
}
func Test_MultipleKeywords_1(t *testing.T) {
	p := New()
	p.AddKeyword("foo")
	p.AddKeyword("foot")
	p.AddKeyword("fog")
	expectIsKeyword(p, "foo", t)
	expectIsKeyword(p, "foot", t)
	expectIsKeyword(p, "fog", t)
}

func Test_MultileKeywords_2(t *testing.T) {
	p := New()
	p.AddKeyword("do")
	p.AddKeyword("defmacro")
	p.AddKeyword("defun")
	p.AddKeyword("func")
	p.AddKeyword("car")
	p.AddKeyword("cdr")
	p.AddKeyword("first")
	p.AddKeyword("rest")
	expectIsKeyword(p, "func", t)
	expectIsKeyword(p, "defun", t)
	expectIsKeyword(p, "car", t)
}
