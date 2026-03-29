package clob

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"
)

const (
	protocolName    = "Polymarket CTF Exchange"
	protocolVersion = "1"
	zeroAddress     = "0x0000000000000000000000000000000000000000"
	collateralTokenDecimals = 6
)

type contractConfig struct {
	exchange        string
	negRiskExchange string
}

func getContractConfig(chainID Chain) (contractConfig, error) {
	switch chainID {
	case ChainPolygon:
		return contractConfig{exchange: "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E", negRiskExchange: "0xC5d563A36AE78145C45a50134d48A1215220f80a"}, nil
	case ChainAmoy:
		return contractConfig{exchange: "0xdFE02Eb6733538f8Ea35D585af8DE5958AD99E40", negRiskExchange: "0xC5d563A36AE78145C45a50134d48A1215220f80a"}, nil
	default:
		return contractConfig{}, fmt.Errorf("invalid network")
	}
}

type roundConfig struct {
	price  int
	size   int
	amount int
}

var roundingConfig = map[TickSize]roundConfig{
	"0.1":    {price: 1, size: 2, amount: 3},
	"0.01":   {price: 2, size: 2, amount: 4},
	"0.001":  {price: 3, size: 2, amount: 5},
	"0.0001": {price: 4, size: 2, amount: 6},
}

func decimalPlaces(num float64) int {
	s := strconv.FormatFloat(num, 'f', -1, 64)
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			return len(s) - i - 1
		}
	}
	return 0
}

func roundNormal(num float64, decimals int) float64 {
	if decimalPlaces(num) <= decimals {
		return num
	}
	pow := math.Pow10(decimals)
	return math.Round((num+math.SmallestNonzeroFloat64)*pow) / pow
}

func roundDown(num float64, decimals int) float64 {
	if decimalPlaces(num) <= decimals {
		return num
	}
	pow := math.Pow10(decimals)
	return math.Floor(num*pow) / pow
}

func roundUp(num float64, decimals int) float64 {
	if decimalPlaces(num) <= decimals {
		return num
	}
	pow := math.Pow10(decimals)
	return math.Ceil(num*pow) / pow
}

func parseUnits(value float64, decimals int) string {
	pow := math.Pow10(decimals)
	return strconv.FormatInt(int64(math.Round(value*pow)), 10)
}

func generateOrderSalt() string {
	return strconv.FormatInt(time.Now().UnixNano()%9223372036854775807, 10)
}

func (c *Client) defaultCreateOrder(ctx context.Context, userOrder UserOrder, options *CreateOrderOptions) (SignedOrder, error) {
	if err := c.ensureL1Auth(); err != nil {
		return SignedOrder{}, err
	}
	if options == nil {
		return SignedOrder{}, fmt.Errorf("options required")
	}
	rc, ok := roundingConfig[options.TickSize]
	if !ok {
		return SignedOrder{}, fmt.Errorf("unsupported tick size: %s", options.TickSize)
	}
	makerAddress, err := c.signer.Address(ctx)
	if err != nil {
		return SignedOrder{}, err
	}
	if userOrder.Taker == "" {
		userOrder.Taker = zeroAddress
	}
	feeRateBps := "0"
	if userOrder.FeeRateBPS != nil {
		feeRateBps = strconv.Itoa(*userOrder.FeeRateBPS)
	}
	nonce := "0"
	if userOrder.Nonce != nil {
		nonce = strconv.FormatInt(*userOrder.Nonce, 10)
	}
	expiration := "0"
	if userOrder.Expiration != nil {
		expiration = strconv.FormatInt(*userOrder.Expiration, 10)
	}

	rawPrice := roundNormal(userOrder.Price, rc.price)
	var rawMakerAmt float64
	var rawTakerAmt float64
	if userOrder.Side == SideBUY {
		rawTakerAmt = roundDown(userOrder.Size, rc.size)
		rawMakerAmt = rawTakerAmt * rawPrice
		if decimalPlaces(rawMakerAmt) > rc.amount {
			rawMakerAmt = roundUp(rawMakerAmt, rc.amount+4)
			if decimalPlaces(rawMakerAmt) > rc.amount {
				rawMakerAmt = roundDown(rawMakerAmt, rc.amount)
			}
		}
	} else {
		rawMakerAmt = roundDown(userOrder.Size, rc.size)
		rawTakerAmt = rawMakerAmt * rawPrice
		if decimalPlaces(rawTakerAmt) > rc.amount {
			rawTakerAmt = roundUp(rawTakerAmt, rc.amount+4)
			if decimalPlaces(rawTakerAmt) > rc.amount {
				rawTakerAmt = roundDown(rawTakerAmt, rc.amount)
			}
		}
	}

	order := SignedOrder{
		Salt:          generateOrderSalt(),
		Maker:         makerAddress,
		Signer:        makerAddress,
		Taker:         userOrder.Taker,
		TokenID:       userOrder.TokenID,
		MakerAmount:   parseUnits(rawMakerAmt, collateralTokenDecimals),
		TakerAmount:   parseUnits(rawTakerAmt, collateralTokenDecimals),
		Expiration:    expiration,
		Nonce:         nonce,
		FeeRateBPS:    feeRateBps,
		Side:          userOrder.Side,
		SignatureType: int(c.signatureType),
	}
	typedData, err := c.buildOrderTypedData(ctx, order, options)
	if err != nil {
		return SignedOrder{}, err
	}
	sig, err := c.signer.SignOrderTypedData(ctx, typedData)
	if err != nil {
		return SignedOrder{}, err
	}
	order.Signature = sig
	return order, nil
}

func (c *Client) defaultCreateMarketOrder(ctx context.Context, userOrder UserMarketOrder, options *CreateOrderOptions) (SignedOrder, error) {
	if userOrder.Price == nil {
		price, err := c.CalculateMarketPrice(ctx, userOrder.TokenID, userOrder.Side, userOrder.Amount, userOrder.OrderType)
		if err != nil {
			return SignedOrder{}, err
		}
		userOrder.Price = &price
	}
	fee := userOrder.FeeRateBPS
	nonce := userOrder.Nonce
	return c.defaultCreateOrder(ctx, UserOrder{
		TokenID:    userOrder.TokenID,
		Price:      *userOrder.Price,
		Size:       userOrder.Amount,
		Side:       userOrder.Side,
		FeeRateBPS: fee,
		Nonce:      nonce,
		Taker:      userOrder.Taker,
	}, options)
}

func (c *Client) buildOrderTypedData(ctx context.Context, order SignedOrder, options *CreateOrderOptions) (OrderTypedDataPayload, error) {
	_ = ctx
	cfg, err := getContractConfig(c.chainID)
	if err != nil {
		return OrderTypedDataPayload{}, err
	}
	verifyingContract := cfg.exchange
	if options != nil && options.NegRisk != nil && *options.NegRisk {
		verifyingContract = cfg.negRiskExchange
	}
	side := 1
	if order.Side == SideBUY {
		side = 0
	}
	return OrderTypedDataPayload{
		PrimaryType: "Order",
		Domain: map[string]any{
			"name":              protocolName,
			"version":           protocolVersion,
			"chainId":           int(c.chainID),
			"verifyingContract": verifyingContract,
		},
		Types: map[string][]map[string]string{
			"Order": {
				{"name": "salt", "type": "uint256"},
				{"name": "maker", "type": "address"},
				{"name": "signer", "type": "address"},
				{"name": "taker", "type": "address"},
				{"name": "tokenId", "type": "uint256"},
				{"name": "makerAmount", "type": "uint256"},
				{"name": "takerAmount", "type": "uint256"},
				{"name": "expiration", "type": "uint256"},
				{"name": "nonce", "type": "uint256"},
				{"name": "feeRateBps", "type": "uint256"},
				{"name": "side", "type": "uint8"},
				{"name": "signatureType", "type": "uint8"},
			},
		},
		Message: map[string]any{
			"salt":          order.Salt,
			"maker":         order.Maker,
			"signer":        order.Signer,
			"taker":         order.Taker,
			"tokenId":       order.TokenID,
			"makerAmount":   order.MakerAmount,
			"takerAmount":   order.TakerAmount,
			"expiration":    order.Expiration,
			"nonce":         order.Nonce,
			"feeRateBps":    order.FeeRateBPS,
			"side":          side,
			"signatureType": order.SignatureType,
		},
	}, nil
}

func (c *Client) CalculateMarketPrice(ctx context.Context, tokenID string, side Side, amount float64, orderType OrderType) (float64, error) {
	book, err := c.GetOrderBook(ctx, tokenID)
	if err != nil {
		return 0, err
	}
	if side == SideBUY {
		if len(book.Asks) == 0 {
			return 0, fmt.Errorf("no match")
		}
		sum := 0.0
		for i := len(book.Asks) - 1; i >= 0; i-- {
			p, _ := strconv.ParseFloat(book.Asks[i].Price, 64)
			s, _ := strconv.ParseFloat(book.Asks[i].Size, 64)
			sum += s * p
			if sum >= amount {
				return p, nil
			}
		}
		if orderType == OrderTypeFOK {
			return 0, fmt.Errorf("no match")
		}
		p, _ := strconv.ParseFloat(book.Asks[0].Price, 64)
		return p, nil
	}
	if len(book.Bids) == 0 {
		return 0, fmt.Errorf("no match")
	}
	sum := 0.0
	for i := len(book.Bids) - 1; i >= 0; i-- {
		p, _ := strconv.ParseFloat(book.Bids[i].Price, 64)
		s, _ := strconv.ParseFloat(book.Bids[i].Size, 64)
		sum += s
		if sum >= amount {
			return p, nil
		}
	}
	if orderType == OrderTypeFOK {
		return 0, fmt.Errorf("no match")
	}
	p, _ := strconv.ParseFloat(book.Bids[0].Price, 64)
	return p, nil
}

func (c *Client) resolveCreateOrderFactory() func(ctx context.Context, order UserOrder, options *CreateOrderOptions) (SignedOrder, error) {
	if c.createOrderFn != nil {
		return c.createOrderFn
	}
	return c.defaultCreateOrder
}

func (c *Client) resolveCreateMarketOrderFactory() func(ctx context.Context, order UserMarketOrder, options *CreateOrderOptions) (SignedOrder, error) {
	if c.createMarketOrderFn != nil {
		return c.createMarketOrderFn
	}
	return c.defaultCreateMarketOrder
}

func (c *Client) CreateOrder(ctx context.Context, userOrder UserOrder, options *CreateOrderOptions) (SignedOrder, error) {
	if options == nil || options.TickSize == "" {
		tick, err := c.GetTickSize(ctx, userOrder.TokenID)
		if err != nil {
			return SignedOrder{}, err
		}
		options = &CreateOrderOptions{TickSize: tick}
	}
	if userOrder.FeeRateBPS == nil {
		fee, err := c.GetFeeRateBps(ctx, userOrder.TokenID)
		if err != nil {
			return SignedOrder{}, err
		}
		userOrder.FeeRateBPS = &fee
	}
	if options.NegRisk == nil {
		neg, err := c.GetNegRisk(ctx, userOrder.TokenID)
		if err != nil {
			return SignedOrder{}, err
		}
		options.NegRisk = &neg
	}
	priceTick, err := strconv.ParseFloat(string(options.TickSize), 64)
	if err != nil {
		return SignedOrder{}, err
	}
	if userOrder.Price < priceTick || userOrder.Price > 1-priceTick {
		return SignedOrder{}, fmt.Errorf("invalid price (%v), min: %v - max: %v", userOrder.Price, priceTick, 1-priceTick)
	}
	fn := c.resolveCreateOrderFactory()
	return fn(ctx, userOrder, options)
}

func (c *Client) CreateMarketOrder(ctx context.Context, userOrder UserMarketOrder, options *CreateOrderOptions) (SignedOrder, error) {
	if options == nil || options.TickSize == "" {
		tick, err := c.GetTickSize(ctx, userOrder.TokenID)
		if err != nil {
			return SignedOrder{}, err
		}
		options = &CreateOrderOptions{TickSize: tick}
	}
	if userOrder.FeeRateBPS == nil {
		fee, err := c.GetFeeRateBps(ctx, userOrder.TokenID)
		if err != nil {
			return SignedOrder{}, err
		}
		userOrder.FeeRateBPS = &fee
	}
	if options.NegRisk == nil {
		neg, err := c.GetNegRisk(ctx, userOrder.TokenID)
		if err != nil {
			return SignedOrder{}, err
		}
		options.NegRisk = &neg
	}
	if userOrder.Price == nil {
		orderType := userOrder.OrderType
		if orderType == "" {
			orderType = OrderTypeFOK
		}
		p, err := c.CalculateMarketPrice(ctx, userOrder.TokenID, userOrder.Side, userOrder.Amount, orderType)
		if err != nil {
			return SignedOrder{}, err
		}
		userOrder.Price = &p
	}
	priceTick, err := strconv.ParseFloat(string(options.TickSize), 64)
	if err != nil {
		return SignedOrder{}, err
	}
	if *userOrder.Price < priceTick || *userOrder.Price > 1-priceTick {
		return SignedOrder{}, fmt.Errorf("invalid price (%v), min: %v - max: %v", *userOrder.Price, priceTick, 1-priceTick)
	}
	fn := c.resolveCreateMarketOrderFactory()
	return fn(ctx, userOrder, options)
}

func (c *Client) CreateAndPostOrder(ctx context.Context, userOrder UserOrder, options *CreateOrderOptions, orderType OrderType, deferExec bool, postOnly *bool) (any, error) {
	if orderType == "" {
		orderType = OrderTypeGTC
	}
	order, err := c.CreateOrder(ctx, userOrder, options)
	if err != nil {
		return nil, err
	}
	return c.PostOrder(ctx, order, orderType, deferExec, postOnly)
}

func (c *Client) CreateAndPostMarketOrder(ctx context.Context, userOrder UserMarketOrder, options *CreateOrderOptions, orderType OrderType, deferExec bool) (any, error) {
	if orderType == "" {
		orderType = OrderTypeFOK
	}
	order, err := c.CreateMarketOrder(ctx, userOrder, options)
	if err != nil {
		return nil, err
	}
	return c.PostOrder(ctx, order, orderType, deferExec, nil)
}
