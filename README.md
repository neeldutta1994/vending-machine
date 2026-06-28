# Vending Machine

A vending machine you drive from the command line. It sells products, takes UK
coins, returns the right change (or asks for more money), and keeps its stock and
cash float up to date. State lives in memory — there is no database and no GUI.

The code is organised along Domain-Driven Design lines so a team picking it up
can find their way around quickly and extend it without surprises.

## What it does

- Loads its product inventory from a file (`input.txt`).
- Asks, on startup, how much money the machine already holds (the change float).
- Lets you select a product, insert coins, and get the product plus any change.
- Tells you how much more is needed when too little is inserted.
- Refuses a sale and hands the money back if it cannot make the exact change.
- Can be restocked and have its float topped up while running.
- Keeps the running state of products and change correct after every operation.

Coins are the eight UK denominations: 1p, 2p, 5p, 10p, 20p, 50p, £1, £2.

## Running it

Requires Go (see `go.mod` for the version).

```bash
go run .              # uses input.txt in the current directory
go run . myfile.txt   # or point it at another inventory file
```

```bash
go test ./...         # run the test suite
go vet ./...          # static checks
```

### The inventory file

The machine reads its products from a plain text file, one product per line:

```
slot,product name,price,quantity
```

The price is in pence. Blank lines and lines starting with `#` are ignored.
The bundled `input.txt`:

```
# Vending machine inventory.
# One product per line: slot,product name,price (in pence),quantity
A1,Water,80,10
A2,Cola,120,8
A3,Diet Cola,120,8
B1,Crisps,95,6
B2,Chocolate,110,5
B3,Gum,45,12
C1,Sparkling Water,90,7
```

### Starting the machine

On startup the program asks how many of each coin the machine already holds.
Press Enter on its own to mean "none". For example, loading five each of 50p,
20p and 10p (a £4 float):

```
Enter how much money the machine already holds.
Type the number of each coin and press Enter (Enter on its own means none):
  £2   x 0
  £1   x 0
  50p  x 5
  20p  x 5
  10p  x 5
  5p   x
  2p   x
  1p   x

Machine ready. Starting float: £4.
```

### Commands

Once running, type `help` to see:

```
  list                              show products and prices
  select <slot>                     choose a product (e.g. select A1)
  insert <pence>                    insert a coin: 1 2 5 10 20 50 100 200
  balance                           show money inserted so far
  cancel                            cancel the sale and get your coins back
  state                             show stock and the change float
  reload-product <slot,name,price,quantity>   add or top up a product
  reload-coins                      add coins to the change float
  help                              show this list
  quit                              exit
```

A coin is inserted by its value in pence: `insert 100` is a £1 coin.

### A worked example

Buying the 80p water with a £1 coin:

```
> select A1
Balance 0p. Please insert 80p more.
> insert 100
Dispensed Water. Change: 1x20p.
```

You can also insert money first and select afterwards — the sale completes the
moment both a product is chosen and the balance reaches the price.

Restocking and topping up the float while the machine is running:

```
> reload-product A1,Water,80,10
Loaded 10 x Water into slot A1.
> reload-coins
Add coins to the float. Number of each coin (Enter on its own means none):
  £2   x 5
  ...
Added £10. Float is now £14.
```

If too much money is inserted and the float can't make the exact change, the
sale is refused and the coins are returned:

```
> insert 200
Can't make the right change, sorry. Returning 1x£2.
```

## Layout

```
.
├── main.go                     # entry point: load file, start the CLI
├── input.txt                   # product inventory (slot,name,price,quantity)
├── internal/
│   ├── cli/                    # command-line front end
│   │   ├── cli.go              #   the interactive command loop
│   │   ├── loader.go           #   reads/parses the inventory file
│   │   └── format.go           #   small printing/parsing helpers
│   ├── money/                  # the money domain: amounts, coins, change-making
│   │   ├── money.go            #   Money — an amount in pence
│   │   ├── denomination.go     #   Denomination — an accepted coin
│   │   ├── coins.go            #   Coins — a value-object multiset of coins
│   │   └── change.go           #   MakeChange — works out coins to give back
│   ├── catalog/
│   │   └── product.go          # Product — code, name, price
│   └── vending/
│       ├── machine.go          # Machine — the aggregate root
│       ├── inventory.go        # stock held by the machine (internal)
│       └── errors.go           # domain errors
└── README.md
```

### Why it is split this way

The brief is small, but it is the kind of thing that grows (card payments,
telemetry, multiple machines, a real persistence layer). DDD keeps the rules of
the business in one place and the plumbing out of the way.

- **`money`** is a self-contained domain with no knowledge of vending. `Money`,
  `Denomination` and `Coins` are *value objects*: comparing two by value is
  meaningful, and they are effectively immutable — `Coins.Add`, `.With` and
  `.Remove` return a new collection rather than changing the one you hold. That
  immutability is what lets the machine compute change against a candidate pool
  of coins and only commit once it knows the sale will go through.
- **`catalog`** holds the `Product` value object. It is separate from the
  machine because a catalogue of products is a concept in its own right and is
  likely to gain fields (barcodes, categories, supplier).
- **`vending`** holds the `Machine` *aggregate root*. The machine is the
  consistency boundary: stock and float only ever change through it, and never
  half-way. A product leaves stock in the same step its change is committed, so
  the two can't drift apart.
- **`cli`** is the presentation layer: prompts, command parsing and file
  loading live here, so the domain never touches stdin, files or formatting. It
  takes an `io.Reader`/`io.Writer`, which is what lets the session be scripted
  in tests.
- **`main.go`** is the composition root: it wires the file loader to the CLI. In
  a larger system an HTTP handler, a hardware driver, or a message consumer
  would sit alongside the CLI here.

`internal/` is used so these packages can't be imported by unrelated modules
while the public surface is still being shaped.

## Using the machine as a library

The CLI is a thin shell over `vending.Machine`; the same API is available to any
caller.

```go
water, _ := catalog.NewProduct("A1", "Water", 80) // 80p

m := vending.New(
    []vending.Stock{{Product: water, Quantity: 5}},
    money.NewCoins(map[money.Denomination]int{money.TwentyPence: 10}),
)

m.Select("A1")
r, _ := m.InsertCoin(money.OnePound)
// r.Dispensed == true, r.Product.Name == "Water", r.Change.Total() == 20p
```

`Result` reports where a sale stands after each `Select`/`InsertCoin`:

| Field         | Meaning                                                        |
|---------------|---------------------------------------------------------------|
| `Dispensed`   | the product and any change have been handed over               |
| `Product`     | what was dispensed (when `Dispensed`)                          |
| `Change`      | coins returned (when `Dispensed`; may be empty)               |
| `Balance`     | total inserted for the current sale                            |
| `Outstanding` | how much more is needed (0 once dispensed, or before selection)|

Other methods: `Cancel()` refunds the inserted coins, `LoadProducts(...)` and
`LoadCoins(...)` restock and top up the float, and `Stock()` / `Float()` /
`Balance()` read current state.

## Design decisions and trade-offs

- **Pence, not floats.** Every amount is a whole number of pence (`money.Money`).
  Money never accumulates rounding error, and coins and prices stay exact.
- **The inserted coins are available as change.** Before working out change, the
  coins just inserted are pooled with the float. This matches a real machine and
  means, for example, a customer's £1 can itself become someone's change.
- **Greedy change-making.** `MakeChange` takes the largest coins first. For the
  UK coin set this gives the optimal (fewest-coins) result when coins are
  plentiful. When the float has run low on a particular coin, greedy can fail to
  find a combination that an exhaustive search would — in that case it returns
  `ErrNoChange`, the sale is refused, and the money is returned rather than
  giving wrong change. An exhaustive search was deliberately not used: it is more
  code to maintain for a case a sensibly stocked float rarely hits, and the
  failure mode (refuse + refund) is safe either way. Swapping the algorithm means
  changing one function.
- **No concurrency.** A vending machine serves one customer at a time, so there
  is no locking. If that assumption changes, the single aggregate root is the one
  place to add synchronisation.
- **In-memory state.** State lives in the `Machine`. Persistence, if ever needed,
  fits behind a repository that loads and saves the aggregate, without touching
  the domain logic.

## Tests

- `internal/money` — formatting, coin arithmetic (including that value objects
  don't mutate), and change-making (happy path, falling back to smaller coins,
  and the impossible case).
- `internal/vending` — the machine end to end: exact payment, overpayment with
  change, insufficient funds, cancel/refund, the can't-make-change refusal,
  inserted coins funding change, selecting before or after paying, rejecting
  invalid coins, out-of-stock and unknown products, and reloading.
- `internal/cli` — inventory-file parsing, and full scripted sessions covering a
  sale with change, insufficient funds then cancel, the can't-make-change path,
  and reloading a product.
