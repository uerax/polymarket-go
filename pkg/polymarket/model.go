package polymarket

type Market struct {
	ID             string  `json:"id"`
	Question       string  `json:"question"`
	Slug           string  `json:"slug"`
	Description    string  `json:"description"`
	EndDate        string  `json:"endDate"`
	Active         bool    `json:"active"`
	Closed         bool    `json:"closed"`
	Volume         string  `json:"volume"`
	Liquidity      string  `json:"liquidity"`
	VolumeNum      float64 `json:"volumeNum"`
	LiquidityNum   float64 `json:"liquidityNum"`
	Outcomes       string  `json:"outcomes"`
	OutcomePrices  string  `json:"outcomePrices"`
	ClobTokenIDs   string  `json:"clobTokenIds"`
	BestBid        float64 `json:"bestBid"`
	BestAsk        float64 `json:"bestAsk"`
	LastTradePrice float64 `json:"lastTradePrice"`
}

type Event struct {
	ID          string   `json:"id"`
	Ticker      string   `json:"ticker"`
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	StartDate   string   `json:"startDate"`
	EndDate     string   `json:"endDate"`
	Category    string   `json:"category"`
	Active      bool     `json:"active"`
	Closed      bool     `json:"closed"`
	Markets     []Market `json:"markets"`
}

type Price struct {
	Price string `json:"price"`
}

type BookLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

type OrderBook struct {
	Market         string      `json:"market"`
	AssetID        string      `json:"asset_id"`
	Timestamp      string      `json:"timestamp"`
	Hash           string      `json:"hash"`
	Bids           []BookLevel `json:"bids"`
	Asks           []BookLevel `json:"asks"`
	MinOrderSize   string      `json:"min_order_size"`
	TickSize       string      `json:"tick_size"`
	NegRisk        bool        `json:"neg_risk"`
	LastTradePrice string      `json:"last_trade_price"`
}
