package money

import "fmt"

// Denomination is a coin the machine accepts: the eight UK coins from 1p to £2.
type Denomination Money

const (
	OnePenny    Denomination = 1
	TwoPence    Denomination = 2
	FivePence   Denomination = 5
	TenPence    Denomination = 10
	TwentyPence Denomination = 20
	FiftyPence  Denomination = 50
	OnePound    Denomination = 100
	TwoPounds   Denomination = 200
)

// denominations is largest first; the change algorithm relies on that order.
var denominations = []Denomination{
	TwoPounds, OnePound, FiftyPence, TwentyPence, TenPence, FivePence, TwoPence, OnePenny,
}

// Denominations returns a copy of the accepted coins, largest first.
func Denominations() []Denomination {
	out := make([]Denomination, len(denominations))
	copy(out, denominations)
	return out
}

func (d Denomination) Value() Money   { return Money(d) }
func (d Denomination) String() string { return Money(d).String() }

// ParseDenomination turns a pence value into a Denomination, rejecting anything
// that is not a coin the machine takes.
func ParseDenomination(pence int) (Denomination, error) {
	for _, d := range denominations {
		if int(d) == pence {
			return d, nil
		}
	}
	return 0, fmt.Errorf("%dp is not an accepted coin", pence)
}
