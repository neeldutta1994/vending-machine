package cli

import (
	"bytes"
	"strings"
	"testing"
)

// session runs the CLI against scripted input and returns everything printed.
// The first eight lines of every script answer the starting-float prompt, in
// the order the coins are asked for: £2, £1, 50p, 20p, 10p, 5p, 2p, 1p.
func session(t *testing.T, script string) string {
	t.Helper()
	products, err := parseStock(strings.NewReader("A1,Water,80,10\nB1,Crisps,95,6\n"), "test")
	if err != nil {
		t.Fatalf("parseStock: %v", err)
	}
	var out bytes.Buffer
	if err := Run(strings.NewReader(script), &out, products); err != nil {
		t.Fatalf("Run: %v", err)
	}
	return out.String()
}

const floatWithFivers = "0\n0\n5\n5\n5\n0\n0\n0\n" // 5x50p,20p,10p = £4.00

func TestSessionBuyWithChange(t *testing.T) {
	out := session(t, floatWithFivers+
		"select A1\n"+
		"insert 100\n"+
		"quit\n")

	if !strings.Contains(out, "Starting float: £4") {
		t.Errorf("float not reported correctly:\n%s", out)
	}
	if !strings.Contains(out, "Dispensed Water") {
		t.Errorf("product was not dispensed:\n%s", out)
	}
	if !strings.Contains(out, "Change: 1x20p") {
		t.Errorf("change was not 1x20p:\n%s", out)
	}
}

func TestSessionInsufficientThenCancel(t *testing.T) {
	out := session(t, floatWithFivers+
		"select A1\n"+
		"insert 50\n"+
		"cancel\n"+
		"quit\n")

	if !strings.Contains(out, "Please insert 30p more") {
		t.Errorf("expected a prompt for 30p more:\n%s", out)
	}
	if !strings.Contains(out, "Returned 1x50p") {
		t.Errorf("cancel should have returned the 50p:\n%s", out)
	}
}

func TestSessionCannotMakeChange(t *testing.T) {
	emptyFloat := "0\n0\n0\n0\n0\n0\n0\n0\n"
	out := session(t, emptyFloat+
		"select A1\n"+
		"insert 200\n"+
		"quit\n")

	if !strings.Contains(out, "Can't make the right change") {
		t.Errorf("expected a can't-make-change message:\n%s", out)
	}
	if !strings.Contains(out, "Returning 1x£2") {
		t.Errorf("the £2 should have been returned:\n%s", out)
	}
}

func TestSessionReload(t *testing.T) {
	out := session(t, floatWithFivers+
		"reload-product C1,Juice,130,4\n"+
		"list\n"+
		"quit\n")

	if !strings.Contains(out, "Loaded 4 x Juice into slot C1") {
		t.Errorf("reload not confirmed:\n%s", out)
	}
	if !strings.Contains(out, "Juice") {
		t.Errorf("new product not listed:\n%s", out)
	}
}
