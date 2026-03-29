# polymarket-go (CLOB SDK)

Go SDK migrated from `clob-client` semantics for Polymarket CLOB.

## Scope

This repository now focuses on a single package:

- `clob`: CLOB client SDK (public endpoints + auth flows + order/rewards/builder/RFQ related methods)

Legacy `pkg/polymarket` and old CLI entrypoint were removed.

## Requirements

- Go 1.23.5+

## Install (as library)

```bash
go get github.com/uerax/polymarket-go/clob
```

## Run tests

```bash
go test ./...
```

## Package layout

```text
clob/
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

## Notes on TS parity

This SDK is aligned to TS `clob-client` behavior in key areas:

- Cursor pagination (`MA==` / `LTE=`)
- Error object mapping + optional `throwOnError`
- L1/L2 header flow
- Query serialization details (including repeated RFQ style query building)

Further parity improvements should continue inside `clob` only.
