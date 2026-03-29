package polymarket

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type requestCapture struct {
	RawQuery string
}

type builderHeaderProviderStub struct{}

func (b *builderHeaderProviderStub) GenerateBuilderHeaders(method string, path string, body string) (map[string]string, error) {
	return map[string]string{"X-BUILDER": "ok"}, nil
}

func (b *builderHeaderProviderStub) IsValid() bool {
	return true
}

type flakyRoundTripper struct {
	calls int
}

func (f *flakyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	f.calls++
	if f.calls == 1 {
		return nil, errors.New("temporary network error")
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`[{"asset_id":"1","tick_size":"0.01"}]`)),
		Request:    req,
	}, nil
}

type testSigner struct{}

func (s *testSigner) Address(ctx context.Context) (string, error) {
	return "0xabc", nil
}

func (s *testSigner) SignClobAuth(ctx context.Context, chainID Chain, timestamp int64, nonce int64) (string, error) {
	return "0xsig", nil
}

func (s *testSigner) SignOrderTypedData(ctx context.Context, payload OrderTypedDataPayload) (string, error) {
	return "0xordersig", nil
}

func TestAddQueryParamsCommaJoin(t *testing.T) {
	url, err := addQueryParams("https://example.com/test", &RequestOptions{Params: QueryParams{
		"ids": []string{"a", "b", "c"},
	}})
	if err != nil {
		t.Fatalf("addQueryParams() error = %v", err)
	}
	if !strings.Contains(url, "ids=a%2Cb%2Cc") {
		t.Fatalf("url = %s, want ids comma-joined", url)
	}
}

func TestThrowOnErrorReturnsApiError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "bad request"})
	}))
	defer server.Close()

	c := NewClient(server.URL, ChainPolygon, &testSigner{}, nil, WithThrowOnError(true))
	_, err := c.GetPrice(context.Background(), "token-1", "buy")
	if err == nil {
		t.Fatal("GetPrice() error = nil, want ApiError")
	}
	apiErr, ok := err.(*ApiError)
	if !ok {
		t.Fatalf("error type = %T, want *ApiError", err)
	}
	if apiErr.Status != http.StatusBadRequest {
		t.Fatalf("ApiError.Status = %d, want %d", apiErr.Status, http.StatusBadRequest)
	}
	if apiErr.Message != "bad request" {
		t.Fatalf("ApiError.Message = %q, want %q", apiErr.Message, "bad request")
	}
}

func TestGetTradesPaginationLoop(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"next_cursor": "NQ==",
				"data": []map[string]any{{"id": "1", "side": "BUY"}},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"next_cursor": EndCursor,
			"data": []map[string]any{{"id": "2", "side": "SELL"}},
		})
	}))
	defer server.Close()

	creds := &ApiKeyCreds{Key: "k", Secret: "c2VjcmV0", Passphrase: "p"}
	c := NewClient(server.URL, ChainPolygon, &testSigner{}, creds)
	trades, err := c.GetTrades(context.Background(), nil, false, "")
	if err != nil {
		t.Fatalf("GetTrades() error = %v", err)
	}
	if len(trades) != 2 {
		t.Fatalf("GetTrades() len = %d, want 2", len(trades))
	}
}

func TestGetRfqRequestsUsesCommaJoinedArrayParams(t *testing.T) {
	capture := &requestCapture{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capture.RawQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":        []map[string]any{},
			"next_cursor": EndCursor,
			"limit":       1,
			"count":       0,
		})
	}))
	defer server.Close()

	creds := &ApiKeyCreds{Key: "k", Secret: "c2VjcmV0", Passphrase: "p"}
	c := NewClient(server.URL, ChainPolygon, &testSigner{}, creds)
	_, err := c.GetRfqRequests(context.Background(), &GetRfqRequestsParams{
		RequestIDs: []string{"r1", "r2"},
		Markets:    []string{"m1", "m2"},
	})
	if err != nil {
		t.Fatalf("GetRfqRequests() error = %v", err)
	}
	if !strings.Contains(capture.RawQuery, "requestIds=r1%2Cr2") {
		t.Fatalf("raw query = %q, want comma-joined requestIds", capture.RawQuery)
	}
	if !strings.Contains(capture.RawQuery, "markets=m1%2Cm2") {
		t.Fatalf("raw query = %q, want comma-joined markets", capture.RawQuery)
	}
}

func TestGetRfqRequesterQuotesUsesCommaJoinedArrayParams(t *testing.T) {
	capture := &requestCapture{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capture.RawQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":        []map[string]any{},
			"next_cursor": EndCursor,
			"limit":       1,
			"count":       0,
		})
	}))
	defer server.Close()

	creds := &ApiKeyCreds{Key: "k", Secret: "c2VjcmV0", Passphrase: "p"}
	c := NewClient(server.URL, ChainPolygon, &testSigner{}, creds)
	_, err := c.GetRfqRequesterQuotes(context.Background(), &GetRfqQuotesParams{
		QuoteIDs:   []string{"q1", "q2"},
		RequestIDs: []string{"r1", "r2"},
		Markets:    []string{"m1", "m2"},
	})
	if err != nil {
		t.Fatalf("GetRfqRequesterQuotes() error = %v", err)
	}
	if !strings.Contains(capture.RawQuery, "quoteIds=q1%2Cq2") {
		t.Fatalf("raw query = %q, want comma-joined quoteIds", capture.RawQuery)
	}
	if !strings.Contains(capture.RawQuery, "requestIds=r1%2Cr2") {
		t.Fatalf("raw query = %q, want comma-joined requestIds", capture.RawQuery)
	}
	if !strings.Contains(capture.RawQuery, "markets=m1%2Cm2") {
		t.Fatalf("raw query = %q, want comma-joined markets", capture.RawQuery)
	}
}

func TestAreOrdersScoringNilParamsSendsEmptyBody(t *testing.T) {
	requestBody := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		buf, _ := io.ReadAll(r.Body)
		requestBody = string(buf)
		_ = json.NewEncoder(w).Encode(map[string]bool{"order-1": true})
	}))
	defer server.Close()

	creds := &ApiKeyCreds{Key: "k", Secret: "c2VjcmV0", Passphrase: "p"}
	c := NewClient(server.URL, ChainPolygon, &testSigner{}, creds)
	_, err := c.AreOrdersScoring(context.Background(), nil)
	if err != nil {
		t.Fatalf("AreOrdersScoring() error = %v", err)
	}
	if requestBody != "" {
		t.Fatalf("request body = %q, want empty", requestBody)
	}
}

func TestPostRetriesOnTransient5xx(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "temporary"})
			return
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{"asset_id": "1", "tick_size": "0.01"}})
	}))
	defer server.Close()

	c := NewClient(server.URL, ChainPolygon, &testSigner{}, nil, WithRetryOnError(true))
	_, err := c.GetOrderBooks(context.Background(), []BookParams{{TokenID: "1"}})
	if err != nil {
		t.Fatalf("GetOrderBooks() error = %v", err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestBuilderEndpointsDoNotRequireL2Creds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-BUILDER") != "ok" {
			t.Fatalf("missing builder header")
		}
		switch {
		case r.Method == http.MethodDelete && r.URL.Path == RevokeBuilderAPIKeyEndpoint:
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
		case r.Method == http.MethodGet && r.URL.Path == GetBuilderTradesEndpoint:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":        []map[string]any{},
				"next_cursor": EndCursor,
				"limit":       0,
				"count":       0,
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, ChainPolygon, &testSigner{}, nil, WithBuilderHeaderProvider(&builderHeaderProviderStub{}))
	if _, err := c.RevokeBuilderAPIKey(context.Background()); err != nil {
		t.Fatalf("RevokeBuilderAPIKey() error = %v", err)
	}
	if _, _, _, _, err := c.GetBuilderTrades(context.Background(), nil, ""); err != nil {
		t.Fatalf("GetBuilderTrades() error = %v", err)
	}
}

func TestGetNegRiskUsesCache(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(map[string]any{"neg_risk": true})
	}))
	defer server.Close()

	c := NewClient(server.URL, ChainPolygon, &testSigner{}, nil)
	v1, err := c.GetNegRisk(context.Background(), "token-1")
	if err != nil {
		t.Fatalf("GetNegRisk() first error = %v", err)
	}
	v2, err := c.GetNegRisk(context.Background(), "token-1")
	if err != nil {
		t.Fatalf("GetNegRisk() second error = %v", err)
	}
	if !v1 || !v2 {
		t.Fatalf("cached neg risk values = %v, %v, want true,true", v1, v2)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestGetFeeRateBpsUsesCache(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(map[string]any{"base_fee": 17})
	}))
	defer server.Close()

	c := NewClient(server.URL, ChainPolygon, &testSigner{}, nil)
	v1, err := c.GetFeeRateBps(context.Background(), "token-1")
	if err != nil {
		t.Fatalf("GetFeeRateBps() first error = %v", err)
	}
	v2, err := c.GetFeeRateBps(context.Background(), "token-1")
	if err != nil {
		t.Fatalf("GetFeeRateBps() second error = %v", err)
	}
	if v1 != 17 || v2 != 17 {
		t.Fatalf("cached fee values = %d, %d, want 17,17", v1, v2)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestPostRetriesOnTransientNetworkError(t *testing.T) {
	transport := &flakyRoundTripper{}
	httpClient := &http.Client{Transport: transport}
	c := NewClient("https://example.com", ChainPolygon, &testSigner{}, nil, WithRetryOnError(true), WithHTTPClient(httpClient))
	_, err := c.GetOrderBooks(context.Background(), []BookParams{{TokenID: "1"}})
	if err != nil {
		t.Fatalf("GetOrderBooks() error = %v", err)
	}
	if transport.calls != 2 {
		t.Fatalf("transport calls = %d, want 2", transport.calls)
	}
}
