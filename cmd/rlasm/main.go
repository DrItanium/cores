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
				fmt.Printf("\t%d: %s\n", ind, str)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

type statement []string

func carveLine(line string) statement {
	var s statement
	oldStart := 0
	start := 0
	// skip the strings at the beginning
	for width := 0; start < len(line); start += width {
		var r rune
		r, width = utf8.DecodeRuneInString(line[start:])
		if unicode.IsSpace(r) || r == '=' || r == ',' {
			str := strings.TrimSpace(line[oldStart:start])
			if len(str) > 0 {
				s = append(s, str)
			}
			oldStart = start
		} else if r == ';' {
			// consume the rest of the line
			preComment := strings.TrimSpace(line[oldStart:start])
			if len(preComment) > 0 {
				s = append(s, preComment)
			}
			// then capture the comment
			s = append(s, line[start:])
			oldStart = start
			break
		}
	}
	if oldStart < start {
		s = append(s, line[oldStart:])
	}
	return s
}
