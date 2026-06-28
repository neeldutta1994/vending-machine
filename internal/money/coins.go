package money

import (
	"fmt"
	"sort"
)

// Coins is a multiset of coins counted by denomination, used for the float, the
// coins a customer inserts, and the change handed back.
//
// It behaves as a value object: Add, With and Remove return a new Coins rather
// than mutating the receiver, so a Coins can be shared or held as a snapshot
// without anyone changing it underneath you.
type Coins struct {
	counts map[Denomination]int
}

// NewCoins builds a Coins from a denomination->quantity map, ignoring nil, zero
// and negative quantities. The input map is copied, not retained.
func NewCoins(counts map[Denomination]int) Coins {
	c := Coins{counts: map[Denomination]int{}}
	for d, n := range counts {
		if n > 0 {
			c.counts[d] = n
		}
	}
	return c
}

func (c Coins) Count(d Denomination) int { return c.counts[d] }

func (c Coins) Total() Money {
	var total Money
	for d, n := range c.counts {
		total += d.Value() * Money(n)
	}
	return total
}

func (c Coins) IsEmpty() bool {
	for _, n := range c.counts {
		if n > 0 {
			return false
		}
	}
	return true
}

// Add returns this collection plus other.
func (c Coins) Add(other Coins) Coins {
	merged := c.clone()
	for d, n := range other.counts {
		merged[d] += n
	}
	return Coins{counts: merged}
}

// With returns this collection plus n more of one coin.
func (c Coins) With(d Denomination, n int) Coins {
	merged := c.clone()
	merged[d] += n
	return Coins{counts: merged}
}

// Remove returns this collection minus other, erroring if it does not hold
// enough of any coin.
func (c Coins) Remove(other Coins) (Coins, error) {
	merged := c.clone()
	for d, n := range other.counts {
		if merged[d] < n {
			return Coins{}, fmt.Errorf("cannot remove %d x %s: only %d held", n, d, merged[d])
		}
		if merged[d] -= n; merged[d] == 0 {
			delete(merged, d)
		}
	}
	return Coins{counts: merged}, nil
}

// Holding is a quantity of a single denomination.
type Holding struct {
	Denomination Denomination
	Quantity     int
}

// Breakdown lists the coins held, largest denomination first.
func (c Coins) Breakdown() []Holding {
	out := make([]Holding, 0, len(c.counts))
	for d, n := range c.counts {
		if n > 0 {
			out = append(out, Holding{Denomination: d, Quantity: n})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Denomination > out[j].Denomination
	})
	return out
}

func (c Coins) clone() map[Denomination]int {
	m := make(map[Denomination]int, len(c.counts))
	for d, n := range c.counts {
		m[d] = n
	}
	return m
}
