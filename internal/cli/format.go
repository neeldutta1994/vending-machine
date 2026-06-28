package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/vending-machine/internal/money"
	"github.com/vending-machine/internal/vending"
)

// splitCommand splits a line into its first word (the command) and the rest
// (its argument), trimming surrounding spaces.
func splitCommand(line string) (command, arg string) {
	fields := strings.SplitN(line, " ", 2)
	if len(fields) == 1 {
		return fields[0], ""
	}
	return fields[0], strings.TrimSpace(fields[1])
}

// describe renders a set of coins as "2x£1, 1x20p", largest coin first, or
// "none" when empty.
func describe(c money.Coins) string {
	if c.IsEmpty() {
		return "none"
	}
	parts := make([]string, 0)
	for _, h := range c.Breakdown() {
		parts = append(parts, fmt.Sprintf("%dx%s", h.Quantity, h.Denomination))
	}
	return strings.Join(parts, ", ")
}

// sortedStock returns the machine's stock ordered by slot, so listings are
// stable and easy to read.
func sortedStock(m *vending.Machine) []vending.Stock {
	stock := m.Stock()
	sort.Slice(stock, func(i, j int) bool {
		return stock[i].Product.Code < stock[j].Product.Code
	})
	return stock
}
