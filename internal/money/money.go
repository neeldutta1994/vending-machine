package money

import "fmt"

// Money is an amount held as a whole number of pence, so coins and prices stay
// exact and never drift the way floating-point pounds would.
type Money int

// String formats an amount the way a price label reads: "5p", "£1", "£1.50".
func (m Money) String() string {
	if m < 100 {
		return fmt.Sprintf("%dp", int(m))
	}
	if m%100 == 0 {
		return fmt.Sprintf("£%d", int(m/100))
	}
	return fmt.Sprintf("£%d.%02d", int(m/100), int(m%100))
}
