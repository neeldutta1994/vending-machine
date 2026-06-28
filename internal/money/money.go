package money

import "fmt"

// Money is an amount of money held as a whole number of pence.
//
// Everything in the machine is counted in pence rather than pounds-as-floats.
// Coins and prices are exact, and adding them up or taking change away never
// drifts the way binary floating point does (0.1 + 0.2 != 0.3).
type Money int

// String formats an amount the way a price label would read: "5p", "50p",
// "£1" or "£1.50".
func (m Money) String() string {
	if m < 100 {
		return fmt.Sprintf("%dp", int(m))
	}
	pounds := m / 100
	pence := m % 100
	if pence == 0 {
		return fmt.Sprintf("£%d", int(pounds))
	}
	return fmt.Sprintf("£%d.%02d", int(pounds), int(pence))
}
