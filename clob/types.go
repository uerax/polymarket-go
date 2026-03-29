package clob

type ApiKeyCreds struct {
	Key        string `json:"key"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

type ApiKeyRaw struct {
	APIKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

type ReadonlyAPIKeyResponse struct {
	APIKey string `json:"apiKey"`
}

type L2HeaderArgs struct {
	Method      string
	RequestPath string
	Body        string
}

type Side string

const (
	SideBUY  Side = "BUY"
	SideSELL Side = "SELL"
)

type OrderType string

const (
	OrderTypeGTC OrderType = "GTC"
	OrderTypeFOK OrderType = "FOK"
	OrderTypeGTD OrderType = "GTD"
	OrderTypeFAK OrderType = "FAK"
)

type Chain int

const (
	ChainPolygon Chain = 137
	ChainAmoy    Chain = 80002
)

type TickSize string

type UserOrder struct {
	TokenID    string `json:"tokenID"`
	Price      float64 `json:"price"`
	Size       float64 `json:"size"`
	Side       Side    `json:"side"`
	FeeRateBPS *int    `json:"feeRateBps,omitempty"`
	Nonce      *int64  `json:"nonce,omitempty"`
	Expiration *int64  `json:"expiration,omitempty"`
	Taker      string  `json:"taker,omitempty"`
}

type UserMarketOrder struct {
	TokenID    string   `json:"tokenID"`
	Price      *float64 `json:"price,omitempty"`
	Amount     float64  `json:"amount"`
	Side       Side     `json:"side"`
	FeeRateBPS *int     `json:"feeRateBps,omitempty"`
	Nonce      *int64   `json:"nonce,omitempty"`
	Taker      string   `json:"taker,omitempty"`
	OrderType  OrderType `json:"orderType,omitempty"`
}

type CreateOrderOptions struct {
	TickSize TickSize `json:"tickSize"`
	NegRisk  *bool    `json:"negRisk,omitempty"`
}

type RfqUserOrder struct {
	TokenID string  `json:"tokenID"`
	Price   float64 `json:"price"`
	Size    float64 `json:"size"`
	Side    Side    `json:"side"`
}

type RfqUserQuote struct {
	RequestID string  `json:"requestId"`
	TokenID   string  `json:"tokenID"`
	Price     float64 `json:"price"`
	Size      float64 `json:"size"`
	Side      Side    `json:"side"`
}

type RfqRequestOrderCreationPayload struct {
	Token string  `json:"token"`
	Side  Side    `json:"side"`
	Size  string  `json:"size"`
	Price float64 `json:"price"`
}

type PriceHistoryInterval string

const (
	PriceHistoryIntervalMax     PriceHistoryInterval = "max"
	PriceHistoryIntervalOneWeek PriceHistoryInterval = "1w"
	PriceHistoryIntervalOneDay  PriceHistoryInterval = "1d"
	PriceHistoryIntervalSixHour PriceHistoryInterval = "6h"
	PriceHistoryIntervalOneHour PriceHistoryInterval = "1h"
)

type AssetType string

const (
	AssetTypeCollateral  AssetType = "COLLATERAL"
	AssetTypeConditional AssetType = "CONDITIONAL"
)

type RfqListState string

type RfqSortDir string

type RfqRequestsSortBy string

type RfqQuotesSortBy string

const (
	RfqListStateActive   RfqListState = "active"
	RfqListStateInactive RfqListState = "inactive"

	RfqSortDirAsc  RfqSortDir = "asc"
	RfqSortDirDesc RfqSortDir = "desc"

	RfqRequestsSortByPrice   RfqRequestsSortBy = "price"
	RfqRequestsSortByExpiry  RfqRequestsSortBy = "expiry"
	RfqRequestsSortBySize    RfqRequestsSortBy = "size"
	RfqRequestsSortByCreated RfqRequestsSortBy = "created"

	RfqQuotesSortByPrice   RfqQuotesSortBy = "price"
	RfqQuotesSortByExpiry  RfqQuotesSortBy = "expiry"
	RfqQuotesSortByCreated RfqQuotesSortBy = "created"
)

type RfqMatchType string

const (
	RfqMatchTypeComplementary RfqMatchType = "COMPLEMENTARY"
	RfqMatchTypeMerge         RfqMatchType = "MERGE"
	RfqMatchTypeMint          RfqMatchType = "MINT"
)

type QueryParams map[string]any

type RequestOptions struct {
	Headers map[string]string
	Data    any
	Params  QueryParams
}

type BanStatus struct {
	ClosedOnly bool `json:"closed_only"`
}

type ApiKeysResponse struct {
	APIKeys []ApiKeyCreds `json:"apiKeys"`
}

type OrderPayload struct {
	OrderID string `json:"orderID"`
}

type OrderMarketCancelParams struct {
	Market  string `json:"market,omitempty"`
	AssetID string `json:"asset_id,omitempty"`
}

type PostOrdersArg struct {
	Order     SignedOrder `json:"order"`
	OrderType OrderType   `json:"orderType"`
	PostOnly  *bool       `json:"postOnly,omitempty"`
}

type NewOrder struct {
	Order struct {
		Salt          int64  `json:"salt"`
		Maker         string `json:"maker"`
		Signer        string `json:"signer"`
		Taker         string `json:"taker"`
		TokenID       string `json:"tokenId"`
		MakerAmount   string `json:"makerAmount"`
		TakerAmount   string `json:"takerAmount"`
		Expiration    string `json:"expiration"`
		Nonce         string `json:"nonce"`
		FeeRateBPS    string `json:"feeRateBps"`
		Side          Side   `json:"side"`
		SignatureType int    `json:"signatureType"`
		Signature     string `json:"signature"`
	} `json:"order"`
	Owner    string    `json:"owner"`
	OrderType OrderType `json:"orderType"`
	DeferExec bool      `json:"deferExec"`
	PostOnly  *bool     `json:"postOnly,omitempty"`
}

type SignedOrder struct {
	Salt          string `json:"salt"`
	Maker         string `json:"maker"`
	Signer        string `json:"signer"`
	Taker         string `json:"taker"`
	TokenID       string `json:"tokenId"`
	MakerAmount   string `json:"makerAmount"`
	TakerAmount   string `json:"takerAmount"`
	Expiration    string `json:"expiration"`
	Nonce         string `json:"nonce"`
	FeeRateBPS    string `json:"feeRateBps"`
	Side          Side   `json:"side"`
	SignatureType int    `json:"signatureType"`
	Signature     string `json:"signature"`
}

type OpenOrder struct {
	ID             string   `json:"id"`
	Status         string   `json:"status"`
	Owner          string   `json:"owner"`
	MakerAddress   string   `json:"maker_address"`
	Market         string   `json:"market"`
	AssetID        string   `json:"asset_id"`
	Side           string   `json:"side"`
	OriginalSize   string   `json:"original_size"`
	SizeMatched    string   `json:"size_matched"`
	Price          string   `json:"price"`
	AssociateTrades []string `json:"associate_trades"`
	Outcome        string   `json:"outcome"`
	CreatedAt      int64    `json:"created_at"`
	Expiration     string   `json:"expiration"`
	OrderType      string   `json:"order_type"`
}

type Trade struct {
	ID              string       `json:"id"`
	TakerOrderID    string       `json:"taker_order_id"`
	Market          string       `json:"market"`
	AssetID         string       `json:"asset_id"`
	Side            Side         `json:"side"`
	Size            string       `json:"size"`
	FeeRateBPS      string       `json:"fee_rate_bps"`
	Price           string       `json:"price"`
	Status          string       `json:"status"`
	MatchTime       string       `json:"match_time"`
	LastUpdate      string       `json:"last_update"`
	Outcome         string       `json:"outcome"`
	BucketIndex     int          `json:"bucket_index"`
	Owner           string       `json:"owner"`
	MakerAddress    string       `json:"maker_address"`
	MakerOrders     []MakerOrder `json:"maker_orders"`
	TransactionHash string       `json:"transaction_hash"`
	TraderSide      string       `json:"trader_side"`
}

type MakerOrder struct {
	OrderID       string `json:"order_id"`
	Owner         string `json:"owner"`
	MakerAddress  string `json:"maker_address"`
	MatchedAmount string `json:"matched_amount"`
	Price         string `json:"price"`
	FeeRateBPS    string `json:"fee_rate_bps"`
	AssetID       string `json:"asset_id"`
	Outcome       string `json:"outcome"`
	Side          Side   `json:"side"`
}

type TradeParams struct {
	ID           string `json:"id,omitempty"`
	MakerAddress string `json:"maker_address,omitempty"`
	Market       string `json:"market,omitempty"`
	AssetID      string `json:"asset_id,omitempty"`
	Before       string `json:"before,omitempty"`
	After        string `json:"after,omitempty"`
}

type OpenOrderParams struct {
	ID      string `json:"id,omitempty"`
	Market  string `json:"market,omitempty"`
	AssetID string `json:"asset_id,omitempty"`
}

type TradesPaginatedResponse struct {
	Trades      []Trade `json:"trades"`
	NextCursor  string  `json:"next_cursor"`
	Limit       int     `json:"limit"`
	Count       int     `json:"count"`
}

type PaginationPayload struct {
	Limit      int    `json:"limit"`
	Count      int    `json:"count"`
	NextCursor string `json:"next_cursor"`
	Data       []any  `json:"data"`
}

type OrderSummary struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

type OrderBookSummary struct {
	Market         string         `json:"market"`
	AssetID        string         `json:"asset_id"`
	Timestamp      string         `json:"timestamp"`
	Bids           []OrderSummary `json:"bids"`
	Asks           []OrderSummary `json:"asks"`
	MinOrderSize   string         `json:"min_order_size"`
	TickSize       string         `json:"tick_size"`
	NegRisk        bool           `json:"neg_risk"`
	LastTradePrice string         `json:"last_trade_price"`
	Hash           string         `json:"hash"`
}

type BookParams struct {
	TokenID string `json:"token_id"`
	Side    Side   `json:"side"`
}

type MarketPrice struct {
	T int64   `json:"t"`
	P float64 `json:"p"`
}

type PriceHistoryFilterParams struct {
	Market   string               `json:"market,omitempty"`
	StartTS  int64                `json:"startTs,omitempty"`
	EndTS    int64                `json:"endTs,omitempty"`
	Fidelity int                  `json:"fidelity,omitempty"`
	Interval PriceHistoryInterval `json:"interval,omitempty"`
}

type Notification struct {
	Type  int    `json:"type"`
	Owner string `json:"owner"`
	Payload any  `json:"payload"`
}

type DropNotificationParams struct {
	IDs []string `json:"ids"`
}

type BalanceAllowanceParams struct {
	AssetType AssetType `json:"asset_type"`
	TokenID   string    `json:"token_id,omitempty"`
}

type BalanceAllowanceResponse struct {
	Balance   string `json:"balance"`
	Allowance string `json:"allowance"`
}

type OrderScoringParams struct {
	OrderID string `json:"order_id"`
}

type OrderScoring struct {
	Scoring bool `json:"scoring"`
}

type OrdersScoringParams struct {
	OrderIDs []string `json:"orderIds"`
}

type OrdersScoring map[string]bool

type MarketTradeEvent struct {
	EventType string `json:"event_type"`
	Market    struct {
		ConditionID string `json:"condition_id"`
		AssetID     string `json:"asset_id"`
		Question    string `json:"question"`
		Icon        string `json:"icon"`
		Slug        string `json:"slug"`
	} `json:"market"`
	User struct {
		Address                string `json:"address"`
		Username               string `json:"username"`
		ProfilePicture         string `json:"profile_picture"`
		OptimizedProfilePicture string `json:"optimized_profile_picture"`
		Pseudonym              string `json:"pseudonym"`
	} `json:"user"`
	Side            Side   `json:"side"`
	Size            string `json:"size"`
	FeeRateBPS      string `json:"fee_rate_bps"`
	Price           string `json:"price"`
	Outcome         string `json:"outcome"`
	OutcomeIndex    int    `json:"outcome_index"`
	TransactionHash string `json:"transaction_hash"`
	Timestamp       string `json:"timestamp"`
}

type UserEarning struct {
	Date        string  `json:"date"`
	ConditionID string  `json:"condition_id"`
	AssetAddress string `json:"asset_address"`
	MakerAddress string `json:"maker_address"`
	Earnings    float64 `json:"earnings"`
	AssetRate   float64 `json:"asset_rate"`
}

type TotalUserEarning struct {
	Date         string  `json:"date"`
	AssetAddress string  `json:"asset_address"`
	MakerAddress string  `json:"maker_address"`
	Earnings     float64 `json:"earnings"`
	AssetRate    float64 `json:"asset_rate"`
}

type RewardsPercentages map[string]float64

type Token struct {
	TokenID string  `json:"token_id"`
	Outcome string  `json:"outcome"`
	Price   float64 `json:"price"`
}

type RewardsConfig struct {
	AssetAddress string  `json:"asset_address"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
	RatePerDay   float64 `json:"rate_per_day"`
	TotalRewards float64 `json:"total_rewards"`
}

type MarketReward struct {
	ConditionID      string          `json:"condition_id"`
	Question         string          `json:"question"`
	MarketSlug       string          `json:"market_slug"`
	EventSlug        string          `json:"event_slug"`
	Image            string          `json:"image"`
	RewardsMaxSpread float64         `json:"rewards_max_spread"`
	RewardsMinSize   float64         `json:"rewards_min_size"`
	Tokens           []Token         `json:"tokens"`
	RewardsConfig    []RewardsConfig `json:"rewards_config"`
}

type Earning struct {
	AssetAddress string  `json:"asset_address"`
	Earnings     float64 `json:"earnings"`
	AssetRate    float64 `json:"asset_rate"`
}

type UserRewardsEarning struct {
	ConditionID          string          `json:"condition_id"`
	Question             string          `json:"question"`
	MarketSlug           string          `json:"market_slug"`
	EventSlug            string          `json:"event_slug"`
	Image                string          `json:"image"`
	RewardsMaxSpread     float64         `json:"rewards_max_spread"`
	RewardsMinSize       float64         `json:"rewards_min_size"`
	MarketCompetitiveness float64        `json:"market_competitiveness"`
	Tokens               []Token         `json:"tokens"`
	RewardsConfig        []RewardsConfig `json:"rewards_config"`
	MakerAddress         string          `json:"maker_address"`
	EarningPercentage    float64         `json:"earning_percentage"`
	Earnings             []Earning       `json:"earnings"`
}

type BuilderAPIKey struct {
	Key        string `json:"key"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

type BuilderAPIKeyResponse struct {
	Key       string `json:"key"`
	CreatedAt string `json:"createdAt,omitempty"`
	RevokedAt string `json:"revokedAt,omitempty"`
}

type BuilderTrade struct {
	ID              string `json:"id"`
	TradeType       string `json:"tradeType"`
	TakerOrderHash  string `json:"takerOrderHash"`
	Builder         string `json:"builder"`
	Market          string `json:"market"`
	AssetID         string `json:"assetId"`
	Side            string `json:"side"`
	Size            string `json:"size"`
	SizeUSDC        string `json:"sizeUsdc"`
	Price           string `json:"price"`
	Status          string `json:"status"`
	Outcome         string `json:"outcome"`
	OutcomeIndex    int    `json:"outcomeIndex"`
	Owner           string `json:"owner"`
	Maker           string `json:"maker"`
	TransactionHash string `json:"transactionHash"`
	MatchTime       string `json:"matchTime"`
	BucketIndex     int    `json:"bucketIndex"`
	Fee             string `json:"fee"`
	FeeUSDC         string `json:"feeUsdc"`
	ErrMsg          string `json:"err_msg,omitempty"`
	CreatedAt       string `json:"createdAt,omitempty"`
	UpdatedAt       string `json:"updatedAt,omitempty"`
}

type HeartbeatResponse struct {
	HeartbeatID string `json:"heartbeat_id"`
	Error       string `json:"error,omitempty"`
}

type CancelRfqRequestParams struct {
	RequestID string `json:"requestId"`
}

type CreateRfqRequestParams struct {
	AssetIn  string `json:"assetIn"`
	AssetOut string `json:"assetOut"`
	AmountIn string `json:"amountIn"`
	AmountOut string `json:"amountOut"`
	UserType int    `json:"userType"`
}

type CreateRfqQuoteParams struct {
	RequestID string `json:"requestId"`
	AssetIn   string `json:"assetIn"`
	AssetOut  string `json:"assetOut"`
	AmountIn  string `json:"amountIn"`
	AmountOut string `json:"amountOut"`
	UserType  int    `json:"userType,omitempty"`
}

type CancelRfqQuoteParams struct {
	QuoteID string `json:"quoteId"`
}

type AcceptQuoteParams struct {
	RequestID  string `json:"requestId"`
	QuoteID    string `json:"quoteId"`
	Expiration int64  `json:"expiration"`
}

type ApproveOrderParams struct {
	RequestID  string `json:"requestId"`
	QuoteID    string `json:"quoteId"`
	Expiration int64  `json:"expiration"`
}

type GetRfqBestQuoteParams struct {
	RequestID string `json:"requestId,omitempty"`
}

type GetRfqQuotesParams struct {
	Offset     string        `json:"offset,omitempty"`
	Limit      int           `json:"limit,omitempty"`
	State      RfqListState  `json:"state,omitempty"`
	QuoteIDs   []string      `json:"quoteIds,omitempty"`
	RequestIDs []string      `json:"requestIds,omitempty"`
	Markets    []string      `json:"markets,omitempty"`
	SizeMin    float64       `json:"sizeMin,omitempty"`
	SizeMax    float64       `json:"sizeMax,omitempty"`
	SizeUSDCMin float64      `json:"sizeUsdcMin,omitempty"`
	SizeUSDCMax float64      `json:"sizeUsdcMax,omitempty"`
	PriceMin   float64       `json:"priceMin,omitempty"`
	PriceMax   float64       `json:"priceMax,omitempty"`
	SortBy     RfqQuotesSortBy `json:"sortBy,omitempty"`
	SortDir    RfqSortDir    `json:"sortDir,omitempty"`
}

type GetRfqRequestsParams struct {
	Offset      string           `json:"offset,omitempty"`
	Limit       int              `json:"limit,omitempty"`
	State       RfqListState     `json:"state,omitempty"`
	RequestIDs  []string         `json:"requestIds,omitempty"`
	Markets     []string         `json:"markets,omitempty"`
	SizeMin     float64          `json:"sizeMin,omitempty"`
	SizeMax     float64          `json:"sizeMax,omitempty"`
	SizeUSDCMin float64          `json:"sizeUsdcMin,omitempty"`
	SizeUSDCMax float64          `json:"sizeUsdcMax,omitempty"`
	PriceMin    float64          `json:"priceMin,omitempty"`
	PriceMax    float64          `json:"priceMax,omitempty"`
	SortBy      RfqRequestsSortBy `json:"sortBy,omitempty"`
	SortDir     RfqSortDir       `json:"sortDir,omitempty"`
}

type RfqPaginatedResponse[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor"`
	Limit      int    `json:"limit"`
	Count      int    `json:"count"`
	TotalCount int    `json:"total_count,omitempty"`
}

type RfqRequest struct {
	RequestID   string  `json:"requestId"`
	UserAddress string  `json:"userAddress"`
	ProxyAddress string `json:"proxyAddress"`
	Token       string  `json:"token"`
	Complement  string  `json:"complement"`
	Condition   string  `json:"condition"`
	Side        string  `json:"side"`
	SizeIn      string  `json:"sizeIn"`
	SizeOut     string  `json:"sizeOut"`
	Price       float64 `json:"price"`
	AcceptedQuoteID string `json:"acceptedQuoteId"`
	State       string  `json:"state"`
	Expiry      string  `json:"expiry"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type RfqQuote struct {
	QuoteID     string  `json:"quoteId"`
	RequestID   string  `json:"requestId"`
	UserAddress string  `json:"userAddress"`
	ProxyAddress string `json:"proxyAddress"`
	Complement  string  `json:"complement"`
	Condition   string  `json:"condition"`
	Token       string  `json:"token"`
	Side        string  `json:"side"`
	SizeIn      string  `json:"sizeIn"`
	SizeOut     string  `json:"sizeOut"`
	Price       float64 `json:"price"`
	State       string  `json:"state"`
	Expiry      string  `json:"expiry"`
	MatchType   string  `json:"matchType"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type RfqRequestsResponse = RfqPaginatedResponse[RfqRequest]

type RfqQuotesResponse = RfqPaginatedResponse[RfqQuote]

type RfqRequestResponse struct {
	RequestID string `json:"requestId"`
	Error     string `json:"error,omitempty"`
}

type RfqQuoteResponse struct {
	QuoteID string `json:"quoteId"`
	Error   string `json:"error,omitempty"`
}
