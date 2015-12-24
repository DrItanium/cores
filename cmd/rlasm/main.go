// assembler for iris1
package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/DrItanium/cores/registration"
	"github.com/DrItanium/cores/registration/parser"
	"os"
	"strings"
)

var target = flag.String("target", "", "target backend (required)")
var input = flag.String("input", "", "input file to be processed (leave blank for stdin)")
var output = flag.String("output", "", "output file (leave blank for stdout)")
var listTargets = flag.Bool("list-targets", false, "display registered targets and exit")
var debug = flag.Bool("debug", false, "enable debug")

func init() {
	registration.Register()
}

func listRegisteredTargets() {
	fmt.Println("Supported targets: ")
	for _, val := range parser.GetRegistered() {
		fmt.Println("\t - ", val)
	}
}
func main() {
	flag.Parse()
	if *listTargets {
		listRegisteredTargets()
		return
	}
	if *target == "" {
		fmt.Println("Did not specify target backend")
		flag.Usage()
		listRegisteredTargets()
	} else if !parser.IsRegistered(*target) {
		fmt.Println("%s is not a registered target backend!", *target)
		flag.Usage()
		listRegisteredTargets()
	} else {
		var scanner *bufio.Scanner
		var o *os.File
		if *input == "" {
			scanner = bufio.NewScanner(os.Stdin)
		} else {
			if file, err := os.Open(*input); err != nil {
				fmt.Println(err)
				return
			} else {
				defer file.Close()
				scanner = bufio.NewScanner(file)
			}
		}
		if *output == "" {
			o = os.Stdout
		} else {
			if file, err := os.Create(*output); err != nil {
				fmt.Println(err)
				return
			} else {
				defer file.Close()
				o = file
			}
		}
		if p, err := parser.New(*target); err != nil {
			fmt.Println(err)
			return
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
			}(b, e3, o)
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
			}(c, e, b, p)
			for i := 0; i < 3; i++ {
				select {
				case err := <-e:
					if err != nil {
						fmt.Println(err)
						return
					}
				case err := <-e2:
					if err != nil {
						fmt.Println(err)
						return
					}
				case err := <-e3:
					if err != nil {
						fmt.Println(err)
						return
					}
				}
			}
		}
	}
}
