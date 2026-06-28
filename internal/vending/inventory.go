package vending

import "github.com/vending-machine/internal/catalog"

// slot is one product line and how many of it remain.
type slot struct {
	product  catalog.Product
	quantity int
}

// inventory is the set of product lines the machine holds, keyed by product
// code. It is an internal helper of the Machine aggregate, not used directly by
// callers.
type inventory struct {
	slots map[string]*slot
}

func newInventory() *inventory {
	return &inventory{slots: map[string]*slot{}}
}

// load tops up an existing line or starts a new one. Reloading an existing code
// also refreshes the product details, so a price change can be rolled out by
// reloading the line.
func (inv *inventory) load(p catalog.Product, qty int) {
	if s, ok := inv.slots[p.Code]; ok {
		s.product = p
		s.quantity += qty
		return
	}
	inv.slots[p.Code] = &slot{product: p, quantity: qty}
}

func (inv *inventory) get(code string) (*slot, bool) {
	s, ok := inv.slots[code]
	return s, ok
}

// Stock describes a product and the quantity held. It is the shape used both
// for loading the machine and for reading back its current contents.
type Stock struct {
	Product  catalog.Product
	Quantity int
}
