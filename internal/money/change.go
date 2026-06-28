package money

import "errors"

// ErrNoChange means the available coins cannot make the exact change required.
var ErrNoChange = errors.New("insufficient coins to make exact change")

// MakeChange returns the coins to give back for amount, drawing only on float
// and taking the largest coins first.
//
// Greedy is optimal for the UK coin set when coins are plentiful. If the float
// has run low it may return ErrNoChange even where some combination exists; the
// caller should then refuse the sale rather than give wrong change. See the
// README for why greedy over an exhaustive search.
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
