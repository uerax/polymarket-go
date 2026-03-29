package clob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	host          string
	chainID       Chain
	httpClient    *http.Client
	signer        ClobSigner
	creds         *ApiKeyCreds
	geoBlockToken string
	useServerTime bool
	retryOnError  bool
	throwOnError  bool

	signatureType SignatureType

	builderHeaderProvider BuilderHeaderProvider
	createOrderFn         func(ctx context.Context, order UserOrder, options *CreateOrderOptions) (SignedOrder, error)
	createMarketOrderFn   func(ctx context.Context, order UserMarketOrder, options *CreateOrderOptions) (SignedOrder, error)

	tickSizes          map[string]TickSize
	tickSizeTimestamps map[string]int64
	tickSizeTTL        time.Duration
	negRisks           map[string]bool
	feeRates           map[string]int
}

type ClientOption func(*Client)

func WithHTTPClient(h *http.Client) ClientOption {
	return func(c *Client) { c.httpClient = h }
}

func WithGeoBlockToken(token string) ClientOption {
	return func(c *Client) { c.geoBlockToken = token }
}

func WithUseServerTime(v bool) ClientOption {
	return func(c *Client) { c.useServerTime = v }
}

func WithRetryOnError(v bool) ClientOption {
	return func(c *Client) { c.retryOnError = v }
}

func WithThrowOnError(v bool) ClientOption {
	return func(c *Client) { c.throwOnError = v }
}

func WithTickSizeTTL(d time.Duration) ClientOption {
	return func(c *Client) { c.tickSizeTTL = d }
}

func WithSignatureType(t SignatureType) ClientOption {
	return func(c *Client) { c.signatureType = t }
}

func WithBuilderHeaderProvider(provider BuilderHeaderProvider) ClientOption {
	return func(c *Client) { c.builderHeaderProvider = provider }
}

func WithOrderFactories(
	createOrder func(ctx context.Context, order UserOrder, options *CreateOrderOptions) (SignedOrder, error),
	createMarketOrder func(ctx context.Context, order UserMarketOrder, options *CreateOrderOptions) (SignedOrder, error),
) ClientOption {
	return func(c *Client) {
		c.createOrderFn = createOrder
		c.createMarketOrderFn = createMarketOrder
	}
}

func NewClient(host string, chainID Chain, signer ClobSigner, creds *ApiKeyCreds, opts ...ClientOption) *Client {
	c := &Client{
		host:               trimSlash(host),
		chainID:            chainID,
		httpClient:         &http.Client{Timeout: 10 * time.Second},
		signer:             signer,
		creds:              creds,
		signatureType:      SignatureTypeEOA,
		tickSizes:          map[string]TickSize{},
		tickSizeTimestamps: map[string]int64{},
		tickSizeTTL:        5 * time.Minute,
		negRisks:           map[string]bool{},
		feeRates:           map[string]int{},
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return c
}

func trimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

func (c *Client) endpoint(path string) string {
	return c.host + path
}

func (c *Client) requestOptions(headers map[string]string, params QueryParams, data any) *RequestOptions {
	if params == nil {
		params = QueryParams{}
	}
	if c.geoBlockToken != "" {
		params["geo_block_token"] = c.geoBlockToken
	}
	return &RequestOptions{Headers: headers, Params: params, Data: data}
}

func decodeAny[T any](v any) (T, error) {
	var out T
	b, err := json.Marshal(v)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(b, &out)
	return out, err
}

func asMap(v any) map[string]any {
	m, ok := v.(map[string]any)
	if ok {
		return m
	}
	return map[string]any{}
}

func (c *Client) getServerTimestamp(ctx context.Context) (*int64, error) {
	if !c.useServerTime {
		return nil, nil
	}
	ts, err := c.GetServerTime(ctx)
	if err != nil {
		return nil, err
	}
	return &ts, nil
}

func (c *Client) ensureL1Auth() error {
	if c.signer == nil {
		return ErrL1AuthUnavailable
	}
	return nil
}

func (c *Client) ensureL2Auth() error {
	if err := c.ensureL1Auth(); err != nil {
		return err
	}
	if c.creds == nil {
		return ErrL2AuthNotAvailable
	}
	return nil
}

func (c *Client) canBuilderAuth() bool {
	return c.builderHeaderProvider != nil && c.builderHeaderProvider.IsValid()
}

func (c *Client) buildRequestHeaders(ctx context.Context, method string, requestPath string, body string, requireL2 bool) (map[string]string, error) {
	if requireL2 {
		if err := c.ensureL2Auth(); err != nil {
			return nil, err
		}
		headers, err := c.l2Headers(ctx, method, requestPath, body)
		if err != nil {
			return nil, err
		}
		if c.canBuilderAuth() {
			builderHeaders, bErr := c.builderHeaderProvider.GenerateBuilderHeaders(method, requestPath, body)
			if bErr != nil {
				return nil, bErr
			}
			for k, v := range builderHeaders {
				headers[k] = v
			}
		}
		return headers, nil
	}
	return nil, nil
}

func (c *Client) buildBuilderOnlyHeaders(method string, requestPath string, body string) (map[string]string, error) {
	if !c.canBuilderAuth() {
		return nil, ErrBuilderAuthRequired
	}
	headers, err := c.builderHeaderProvider.GenerateBuilderHeaders(method, requestPath, body)
	if err != nil {
		return nil, err
	}
	if headers == nil {
		return nil, ErrBuilderAuthFailed
	}
	return headers, nil
}

func (c *Client) toL2OrderPayload(order SignedOrder, orderType OrderType, deferExec bool, postOnly *bool) (NewOrder, error) {
	if postOnly != nil && *postOnly && orderType != OrderTypeGTC && orderType != OrderTypeGTD {
		return NewOrder{}, fmt.Errorf("postOnly is only supported for GTC and GTD orders")
	}
	salt, err := strconv.ParseInt(order.Salt, 10, 64)
	if err != nil {
		return NewOrder{}, err
	}
	var payload NewOrder
	payload.DeferExec = deferExec
	payload.Owner = ""
	if c.creds != nil {
		payload.Owner = c.creds.Key
	}
	payload.OrderType = orderType
	payload.PostOnly = postOnly
	payload.Order.Salt = salt
	payload.Order.Maker = order.Maker
	payload.Order.Signer = order.Signer
	payload.Order.Taker = order.Taker
	payload.Order.TokenID = order.TokenID
	payload.Order.MakerAmount = order.MakerAmount
	payload.Order.TakerAmount = order.TakerAmount
	payload.Order.Side = order.Side
	payload.Order.Expiration = order.Expiration
	payload.Order.Nonce = order.Nonce
	payload.Order.FeeRateBPS = order.FeeRateBPS
	payload.Order.SignatureType = order.SignatureType
	payload.Order.Signature = order.Signature
	return payload, nil
}

func (c *Client) GetOK(ctx context.Context) (any, error) {
	_ = ctx
	return c.get(c.endpoint("/"), c.requestOptions(nil, nil, nil))
}

func (c *Client) GetServerTime(ctx context.Context) (int64, error) {
	_ = ctx
	res, err := c.get(c.endpoint(TimeEndpoint), c.requestOptions(nil, nil, nil))
	if err != nil {
		return 0, err
	}
	switch v := res.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case string:
		i, pErr := strconv.ParseInt(v, 10, 64)
		if pErr != nil {
			return 0, pErr
		}
		return i, nil
	default:
		return 0, fmt.Errorf("unexpected server time response")
	}
}

func (c *Client) GetSamplingSimplifiedMarkets(ctx context.Context, nextCursor string) (PaginationPayload, error) {
	return c.getPagination(ctx, GetSamplingSimplifiedMarketsEndpoint, nextCursor)
}

func (c *Client) GetSamplingMarkets(ctx context.Context, nextCursor string) (PaginationPayload, error) {
	return c.getPagination(ctx, GetSamplingMarketsEndpoint, nextCursor)
}

func (c *Client) GetSimplifiedMarkets(ctx context.Context, nextCursor string) (PaginationPayload, error) {
	return c.getPagination(ctx, GetSimplifiedMarketsEndpoint, nextCursor)
}

func (c *Client) GetMarkets(ctx context.Context, nextCursor string) (PaginationPayload, error) {
	return c.getPagination(ctx, GetMarketsEndpoint, nextCursor)
}

func (c *Client) getPagination(ctx context.Context, endpoint string, nextCursor string) (PaginationPayload, error) {
	_ = ctx
	if nextCursor == "" {
		nextCursor = InitialCursor
	}
	res, err := c.get(c.endpoint(endpoint), c.requestOptions(nil, QueryParams{"next_cursor": nextCursor}, nil))
	if err != nil {
		return PaginationPayload{}, err
	}
	return decodeAny[PaginationPayload](res)
}

func (c *Client) GetMarket(ctx context.Context, conditionID string) (any, error) {
	_ = ctx
	return c.get(c.endpoint(GetMarketEndpoint+conditionID), c.requestOptions(nil, nil, nil))
}

func (c *Client) GetOrderBook(ctx context.Context, tokenID string) (OrderBookSummary, error) {
	_ = ctx
	res, err := c.get(c.endpoint(GetOrderBookEndpoint), c.requestOptions(nil, QueryParams{"token_id": tokenID}, nil))
	if err != nil {
		return OrderBookSummary{}, err
	}
	book, err := decodeAny[OrderBookSummary](res)
	if err != nil {
		return OrderBookSummary{}, err
	}
	if book.AssetID != "" && book.TickSize != "" {
		c.tickSizes[book.AssetID] = TickSize(book.TickSize)
		c.tickSizeTimestamps[book.AssetID] = time.Now().UnixMilli()
	}
	return book, nil
}

func (c *Client) GetOrderBooks(ctx context.Context, params []BookParams) ([]OrderBookSummary, error) {
	_ = ctx
	res, err := c.post(c.endpoint(GetOrderBooksEndpoint), c.requestOptions(nil, nil, params))
	if err != nil {
		return nil, err
	}
	books, err := decodeAny[[]OrderBookSummary](res)
	if err != nil {
		return nil, err
	}
	for _, b := range books {
		if b.AssetID != "" && b.TickSize != "" {
			c.tickSizes[b.AssetID] = TickSize(b.TickSize)
			c.tickSizeTimestamps[b.AssetID] = time.Now().UnixMilli()
		}
	}
	return books, nil
}

func (c *Client) GetTickSize(ctx context.Context, tokenID string) (TickSize, error) {
	_ = ctx
	if ts, ok := c.tickSizeTimestamps[tokenID]; ok {
		if v, ok2 := c.tickSizes[tokenID]; ok2 && time.Since(time.UnixMilli(ts)) < c.tickSizeTTL {
			return v, nil
		}
	}
	res, err := c.get(c.endpoint(GetTickSizeEndpoint), c.requestOptions(nil, QueryParams{"token_id": tokenID}, nil))
	if err != nil {
		return "", err
	}
	m := asMap(res)
	if errValue, ok := m["error"]; ok {
		return "", fmt.Errorf("%v", errValue)
	}
	v := fmt.Sprintf("%v", m["minimum_tick_size"])
	c.tickSizes[tokenID] = TickSize(v)
	c.tickSizeTimestamps[tokenID] = time.Now().UnixMilli()
	return TickSize(v), nil
}

func (c *Client) ClearTickSizeCache(tokenID string) {
	if tokenID != "" {
		delete(c.tickSizes, tokenID)
		delete(c.tickSizeTimestamps, tokenID)
		return
	}
	c.tickSizes = map[string]TickSize{}
	c.tickSizeTimestamps = map[string]int64{}
}

func (c *Client) GetNegRisk(ctx context.Context, tokenID string) (bool, error) {
	_ = ctx
	if v, ok := c.negRisks[tokenID]; ok {
		return v, nil
	}
	res, err := c.get(c.endpoint(GetNegRiskEndpoint), c.requestOptions(nil, QueryParams{"token_id": tokenID}, nil))
	if err != nil {
		return false, err
	}
	m := asMap(res)
	if errValue, ok := m["error"]; ok {
		return false, fmt.Errorf("%v", errValue)
	}
	v, _ := m["neg_risk"].(bool)
	c.negRisks[tokenID] = v
	return v, nil
}

func (c *Client) GetFeeRateBps(ctx context.Context, tokenID string) (int, error) {
	_ = ctx
	if v, ok := c.feeRates[tokenID]; ok {
		return v, nil
	}
	res, err := c.get(c.endpoint(GetFeeRateEndpoint), c.requestOptions(nil, QueryParams{"token_id": tokenID}, nil))
	if err != nil {
		return 0, err
	}
	m := asMap(res)
	if errValue, ok := m["error"]; ok {
		return 0, fmt.Errorf("%v", errValue)
	}
	s := fmt.Sprintf("%v", m["base_fee"])
	f, pErr := strconv.ParseFloat(s, 64)
	if pErr != nil {
		return 0, pErr
	}
	c.feeRates[tokenID] = int(f)
	return int(f), nil
}

func (c *Client) GetMidpoint(ctx context.Context, tokenID string) (any, error) {
	_ = ctx
	return c.get(c.endpoint(GetMidpointEndpoint), c.requestOptions(nil, QueryParams{"token_id": tokenID}, nil))
}

func (c *Client) GetMidpoints(ctx context.Context, params []BookParams) (any, error) {
	_ = ctx
	return c.post(c.endpoint(GetMidpointsEndpoint), c.requestOptions(nil, nil, params))
}

func (c *Client) GetPrice(ctx context.Context, tokenID string, side string) (any, error) {
	_ = ctx
	return c.get(c.endpoint(GetPriceEndpoint), c.requestOptions(nil, QueryParams{"token_id": tokenID, "side": side}, nil))
}

func (c *Client) GetPrices(ctx context.Context, params []BookParams) (any, error) {
	_ = ctx
	return c.post(c.endpoint(GetPricesEndpoint), c.requestOptions(nil, nil, params))
}

func (c *Client) GetSpread(ctx context.Context, tokenID string) (any, error) {
	_ = ctx
	return c.get(c.endpoint(GetSpreadEndpoint), c.requestOptions(nil, QueryParams{"token_id": tokenID}, nil))
}

func (c *Client) GetSpreads(ctx context.Context, params []BookParams) (any, error) {
	_ = ctx
	return c.post(c.endpoint(GetSpreadsEndpoint), c.requestOptions(nil, nil, params))
}

func (c *Client) GetLastTradePrice(ctx context.Context, tokenID string) (any, error) {
	_ = ctx
	return c.get(c.endpoint(GetLastTradePriceEndpoint), c.requestOptions(nil, QueryParams{"token_id": tokenID}, nil))
}

func (c *Client) GetLastTradesPrices(ctx context.Context, params []BookParams) (any, error) {
	_ = ctx
	return c.post(c.endpoint(GetLastTradesPricesEndpoint), c.requestOptions(nil, nil, params))
}

func (c *Client) GetPricesHistory(ctx context.Context, params PriceHistoryFilterParams) ([]MarketPrice, error) {
	_ = ctx
	q := QueryParams{}
	if params.Market != "" {
		q["market"] = params.Market
	}
	if params.StartTS > 0 {
		q["startTs"] = params.StartTS
	}
	if params.EndTS > 0 {
		q["endTs"] = params.EndTS
	}
	if params.Fidelity > 0 {
		q["fidelity"] = params.Fidelity
	}
	if params.Interval != "" {
		q["interval"] = string(params.Interval)
	}
	res, err := c.get(c.endpoint(GetPricesHistoryEndpoint), c.requestOptions(nil, q, nil))
	if err != nil {
		return nil, err
	}
	return decodeAny[[]MarketPrice](res)
}

func (c *Client) CreateAPIKey(ctx context.Context, nonce *int64) (ApiKeyCreds, error) {
	if err := c.ensureL1Auth(); err != nil {
		return ApiKeyCreds{}, err
	}
	ts, err := c.getServerTimestamp(ctx)
	if err != nil {
		return ApiKeyCreds{}, err
	}
	headers, err := createL1Headers(ctx, c.signer, c.chainID, nonce, ts)
	if err != nil {
		return ApiKeyCreds{}, err
	}
	res, err := c.post(c.endpoint(CreateAPIKeyEndpoint), c.requestOptions(headers, nil, nil))
	if err != nil {
		return ApiKeyCreds{}, err
	}
	raw, err := decodeAny[ApiKeyRaw](res)
	if err != nil {
		return ApiKeyCreds{}, err
	}
	return ApiKeyCreds{Key: raw.APIKey, Secret: raw.Secret, Passphrase: raw.Passphrase}, nil
}

func (c *Client) DeriveAPIKey(ctx context.Context, nonce *int64) (ApiKeyCreds, error) {
	if err := c.ensureL1Auth(); err != nil {
		return ApiKeyCreds{}, err
	}
	ts, err := c.getServerTimestamp(ctx)
	if err != nil {
		return ApiKeyCreds{}, err
	}
	headers, err := createL1Headers(ctx, c.signer, c.chainID, nonce, ts)
	if err != nil {
		return ApiKeyCreds{}, err
	}
	res, err := c.get(c.endpoint(DeriveAPIKeyEndpoint), c.requestOptions(headers, nil, nil))
	if err != nil {
		return ApiKeyCreds{}, err
	}
	raw, err := decodeAny[ApiKeyRaw](res)
	if err != nil {
		return ApiKeyCreds{}, err
	}
	return ApiKeyCreds{Key: raw.APIKey, Secret: raw.Secret, Passphrase: raw.Passphrase}, nil
}

func (c *Client) CreateOrDeriveAPIKey(ctx context.Context, nonce *int64) (ApiKeyCreds, error) {
	created, err := c.CreateAPIKey(ctx, nonce)
	if err != nil {
		return ApiKeyCreds{}, err
	}
	if created.Key == "" {
		return c.DeriveAPIKey(ctx, nonce)
	}
	return created, nil
}

func (c *Client) l2Headers(ctx context.Context, method string, requestPath string, body string) (map[string]string, error) {
	ts, err := c.getServerTimestamp(ctx)
	if err != nil {
		return nil, err
	}
	return createL2Headers(ctx, c.signer, *c.creds, L2HeaderArgs{Method: method, RequestPath: requestPath, Body: body}, ts)
}

func (c *Client) GetAPIKeys(ctx context.Context) (ApiKeysResponse, error) {
	if err := c.ensureL2Auth(); err != nil {
		return ApiKeysResponse{}, err
	}
	headers, err := c.l2Headers(ctx, http.MethodGet, GetAPIKeysEndpoint, "")
	if err != nil {
		return ApiKeysResponse{}, err
	}
	res, err := c.get(c.endpoint(GetAPIKeysEndpoint), c.requestOptions(headers, nil, nil))
	if err != nil {
		return ApiKeysResponse{}, err
	}
	return decodeAny[ApiKeysResponse](res)
}

func (c *Client) GetClosedOnlyMode(ctx context.Context) (BanStatus, error) {
	if err := c.ensureL2Auth(); err != nil {
		return BanStatus{}, err
	}
	headers, err := c.l2Headers(ctx, http.MethodGet, ClosedOnlyEndpoint, "")
	if err != nil {
		return BanStatus{}, err
	}
	res, err := c.get(c.endpoint(ClosedOnlyEndpoint), c.requestOptions(headers, nil, nil))
	if err != nil {
		return BanStatus{}, err
	}
	return decodeAny[BanStatus](res)
}

func (c *Client) DeleteAPIKey(ctx context.Context) (any, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.l2Headers(ctx, http.MethodDelete, DeleteAPIKeyEndpoint, "")
	if err != nil {
		return nil, err
	}
	return c.del(c.endpoint(DeleteAPIKeyEndpoint), c.requestOptions(headers, nil, nil))
}

func (c *Client) CreateReadonlyAPIKey(ctx context.Context) (ReadonlyAPIKeyResponse, error) {
	if err := c.ensureL2Auth(); err != nil {
		return ReadonlyAPIKeyResponse{}, err
	}
	headers, err := c.l2Headers(ctx, http.MethodPost, CreateReadonlyAPIKeyEndpoint, "")
	if err != nil {
		return ReadonlyAPIKeyResponse{}, err
	}
	res, err := c.post(c.endpoint(CreateReadonlyAPIKeyEndpoint), c.requestOptions(headers, nil, nil))
	if err != nil {
		return ReadonlyAPIKeyResponse{}, err
	}
	return decodeAny[ReadonlyAPIKeyResponse](res)
}

func (c *Client) GetReadonlyAPIKeys(ctx context.Context) ([]string, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.l2Headers(ctx, http.MethodGet, GetReadonlyAPIKeysEndpoint, "")
	if err != nil {
		return nil, err
	}
	res, err := c.get(c.endpoint(GetReadonlyAPIKeysEndpoint), c.requestOptions(headers, nil, nil))
	if err != nil {
		return nil, err
	}
	return decodeAny[[]string](res)
}

func (c *Client) DeleteReadonlyAPIKey(ctx context.Context, key string) (bool, error) {
	if err := c.ensureL2Auth(); err != nil {
		return false, err
	}
	payload := map[string]string{"key": key}
	bodyBytes, _ := json.Marshal(payload)
	headers, err := c.l2Headers(ctx, http.MethodDelete, DeleteReadonlyAPIKeyEndpoint, string(bodyBytes))
	if err != nil {
		return false, err
	}
	res, err := c.del(c.endpoint(DeleteReadonlyAPIKeyEndpoint), c.requestOptions(headers, nil, payload))
	if err != nil {
		return false, err
	}
	b, ok := res.(bool)
	if ok {
		return b, nil
	}
	return false, nil
}

func (c *Client) ValidateReadonlyAPIKey(ctx context.Context, address string, key string) (string, error) {
	_ = ctx
	res, err := c.get(c.endpoint(ValidateReadonlyAPIKeyEndpoint), c.requestOptions(nil, QueryParams{"address": address, "key": key}, nil))
	if err != nil {
		return "", err
	}
	if s, ok := res.(string); ok {
		return s, nil
	}
	return fmt.Sprintf("%v", res), nil
}

func (c *Client) GetOrder(ctx context.Context, orderID string) (OpenOrder, error) {
	if err := c.ensureL2Auth(); err != nil {
		return OpenOrder{}, err
	}
	path := GetOrderEndpoint + orderID
	headers, err := c.l2Headers(ctx, http.MethodGet, path, "")
	if err != nil {
		return OpenOrder{}, err
	}
	res, err := c.get(c.endpoint(path), c.requestOptions(headers, nil, nil))
	if err != nil {
		return OpenOrder{}, err
	}
	return decodeAny[OpenOrder](res)
}

func (c *Client) GetTrades(ctx context.Context, params *TradeParams, onlyFirstPage bool, nextCursor string) ([]Trade, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.l2Headers(ctx, http.MethodGet, GetTradesEndpoint, "")
	if err != nil {
		return nil, err
	}
	if nextCursor == "" {
		nextCursor = InitialCursor
	}
	results := make([]Trade, 0)
	for nextCursor != EndCursor && (nextCursor == InitialCursor || !onlyFirstPage) {
		query := QueryParams{"next_cursor": nextCursor}
		if params != nil {
			if params.ID != "" {
				query["id"] = params.ID
			}
			if params.MakerAddress != "" {
				query["maker_address"] = params.MakerAddress
			}
			if params.Market != "" {
				query["market"] = params.Market
			}
			if params.AssetID != "" {
				query["asset_id"] = params.AssetID
			}
			if params.Before != "" {
				query["before"] = params.Before
			}
			if params.After != "" {
				query["after"] = params.After
			}
		}
		res, rErr := c.get(c.endpoint(GetTradesEndpoint), c.requestOptions(headers, query, nil))
		if rErr != nil {
			return nil, rErr
		}
		m := asMap(res)
		nextCursor = fmt.Sprintf("%v", m["next_cursor"])
		pageTrades, dErr := decodeAny[[]Trade](m["data"])
		if dErr != nil {
			return nil, dErr
		}
		results = append(results, pageTrades...)
	}
	return results, nil
}

func (c *Client) GetTradesPaginated(ctx context.Context, params *TradeParams, nextCursor string) (TradesPaginatedResponse, error) {
	if err := c.ensureL2Auth(); err != nil {
		return TradesPaginatedResponse{}, err
	}
	headers, err := c.l2Headers(ctx, http.MethodGet, GetTradesEndpoint, "")
	if err != nil {
		return TradesPaginatedResponse{}, err
	}
	if nextCursor == "" {
		nextCursor = InitialCursor
	}
	query := QueryParams{"next_cursor": nextCursor}
	if params != nil {
		if params.ID != "" {
			query["id"] = params.ID
		}
		if params.MakerAddress != "" {
			query["maker_address"] = params.MakerAddress
		}
		if params.Market != "" {
			query["market"] = params.Market
		}
		if params.AssetID != "" {
			query["asset_id"] = params.AssetID
		}
		if params.Before != "" {
			query["before"] = params.Before
		}
		if params.After != "" {
			query["after"] = params.After
		}
	}
	res, err := c.get(c.endpoint(GetTradesEndpoint), c.requestOptions(headers, query, nil))
	if err != nil {
		return TradesPaginatedResponse{}, err
	}
	m := asMap(res)
	trades, err := decodeAny[[]Trade](m["data"])
	if err != nil {
		return TradesPaginatedResponse{}, err
	}
	limit, _ := strconv.Atoi(fmt.Sprintf("%v", m["limit"]))
	count, _ := strconv.Atoi(fmt.Sprintf("%v", m["count"]))
	return TradesPaginatedResponse{Trades: trades, NextCursor: fmt.Sprintf("%v", m["next_cursor"]), Limit: limit, Count: count}, nil
}

func (c *Client) GetOpenOrders(ctx context.Context, params *OpenOrderParams, onlyFirstPage bool, nextCursor string) ([]OpenOrder, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodGet, GetOpenOrdersEndpoint, "", true)
	if err != nil {
		return nil, err
	}
	if nextCursor == "" {
		nextCursor = InitialCursor
	}
	results := make([]OpenOrder, 0)
	for nextCursor != EndCursor && (nextCursor == InitialCursor || !onlyFirstPage) {
		query := QueryParams{"next_cursor": nextCursor}
		if params != nil {
			if params.ID != "" {
				query["id"] = params.ID
			}
			if params.Market != "" {
				query["market"] = params.Market
			}
			if params.AssetID != "" {
				query["asset_id"] = params.AssetID
			}
		}
		res, rErr := c.get(c.endpoint(GetOpenOrdersEndpoint), c.requestOptions(headers, query, nil))
		if rErr != nil {
			return nil, rErr
		}
		m := asMap(res)
		nextCursor = fmt.Sprintf("%v", m["next_cursor"])
		pageOrders, dErr := decodeAny[[]OpenOrder](m["data"])
		if dErr != nil {
			return nil, dErr
		}
		results = append(results, pageOrders...)
	}
	return results, nil
}

func (c *Client) PostOrder(ctx context.Context, order SignedOrder, orderType OrderType, deferExec bool, postOnly *bool) (any, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	payload, err := c.toL2OrderPayload(order, orderType, deferExec, postOnly)
	if err != nil {
		return nil, err
	}
	bodyBytes, _ := json.Marshal(payload)
	headers, err := c.buildRequestHeaders(ctx, http.MethodPost, PostOrderEndpoint, string(bodyBytes), true)
	if err != nil {
		return nil, err
	}
	return c.post(c.endpoint(PostOrderEndpoint), c.requestOptions(headers, nil, payload))
}

func (c *Client) PostOrders(ctx context.Context, args []PostOrdersArg, deferExec bool, defaultPostOnly bool) (any, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	payloads := make([]NewOrder, 0, len(args))
	for _, a := range args {
		postOnly := a.PostOnly
		if postOnly == nil {
			v := defaultPostOnly
			postOnly = &v
		}
		p, err := c.toL2OrderPayload(a.Order, a.OrderType, deferExec, postOnly)
		if err != nil {
			return nil, err
		}
		payloads = append(payloads, p)
	}
	bodyBytes, _ := json.Marshal(payloads)
	headers, err := c.buildRequestHeaders(ctx, http.MethodPost, PostOrdersEndpoint, string(bodyBytes), true)
	if err != nil {
		return nil, err
	}
	return c.post(c.endpoint(PostOrdersEndpoint), c.requestOptions(headers, nil, payloads))
}

func (c *Client) CancelOrder(ctx context.Context, payload OrderPayload) (any, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	bodyBytes, _ := json.Marshal(payload)
	headers, err := c.buildRequestHeaders(ctx, http.MethodDelete, CancelOrderEndpoint, string(bodyBytes), true)
	if err != nil {
		return nil, err
	}
	return c.del(c.endpoint(CancelOrderEndpoint), c.requestOptions(headers, nil, payload))
}

func (c *Client) CancelOrders(ctx context.Context, orderHashes []string) (any, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	bodyBytes, _ := json.Marshal(orderHashes)
	headers, err := c.buildRequestHeaders(ctx, http.MethodDelete, CancelOrdersEndpoint, string(bodyBytes), true)
	if err != nil {
		return nil, err
	}
	return c.del(c.endpoint(CancelOrdersEndpoint), c.requestOptions(headers, nil, orderHashes))
}

func (c *Client) CancelAll(ctx context.Context) (any, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodDelete, CancelAllEndpoint, "", true)
	if err != nil {
		return nil, err
	}
	return c.del(c.endpoint(CancelAllEndpoint), c.requestOptions(headers, nil, nil))
}

func (c *Client) CancelMarketOrders(ctx context.Context, payload OrderMarketCancelParams) (any, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	bodyBytes, _ := json.Marshal(payload)
	headers, err := c.buildRequestHeaders(ctx, http.MethodDelete, CancelMarketOrdersEndpoint, string(bodyBytes), true)
	if err != nil {
		return nil, err
	}
	return c.del(c.endpoint(CancelMarketOrdersEndpoint), c.requestOptions(headers, nil, payload))
}

func (c *Client) IsOrderScoring(ctx context.Context, params *OrderScoringParams) (OrderScoring, error) {
	if err := c.ensureL2Auth(); err != nil {
		return OrderScoring{}, err
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodGet, IsOrderScoringEndpoint, "", true)
	if err != nil {
		return OrderScoring{}, err
	}
	q := QueryParams{}
	if params != nil && params.OrderID != "" {
		q["order_id"] = params.OrderID
	}
	res, err := c.get(c.endpoint(IsOrderScoringEndpoint), c.requestOptions(headers, q, nil))
	if err != nil {
		return OrderScoring{}, err
	}
	return decodeAny[OrderScoring](res)
}

func (c *Client) AreOrdersScoring(ctx context.Context, params *OrdersScoringParams) (OrdersScoring, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	var payload any
	body := ""
	if params != nil {
		payload = params.OrderIDs
		bodyBytes, _ := json.Marshal(params.OrderIDs)
		body = string(bodyBytes)
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodPost, AreOrdersScoringEndpoint, body, true)
	if err != nil {
		return nil, err
	}
	res, err := c.post(c.endpoint(AreOrdersScoringEndpoint), c.requestOptions(headers, nil, payload))
	if err != nil {
		return nil, err
	}
	return decodeAny[OrdersScoring](res)
}

func (c *Client) GetNotifications(ctx context.Context) ([]Notification, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodGet, GetNotificationsEndpoint, "", true)
	if err != nil {
		return nil, err
	}
	res, err := c.get(c.endpoint(GetNotificationsEndpoint), c.requestOptions(headers, QueryParams{"signature_type": int(c.signatureType)}, nil))
	if err != nil {
		return nil, err
	}
	return decodeAny[[]Notification](res)
}

func (c *Client) DropNotifications(ctx context.Context, params *DropNotificationParams) error {
	if err := c.ensureL2Auth(); err != nil {
		return err
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodDelete, DropNotificationsEndpoint, "", true)
	if err != nil {
		return err
	}
	q := QueryParams{}
	if params != nil && len(params.IDs) > 0 {
		q["ids"] = params.IDs
	}
	_, err = c.del(c.endpoint(DropNotificationsEndpoint), c.requestOptions(headers, q, nil))
	return err
}

func (c *Client) GetBalanceAllowance(ctx context.Context, params *BalanceAllowanceParams) (BalanceAllowanceResponse, error) {
	if err := c.ensureL2Auth(); err != nil {
		return BalanceAllowanceResponse{}, err
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodGet, GetBalanceAllowanceEndpoint, "", true)
	if err != nil {
		return BalanceAllowanceResponse{}, err
	}
	q := QueryParams{"signature_type": int(c.signatureType)}
	if params != nil {
		if params.AssetType != "" {
			q["asset_type"] = string(params.AssetType)
		}
		if params.TokenID != "" {
			q["token_id"] = params.TokenID
		}
	}
	res, err := c.get(c.endpoint(GetBalanceAllowanceEndpoint), c.requestOptions(headers, q, nil))
	if err != nil {
		return BalanceAllowanceResponse{}, err
	}
	return decodeAny[BalanceAllowanceResponse](res)
}

func (c *Client) UpdateBalanceAllowance(ctx context.Context, params *BalanceAllowanceParams) error {
	if err := c.ensureL2Auth(); err != nil {
		return err
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodGet, UpdateBalanceAllowanceEndpoint, "", true)
	if err != nil {
		return err
	}
	q := QueryParams{"signature_type": int(c.signatureType)}
	if params != nil {
		if params.AssetType != "" {
			q["asset_type"] = string(params.AssetType)
		}
		if params.TokenID != "" {
			q["token_id"] = params.TokenID
		}
	}
	_, err = c.get(c.endpoint(UpdateBalanceAllowanceEndpoint), c.requestOptions(headers, q, nil))
	return err
}

func (c *Client) PostHeartbeat(ctx context.Context, heartbeatID *string) (HeartbeatResponse, error) {
	if err := c.ensureL2Auth(); err != nil {
		return HeartbeatResponse{}, err
	}
	body := map[string]any{"heartbeat_id": nil}
	if heartbeatID != nil {
		body["heartbeat_id"] = *heartbeatID
	}
	bodyBytes, _ := json.Marshal(body)
	headers, err := c.buildRequestHeaders(ctx, http.MethodPost, PostHeartbeatEndpoint, string(bodyBytes), true)
	if err != nil {
		return HeartbeatResponse{}, err
	}
	res, err := c.post(c.endpoint(PostHeartbeatEndpoint), c.requestOptions(headers, nil, string(bodyBytes)))
	if err != nil {
		return HeartbeatResponse{}, err
	}
	return decodeAny[HeartbeatResponse](res)
}

func (c *Client) GetEarningsForUserForDay(ctx context.Context, date string) ([]UserEarning, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodGet, GetEarningsForUserForDayEndpoint, "", true)
	if err != nil {
		return nil, err
	}
	results := make([]UserEarning, 0)
	nextCursor := InitialCursor
	for nextCursor != EndCursor {
		res, err := c.get(c.endpoint(GetEarningsForUserForDayEndpoint), c.requestOptions(headers, QueryParams{"date": date, "signature_type": int(c.signatureType), "next_cursor": nextCursor}, nil))
		if err != nil {
			return nil, err
		}
		m := asMap(res)
		nextCursor = fmt.Sprintf("%v", m["next_cursor"])
		page, err := decodeAny[[]UserEarning](m["data"])
		if err != nil {
			return nil, err
		}
		results = append(results, page...)
	}
	return results, nil
}

func (c *Client) GetTotalEarningsForUserForDay(ctx context.Context, date string) ([]TotalUserEarning, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodGet, GetTotalEarningsForUserForDayEndpoint, "", true)
	if err != nil {
		return nil, err
	}
	res, err := c.get(c.endpoint(GetTotalEarningsForUserForDayEndpoint), c.requestOptions(headers, QueryParams{"date": date, "signature_type": int(c.signatureType)}, nil))
	if err != nil {
		return nil, err
	}
	return decodeAny[[]TotalUserEarning](res)
}

func (c *Client) GetUserEarningsAndMarketsConfig(ctx context.Context, date string, orderBy string, position string, noCompetition bool) ([]UserRewardsEarning, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodGet, GetRewardsEarningsPercentagesEndpoint, "", true)
	if err != nil {
		return nil, err
	}
	results := make([]UserRewardsEarning, 0)
	nextCursor := InitialCursor
	for nextCursor != EndCursor {
		res, err := c.get(c.endpoint(GetRewardsEarningsPercentagesEndpoint), c.requestOptions(headers, QueryParams{"date": date, "signature_type": int(c.signatureType), "next_cursor": nextCursor, "order_by": orderBy, "position": position, "no_competition": noCompetition}, nil))
		if err != nil {
			return nil, err
		}
		m := asMap(res)
		nextCursor = fmt.Sprintf("%v", m["next_cursor"])
		page, err := decodeAny[[]UserRewardsEarning](m["data"])
		if err != nil {
			return nil, err
		}
		results = append(results, page...)
	}
	return results, nil
}

func (c *Client) GetRewardPercentages(ctx context.Context) (RewardsPercentages, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodGet, GetLiquidityRewardPercentagesEndpoint, "", true)
	if err != nil {
		return nil, err
	}
	res, err := c.get(c.endpoint(GetLiquidityRewardPercentagesEndpoint), c.requestOptions(headers, QueryParams{"signature_type": int(c.signatureType)}, nil))
	if err != nil {
		return nil, err
	}
	return decodeAny[RewardsPercentages](res)
}

func (c *Client) GetCurrentRewards(ctx context.Context) ([]MarketReward, error) {
	results := make([]MarketReward, 0)
	nextCursor := InitialCursor
	for nextCursor != EndCursor {
		res, err := c.get(c.endpoint(GetRewardsMarketsCurrentEndpoint), c.requestOptions(nil, QueryParams{"next_cursor": nextCursor}, nil))
		if err != nil {
			return nil, err
		}
		m := asMap(res)
		nextCursor = fmt.Sprintf("%v", m["next_cursor"])
		page, err := decodeAny[[]MarketReward](m["data"])
		if err != nil {
			return nil, err
		}
		results = append(results, page...)
	}
	return results, nil
}

func (c *Client) GetRawRewardsForMarket(ctx context.Context, conditionID string) ([]MarketReward, error) {
	results := make([]MarketReward, 0)
	nextCursor := InitialCursor
	path := GetRewardsMarketsEndpoint + conditionID
	for nextCursor != EndCursor {
		res, err := c.get(c.endpoint(path), c.requestOptions(nil, QueryParams{"next_cursor": nextCursor}, nil))
		if err != nil {
			return nil, err
		}
		m := asMap(res)
		nextCursor = fmt.Sprintf("%v", m["next_cursor"])
		page, err := decodeAny[[]MarketReward](m["data"])
		if err != nil {
			return nil, err
		}
		results = append(results, page...)
	}
	return results, nil
}

func (c *Client) GetMarketTradesEvents(ctx context.Context, conditionID string) ([]MarketTradeEvent, error) {
	_ = ctx
	res, err := c.get(c.endpoint(GetMarketTradesEventsEndpoint+conditionID), c.requestOptions(nil, nil, nil))
	if err != nil {
		return nil, err
	}
	return decodeAny[[]MarketTradeEvent](res)
}

func (c *Client) CreateBuilderAPIKey(ctx context.Context) (BuilderAPIKey, error) {
	if err := c.ensureL2Auth(); err != nil {
		return BuilderAPIKey{}, err
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodPost, CreateBuilderAPIKeyEndpoint, "", true)
	if err != nil {
		return BuilderAPIKey{}, err
	}
	res, err := c.post(c.endpoint(CreateBuilderAPIKeyEndpoint), c.requestOptions(headers, nil, nil))
	if err != nil {
		return BuilderAPIKey{}, err
	}
	return decodeAny[BuilderAPIKey](res)
}

func (c *Client) GetBuilderAPIKeys(ctx context.Context) ([]BuilderAPIKeyResponse, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodGet, GetBuilderAPIKeysEndpoint, "", true)
	if err != nil {
		return nil, err
	}
	res, err := c.get(c.endpoint(GetBuilderAPIKeysEndpoint), c.requestOptions(headers, nil, nil))
	if err != nil {
		return nil, err
	}
	return decodeAny[[]BuilderAPIKeyResponse](res)
}

func (c *Client) RevokeBuilderAPIKey(ctx context.Context) (any, error) {
	_ = ctx
	headers, err := c.buildBuilderOnlyHeaders(http.MethodDelete, RevokeBuilderAPIKeyEndpoint, "")
	if err != nil {
		return nil, err
	}
	return c.del(c.endpoint(RevokeBuilderAPIKeyEndpoint), c.requestOptions(headers, nil, nil))
}

func (c *Client) GetBuilderTrades(ctx context.Context, params *TradeParams, nextCursor string) ([]BuilderTrade, string, int, int, error) {
	_ = ctx
	headers, err := c.buildBuilderOnlyHeaders(http.MethodGet, GetBuilderTradesEndpoint, "")
	if err != nil {
		return nil, "", 0, 0, err
	}
	if nextCursor == "" {
		nextCursor = InitialCursor
	}
	query := QueryParams{"next_cursor": nextCursor}
	if params != nil {
		if params.ID != "" {
			query["id"] = params.ID
		}
		if params.MakerAddress != "" {
			query["maker_address"] = params.MakerAddress
		}
		if params.Market != "" {
			query["market"] = params.Market
		}
		if params.AssetID != "" {
			query["asset_id"] = params.AssetID
		}
	}
	res, err := c.get(c.endpoint(GetBuilderTradesEndpoint), c.requestOptions(headers, query, nil))
	if err != nil {
		return nil, "", 0, 0, err
	}
	m := asMap(res)
	trades, err := decodeAny[[]BuilderTrade](m["data"])
	if err != nil {
		return nil, "", 0, 0, err
	}
	limit, _ := strconv.Atoi(fmt.Sprintf("%v", m["limit"]))
	count, _ := strconv.Atoi(fmt.Sprintf("%v", m["count"]))
	return trades, fmt.Sprintf("%v", m["next_cursor"]), limit, count, nil
}

func (c *Client) buildRfqQueryParams(params map[string]any) QueryParams {
	q := QueryParams{}
	for k, v := range params {
		if v == nil {
			continue
		}
		switch vv := v.(type) {
		case string:
			if vv != "" {
				q[k] = vv
			}
		case int:
			if vv != 0 {
				q[k] = vv
			}
		case float64:
			if vv != 0 {
				q[k] = vv
			}
		case []string:
			if len(vv) > 0 {
				q[k] = strings.Join(vv, ",")
			}
		default:
			if fmt.Sprintf("%v", vv) != "" {
				q[k] = vv
			}
		}
	}
	return q
}

func (c *Client) ensureOrderFactory() error {
	if c.signer == nil && c.createOrderFn == nil {
		return ErrOrderSignerMissing
	}
	return nil
}

func (c *Client) CreateRfqRequest(ctx context.Context, userOrder RfqUserOrder, options *CreateOrderOptions) (RfqRequestResponse, error) {
	if err := c.ensureL2Auth(); err != nil {
		return RfqRequestResponse{}, err
	}
	if options == nil || options.TickSize == "" {
		tick, err := c.GetTickSize(ctx, userOrder.TokenID)
		if err != nil {
			return RfqRequestResponse{}, err
		}
		options = &CreateOrderOptions{TickSize: tick}
	}
	size := userOrder.Size
	price := userOrder.Price
	var amountIn string
	var amountOut string
	var assetIn string
	var assetOut string
	if userOrder.Side == SideBUY {
		amountIn = fmt.Sprintf("%.0f", size)
		amountOut = fmt.Sprintf("%.0f", size*price)
		assetIn = userOrder.TokenID
		assetOut = "0"
	} else {
		amountIn = fmt.Sprintf("%.0f", size*price)
		amountOut = fmt.Sprintf("%.0f", size)
		assetIn = "0"
		assetOut = userOrder.TokenID
	}
	payload := CreateRfqRequestParams{AssetIn: assetIn, AssetOut: assetOut, AmountIn: amountIn, AmountOut: amountOut, UserType: int(c.signatureType)}
	bodyBytes, _ := json.Marshal(payload)
	headers, err := c.buildRequestHeaders(ctx, http.MethodPost, CreateRFQRequestEndpoint, string(bodyBytes), true)
	if err != nil {
		return RfqRequestResponse{}, err
	}
	res, err := c.post(c.endpoint(CreateRFQRequestEndpoint), c.requestOptions(headers, nil, payload))
	if err != nil {
		return RfqRequestResponse{}, err
	}
	return decodeAny[RfqRequestResponse](res)
}

func (c *Client) CancelRfqRequest(ctx context.Context, req CancelRfqRequestParams) (string, error) {
	if err := c.ensureL2Auth(); err != nil {
		return "", err
	}
	bodyBytes, _ := json.Marshal(req)
	headers, err := c.buildRequestHeaders(ctx, http.MethodDelete, CancelRFQRequestEndpoint, string(bodyBytes), true)
	if err != nil {
		return "", err
	}
	res, err := c.del(c.endpoint(CancelRFQRequestEndpoint), c.requestOptions(headers, nil, req))
	if err != nil {
		return "", err
	}
	if s, ok := res.(string); ok {
		return s, nil
	}
	return "OK", nil
}

func (c *Client) GetRfqRequests(ctx context.Context, params *GetRfqRequestsParams) (RfqRequestsResponse, error) {
	if err := c.ensureL2Auth(); err != nil {
		return RfqRequestsResponse{}, err
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodGet, GetRFQRequestsEndpoint, "", true)
	if err != nil {
		return RfqRequestsResponse{}, err
	}
	m := map[string]any{}
	if params != nil {
		m["offset"] = params.Offset
		m["limit"] = params.Limit
		m["state"] = params.State
		m["requestIds"] = params.RequestIDs
		m["markets"] = params.Markets
		m["sizeMin"] = params.SizeMin
		m["sizeMax"] = params.SizeMax
		m["sizeUsdcMin"] = params.SizeUSDCMin
		m["sizeUsdcMax"] = params.SizeUSDCMax
		m["priceMin"] = params.PriceMin
		m["priceMax"] = params.PriceMax
		m["sortBy"] = params.SortBy
		m["sortDir"] = params.SortDir
	}
	q := c.buildRfqQueryParams(m)
	res, err := c.get(c.endpoint(GetRFQRequestsEndpoint), c.requestOptions(headers, q, nil))
	if err != nil {
		return RfqRequestsResponse{}, err
	}
	return decodeAny[RfqRequestsResponse](res)
}

func (c *Client) CreateRfqQuote(ctx context.Context, quote RfqUserQuote, options *CreateOrderOptions) (RfqQuoteResponse, error) {
	if err := c.ensureL2Auth(); err != nil {
		return RfqQuoteResponse{}, err
	}
	size := quote.Size
	price := quote.Price
	var amountIn string
	var amountOut string
	var assetIn string
	var assetOut string
	if quote.Side == SideSELL {
		amountIn = fmt.Sprintf("%.0f", size*price)
		amountOut = fmt.Sprintf("%.0f", size)
		assetIn = "0"
		assetOut = quote.TokenID
	} else {
		amountIn = fmt.Sprintf("%.0f", size)
		amountOut = fmt.Sprintf("%.0f", size*price)
		assetIn = quote.TokenID
		assetOut = "0"
	}
	payload := CreateRfqQuoteParams{RequestID: quote.RequestID, AssetIn: assetIn, AssetOut: assetOut, AmountIn: amountIn, AmountOut: amountOut, UserType: int(c.signatureType)}
	bodyBytes, _ := json.Marshal(payload)
	headers, err := c.buildRequestHeaders(ctx, http.MethodPost, CreateRFQQuoteEndpoint, string(bodyBytes), true)
	if err != nil {
		return RfqQuoteResponse{}, err
	}
	res, err := c.post(c.endpoint(CreateRFQQuoteEndpoint), c.requestOptions(headers, nil, payload))
	if err != nil {
		return RfqQuoteResponse{}, err
	}
	return decodeAny[RfqQuoteResponse](res)
}

func (c *Client) getRfqQuotes(ctx context.Context, endpoint string, params *GetRfqQuotesParams) (RfqQuotesResponse, error) {
	if err := c.ensureL2Auth(); err != nil {
		return RfqQuotesResponse{}, err
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodGet, endpoint, "", true)
	if err != nil {
		return RfqQuotesResponse{}, err
	}
	m := map[string]any{}
	if params != nil {
		m["offset"] = params.Offset
		m["limit"] = params.Limit
		m["state"] = params.State
		m["quoteIds"] = params.QuoteIDs
		m["requestIds"] = params.RequestIDs
		m["markets"] = params.Markets
		m["sizeMin"] = params.SizeMin
		m["sizeMax"] = params.SizeMax
		m["sizeUsdcMin"] = params.SizeUSDCMin
		m["sizeUsdcMax"] = params.SizeUSDCMax
		m["priceMin"] = params.PriceMin
		m["priceMax"] = params.PriceMax
		m["sortBy"] = params.SortBy
		m["sortDir"] = params.SortDir
	}
	q := c.buildRfqQueryParams(m)
	res, err := c.get(c.endpoint(endpoint), c.requestOptions(headers, q, nil))
	if err != nil {
		return RfqQuotesResponse{}, err
	}
	return decodeAny[RfqQuotesResponse](res)
}

func (c *Client) GetRfqRequesterQuotes(ctx context.Context, params *GetRfqQuotesParams) (RfqQuotesResponse, error) {
	return c.getRfqQuotes(ctx, GetRFQRequesterQuotesEndpoint, params)
}

func (c *Client) GetRfqQuoterQuotes(ctx context.Context, params *GetRfqQuotesParams) (RfqQuotesResponse, error) {
	return c.getRfqQuotes(ctx, GetRFQQuoterQuotesEndpoint, params)
}

func (c *Client) GetRfqBestQuote(ctx context.Context, params *GetRfqBestQuoteParams) (RfqQuote, error) {
	if err := c.ensureL2Auth(); err != nil {
		return RfqQuote{}, err
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodGet, GetRFQBestQuoteEndpoint, "", true)
	if err != nil {
		return RfqQuote{}, err
	}
	q := QueryParams{}
	if params != nil && params.RequestID != "" {
		q["requestId"] = params.RequestID
	}
	res, err := c.get(c.endpoint(GetRFQBestQuoteEndpoint), c.requestOptions(headers, q, nil))
	if err != nil {
		return RfqQuote{}, err
	}
	return decodeAny[RfqQuote](res)
}

func (c *Client) CancelRfqQuote(ctx context.Context, quote CancelRfqQuoteParams) (string, error) {
	if err := c.ensureL2Auth(); err != nil {
		return "", err
	}
	bodyBytes, _ := json.Marshal(quote)
	headers, err := c.buildRequestHeaders(ctx, http.MethodDelete, CancelRFQQuoteEndpoint, string(bodyBytes), true)
	if err != nil {
		return "", err
	}
	res, err := c.del(c.endpoint(CancelRFQQuoteEndpoint), c.requestOptions(headers, nil, quote))
	if err != nil {
		return "", err
	}
	if s, ok := res.(string); ok {
		return s, nil
	}
	return "OK", nil
}

func (c *Client) RFQConfig(ctx context.Context) (any, error) {
	if err := c.ensureL2Auth(); err != nil {
		return nil, err
	}
	headers, err := c.buildRequestHeaders(ctx, http.MethodGet, RFQConfigEndpoint, "", true)
	if err != nil {
		return nil, err
	}
	return c.get(c.endpoint(RFQConfigEndpoint), c.requestOptions(headers, nil, nil))
}

func (c *Client) getRfqRequestOrderCreationPayload(quote RfqQuote) (RfqRequestOrderCreationPayload, error) {
	quoteSide := quote.Side
	matchType := quote.MatchType
	switch matchType {
	case string(RfqMatchTypeComplementary):
		side := SideBUY
		if quoteSide == "BUY" {
			side = SideSELL
		}
		size := quote.SizeIn
		if side == SideBUY {
			size = quote.SizeOut
		}
		return RfqRequestOrderCreationPayload{Token: quote.Token, Side: side, Size: size, Price: quote.Price}, nil
	case string(RfqMatchTypeMint), string(RfqMatchTypeMerge):
		side := SideSELL
		if quoteSide == "BUY" {
			side = SideBUY
		}
		size := quote.SizeOut
		if side == SideBUY {
			size = quote.SizeIn
		}
		return RfqRequestOrderCreationPayload{Token: quote.Complement, Side: side, Size: size, Price: 1 - quote.Price}, nil
	default:
		return RfqRequestOrderCreationPayload{}, fmt.Errorf("invalid match type")
	}
}

func (c *Client) AcceptRfqQuote(ctx context.Context, payload AcceptQuoteParams) (string, error) {
	if err := c.ensureOrderFactory(); err != nil {
		return "", err
	}
	rfqQuotes, err := c.GetRfqRequesterQuotes(ctx, &GetRfqQuotesParams{QuoteIDs: []string{payload.QuoteID}})
	if err != nil {
		return "", err
	}
	if len(rfqQuotes.Data) == 0 {
		return "", fmt.Errorf("RFQ quote not found")
	}
	quote := rfqQuotes.Data[0]
	orderPayload, err := c.getRfqRequestOrderCreationPayload(quote)
	if err != nil {
		return "", err
	}
	size, err := strconv.ParseFloat(orderPayload.Size, 64)
	if err != nil {
		return "", err
	}
	exp := payload.Expiration
	order, err := c.resolveCreateOrderFactory()(ctx, UserOrder{TokenID: orderPayload.Token, Price: orderPayload.Price, Size: size, Side: orderPayload.Side, Expiration: &exp}, nil)
	if err != nil {
		return "", err
	}
	owner := ""
	if c.creds != nil {
		owner = c.creds.Key
	}
	salt, _ := strconv.ParseInt(order.Salt, 10, 64)
	expiration, _ := strconv.ParseInt(order.Expiration, 10, 64)
	acceptPayload := map[string]any{"requestId": payload.RequestID, "quoteId": payload.QuoteID, "owner": owner, "salt": salt, "maker": order.Maker, "signer": order.Signer, "taker": order.Taker, "tokenId": order.TokenID, "makerAmount": order.MakerAmount, "takerAmount": order.TakerAmount, "expiration": expiration, "nonce": order.Nonce, "feeRateBps": order.FeeRateBPS, "side": orderPayload.Side, "signatureType": order.SignatureType, "signature": order.Signature}
	bodyBytes, _ := json.Marshal(acceptPayload)
	headers, err := c.buildRequestHeaders(ctx, http.MethodPost, RFQRequestsAcceptEndpoint, string(bodyBytes), true)
	if err != nil {
		return "", err
	}
	res, err := c.post(c.endpoint(RFQRequestsAcceptEndpoint), c.requestOptions(headers, nil, acceptPayload))
	if err != nil {
		return "", err
	}
	if s, ok := res.(string); ok {
		return s, nil
	}
	return "OK", nil
}

func (c *Client) ApproveRfqOrder(ctx context.Context, payload ApproveOrderParams) (string, error) {
	if err := c.ensureOrderFactory(); err != nil {
		return "", err
	}
	rfqQuotes, err := c.GetRfqQuoterQuotes(ctx, &GetRfqQuotesParams{QuoteIDs: []string{payload.QuoteID}})
	if err != nil {
		return "", err
	}
	if len(rfqQuotes.Data) == 0 {
		return "", fmt.Errorf("RFQ quote not found")
	}
	quote := rfqQuotes.Data[0]
	side := SideSELL
	if quote.Side == "BUY" {
		side = SideBUY
	}
	sizeRaw := quote.SizeOut
	if quote.Side == "BUY" {
		sizeRaw = quote.SizeIn
	}
	size, err := strconv.ParseFloat(sizeRaw, 64)
	if err != nil {
		return "", err
	}
	exp := payload.Expiration
	order, err := c.resolveCreateOrderFactory()(ctx, UserOrder{TokenID: quote.Token, Price: quote.Price, Size: size, Side: side, Expiration: &exp}, nil)
	if err != nil {
		return "", err
	}
	owner := ""
	if c.creds != nil {
		owner = c.creds.Key
	}
	salt, _ := strconv.ParseInt(order.Salt, 10, 64)
	expiration, _ := strconv.ParseInt(order.Expiration, 10, 64)
	approvePayload := map[string]any{"requestId": payload.RequestID, "quoteId": payload.QuoteID, "owner": owner, "salt": salt, "maker": order.Maker, "signer": order.Signer, "taker": order.Taker, "tokenId": order.TokenID, "makerAmount": order.MakerAmount, "takerAmount": order.TakerAmount, "expiration": expiration, "nonce": order.Nonce, "feeRateBps": order.FeeRateBPS, "side": side, "signatureType": order.SignatureType, "signature": order.Signature}
	bodyBytes, _ := json.Marshal(approvePayload)
	headers, err := c.buildRequestHeaders(ctx, http.MethodPost, RFQQuoteApproveEndpoint, string(bodyBytes), true)
	if err != nil {
		return "", err
	}
	res, err := c.post(c.endpoint(RFQQuoteApproveEndpoint), c.requestOptions(headers, nil, approvePayload))
	if err != nil {
		return "", err
	}
	if s, ok := res.(string); ok {
		return s, nil
	}
	return "OK", nil
}
