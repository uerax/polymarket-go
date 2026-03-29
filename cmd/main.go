package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	polymarket "github.com/uerax/polymarket-go/pkg/polymarket"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	client := polymarket.NewClient(polymarket.Config{Timeout: 10 * time.Second, DefaultLimit: 10})

	var err error
	switch os.Args[1] {
	case "markets":
		err = runMarkets(client, os.Args[2:])
	case "market":
		err = runMarket(client, os.Args[2:])
	case "price":
		err = runPrice(client, os.Args[2:])
	case "book":
		err = runBook(client, os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runMarkets(client *polymarket.Client, args []string) error {
	fs := flag.NewFlagSet("markets", flag.ContinueOnError)
	limit := fs.Int("limit", 10, "maximum number of markets")
	query := fs.String("query", "", "search query")
	active := fs.Bool("active", false, "filter active markets")
	closed := fs.Bool("closed", false, "filter closed markets")
	if err := fs.Parse(args); err != nil {
		return err
	}

	opts := polymarket.ListMarketsOptions{Limit: *limit, Query: *query}
	if fs.Lookup("active").Value.String() == "true" {
		opts.Active = active
	}
	if fs.Lookup("closed").Value.String() == "true" {
		opts.Closed = closed
	}

	markets, err := client.ListMarkets(opts)
	if err != nil {
		return err
	}
	return printJSON(markets)
}

func runMarket(client *polymarket.Client, args []string) error {
	fs := flag.NewFlagSet("market", flag.ContinueOnError)
	id := fs.String("id", "", "market id")
	slug := fs.String("slug", "", "market slug")
	if err := fs.Parse(args); err != nil {
		return err
	}

	switch {
	case *id != "":
		market, err := client.GetMarketByID(*id)
		if err != nil {
			return err
		}
		return printJSON(market)
	case *slug != "":
		market, err := client.GetMarketBySlug(*slug)
		if err != nil {
			return err
		}
		return printJSON(market)
	default:
		return fmt.Errorf("market requires --id or --slug")
	}
}

func runPrice(client *polymarket.Client, args []string) error {
	fs := flag.NewFlagSet("price", flag.ContinueOnError)
	tokenID := fs.String("token-id", "", "CLOB token id")
	side := fs.String("side", polymarket.PriceSideBuy, "price side: buy or sell")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *tokenID == "" {
		return fmt.Errorf("price requires --token-id")
	}

	price, err := client.GetTokenPrice(*tokenID, *side)
	if err != nil {
		return err
	}
	return printJSON(price)
}

func runBook(client *polymarket.Client, args []string) error {
	fs := flag.NewFlagSet("book", flag.ContinueOnError)
	tokenID := fs.String("token-id", "", "CLOB token id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *tokenID == "" {
		return fmt.Errorf("book requires --token-id")
	}

	book, err := client.GetOrderBook(*tokenID)
	if err != nil {
		return err
	}
	return printJSON(book)
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  polymarket-go markets [--limit N] [--query TEXT] [--active] [--closed]\n")
	fmt.Fprintf(os.Stderr, "  polymarket-go market --id ID | --slug SLUG\n")
	fmt.Fprintf(os.Stderr, "  polymarket-go price --token-id TOKEN_ID [--side buy|sell]\n")
	fmt.Fprintf(os.Stderr, "  polymarket-go book --token-id TOKEN_ID\n")
}
