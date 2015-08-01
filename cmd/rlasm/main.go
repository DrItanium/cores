// assembler for iris1
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := bufio.NewScanner(strings.NewReader(scanner.Text()))
		line.Split(bufio.ScanWords)
		for line.Scan() {
			fmt.Println(line.Text())
		}
		if err := line.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading line:", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
