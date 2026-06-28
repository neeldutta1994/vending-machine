package vending

import (
	"github.com/vending-machine/internal/catalog"
	"github.com/vending-machine/internal/money"
)

// Machine is the vending machine aggregate root. It owns the product inventory,
// the coin float used to give change, and the state of the sale in progress.
//
// A machine serves one customer at a time, so there is deliberately no locking.
// All money handling goes through this type, which keeps the inventory and the
// float consistent: a product only leaves stock when its change has already
// been worked out and committed.
type Machine struct {
	inv   *inventory
	float money.Coins

	// State of the current sale.
	selected *slot
	inserted money.Coins
}

// New builds a machine with an initial load of products and an initial change
// float.
func New(products []Stock, float money.Coins) *Machine {
	m := &Machine{
		inv:      newInventory(),
		float:    float,
		inserted: money.NewCoins(nil),
	}
	for _, s := range products {
		m.inv.load(s.Product, s.Quantity)
	}
	return m
}

// Result describes the outcome of a Select or InsertCoin call.
type Result struct {
	// Dispensed is true once the product and any change have been handed over.
	Dispensed bool
	// Product is the item dispensed. Only meaningful when Dispensed is true.
	Product catalog.Product
	// Change is the coins returned with the product. Only meaningful when
	// Dispensed is true; it may be empty if exact money was inserted.
	Change money.Coins

	// Balance is the total inserted for the current sale.
	Balance money.Money
	// Outstanding is how much more money is still needed. It is zero once the
	// product is dispensed, and also zero before a product has been selected
	// (the price is not yet known).
	Outstanding money.Money
}

// Select chooses a product by its keypad code. Money may be inserted before or
// after selecting; if enough has already been inserted, the sale completes and
// the product is dispensed straight away.
func (m *Machine) Select(code string) (Result, error) {
	s, ok := m.inv.get(code)
	if !ok {
		return Result{}, ErrUnknownProduct
	}
	if s.quantity == 0 {
		return Result{}, ErrOutOfStock
	}
	m.selected = s
	return m.settle()
}

// InsertCoin accepts a single coin towards the current sale. When a product is
// selected and enough money is in, the sale completes and change is returned.
func (m *Machine) InsertCoin(d money.Denomination) (Result, error) {
	if _, err := money.ParseDenomination(int(d)); err != nil {
		return Result{}, err
	}
	m.inserted = m.inserted.With(d, 1)
	return m.settle()
}

// settle checks whether the current sale can complete, and completes it if so.
// It is the single place where stock and the float are changed, so the two can
// never drift apart.
func (m *Machine) settle() (Result, error) {
	balance := m.inserted.Total()

	if m.selected == nil {
		return Result{Balance: balance}, nil
	}

	price := m.selected.product.Price
	if balance < price {
		return Result{Balance: balance, Outstanding: price - balance}, nil
	}

	// Enough money is in. The coins just inserted are part of the float for the
	// purpose of giving change, so fold them in before working change out.
	due := balance - price
	pooled := m.float.Add(m.inserted)
	change, err := money.MakeChange(pooled, due)
	if err != nil {
		// Cannot give correct change. Leave the sale untouched so the customer
		// can Cancel to get their coins back rather than be short-changed.
		return Result{Balance: balance}, ErrCannotMakeChange
	}

	// Commit. Removing the change from the pooled coins leaves the machine
	// holding exactly its old float plus the product's price.
	newFloat, err := pooled.Remove(change)
	if err != nil {
		// Unreachable: MakeChange only ever uses coins that are present.
		return Result{}, err
	}
	m.float = newFloat
	m.selected.quantity--
	product := m.selected.product
	m.reset()

	return Result{
		Dispensed: true,
		Product:   product,
		Change:    change,
		Balance:   balance,
	}, nil
}

// Cancel abandons the current sale and returns the coins the customer inserted.
func (m *Machine) Cancel() money.Coins {
	refund := m.inserted
	m.reset()
	return refund
}

func (m *Machine) reset() {
	m.selected = nil
	m.inserted = money.NewCoins(nil)
}

// LoadProducts adds products to the machine, topping up existing lines or
// starting new ones. It can be called at any time to restock.
func (m *Machine) LoadProducts(stock ...Stock) {
	for _, s := range stock {
		if s.Quantity > 0 {
			m.inv.load(s.Product, s.Quantity)
		}
	}
}

// LoadCoins adds coins to the change float. It can be called at any time to top
// the float up.
func (m *Machine) LoadCoins(c money.Coins) {
	m.float = m.float.Add(c)
}

// Float returns a snapshot of the coins held for change.
func (m *Machine) Float() money.Coins { return m.float }

// Balance is the total inserted for the sale currently in progress.
func (m *Machine) Balance() money.Money { return m.inserted.Total() }

// Stock returns a snapshot of the product lines held, in no particular order.
func (m *Machine) Stock() []Stock {
	out := make([]Stock, 0, len(m.inv.slots))
	for _, s := range m.inv.slots {
		out = append(out, Stock{Product: s.product, Quantity: s.quantity})
	}
	return out
}
