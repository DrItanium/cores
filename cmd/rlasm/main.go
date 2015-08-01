// assembler for iris1
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		} else {
			fmt.Println(line)
			stmt := carveLine(line)
			for ind, str := range stmt {
				fmt.Printf("\t%d: %s\n", ind, str.Value)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

type NodeType int

const (
	TypeUnknown NodeType = iota
	TypeEquals
	TypeComma
	TypeLabel
	TypeRegister
	TypeImmediate
	TypeComment
)

type Node struct {
	Value string
	Type  NodeType
}

func (this *Node) IsComment() bool {
	return this.Type == TypeComment
}
func (this *Node) IsLabel() bool {
	return this.Type == TypeLabel
}

func (this *Node) NeedsAnalysis() bool {
	return this.Type == TypeUnknown
}

type Statement []Node

func carveLine(line string) Statement {
	// trim the damn line first
	captureSequence := func(nodes Statement, dat string, from, to int) Statement {
		str := strings.TrimSpace(dat[from:to])
		if len(str) > 0 {
			return append(nodes, Node{Value: str, Type: TypeUnknown})
		} else {
			return nodes
		}
	}
	data := strings.TrimSpace(line)
	var s Statement
	if len(data) == 0 {
		return s
	}
	oldStart := 0
	start := 0
	// skip the strings at the beginning
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRuneInString(data[start:])
		if unicode.IsSpace(r) {
			s = captureSequence(s, data, oldStart, start)
			oldStart = start
		} else if r == '=' {
			s = captureSequence(s, data, oldStart, start)
			s = append(s, Node{Value: "=", Type: TypeEquals})
			oldStart = start + width
		} else if r == ',' {
			s = captureSequence(s, data, oldStart, start)
			s = append(s, Node{Value: ",", Type: TypeComma})
			oldStart = start + width
		} else if r == ';' {
			// consume the rest of the data
			s = captureSequence(s, data, oldStart, start)
			// then capture the comment
			s = append(s, Node{Value: data[start:], Type: TypeComment})
			oldStart = start
			break
		}
	}
	if oldStart < start {
		s = append(s, Node{Value: data[oldStart:]})
	}
	return s
}
