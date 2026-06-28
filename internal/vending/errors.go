package vending

import "errors"

var (
	ErrUnknownProduct = errors.New("no product with that code")
	ErrOutOfStock     = errors.New("selected product is out of stock")
	// ErrCannotMakeChange: the sale is abandoned and the inserted coins stay
	// claimable through Cancel.
	ErrCannotMakeChange = errors.New("machine cannot make the right change")
)
