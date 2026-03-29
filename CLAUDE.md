# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This repo is **clob-only** now.

- Primary package: `clob`
- Goal: align Go SDK behavior with TS `clob-client` semantics
- Legacy package/CLI from old `pkg/polymarket` flow was removed and should not be reintroduced.

## Commands

- Run tests:
  - `go test ./...`
- Run package tests only:
  - `go test ./clob -v`
- Run one test:
  - `go test ./clob -run TestThrowOnErrorReturnsApiError -v`

## Architecture (high level)

- `clob/client.go`
  - Main API surface (public, auth, orders, rewards, builder, RFQ methods)
  - Cursor pagination behavior and response mapping
- `clob/http_helpers.go`
  - Shared HTTP wrapper, non-2xx error object mapping, optional throw-on-error, POST retry
- `clob/headers.go`
  - L1/L2 auth headers and HMAC signature assembly
- `clob/types.go`
  - Wire models and request payloads
- `clob/constants.go`
  - Endpoint constants and cursor constants
- `clob/errors.go`
  - `ApiError` and auth/builder error constants
- `clob/signer.go`, `clob/order_types.go`
  - Signer and builder integration abstractions

## Important Constraints

- Preserve TS parity for:
  - error model (`status`, body mapping, throwOnError behavior)
  - cursor pagination (`MA==`/`LTE=` semantics)
  - naming and response structures where practical in Go
  - L1/L2 auth header behavior

- Keep new work inside `clob`.
- Do not add back old `pkg/polymarket` package or old CLI entrypoint unless explicitly requested.
