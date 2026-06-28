package money

import "testing"

func TestMoneyString(t *testing.T) {
	cases := []struct {
		amount Money
		want   string
	}{
		{1, "1p"},
		{50, "50p"},
		{99, "99p"},
		{100, "£1"},
		{120, "£1.20"},
		{105, "£1.05"},
		{250, "£2.50"},
	}
	for _, c := range cases {
		if got := c.amount.String(); got != c.want {
			t.Errorf("Money(%d).String() = %q, want %q", c.amount, got, c.want)
		}
	}
}

func TestParseDenomination(t *testing.T) {
	if d, err := ParseDenomination(50); err != nil || d != FiftyPence {
		t.Errorf("ParseDenomination(50) = %v, %v; want 50p, nil", d, err)
	}
	if _, err := ParseDenomination(3); err == nil {
		t.Error("ParseDenomination(3) should reject 3p as not an accepted coin")
	}
}
