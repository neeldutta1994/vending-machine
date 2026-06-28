// Command vending-machine is the command-line vending machine.
//
// It loads its product inventory from a file (input.txt by default, or the
// first argument), asks how much money the machine already holds, and then
// reads commands from standard input. Type "help" once it is running for the
// list of commands.
package main

import (
	"fmt"
	"os"

	"github.com/vending-machine/internal/cli"
)

func main() {
	path := "input.txt"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	products, err := cli.LoadStock(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading inventory:", err)
		os.Exit(1)
	}

	if err := cli.Run(os.Stdin, os.Stdout, products); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
