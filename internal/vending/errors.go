package vending

import "errors"

var (
	// ErrUnknownProduct is returned when a selection code matches no product.
	ErrUnknownProduct = errors.New("no product with that code")
	// ErrOutOfStock is returned when the selected product has no stock left.
	ErrOutOfStock = errors.New("selected product is out of stock")
	// ErrCannotMakeChange is returned when the sale would need change the float
	// cannot produce. The sale is abandoned and the inserted coins stay
	// claimable through Cancel.
	ErrCannotMakeChange = errors.New("machine cannot make the right change")
)
