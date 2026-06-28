package cli

import (
	"fmt"
	"strings"

	"github.com/vending-machine/internal/money"
)

// splitCommand splits a line into its first word and the rest.
func splitCommand(line string) (command, arg string) {
	fields := strings.SplitN(line, " ", 2)
	if len(fields) == 1 {
		return fields[0], ""
	}
	return fields[0], strings.TrimSpace(fields[1])
}

// describe renders coins as "2x£1, 1x20p", largest first, or "none" when empty.
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
