package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/vending-machine/internal/catalog"
	"github.com/vending-machine/internal/money"
	"github.com/vending-machine/internal/vending"
)

// LoadStock reads an inventory file of "slot,name,price,quantity" lines (price
// in pence), e.g. "A2,Water,80,10". Blank lines and '#' comments are ignored.
func LoadStock(path string) ([]vending.Stock, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseStock(f, path)
}

func parseStock(r io.Reader, name string) ([]vending.Stock, error) {
	scanner := bufio.NewScanner(r)
	var stock []vending.Stock
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		s, err := parseStockLine(text)
		if err != nil {
			return nil, fmt.Errorf("%s line %d: %w", name, line, err)
		}
		stock = append(stock, s)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(stock) == 0 {
		return nil, fmt.Errorf("%s lists no products", name)
	}
	return stock, nil
}

func parseStockLine(line string) (vending.Stock, error) {
	parts := strings.Split(line, ",")
	if len(parts) != 4 {
		return vending.Stock{}, fmt.Errorf("expected slot,name,price,quantity but got %d fields", len(parts))
	}
	price, err := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil {
		return vending.Stock{}, fmt.Errorf("price %q is not a whole number of pence", parts[2])
	}
	quantity, err := strconv.Atoi(strings.TrimSpace(parts[3]))
	if err != nil {
		return vending.Stock{}, fmt.Errorf("quantity %q is not a whole number", parts[3])
	}
	if quantity < 0 {
		return vending.Stock{}, fmt.Errorf("quantity %d cannot be negative", quantity)
	}

	product, err := catalog.NewProduct(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), money.Money(price))
	if err != nil {
		return vending.Stock{}, err
	}
	return vending.Stock{Product: product, Quantity: quantity}, nil
}
