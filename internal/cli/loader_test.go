package cli

import (
	"strings"
	"testing"
)

func TestParseStockLine(t *testing.T) {
	s, err := parseStockLine("A2,Water,80,10")
	if err != nil {
		t.Fatalf("parseStockLine: %v", err)
	}
	if s.Product.Code != "A2" || s.Product.Name != "Water" {
		t.Errorf("got code=%q name=%q", s.Product.Code, s.Product.Name)
	}
	if s.Product.Price != 80 {
		t.Errorf("price = %s, want 80p", s.Product.Price)
	}
	if s.Quantity != 10 {
		t.Errorf("quantity = %d, want 10", s.Quantity)
	}
}

func TestParseStockLineTrimsAndKeepsSpacedNames(t *testing.T) {
	s, err := parseStockLine("  C1 , Sparkling Water , 90 , 7 ")
	if err != nil {
		t.Fatalf("parseStockLine: %v", err)
	}
	if s.Product.Code != "C1" || s.Product.Name != "Sparkling Water" || s.Product.Price != 90 || s.Quantity != 7 {
		t.Errorf("unexpected parse: %+v", s)
	}
}

func TestParseStockLineErrors(t *testing.T) {
	bad := []string{
		"A1,Water,80",      // too few fields
		"A1,Water,80,5,x",  // too many fields
		"A1,Water,free,5",  // price not a number
		"A1,Water,80,lots", // quantity not a number
		"A1,Water,80,-1",   // negative quantity
		"A1,Water,0,5",     // zero price rejected by NewProduct
		",Water,80,5",      // empty slot rejected by NewProduct
	}
	for _, line := range bad {
		if _, err := parseStockLine(line); err == nil {
			t.Errorf("parseStockLine(%q) should have failed", line)
		}
	}
}

func TestParseStockSkipsBlankAndComments(t *testing.T) {
	in := strings.NewReader("# header\n\nA1,Water,80,10\n  \n# note\nB1,Crisps,95,6\n")
	stock, err := parseStock(in, "test")
	if err != nil {
		t.Fatalf("parseStock: %v", err)
	}
	if len(stock) != 2 {
		t.Fatalf("got %d products, want 2", len(stock))
	}
}

func TestParseStockEmptyIsError(t *testing.T) {
	if _, err := parseStock(strings.NewReader("# only comments\n\n"), "test"); err == nil {
		t.Error("an inventory with no products should be an error")
	}
}
