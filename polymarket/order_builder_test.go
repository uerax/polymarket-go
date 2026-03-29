package polymarket

import (
	"context"
	"testing"
)

type orderSignerStub struct{}

func (s *orderSignerStub) Address(ctx context.Context) (string, error) {
	return "0x1111111111111111111111111111111111111111", nil
}

func (s *orderSignerStub) SignClobAuth(ctx context.Context, chainID Chain, timestamp int64, nonce int64) (string, error) {
	return "0xauth", nil
}

func (s *orderSignerStub) SignOrderTypedData(ctx context.Context, payload OrderTypedDataPayload) (string, error) {
	return "0xordersig", nil
}

func TestDefaultCreateOrderBuildsSignature(t *testing.T) {
	c := NewClient("https://clob.polymarket.com", ChainPolygon, &orderSignerStub{}, nil)
	fee := 10
	order, err := c.defaultCreateOrder(context.Background(), UserOrder{
		TokenID:    "1001",
		Price:      0.42,
		Size:       5,
		Side:       SideBUY,
		FeeRateBPS: &fee,
	}, &CreateOrderOptions{TickSize: "0.01"})
	if err != nil {
		t.Fatalf("defaultCreateOrder() error = %v", err)
	}
	if order.Signature != "0xordersig" {
		t.Fatalf("signature = %s, want 0xordersig", order.Signature)
	}
	if order.Maker == "" || order.Signer == "" {
		t.Fatalf("maker/signer should be populated")
	}
}

func TestBuildOrderTypedDataUsesNegRiskContract(t *testing.T) {
	c := NewClient("https://clob.polymarket.com", ChainPolygon, &orderSignerStub{}, nil)
	neg := true
	payload, err := c.buildOrderTypedData(context.Background(), SignedOrder{
		Salt:          "1",
		Maker:         "0x1111111111111111111111111111111111111111",
		Signer:        "0x1111111111111111111111111111111111111111",
		Taker:         "0x0000000000000000000000000000000000000000",
		TokenID:       "1001",
		MakerAmount:   "1000",
		TakerAmount:   "2000",
		Expiration:    "0",
		Nonce:         "0",
		FeeRateBPS:    "0",
		Side:          SideBUY,
		SignatureType: 0,
	}, &CreateOrderOptions{TickSize: "0.01", NegRisk: &neg})
	if err != nil {
		t.Fatalf("buildOrderTypedData() error = %v", err)
	}
	contract, ok := payload.Domain["verifyingContract"].(string)
	if !ok || contract == "" {
		t.Fatalf("verifyingContract missing")
	}
	if contract != "0xC5d563A36AE78145C45a50134d48A1215220f80a" {
		t.Fatalf("verifyingContract = %s, want neg risk exchange", contract)
	}
}
