package money

import "testing"

func TestCoinsTotal(t *testing.T) {
	c := NewCoins(map[Denomination]int{
		OnePound:    2,
		TwentyPence: 3,
		FivePence:   1,
	})
	if got, want := c.Total(), Money(265); got != want {
		t.Errorf("Total() = %s, want %s", got, want)
	}
}

func TestNewCoinsDropsNonPositive(t *testing.T) {
	c := NewCoins(map[Denomination]int{OnePound: 0, FiftyPence: -2, TenPence: 1})
	if !c.IsEmpty() && c.Total() != 10 {
		t.Fatalf("expected only the 10p to survive, got total %s", c.Total())
	}
	if c.Count(OnePound) != 0 || c.Count(FiftyPence) != 0 {
		t.Error("zero and negative quantities should be dropped")
	}
}

func TestCoinsAddAndWithDoNotMutate(t *testing.T) {
	original := NewCoins(map[Denomination]int{TenPence: 1})

	added := original.Add(NewCoins(map[Denomination]int{TenPence: 2}))
	withOne := original.With(TwentyPence, 1)

	if original.Total() != 10 {
		t.Errorf("original was mutated, total now %s", original.Total())
	}
	if added.Count(TenPence) != 3 {
		t.Errorf("Add() count = %d, want 3", added.Count(TenPence))
	}
	if withOne.Count(TwentyPence) != 1 {
		t.Errorf("With() count = %d, want 1", withOne.Count(TwentyPence))
	}
}

func TestCoinsRemove(t *testing.T) {
	c := NewCoins(map[Denomination]int{OnePound: 2, TenPence: 1})

	left, err := c.Remove(NewCoins(map[Denomination]int{OnePound: 1}))
	if err != nil {
		t.Fatalf("Remove() unexpected error: %v", err)
	}
	if left.Count(OnePound) != 1 || left.Count(TenPence) != 1 {
		t.Errorf("after Remove got %v", left.Breakdown())
	}

	if _, err := c.Remove(NewCoins(map[Denomination]int{FiftyPence: 1})); err == nil {
		t.Error("removing a coin that is not held should error")
	}
}

func TestCoinsBreakdownIsLargestFirst(t *testing.T) {
	c := NewCoins(map[Denomination]int{FivePence: 1, OnePound: 1, TwentyPence: 1})
	b := c.Breakdown()
	want := []Denomination{OnePound, TwentyPence, FivePence}
	if len(b) != len(want) {
		t.Fatalf("Breakdown() length = %d, want %d", len(b), len(want))
	}
	for i, h := range b {
		if h.Denomination != want[i] {
			t.Errorf("Breakdown()[%d] = %s, want %s", i, h.Denomination, want[i])
		}
	}
}
