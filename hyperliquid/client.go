package hyperliquid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"aitrading/indicators"
)

// Client handles Hyperliquid API interactions
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Hyperliquid client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// MarketInfo contains market information
type MarketInfo struct {
	Symbol       string
	CurrentPrice float64
	PriceChange  float64
	Volume24h    float64
	High24h      float64
	Low24h       float64
}

// Position represents a trading position
type Position struct {
	Symbol      string
	Side        string // "LONG" or "SHORT"
	Size        float64
	EntryPrice  float64
	CurrentPnL  float64
	PnLPercent  float64
	OpenTime    time.Time
	HoldingTime time.Duration
}

// GetMarketData fetches current market information
func (c *Client) GetMarketData(symbol string) (*MarketInfo, error) {
	url := fmt.Sprintf("%s/info", c.baseURL)

	// Get current prices using allMids
	req := map[string]interface{}{
		"type": "allMids",
	}

	respData, err := c.doRequest("POST", url, req)
	if err != nil {
		return nil, err
	}

	marketInfo := &MarketInfo{
		Symbol: symbol,
	}

	// Parse allMids response - it's a map of symbol to price
	if priceMap, ok := respData.(map[string]interface{}); ok {
		if priceStr, ok := priceMap[symbol].(string); ok {
			fmt.Sscanf(priceStr, "%f", &marketInfo.CurrentPrice)
		}
	}

	// Get meta and asset contexts for volume data
	req2 := map[string]interface{}{
		"type": "metaAndAssetCtxs",
	}

	respData2, err := c.doRequest("POST", url, req2)
	if err == nil {
		// Parse response array: [meta, assetCtxs]
		if respArray, ok := respData2.([]interface{}); ok && len(respArray) >= 2 {
			// assetCtxs is the second element
			if assetCtxs, ok := respArray[1].([]interface{}); ok {
				// Get metadata to find symbol index
				var symbolIndex int
				if meta, ok := respArray[0].(map[string]interface{}); ok {
					if universe, ok := meta["universe"].([]interface{}); ok {
						for i, asset := range universe {
							if assetMap, ok := asset.(map[string]interface{}); ok {
								if name, ok := assetMap["name"].(string); ok && name == symbol {
									symbolIndex = i
									break
								}
							}
						}
					}
				}

				// Get data from assetCtxs using the index
				if symbolIndex < len(assetCtxs) {
					if ctxMap, ok := assetCtxs[symbolIndex].(map[string]interface{}); ok {
						// Get 24h volume
						if dayNtlVlm, ok := ctxMap["dayNtlVlm"].(string); ok {
							var volume float64
							if _, err := fmt.Sscanf(dayNtlVlm, "%f", &volume); err == nil {
								marketInfo.Volume24h = volume
							}
						}
						// Calculate 24h change from prevDayPx and markPx
						var prevPrice, markPrice float64
						if prevDayPx, ok := ctxMap["prevDayPx"].(string); ok {
							fmt.Sscanf(prevDayPx, "%f", &prevPrice)
						}
						if markPx, ok := ctxMap["markPx"].(string); ok {
							fmt.Sscanf(markPx, "%f", &markPrice)
							// Use mark price if current price is still 0
							if marketInfo.CurrentPrice == 0 {
								marketInfo.CurrentPrice = markPrice
							}
						}
						// Calculate percentage change
						if prevPrice > 0 && markPrice > 0 {
							marketInfo.PriceChange = ((markPrice - prevPrice) / prevPrice) * 100
						}
					}
				}
			}
		}
	}

	return marketInfo, nil
}

// GetCandlestickData fetches historical candlestick data
func (c *Client) GetCandlestickData(symbol, interval string, limit int) ([]indicators.MarketData, error) {
	url := fmt.Sprintf("%s/info", c.baseURL)

	endTime := time.Now().Unix() * 1000
	startTime := endTime - int64(limit*c.intervalToMilliseconds(interval))

	req := map[string]interface{}{
		"type": "candleSnapshot",
		"req": map[string]interface{}{
			"coin":      symbol,
			"interval":  interval,
			"startTime": startTime,
			"endTime":   endTime,
		},
	}

	respData, err := c.doRequest("POST", url, req)
	if err != nil {
		return nil, err
	}

	// Parse candlestick data
	candles := []indicators.MarketData{}

	if candleData, ok := respData.([]interface{}); ok {
		for _, candle := range candleData {
			if c, ok := candle.(map[string]interface{}); ok {
				md := indicators.MarketData{}

				if t, ok := c["t"].(float64); ok {
					md.Timestamp = int64(t)
				}
				if o, ok := c["o"].(string); ok {
					fmt.Sscanf(o, "%f", &md.Open)
				}
				if h, ok := c["h"].(string); ok {
					fmt.Sscanf(h, "%f", &md.High)
				}
				if l, ok := c["l"].(string); ok {
					fmt.Sscanf(l, "%f", &md.Low)
				}
				if cl, ok := c["c"].(string); ok {
					fmt.Sscanf(cl, "%f", &md.Close)
				}
				if v, ok := c["v"].(string); ok {
					fmt.Sscanf(v, "%f", &md.Volume)
				}

				candles = append(candles, md)
			}
		}
	}

	return candles, nil
}

// GetPosition fetches current position for a symbol
func (c *Client) GetPosition(symbol, accountAddress string) (*Position, error) {
	url := fmt.Sprintf("%s/info", c.baseURL)

	req := map[string]interface{}{
		"type": "clearinghouseState",
		"user": accountAddress,
	}

	respData, err := c.doRequest("POST", url, req)
	if err != nil {
		return nil, err
	}

	// Parse position data
	if stateMap, ok := respData.(map[string]interface{}); ok {
		if assetPositions, ok := stateMap["assetPositions"].([]interface{}); ok {
			for _, pos := range assetPositions {
				if posMap, ok := pos.(map[string]interface{}); ok {
					if positionMap, ok := posMap["position"].(map[string]interface{}); ok {
						if coin, ok := positionMap["coin"].(string); ok && coin == symbol {
							position := &Position{
								Symbol:   symbol,
								OpenTime: time.Now(), // Would need to track this separately
							}

							if szi, ok := positionMap["szi"].(string); ok {
								var size float64
								fmt.Sscanf(szi, "%f", &size)
								position.Size = size
								if size > 0 {
									position.Side = "LONG"
								} else if size < 0 {
									position.Side = "SHORT"
									position.Size = -size
								}
							}

							if entryPx, ok := positionMap["entryPx"].(string); ok {
								fmt.Sscanf(entryPx, "%f", &position.EntryPrice)
							}

							if unrealizedPnl, ok := positionMap["unrealizedPnl"].(string); ok {
								fmt.Sscanf(unrealizedPnl, "%f", &position.CurrentPnL)
							}

							if position.EntryPrice > 0 {
								position.PnLPercent = (position.CurrentPnL / position.EntryPrice) * 100
							}

							return position, nil
						}
					}
				}
			}
		}
	}

	// No position found
	return &Position{
		Symbol: symbol,
		Side:   "NONE",
		Size:   0,
	}, nil
}

// GetAccountBalance fetches account balance
func (c *Client) GetAccountBalance(accountAddress string) (float64, error) {
	url := fmt.Sprintf("%s/info", c.baseURL)

	req := map[string]interface{}{
		"type": "clearinghouseState",
		"user": accountAddress,
	}

	respData, err := c.doRequest("POST", url, req)
	if err != nil {
		return 0, err
	}

	// Parse balance
	if stateMap, ok := respData.(map[string]interface{}); ok {
		if marginSummary, ok := stateMap["marginSummary"].(map[string]interface{}); ok {
			if accountValue, ok := marginSummary["accountValue"].(string); ok {
				var balance float64
				fmt.Sscanf(accountValue, "%f", &balance)
				return balance, nil
			}
		}
	}

	return 0, fmt.Errorf("failed to parse account balance")
}

// doRequest performs HTTP request to Hyperliquid API
func (c *Client) doRequest(method, url string, payload interface{}) (interface{}, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

// intervalToMilliseconds converts interval string to milliseconds
func (c *Client) intervalToMilliseconds(interval string) int {
	switch interval {
	case "1m":
		return 60 * 1000
	case "5m":
		return 5 * 60 * 1000
	case "15m":
		return 15 * 60 * 1000
	case "1h":
		return 60 * 60 * 1000
	case "4h":
		return 4 * 60 * 60 * 1000
	case "1d":
		return 24 * 60 * 60 * 1000
	default:
		return 5 * 60 * 1000 // default 5m
	}
}
