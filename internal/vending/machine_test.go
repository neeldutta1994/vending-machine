package vending

import (
	"errors"
	"testing"

	"github.com/vending-machine/internal/catalog"
	"github.com/vending-machine/internal/money"
)

// newTestMachine builds a machine stocked with one 80p water (qty 1) and a
// generous float, unless the caller overrides the float.
func newTestMachine(t *testing.T, float money.Coins) *Machine {
	t.Helper()
	water, err := catalog.NewProduct("A1", "Water", 80)
	if err != nil {
		t.Fatalf("building product: %v", err)
	}
	return New([]Stock{{Product: water, Quantity: 1}}, float)
}

func generousFloat() money.Coins {
	return money.NewCoins(map[money.Denomination]int{
		money.OnePenny: 50, money.TwoPence: 50, money.FivePence: 50,
		money.TenPence: 50, money.TwentyPence: 50, money.FiftyPence: 50,
		money.OnePound: 50, money.TwoPounds: 50,
	})
}

func TestExactMoneyDispensesNoChange(t *testing.T) {
	m := newTestMachine(t, generousFloat())

	if _, err := m.Select("A1"); err != nil {
		t.Fatalf("Select: %v", err)
	}
	r, err := m.InsertCoin(money.FiftyPence)
	if err != nil {
		t.Fatalf("InsertCoin: %v", err)
	}
	if r.Dispensed {
		t.Fatal("50p should not be enough for an 80p item yet")
	}
	if r.Outstanding != 30 {
		t.Errorf("outstanding = %s, want 30p", r.Outstanding)
	}

	r, err = m.InsertCoin(money.TwentyPence)
	if err != nil {
		t.Fatalf("InsertCoin: %v", err)
	}
	if r.Dispensed {
		t.Fatal("70p should still not be enough")
	}

	r, err = m.InsertCoin(money.TenPence)
	if err != nil {
		t.Fatalf("InsertCoin: %v", err)
	}
	if !r.Dispensed {
		t.Fatal("80p exact should dispense")
	}
	if r.Product.Code != "A1" {
		t.Errorf("dispensed %q, want A1", r.Product.Code)
	}
	if !r.Change.IsEmpty() {
		t.Errorf("exact money should give no change, got %v", r.Change.Breakdown())
	}
}

func TestOverpaymentReturnsChange(t *testing.T) {
	m := newTestMachine(t, generousFloat())
	mustSelect(t, m, "A1")

	r, err := m.InsertCoin(money.OnePound)
	if err != nil {
		t.Fatalf("InsertCoin: %v", err)
	}
	if !r.Dispensed {
		t.Fatal("£1 for an 80p item should dispense")
	}
	if r.Change.Total() != 20 {
		t.Errorf("change = %s, want 20p", r.Change.Total())
	}
}

func TestFloatGrowsByPriceAfterSale(t *testing.T) {
	float := generousFloat()
	before := float.Total()
	m := newTestMachine(t, float)
	mustSelect(t, m, "A1")

	if _, err := m.InsertCoin(money.OnePound); err != nil {
		t.Fatalf("InsertCoin: %v", err)
	}
	// Machine kept £1, gave 20p change: net gain is the 80p price.
	if got := m.Float().Total(); got != before+80 {
		t.Errorf("float total = %s, want %s", got, before+80)
	}
}

func TestInsufficientThenCancelRefunds(t *testing.T) {
	m := newTestMachine(t, generousFloat())
	mustSelect(t, m, "A1")

	if _, err := m.InsertCoin(money.TwentyPence); err != nil {
		t.Fatalf("InsertCoin: %v", err)
	}
	refund := m.Cancel()
	if refund.Total() != 20 {
		t.Errorf("refund = %s, want 20p", refund.Total())
	}
	if m.Balance() != 0 {
		t.Errorf("balance after cancel = %s, want 0", m.Balance())
	}
}

func TestCannotMakeChangeAbandonsSale(t *testing.T) {
	// Empty float and an item that needs change: pay £2 for an 80p item.
	m := newTestMachine(t, money.NewCoins(nil))
	mustSelect(t, m, "A1")

	_, err := m.InsertCoin(money.TwoPounds)
	if !errors.Is(err, ErrCannotMakeChange) {
		t.Fatalf("error = %v, want ErrCannotMakeChange", err)
	}
	// The coin is still claimable.
	refund := m.Cancel()
	if refund.Total() != 200 {
		t.Errorf("refund = %s, want £2", refund.Total())
	}
	// Stock untouched.
	if got := stockQty(m, "A1"); got != 1 {
		t.Errorf("stock A1 = %d, want 1 (sale should not have happened)", got)
	}
}

func TestInsertedCoinsCanFundChange(t *testing.T) {
	// Float has only a 20p. Customer pays £1 for an 80p item and is owed 20p:
	// the float's 20p makes that change.
	m := newTestMachine(t, money.NewCoins(map[money.Denomination]int{money.TwentyPence: 1}))
	mustSelect(t, m, "A1")

	r, err := m.InsertCoin(money.OnePound)
	if err != nil {
		t.Fatalf("InsertCoin: %v", err)
	}
	if r.Change.Count(money.TwentyPence) != 1 {
		t.Errorf("change = %v, want 1 x 20p", r.Change.Breakdown())
	}
}

func TestUnknownAndOutOfStock(t *testing.T) {
	m := newTestMachine(t, generousFloat())

	if _, err := m.Select("ZZ"); !errors.Is(err, ErrUnknownProduct) {
		t.Errorf("Select unknown error = %v, want ErrUnknownProduct", err)
	}

	// Drain the only water.
	mustSelect(t, m, "A1")
	if _, err := m.InsertCoin(money.OnePound); err != nil {
		t.Fatalf("InsertCoin: %v", err)
	}
	if _, err := m.Select("A1"); !errors.Is(err, ErrOutOfStock) {
		t.Errorf("Select drained error = %v, want ErrOutOfStock", err)
	}
}

func TestSelectAfterInsertingCompletesImmediately(t *testing.T) {
	m := newTestMachine(t, generousFloat())

	// Money first, no selection yet.
	r, err := m.InsertCoin(money.OnePound)
	if err != nil {
		t.Fatalf("InsertCoin: %v", err)
	}
	if r.Dispensed {
		t.Fatal("nothing should dispense before a product is selected")
	}

	r, err = m.Select("A1")
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if !r.Dispensed {
		t.Fatal("selecting with enough money already in should dispense")
	}
	if r.Change.Total() != 20 {
		t.Errorf("change = %s, want 20p", r.Change.Total())
	}
}

func TestRejectsUnacceptedCoin(t *testing.T) {
	m := newTestMachine(t, generousFloat())
	mustSelect(t, m, "A1")
	if _, err := m.InsertCoin(money.Denomination(7)); err == nil {
		t.Error("inserting a 7p 'coin' should be rejected")
	}
}

func TestReload(t *testing.T) {
	m := newTestMachine(t, money.NewCoins(nil))

	chocolate, _ := catalog.NewProduct("C1", "Chocolate", 110)
	m.LoadProducts(
		Stock{Product: chocolate, Quantity: 4},
		Stock{Product: mustProduct(t, "A1", "Water", 80), Quantity: 2}, // top up existing line
	)
	if got := stockQty(m, "A1"); got != 3 {
		t.Errorf("A1 stock after top-up = %d, want 3", got)
	}
	if got := stockQty(m, "C1"); got != 4 {
		t.Errorf("C1 stock = %d, want 4", got)
	}

	m.LoadCoins(money.NewCoins(map[money.Denomination]int{money.FiftyPence: 4}))
	if got := m.Float().Total(); got != 200 {
		t.Errorf("float after load = %s, want £2", got)
	}
}

// --- helpers ---

func mustSelect(t *testing.T, m *Machine, code string) {
	t.Helper()
	if _, err := m.Select(code); err != nil {
		t.Fatalf("Select(%q): %v", code, err)
	}
}

func mustProduct(t *testing.T, code, name string, price money.Money) catalog.Product {
	t.Helper()
	p, err := catalog.NewProduct(code, name, price)
	if err != nil {
		t.Fatalf("NewProduct: %v", err)
	}
	return p
}

func stockQty(m *Machine, code string) int {
	for _, s := range m.Stock() {
		if s.Product.Code == code {
			return s.Quantity
		}
	}
	return -1
}
