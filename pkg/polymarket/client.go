package polymarket

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var ErrInvalidPriceSide = errors.New("invalid price side")

const (
	DefaultGammaBaseURL = "https://gamma-api.polymarket.com"
	DefaultClobBaseURL  = "https://clob.polymarket.com"
	PriceSideBuy        = "buy"
	PriceSideSell       = "sell"
)

type Config struct {
	GammaBaseURL string
	ClobBaseURL  string
	Timeout      time.Duration
	DefaultLimit int
	HTTPClient   *http.Client
}

type Client struct {
	gammaBaseURL string
	clobBaseURL  string
	defaultLimit int
	client       *http.Client
}

type ListMarketsOptions struct {
	Limit  int
	Active *bool
	Closed *bool
	Query  string
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewClient(cfg Config) *Client {
	gammaBaseURL := strings.TrimRight(cfg.GammaBaseURL, "/")
	if gammaBaseURL == "" {
		gammaBaseURL = DefaultGammaBaseURL
	}

	clobBaseURL := strings.TrimRight(cfg.ClobBaseURL, "/")
	if clobBaseURL == "" {
		clobBaseURL = DefaultClobBaseURL
	}

	defaultLimit := cfg.DefaultLimit
	if defaultLimit <= 0 {
		defaultLimit = 10
	}

	client := cfg.HTTPClient
	if client == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 10 * time.Second
		}
		client = &http.Client{Timeout: timeout}
	}

	return &Client{
		gammaBaseURL: gammaBaseURL,
		clobBaseURL:  clobBaseURL,
		defaultLimit: defaultLimit,
		client:       client,
	}
}

func (c *Client) ListMarkets(opts ListMarketsOptions) ([]Market, error) {
	query := url.Values{}
	query.Set("limit", strconv.Itoa(c.normalizeLimit(opts.Limit)))
	if opts.Active != nil {
		query.Set("active", strconv.FormatBool(*opts.Active))
	}
	if opts.Closed != nil {
		query.Set("closed", strconv.FormatBool(*opts.Closed))
	}

	var markets []Market
	if err := c.getJSON(c.gammaURL("/markets", query), &markets); err != nil {
		return nil, err
	}

	if opts.Query == "" {
		return markets, nil
	}

	return filterMarkets(markets, opts.Query), nil
}

func (c *Client) SearchMarkets(query string, limit int) ([]Market, error) {
	return c.ListMarkets(ListMarketsOptions{Limit: limit, Query: query})
}

func (c *Client) GetMarketByID(id string) (*Market, error) {
	var market Market
	if err := c.getJSON(c.gammaURL("/markets/"+url.PathEscape(id), nil), &market); err != nil {
		return nil, err
	}
	return &market, nil
}

func (c *Client) GetMarketBySlug(slug string) (*Market, error) {
	var market Market
	if err := c.getJSON(c.gammaURL("/markets/slug/"+url.PathEscape(slug), nil), &market); err != nil {
		return nil, err
	}
	return &market, nil
}

func (c *Client) ListEvents(limit int) ([]Event, error) {
	query := url.Values{}
	query.Set("limit", strconv.Itoa(c.normalizeLimit(limit)))

	var events []Event
	if err := c.getJSON(c.gammaURL("/events", query), &events); err != nil {
		return nil, err
	}
	return events, nil
}

func (c *Client) GetTokenPrice(tokenID string, side string) (*Price, error) {
	if side != PriceSideBuy && side != PriceSideSell {
		return nil, ErrInvalidPriceSide
	}

	query := url.Values{}
	query.Set("token_id", tokenID)
	query.Set("side", side)

	var price Price
	if err := c.getJSON(c.clobURL("/price", query), &price); err != nil {
		return nil, err
	}
	return &price, nil
}

func (c *Client) GetOrderBook(tokenID string) (*OrderBook, error) {
	query := url.Values{}
	query.Set("token_id", tokenID)

	var book OrderBook
	if err := c.getJSON(c.clobURL("/book", query), &book); err != nil {
		return nil, err
	}
	return &book, nil
}

func (c *Client) gammaURL(path string, query url.Values) string {
	return buildURL(c.gammaBaseURL, path, query)
}

func (c *Client) clobURL(path string, query url.Values) string {
	return buildURL(c.clobBaseURL, path, query)
}

func buildURL(base string, path string, query url.Values) string {
	u := base + path
	if len(query) == 0 {
		return u
	}
	return u + "?" + query.Encode()
}

func (c *Client) normalizeLimit(limit int) int {
	if limit > 0 {
		return limit
	}
	return c.defaultLimit
}

func filterMarkets(markets []Market, query string) []Market {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return markets
	}

	filtered := make([]Market, 0, len(markets))
	for _, market := range markets {
		if strings.Contains(strings.ToLower(market.Question), query) || strings.Contains(strings.ToLower(market.Slug), query) {
			filtered = append(filtered, market)
		}
	}
	return filtered
}

func (c *Client) getJSON(rawURL string, target any) error {
	resp, err := c.client.Get(rawURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		var apiErr errorResponse
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error != "" {
			return fmt.Errorf("polymarket api error: %s", apiErr.Error)
		}
		return fmt.Errorf("polymarket api status: %d", resp.StatusCode)
	}

	return json.Unmarshal(body, target)
}
