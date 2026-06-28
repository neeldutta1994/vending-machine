// Package cli is the command-line front end. It reads commands and drives the
// vending.Machine, keeping all user interaction out of the domain.
package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/vending-machine/internal/money"
	"github.com/vending-machine/internal/vending"
)

// Run asks how much money the machine already holds, then reads commands until
// "quit" or end of input. Input and output are passed in so it can be scripted
// in tests.
func Run(in io.Reader, out io.Writer, products []vending.Stock) error {
	input := bufio.NewScanner(in)
	input.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	fmt.Fprintln(out, "Vending machine")
	fmt.Fprintf(out, "Loaded %d product line(s).\n", len(products))

	fmt.Fprintln(out, "\nEnter how much money the machine already holds.")
	fmt.Fprintln(out, "Type the number of each coin and press Enter (Enter on its own means none):")
	float := readCoinCounts(input, out)

	machine := vending.New(products, float)
	fmt.Fprintf(out, "\nMachine ready. Starting float: %s.\n", float.Total())
	printHelp(out)

	for {
		fmt.Fprint(out, "\n> ")
		if !input.Scan() {
			fmt.Fprintln(out)
			return nil
		}
		command, arg := splitCommand(strings.TrimSpace(input.Text()))

		switch strings.ToLower(command) {
		case "":
			continue
		case "help", "?":
			printHelp(out)
		case "list":
			printProducts(out, machine)
		case "select":
			doSelect(out, machine, arg)
		case "insert":
			doInsert(out, machine, arg)
		case "balance":
			fmt.Fprintf(out, "Inserted so far: %s\n", machine.Balance())
		case "cancel":
			if refund := machine.Cancel(); refund.IsEmpty() {
				fmt.Fprintln(out, "Nothing to refund.")
			} else {
				fmt.Fprintf(out, "Returned %s.\n", describe(refund))
			}
		case "state":
			printProducts(out, machine)
			fmt.Fprintf(out, "Float: %s (%s)\n", describe(machine.Float()), machine.Float().Total())
		case "reload-product":
			doReloadProduct(out, machine, arg)
		case "reload-coins":
			doReloadCoins(input, out, machine)
		case "quit", "exit":
			fmt.Fprintln(out, "Goodbye.")
			return nil
		default:
			fmt.Fprintf(out, "Unknown command %q. Type 'help' for the list.\n", command)
		}
	}
}

func doSelect(out io.Writer, m *vending.Machine, slot string) {
	if slot == "" {
		fmt.Fprintln(out, "Usage: select <slot>   e.g. select A1")
		return
	}
	slot = strings.ToUpper(slot)
	result, err := m.Select(slot)
	if err != nil {
		fmt.Fprintf(out, "Cannot select %s: %v\n", slot, err)
		return
	}
	report(out, result)
}

func doInsert(out io.Writer, m *vending.Machine, arg string) {
	pence, err := strconv.Atoi(arg)
	if err != nil {
		fmt.Fprintln(out, "Usage: insert <pence>   accepted: 1 2 5 10 20 50 100 200")
		return
	}
	coin, err := money.ParseDenomination(pence)
	if err != nil {
		fmt.Fprintf(out, "%v\n", err)
		return
	}

	result, err := m.InsertCoin(coin)
	if errors.Is(err, vending.ErrCannotMakeChange) {
		fmt.Fprintf(out, "Can't make the right change, sorry. Returning %s.\n", describe(m.Cancel()))
		return
	}
	report(out, result)
}

func doReloadProduct(out io.Writer, m *vending.Machine, arg string) {
	if arg == "" {
		fmt.Fprintln(out, "Usage: reload-product <slot,name,price,quantity>   e.g. reload-product A1,Water,80,10")
		return
	}
	stock, err := parseStockLine(arg)
	if err != nil {
		fmt.Fprintf(out, "%v\n", err)
		return
	}
	m.LoadProducts(stock)
	fmt.Fprintf(out, "Loaded %d x %s into slot %s.\n", stock.Quantity, stock.Product.Name, stock.Product.Code)
}

func doReloadCoins(in *bufio.Scanner, out io.Writer, m *vending.Machine) {
	fmt.Fprintln(out, "Add coins to the float. Number of each coin (Enter on its own means none):")
	added := readCoinCounts(in, out)
	m.LoadCoins(added)
	fmt.Fprintf(out, "Added %s. Float is now %s.\n", added.Total(), m.Float().Total())
}

// report prints where a sale stands after a Select or InsertCoin.
func report(out io.Writer, r vending.Result) {
	switch {
	case r.Dispensed && r.Change.IsEmpty():
		fmt.Fprintf(out, "Dispensed %s. No change.\n", r.Product.Name)
	case r.Dispensed:
		fmt.Fprintf(out, "Dispensed %s. Change: %s.\n", r.Product.Name, describe(r.Change))
	case r.Outstanding > 0:
		fmt.Fprintf(out, "Balance %s. Please insert %s more.\n", r.Balance, r.Outstanding)
	default:
		fmt.Fprintf(out, "Balance %s. Select a product.\n", r.Balance)
	}
}

// readCoinCounts asks for a quantity of each denomination and returns the coins.
func readCoinCounts(in *bufio.Scanner, out io.Writer) money.Coins {
	counts := map[money.Denomination]int{}
	for _, d := range money.Denominations() {
		for {
			fmt.Fprintf(out, "  %-4s x ", d)
			if !in.Scan() {
				return money.NewCoins(counts)
			}
			text := strings.TrimSpace(in.Text())
			if text == "" {
				break
			}
			if n, err := strconv.Atoi(text); err == nil && n >= 0 {
				counts[d] = n
				break
			}
			fmt.Fprintln(out, "    please enter 0 or a positive whole number")
		}
	}
	return money.NewCoins(counts)
}

func printProducts(out io.Writer, m *vending.Machine) {
	stock := m.Stock()
	sort.Slice(stock, func(i, j int) bool { return stock[i].Product.Code < stock[j].Product.Code })

	fmt.Fprintln(out, "Slot  Product           Price   Qty")
	for _, s := range stock {
		fmt.Fprintf(out, "%-5s %-17s %-7s %d\n", s.Product.Code, s.Product.Name, s.Product.Price, s.Quantity)
	}
}

func printHelp(out io.Writer) {
	fmt.Fprint(out, `
Commands:
  list                              show products and prices
  select <slot>                     choose a product (e.g. select A1)
  insert <pence>                    insert a coin: 1 2 5 10 20 50 100 200
  balance                           show money inserted so far
  cancel                            cancel the sale and get your coins back
  state                             show stock and the change float
  reload-product <slot,name,price,quantity>   add or top up a product
  reload-coins                      add coins to the change float
  help                              show this list
  quit                              exit
`)
}
