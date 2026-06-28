package money

import "errors"

// ErrNoChange means the available coins cannot make the exact change required.
var ErrNoChange = errors.New("insufficient coins to make exact change")

// MakeChange works out which coins to give back for amount, drawing only on the
// coins in float. It takes the largest coins first.
//
// Largest-first (greedy) is optimal for the UK coin set when coins are
// plentiful. It can, however, fail to find a valid combination when the float
// has run low on a particular coin — for example needing 30p from a float of
// 1x20p and 2x20p... when only larger coins remain. In that situation it
// returns ErrNoChange rather than the wrong coins, and the caller is expected
// to refuse the sale and return the customer's money. See the README for the
// trade-off behind choosing greedy over an exhaustive search.
func MakeChange(float Coins, amount Money) (Coins, error) {
	if amount < 0 {
		return Coins{}, errors.New("cannot make change for a negative amount")
	}

	change := map[Denomination]int{}
	remaining := amount
	for _, d := range denominations {
		if remaining <= 0 {
			break
		}
		want := int(remaining / d.Value())
		if have := float.Count(d); want > have {
			want = have
		}
		if want > 0 {
			change[d] = want
			remaining -= d.Value() * Money(want)
		}
	}

	if remaining > 0 {
		return Coins{}, ErrNoChange
	}
	return NewCoins(change), nil
}
