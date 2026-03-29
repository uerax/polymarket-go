package polymarket

import "context"

type OrderTypedDataPayload struct {
	PrimaryType string
	Domain      map[string]any
	Types       map[string][]map[string]string
	Message     map[string]any
}

type ClobSigner interface {
	Address(ctx context.Context) (string, error)
	SignClobAuth(ctx context.Context, chainID Chain, timestamp int64, nonce int64) (string, error)
	SignOrderTypedData(ctx context.Context, payload OrderTypedDataPayload) (string, error)
}
