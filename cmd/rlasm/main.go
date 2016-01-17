// assembler for iris1
package main

import (
	"bufio"
	"flag"
	"fmt"
	_ "github.com/DrItanium/cores/registration"
	"github.com/DrItanium/cores/registration/parser"
	"os"
	"strings"
)

var target = flag.String("target", "", "target backend (required)")
var input = flag.String("input", "", "input file to be processed (leave blank for stdin)")
var output = flag.String("output", "", "output file (leave blank for stdout)")
var listTargets = flag.Bool("list-targets", false, "display registered targets and exit")
var debug = flag.Bool("debug", false, "enable debug")

func listRegisteredTargets() {
	fmt.Fprintln(os.Stderr, "Supported targets: ")
	for _, val := range parser.GetRegistered() {
		fmt.Fprintln(os.Stderr, "\t - ", val)
	}
}
func main() {
	// add a layer of indirection to make sure that all of the files are correctly close on an os.Exit call
	if listTargets, listUsage, err, code := body(); err != nil {
		if listUsage {
			flag.Usage()
		}
		if listTargets {
			listRegisteredTargets()
		}
		if str := err.Error(); len(str) > 0 {
			fmt.Fprintln(os.Stderr, str)
		}
		os.Exit(code)
	}
}
func body() (bool, bool, error, int) {
	// list targets, list usage, error message
	flag.Parse()
	if *listTargets {
		return true, false, fmt.Errorf(""), 1
	}
	if *target == "" {
		return true, true, fmt.Errorf("Did not specify a target backend"), 2
	} else if !parser.IsRegistered(*target) {
		return true, true, fmt.Errorf("%s is not a registered target backend!", *target), 3
	} else {
		var scanner *bufio.Scanner
		var o *os.File
		if *input == "" {
			scanner = bufio.NewScanner(os.Stdin)
		} else {
			if file, err := os.Open(*input); err != nil {
				return false, false, err, 4
			} else {
				defer file.Close()
				scanner = bufio.NewScanner(file)
			}
		}
		if *output == "" {
			o = os.Stdout
		} else {
			if file, err := os.Create(*output); err != nil {
				return false, false, err, 5
			} else {
				defer file.Close()
				o = file
			}
		}
		if p, err := parser.New(*target); err != nil {
			return false, false, err, 6
		} else {
			c, e, e2, e3, b := make(chan parser.Entry, 1024), make(chan error), make(chan error), make(chan error), make(chan byte, 512)
			// scanner goroutine
			go func(scanner *bufio.Scanner, c chan parser.Entry, e chan error) {
				for count := 0; scanner.Scan(); count++ {
					line := strings.TrimSpace(scanner.Text())
					if len(line) == 0 {
						continue
					} else {
						c <- parser.Entry{Line: line, Index: count}
					}
				}
				close(c)
				if err := scanner.Err(); err != nil {
					e <- err
				} else {
					e <- nil
				}
			}(scanner, c, e)
			// output goroutine
			go func(c chan byte, e chan error, o *os.File) {
				q := bufio.NewWriter(o)
				if *debug {
					var count int
					for v := range c {
						q.WriteByte(v)
						count++
					}
					fmt.Println("Wrote", count, "bytes")
				} else {
					for v := range c {
						q.WriteByte(v)
					}
				}
				q.Flush()
				e <- nil
			}(b, e2, o)
			// parse process goroutine
			go func(c chan parser.Entry, e chan error, o chan byte, p parser.Parser) {
				if err := p.Parse(c); err != nil {
					e <- err
				} else if err := p.Process(); err != nil {
					e <- err
				} else if err := p.Dump(o); err != nil {
					e <- err
				} else {
					e <- nil
				}
				close(o)
			}(c, e3, b, p)
			for i := 0; i < 3; i++ {
				select {
				case err := <-e:
					if err != nil {
						return false, false, err, 7
					}
				case err := <-e2:
					if err != nil {
						return false, false, err, 8
					}
				case err := <-e3:
					if err != nil {
						return false, false, err, 9
					}
				}
			}
		}
		return false, false, nil, 0
	}
}
