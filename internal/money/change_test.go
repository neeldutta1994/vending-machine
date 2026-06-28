package money

import (
	"errors"
	"testing"
)

func TestMakeChangeZero(t *testing.T) {
	change, err := MakeChange(NewCoins(nil), 0)
	if err != nil {
		t.Fatalf("MakeChange for 0 returned error: %v", err)
	}
	if !change.IsEmpty() {
		t.Errorf("MakeChange for 0 should be empty, got %v", change.Breakdown())
	}
}

func TestMakeChangeGreedy(t *testing.T) {
	float := NewCoins(map[Denomination]int{
		FiftyPence:  5,
		TwentyPence: 5,
		TenPence:    5,
		FivePence:   5,
	})
	change, err := MakeChange(float, 85)
	if err != nil {
		t.Fatalf("MakeChange(85) error: %v", err)
	}
	if change.Total() != 85 {
		t.Errorf("change total = %s, want 85p", change.Total())
	}
	// Largest first: 50 + 20 + 10 + 5.
	want := map[Denomination]int{FiftyPence: 1, TwentyPence: 1, TenPence: 1, FivePence: 1}
	for d, n := range want {
		if change.Count(d) != n {
			t.Errorf("change has %d x %s, want %d", change.Count(d), d, n)
		}
	}
}

func TestMakeChangeUsesSmallerCoinsWhenLargeRunOut(t *testing.T) {
	// No 50p available, so 70p must come from 20p coins.
	float := NewCoins(map[Denomination]int{TwentyPence: 4})
	change, err := MakeChange(float, 60)
	if err != nil {
		t.Fatalf("MakeChange(60) error: %v", err)
	}
	if change.Count(TwentyPence) != 3 {
		t.Errorf("expected 3 x 20p, got %v", change.Breakdown())
	}
}

func TestMakeChangeImpossible(t *testing.T) {
	float := NewCoins(map[Denomination]int{TwentyPence: 1})
	if _, err := MakeChange(float, 30); !errors.Is(err, ErrNoChange) {
		t.Errorf("MakeChange(30) error = %v, want ErrNoChange", err)
	}
}
