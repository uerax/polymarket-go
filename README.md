# polymarket-go (CLOB SDK)

Go SDK for Polymarket CLOB API, migrated from `clob-client` semantics for TypeScript.

## Features

- **Easy-to-use API** for market data, trading, and account management
- **Multiple auth flows**: L1 (signer-based) and L2 (API key-based)
- **Cursor-based pagination** for large result sets
- **Automatic retry logic** for transient failures
- **Request validation** and error handling aligned with TS SDK

## Requirements

- Go 1.23.5+

## Installation

```bash
go get github.com/uerax/polymarket-go/polymarket
```

## Quick Start

### 1. Public Market Data (no auth required)

```go
package main

import (
	"context"
	"fmt"

	"github.com/uerax/polymarket-go/polymarket"
)

func main() {
	ctx := context.Background()

	// Create client (no auth for public data)
	client := polymarket.NewClient(
		"https://clob.polymarket.com",
		polymarket.ChainPolygon,
		nil,  // signer
		nil,  // api key creds
	)

	// Get server time
	ts, err := client.GetServerTime(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Server time: %d\n", ts)

	// Get order book for a token
	book, err := client.GetOrderBook(ctx, "12345")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Bids: %d, Asks: %d\n", len(book.Bids), len(book.Asks))

	// Get market prices
	prices, err := client.GetPrice(ctx, "12345", "BUY")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Prices: %v\n", prices)
}
```

### 2. Create API Key (L1 Auth Only)

```go
// Requires a ClobSigner implementation
client := polymarket.NewClient(
	"https://clob.polymarket.com",
	polymarket.ChainPolygon,
	signer, // Your signer implementation (ClobSigner interface)
	nil,
)

creds, err := client.CreateAPIKey(ctx, nil)
if err != nil {
	panic(err)
}
fmt.Printf("API Key: %s\n", creds.Key)
```

### 3. Query Account & Orders (L2 Auth)

```go
client := polymarket.NewClient(
	"https://clob.polymarket.com",
	polymarket.ChainPolygon,
	signer,  // L1 auth for signing
	&polymarket.ApiKeyCreds{
		Key:        "your-api-key",
		Secret:     "your-secret",
		Passphrase: "your-passphrase",
	},
)

// Get open orders
orders, err := client.GetOpenOrders(ctx, nil, false, "")
if err != nil {
	panic(err)
}
for _, order := range orders {
	fmt.Printf("Order: %s, Side: %s, Size: %s\n", order.ID, order.Side, order.Size)
}

// Get trades
trades, err := client.GetTrades(ctx, &polymarket.TradeParams{}, false, "")
if err != nil {
	panic(err)
}
for _, trade := range trades {
	fmt.Printf("Trade: %s, Price: %s\n", trade.ID, trade.Price)
}
```

### 4. Place an Order (L2 Auth Required)

```go
// First, build and sign an order using your signer
signedOrder, err := client.CreateOrder(ctx, polymarket.UserOrder{
	TokenID:    "12345",
	Price:      0.5,
	Size:       100.0,
	Side:       polymarket.SideBUY,
	Expiration: nil,
}, nil)
if err != nil {
	panic(err)
}

// Post the order
result, err := client.PostOrder(
	ctx,
	signedOrder,
	polymarket.OrderTypeGTC,
	false, // deferExec
	nil,   // postOnly
)
if err != nil {
	panic(err)
}
fmt.Printf("Order placed: %v\n", result)

// Cancel an order
cancelResult, err := client.CancelOrder(ctx, polymarket.OrderPayload{
	OrderID: "order-hash",
})
if err != nil {
	panic(err)
}
fmt.Printf("Cancel result: %v\n", cancelResult)
```

## Client Initialization

### Basic Options

```go
client := polymarket.NewClient(
	host,    // API host URL (e.g., "https://clob.polymarket.com")
	chain,   // Chain ID (ChainPolygon=137 or ChainAmoy=80002)
	signer,  // ClobSigner for L1 auth (nil if no L1 auth needed)
	creds,   // ApiKeyCreds for L2 auth (nil if no L2 auth needed)
	// ... optional client options
)
```

### Optional Client Options

```go
client := polymarket.NewClient(
	host, chain, signer, creds,
	polymarket.WithHTTPClient(customHTTPClient),
	polymarket.WithThrowOnError(true),           // Throw on API errors
	polymarket.WithRetryOnError(true),           // Retry transient errors
	polymarket.WithGeoBlockToken("your-token"),  // For geo-blocking bypass
	polymarket.WithUseServerTime(true),          // Use server time for requests
	polymarket.WithTickSizeTTL(5 * time.Minute), // Cache TTL for tick sizes
)
```

## API Reference

### Markets (Public - No Auth Required)

Market discovery and pricing endpoints:

- **GetOrderBook(ctx, tokenID)** - Get order book for a token
- **GetOrderBooks(ctx, params)** - Get multiple order books
- **GetPrice(ctx, tokenID, side)** - Get price for a token and side
- **GetPrices(ctx, params)** - Get prices for multiple tokens
- **GetSpread(ctx, tokenID)** - Get spread
- **GetMidpoint(ctx, tokenID)** - Get midpoint price
- **GetTickSize(ctx, tokenID)** - Get minimum tick size
- **GetNegRisk(ctx, tokenID)** - Check if token has negative risk
- **GetFeeRateBps(ctx, tokenID)** - Get fee rate in BPS
- **GetPricesHistory(ctx, params)** - Get historical prices
- **GetMarkets(ctx, nextCursor)** - Paginated market list
- **GetMarket(ctx, conditionID)** - Get market details

### Authentication (L1 Auth - Requires Signer)

API key and authentication management:

- **CreateAPIKey(ctx, nonce)** - Create new API key
- **DeriveAPIKey(ctx, nonce)** - Derive existing API key
- **CreateOrDeriveAPIKey(ctx, nonce)** - Create or derive (tries create first)
- **CreateReadonlyAPIKey(ctx)** - Create read-only key
- **GetReadonlyAPIKeys(ctx)** - List read-only keys
- **DeleteReadonlyAPIKey(ctx, key)** - Delete read-only key
- **ValidateReadonlyAPIKey(ctx, address, key)** - Validate key

### Accounts & Orders (L2 Auth - Requires API Key)

Order and account information:

- **GetOrder(ctx, orderID)** - Get single order details
- **GetOpenOrders(ctx, params, onlyFirstPage, nextCursor)** - List open orders
- **GetTrades(ctx, params, onlyFirstPage, nextCursor)** - List trades
- **GetTradesPaginated(ctx, params, nextCursor)** - Single page of trades
- **GetAPIKeys(ctx)** - List all API keys
- **GetClosedOnlyMode(ctx)** - Check if account is in closed-only mode
- **GetBalanceAllowance(ctx, params)** - Get balance and allowance

### Order Management (L2 Auth)

Place and manage orders:

- **PostOrder(ctx, order, orderType, deferExec, postOnly)** - Place single order
- **PostOrders(ctx, args, deferExec, defaultPostOnly)** - Batch place orders
- **CancelOrder(ctx, payload)** - Cancel by order ID
- **CancelOrders(ctx, orderHashes)** - Batch cancel by hash
- **CancelAll(ctx)** - Cancel all orders
- **CancelMarketOrders(ctx, params)** - Cancel by market/asset
- **IsOrderScoring(ctx, params)** - Check if order scored
- **AreOrdersScoring(ctx, params)** - Check multiple orders

### Rewards & Earnings (L2 Auth)

Liquidity rewards and earnings:

- **GetEarningsForUserForDay(ctx, date)** - Daily earnings breakdown
- **GetTotalEarningsForUserForDay(ctx, date)** - Total daily earnings
- **GetRewardPercentages(ctx)** - Reward percentage configuration
- **GetUserEarningsAndMarketsConfig(ctx, ...)** - Detailed earnings with market config
- **GetCurrentRewards(ctx)** - Current reward pools
- **GetRawRewardsForMarket(ctx, conditionID)** - Rewards for specific market

### RFQ (Request for Quote - L2 Auth)

RFQ workflow:

- **CreateRfqRequest(ctx, order, options)** - Create RFQ request
- **CancelRfqRequest(ctx, params)** - Cancel request
- **GetRfqRequests(ctx, params)** - List RFQ requests
- **CreateRfqQuote(ctx, quote, options)** - Create quote for request
- **CancelRfqQuote(ctx, params)** - Cancel quote
- **GetRfqRequesterQuotes(ctx, params)** - List quotes (requester side)
- **GetRfqQuoterQuotes(ctx, params)** - List quotes (quoter side)
- **GetRfqBestQuote(ctx, params)** - Get best quote for request
- **AcceptRfqQuote(ctx, payload)** - Accept a quote
- **ApproveRfqOrder(ctx, payload)** - Approve RFQ order
- **RFQConfig(ctx)** - Get RFQ configuration

### Builder API (Optional - Requires Builder Auth Header)

- **CreateBuilderAPIKey(ctx)** - Create builder key
- **GetBuilderAPIKeys(ctx)** - List builder keys
- **RevokeBuilderAPIKey(ctx)** - Revoke builder key
- **GetBuilderTrades(ctx, params, nextCursor)** - Get builder trades

## Common Patterns

### Pagination

Many list endpoints support cursor-based pagination:

```go
// Iterate through all results
var allResults []YourType
nextCursor := ""

for nextCursor != polymarket.EndCursor {
	result, err := client.SomeListEndpoint(ctx, params, false, nextCursor)
	if err != nil {
		panic(err)
	}
	allResults = append(allResults, result...)
	nextCursor = result.NextCursor // or set from response
}
```

Or fetch single pages:

```go
page, err := client.GetTradesPaginated(ctx, params, "")
if err != nil {
	panic(err)
}
fmt.Printf("Trades: %d, Next: %s\n", len(page.Trades), page.NextCursor)
```

### Error Handling

```go
result, err := client.PostOrder(ctx, order, orderType, false, nil)
if err != nil {
	// Check if it's an ApiError
	if apiErr, ok := err.(*polymarket.ApiError); ok {
		fmt.Printf("API Error: Status %d, Body: %s\n", apiErr.Status, string(apiErr.Body))
	} else {
		fmt.Printf("Error: %v\n", err)
	}
}
```

### Order Types

```go
const (
	OrderTypeGTC OrderType = "GTC"  // Good-til-Cancelled
	OrderTypeGTD OrderType = "GTD"  // Good-til-Date
	OrderTypeFOK OrderType = "FOK"  // Fill-or-Kill
	OrderTypeFAK OrderType = "FAK"  // Fill-and-Kill
)
```

### Authentication Levels

| Endpoint | Required Auth | Example |
|----------|---------------|---------|
| Market data | None | `GetOrderBook`, `GetPrice` |
| Create API Key | L1 (Signer) | `CreateAPIKey` |
| Account info | L2 (API Key) | `GetOpenOrders`, `GetTrades` |
| Place order | L2 (API Key) | `PostOrder` |
| Builder API | Builder Header | `GetBuilderTrades` |

## Implementing ClobSigner

The `ClobSigner` interface handles signature generation:

```go
type ClobSigner interface {
	// Get user's wallet address
	Address(ctx context.Context) (string, error)

	// Sign CLOB auth headers (for L1 auth)
	SignClobAuth(ctx context.Context, chainID Chain, timestamp int64, nonce int64) (string, error)

	// Sign order typed data (for order signing)
	SignOrderTypedData(ctx context.Context, payload OrderTypedDataPayload) (string, error)
}
```

## Testing

Run tests:

```bash
# All tests
go test ./...

# Package tests only
go test ./polymarket -v

# Single test
go test ./polymarket -run TestThrowOnErrorReturnsApiError -v
```

## Package Layout

```text
polymarket/
├── client.go            # Main client and API methods
├── constants.go         # Endpoint and cursor constants
├── types.go             # Request/response models
├── errors.go            # ApiError and auth errors
├── http_helpers.go      # HTTP + error mapping + retry + throwOnError
├── headers.go           # L1/L2 header generation
├── signer.go            # Signer abstraction
├── order_types.go       # Signature/builder related interfaces
└── http_helpers_test.go # Parity-focused behavior tests
```

## Notes on TS Parity

This SDK is aligned to TS `clob-client` behavior in key areas:

- Cursor pagination (`MA==` / `LTE=`)
- Error object mapping + optional `throwOnError`
- L1/L2 header flow
- Query serialization details (including repeated RFQ style query building)

Further parity improvements should continue inside `polymarket` only.

## License

MIT
