// keyword parsing
package keyword

import (
	"unicode/utf8"
)

type node struct {
	canTerminate bool
	children     map[rune]*node
}

func (this *node) sink(input string) {
	numRunes := utf8.RuneCountInString(input)
	if numRunes == 0 {
		return
	}
	r, width := utf8.DecodeRuneInString(input)
	singleRune := numRunes == 1
	if _, ok := this.children[r]; !ok {
		this.children[r] = newNode(singleRune)
	} else {
		this.children[r].canTerminate = this.children[r].canTerminate || singleRune
	}
	if !singleRune {
		this.children[r].sink(input[width:])
	}
}
func (this *node) isKeyword(input string) bool {
	if len(input) == 0 {
		return this.canTerminate
	}

	r, width := utf8.DecodeRuneInString(input)
	if cell, ok := this.children[r]; !ok {
		return false
	} else {
		return cell.isKeyword(input[width:])
	}
}

func newNode(canTerminate bool) *node {
	return &node{canTerminate: canTerminate, children: make(map[rune]*node)}
}

type Parser struct {
	contents map[rune]*node
}

func New() *Parser {
	return &Parser{contents: make(map[rune]*node)}
}
func (this Parser) AddKeyword(input string) {
	// construct a buffer element
	r, width := utf8.DecodeRuneInString(input)
	singleRune := utf8.RuneCountInString(input) == 1
	if _, ok := this.contents[r]; !ok {
		this.contents[r] = newNode(singleRune)
	} else {
		this.contents[r].canTerminate = singleRune || this.contents[r].canTerminate // update canTerminate
	}
	if !singleRune {
		this.contents[r].sink(input[width:])
	}
}

func (this Parser) IsKeyword(input string) bool {
	if len(input) == 0 {
		return false
	}
	r, width := utf8.DecodeRuneInString(input)
	if cell, ok := this.contents[r]; !ok {
		return false
	} else {
		return cell.isKeyword(input[width:])
	}
}

func (this Parser) AddKeywordList(input []string) {
	for _, in := range input {
		this.AddKeyword(in)
	}
}
