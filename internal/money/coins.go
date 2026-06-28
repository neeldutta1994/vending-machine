package money

import (
	"fmt"
	"sort"
)

// Coins is a collection of coins counted by denomination. The same type is used
// for the machine's change float, for the coins a customer has put in, and for
// the change handed back.
//
// A Coins value behaves like a value object: the methods that change the
// contents (Add, With, Remove) return a new Coins and leave the receiver
// untouched. That makes it safe to hand a Coins around or hold onto a snapshot
// without worrying that someone else will mutate it underneath you.
type Coins struct {
	counts map[Denomination]int
}

// NewCoins builds a Coins from a denomination->quantity map. Nil, zero and
// negative quantities are ignored. The input map is copied, not retained.
func NewCoins(counts map[Denomination]int) Coins {
	c := Coins{counts: map[Denomination]int{}}
	for d, n := range counts {
		if n > 0 {
			c.counts[d] = n
		}
	}
	return c
}

// Count is how many of a given coin are held.
func (c Coins) Count(d Denomination) int { return c.counts[d] }

// Total adds up the face value of every coin held.
func (c Coins) Total() Money {
	var total Money
	for d, n := range c.counts {
		total += d.Value() * Money(n)
	}
	return total
}

// IsEmpty reports whether no coins are held.
func (c Coins) IsEmpty() bool {
	for _, n := range c.counts {
		if n > 0 {
			return false
		}
	}
	return true
}

// Add returns a new Coins holding this collection plus other.
func (c Coins) Add(other Coins) Coins {
	merged := c.clone()
	for d, n := range other.counts {
		merged[d] += n
	}
	return Coins{counts: merged}
}

// With returns a new Coins holding n more of the given coin.
func (c Coins) With(d Denomination, n int) Coins {
	merged := c.clone()
	merged[d] += n
	return Coins{counts: merged}
}

// Remove returns a new Coins with other taken away. It errors if this
// collection does not hold enough of any coin to cover the subtraction.
func (c Coins) Remove(other Coins) (Coins, error) {
	merged := c.clone()
	for d, n := range other.counts {
		if merged[d] < n {
			return Coins{}, fmt.Errorf("cannot remove %d x %s: only %d held", n, d, merged[d])
		}
		merged[d] -= n
		if merged[d] == 0 {
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

// Breakdown lists the coins held, largest denomination first. Handy for
// printing and for asserting in tests.
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
