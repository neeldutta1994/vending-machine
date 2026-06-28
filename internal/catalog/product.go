package catalog

import (
	"fmt"

	"github.com/vending-machine/internal/money"
)

// Product is something the machine sells. Code is the keypad selection (e.g.
// "A1"), Name is for display, Price is what it costs.
type Product struct {
	Code  string
	Name  string
	Price money.Money
}

// NewProduct builds a Product, rejecting an empty code or name and a
// non-positive price.
func NewProduct(code, name string, price money.Money) (Product, error) {
	switch {
	case code == "":
		return Product{}, fmt.Errorf("product code must not be empty")
	case name == "":
		return Product{}, fmt.Errorf("product %q must have a name", code)
	case price <= 0:
		return Product{}, fmt.Errorf("product %q must have a positive price", code)
	}
	return Product{Code: code, Name: name, Price: price}, nil
}
