# polymarket-go

English | [中文](./README.zh-CN.md)

A small Go client and CLI for Polymarket Gamma and CLOB APIs.

## Features

- List markets from the Gamma API
- Search/filter markets locally by query
- Get a market by ID
- Get a market by slug
- Get token price from the CLOB API
- Get order book data from the CLOB API

## Requirements

- Go 1.23.5 or newer

## Installation

### Library

```bash
go get github.com/uerax/polymarket-go/pkg/polymarket
```

### CLI

```bash
go build -o polymarket-go ./cmd
```

## Library Usage

```go
package main

import (
    "fmt"
    "log"
    "time"

    polymarket "github.com/uerax/polymarket-go/pkg/polymarket"
)

func main() {
    client := polymarket.NewClient(polymarket.Config{
        Timeout:      10 * time.Second,
        DefaultLimit: 10,
    })

    active := true
    markets, err := client.ListMarkets(polymarket.ListMarketsOptions{
        Limit:  5,
        Active: &active,
        Query:  "bitcoin",
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, market := range markets {
        fmt.Println(market.ID, market.Question)
    }
}
```

## CLI Usage

Build the CLI first:

```bash
go build -o polymarket-go ./cmd
```

### List markets

```bash
./polymarket-go markets --limit 5 --query bitcoin --active
```

### Get market by ID

```bash
./polymarket-go market --id 12345
```

### Get market by slug

```bash
./polymarket-go market --slug will-btc-hit-200k
```

### Get token price

```bash
./polymarket-go price --token-id <TOKEN_ID> --side buy
```

### Get order book

```bash
./polymarket-go book --token-id <TOKEN_ID>
```

## Development

Run tests:

```bash
go test ./...
```

Build all packages:

```bash
go build ./...
```

Run the CLI without building a binary:

```bash
go run ./cmd markets --limit 1
```

## Project Structure

```text
.
├── cmd/
│   └── main.go
├── pkg/
│   └── polymarket/
│       ├── client.go
│       ├── client_test.go
│       └── model.go
├── go.mod
├── README.md
└── README.zh-CN.md
```

## Notes

- Default Gamma base URL: `https://gamma-api.polymarket.com`
- Default CLOB base URL: `https://clob.polymarket.com`
- The CLI outputs formatted JSON to stdout.
