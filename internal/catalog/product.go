package catalog

import (
	"fmt"

	"github.com/vending-machine/internal/money"
)

// Product is something the machine sells.
//
// Code is the keypad selection a customer keys in (for example "A1"), Name is
// for display on a receipt or label, and Price is what it costs.
type Product struct {
	Code  string
	Name  string
	Price money.Money
}

// NewProduct builds a Product, rejecting an empty code or name and a price that
// is not greater than zero.
func NewProduct(code, name string, price money.Money) (Product, error) {
	if code == "" {
		return Product{}, fmt.Errorf("product code must not be empty")
	}
	if name == "" {
		return Product{}, fmt.Errorf("product %q must have a name", code)
	}
	if price <= 0 {
		return Product{}, fmt.Errorf("product %q must have a positive price", code)
	}
	return Product{Code: code, Name: name, Price: price}, nil
}
