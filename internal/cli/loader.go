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

// LoadStock reads an inventory file and returns the product lines it describes.
//
// The file has one product per line in the form
//
//	slot,product name,price,quantity
//
// for example "A2,Water,80,10" (price is in pence). Blank lines and lines
// starting with '#' are ignored.
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
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		s, err := parseStockLine(text)
		if err != nil {
			return nil, fmt.Errorf("%s line %d: %w", name, lineNo, err)
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

// parseStockLine parses one "slot,name,price,quantity" line.
func parseStockLine(line string) (vending.Stock, error) {
	parts := strings.Split(line, ",")
	if len(parts) != 4 {
		return vending.Stock{}, fmt.Errorf("expected slot,name,price,quantity but got %d fields", len(parts))
	}
	slot := strings.TrimSpace(parts[0])
	productName := strings.TrimSpace(parts[1])

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

	product, err := catalog.NewProduct(slot, productName, money.Money(price))
	if err != nil {
		return vending.Stock{}, err
	}
	return vending.Stock{Product: product, Quantity: quantity}, nil
}
