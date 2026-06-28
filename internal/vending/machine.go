package vending

import (
	"github.com/vending-machine/internal/catalog"
	"github.com/vending-machine/internal/money"
)

// Machine is the vending machine aggregate root. It owns the product stock, the
// coin float used for change, and the sale in progress.
//
// Stock and float only ever change through the machine, and never half-way: a
// product leaves stock in the same step its change is committed, so the two
// can't drift apart. It serves one customer at a time, so there is no locking.
type Machine struct {
	slots map[string]*slot
	float money.Coins

	selected *slot
	inserted money.Coins
}

type slot struct {
	product  catalog.Product
	quantity int
}

// Stock is a product and the quantity held: the shape used both to load the
// machine and to read its contents back.
type Stock struct {
	Product  catalog.Product
	Quantity int
}

// New builds a machine with an initial load of products and change.
func New(products []Stock, float money.Coins) *Machine {
	m := &Machine{
		slots:    map[string]*slot{},
		float:    float,
		inserted: money.NewCoins(nil),
	}
	m.LoadProducts(products...)
	return m
}

// Result describes the outcome of a Select or InsertCoin.
type Result struct {
	Dispensed bool            // product and any change have been handed over
	Product   catalog.Product // dispensed item (when Dispensed)
	Change    money.Coins     // coins returned (when Dispensed; may be empty)

	Balance     money.Money // total inserted for the current sale
	Outstanding money.Money // still needed; 0 once dispensed or before a selection
}

// Select chooses a product by its code. Money may be inserted before or after;
// if enough is already in, the sale completes straight away.
func (m *Machine) Select(code string) (Result, error) {
	s, ok := m.slots[code]
	if !ok {
		return Result{}, ErrUnknownProduct
	}
	if s.quantity == 0 {
		return Result{}, ErrOutOfStock
	}
	m.selected = s
	return m.settle()
}

// InsertCoin accepts one coin towards the current sale, completing it and
// giving change once a product is selected and enough money is in.
func (m *Machine) InsertCoin(d money.Denomination) (Result, error) {
	if _, err := money.ParseDenomination(int(d)); err != nil {
		return Result{}, err
	}
	m.inserted = m.inserted.With(d, 1)
	return m.settle()
}

// settle completes the sale if it can. It is the only place stock and the float
// change, so they can never drift apart.
func (m *Machine) settle() (Result, error) {
	balance := m.inserted.Total()
	if m.selected == nil {
		return Result{Balance: balance}, nil
	}

	price := m.selected.product.Price
	if balance < price {
		return Result{Balance: balance, Outstanding: price - balance}, nil
	}

	// The inserted coins are part of the float for the purpose of giving change.
	pooled := m.float.Add(m.inserted)
	change, err := money.MakeChange(pooled, balance-price)
	if err != nil {
		// Leave the sale untouched so Cancel can return the coins.
		return Result{Balance: balance}, ErrCannotMakeChange
	}

	// Removing the change from the pool leaves the old float plus the price.
	m.float, _ = pooled.Remove(change)
	m.selected.quantity--
	product := m.selected.product
	m.reset()

	return Result{Dispensed: true, Product: product, Change: change, Balance: balance}, nil
}

// Cancel abandons the current sale and returns the inserted coins.
func (m *Machine) Cancel() money.Coins {
	refund := m.inserted
	m.reset()
	return refund
}

func (m *Machine) reset() {
	m.selected = nil
	m.inserted = money.NewCoins(nil)
}

// LoadProducts adds products at any time, topping up existing slots or starting
// new ones. Reloading a slot also refreshes its details, so a price change can
// be rolled out by reloading.
func (m *Machine) LoadProducts(stock ...Stock) {
	for _, s := range stock {
		if s.Quantity <= 0 {
			continue
		}
		if existing, ok := m.slots[s.Product.Code]; ok {
			existing.product = s.Product
			existing.quantity += s.Quantity
		} else {
			m.slots[s.Product.Code] = &slot{product: s.Product, quantity: s.Quantity}
		}
	}
}

// LoadCoins adds coins to the change float at any time.
func (m *Machine) LoadCoins(c money.Coins) { m.float = m.float.Add(c) }

// Float is a snapshot of the coins held for change.
func (m *Machine) Float() money.Coins { return m.float }

// Balance is the total inserted for the sale in progress.
func (m *Machine) Balance() money.Money { return m.inserted.Total() }

// Stock is a snapshot of the product slots held, in no particular order.
func (m *Machine) Stock() []Stock {
	out := make([]Stock, 0, len(m.slots))
	for _, s := range m.slots {
		out = append(out, Stock{Product: s.product, Quantity: s.quantity})
	}
	return out
}
