package polymarket

const (
	InitialCursor = "MA=="
	EndCursor     = "LTE="
)

const (
	TimeEndpoint = "/time"

	CreateAPIKeyEndpoint = "/auth/api-key"
	GetAPIKeysEndpoint   = "/auth/api-keys"
	DeleteAPIKeyEndpoint = "/auth/api-key"
	DeriveAPIKeyEndpoint = "/auth/derive-api-key"
	ClosedOnlyEndpoint   = "/auth/ban-status/closed-only"

	CreateReadonlyAPIKeyEndpoint   = "/auth/readonly-api-key"
	GetReadonlyAPIKeysEndpoint     = "/auth/readonly-api-keys"
	DeleteReadonlyAPIKeyEndpoint   = "/auth/readonly-api-key"
	ValidateReadonlyAPIKeyEndpoint = "/auth/validate-readonly-api-key"

	CreateBuilderAPIKeyEndpoint = "/auth/builder-api-key"
	GetBuilderAPIKeysEndpoint   = "/auth/builder-api-key"
	RevokeBuilderAPIKeyEndpoint = "/auth/builder-api-key"

	GetSamplingSimplifiedMarketsEndpoint = "/sampling-simplified-markets"
	GetSamplingMarketsEndpoint           = "/sampling-markets"
	GetSimplifiedMarketsEndpoint         = "/simplified-markets"
	GetMarketsEndpoint                   = "/markets"
	GetMarketEndpoint                    = "/markets/"
	GetOrderBookEndpoint                 = "/book"
	GetOrderBooksEndpoint                = "/books"
	GetMidpointEndpoint                  = "/midpoint"
	GetMidpointsEndpoint                 = "/midpoints"
	GetPriceEndpoint                     = "/price"
	GetPricesEndpoint                    = "/prices"
	GetSpreadEndpoint                    = "/spread"
	GetSpreadsEndpoint                   = "/spreads"
	GetLastTradePriceEndpoint            = "/last-trade-price"
	GetLastTradesPricesEndpoint          = "/last-trades-prices"
	GetTickSizeEndpoint                  = "/tick-size"
	GetNegRiskEndpoint                   = "/neg-risk"
	GetFeeRateEndpoint                   = "/fee-rate"

	PostOrderEndpoint          = "/order"
	PostOrdersEndpoint         = "/orders"
	CancelOrderEndpoint        = "/order"
	CancelOrdersEndpoint       = "/orders"
	GetOrderEndpoint           = "/data/order/"
	CancelAllEndpoint          = "/cancel-all"
	CancelMarketOrdersEndpoint = "/cancel-market-orders"
	GetOpenOrdersEndpoint      = "/data/orders"
	GetTradesEndpoint          = "/data/trades"
	IsOrderScoringEndpoint     = "/order-scoring"
	AreOrdersScoringEndpoint   = "/orders-scoring"

	GetPricesHistoryEndpoint = "/prices-history"

	GetNotificationsEndpoint  = "/notifications"
	DropNotificationsEndpoint = "/notifications"

	GetBalanceAllowanceEndpoint    = "/balance-allowance"
	UpdateBalanceAllowanceEndpoint = "/balance-allowance/update"

	GetMarketTradesEventsEndpoint = "/live-activity/events/"

	GetEarningsForUserForDayEndpoint      = "/rewards/user"
	GetTotalEarningsForUserForDayEndpoint = "/rewards/user/total"
	GetLiquidityRewardPercentagesEndpoint = "/rewards/user/percentages"
	GetRewardsMarketsCurrentEndpoint      = "/rewards/markets/current"
	GetRewardsMarketsEndpoint             = "/rewards/markets/"
	GetRewardsEarningsPercentagesEndpoint = "/rewards/user/markets"

	GetBuilderTradesEndpoint = "/builder/trades"

	PostHeartbeatEndpoint = "/v1/heartbeats"

	CreateRFQRequestEndpoint      = "/rfq/request"
	CancelRFQRequestEndpoint      = "/rfq/request"
	GetRFQRequestsEndpoint        = "/rfq/data/requests"
	CreateRFQQuoteEndpoint        = "/rfq/quote"
	CancelRFQQuoteEndpoint        = "/rfq/quote"
	RFQRequestsAcceptEndpoint     = "/rfq/request/accept"
	RFQQuoteApproveEndpoint       = "/rfq/quote/approve"
	GetRFQRequesterQuotesEndpoint = "/rfq/data/requester/quotes"
	GetRFQQuoterQuotesEndpoint    = "/rfq/data/quoter/quotes"
	GetRFQBestQuoteEndpoint       = "/rfq/data/best-quote"
	RFQConfigEndpoint             = "/rfq/config"
)
