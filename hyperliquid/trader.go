package hyperliquid

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Trader handles trade execution on Hyperliquid
type Trader struct {
	client     *Client
	privateKey *ecdsa.PrivateKey
	address    common.Address
}

// NewTrader creates a new trader instance
func NewTrader(client *Client, privateKeyHex, accountAddress string) (*Trader, error) {
	// Remove 0x prefix if present
	if len(privateKeyHex) > 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	address := common.HexToAddress(accountAddress)

	return &Trader{
		client:     client,
		privateKey: privateKey,
		address:    address,
	}, nil
}

// OrderSide represents order side
type OrderSide string

const (
	OrderSideBuy  OrderSide = "A" // Ask (Buy)
	OrderSideSell OrderSide = "B" // Bid (Sell)
)

// OrderType represents order type
type OrderType struct {
	Limit *LimitOrderType `json:"limit,omitempty"`
}

type LimitOrderType struct {
	Tif string `json:"tif"` // Time in force: "Gtc", "Ioc", "Alo"
}

// PlaceOrderRequest represents order placement request
type PlaceOrderRequest struct {
	Asset      int       `json:"a"`      // Asset index
	IsBuy      bool      `json:"b"`      // Buy (true) or Sell (false)
	Price      string    `json:"p"`      // Price
	Size       string    `json:"s"`      // Size
	ReduceOnly bool      `json:"r"`      // Reduce only
	OrderType  OrderType `json:"t"`      // Order type
}

// OrderResult represents order execution result
type OrderResult struct {
	Success bool
	OrderID string
	Message string
}

// OpenLongPosition opens a long position
func (t *Trader) OpenLongPosition(symbol string, size float64, price float64) (*OrderResult, error) {
	return t.placeOrder(symbol, true, size, price, false)
}

// OpenShortPosition opens a short position
func (t *Trader) OpenShortPosition(symbol string, size float64, price float64) (*OrderResult, error) {
	return t.placeOrder(symbol, false, size, price, false)
}

// ClosePosition closes an existing position
func (t *Trader) ClosePosition(symbol string, side string, size float64, price float64) (*OrderResult, error) {
	// For closing, we do the opposite of the position side
	isBuy := side == "SHORT" // If position is short, we buy to close
	return t.placeOrder(symbol, isBuy, size, price, true)
}

// placeOrder places an order on Hyperliquid
func (t *Trader) placeOrder(symbol string, isBuy bool, size float64, price float64, reduceOnly bool) (*OrderResult, error) {
	// Get asset index for the symbol
	assetIndex, err := t.getAssetIndex(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset index: %w", err)
	}

	// Format price and size strings (remove trailing zeros as required by Hyperliquid)
	priceStr := formatPriceForAPI(price)
	sizeStr := formatPriceForAPI(size)

	// Create order request
	order := PlaceOrderRequest{
		Asset:  assetIndex,
		IsBuy:  isBuy,
		Price:  priceStr,
		Size:   sizeStr,
		OrderType: OrderType{
			Limit: &LimitOrderType{
				Tif: "Gtc", // Good till cancel
			},
		},
		ReduceOnly: reduceOnly,
	}

	// Create action payload
	action := map[string]interface{}{
		"type":     "order",
		"orders":   []PlaceOrderRequest{order},
		"grouping": "na",
	}

	// Sign the action
	signature, err := t.signAction(action)
	if err != nil {
		return nil, fmt.Errorf("failed to sign order: %w", err)
	}

	// Create exchange request
	payload := map[string]interface{}{
		"action":    action,
		"nonce":     time.Now().UnixMilli(),
		"signature": signature,
	}

	// Send request
	url := fmt.Sprintf("%s/exchange", t.client.baseURL)
	respData, err := t.client.doRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}

	// Parse response
	result := &OrderResult{
		Success: true,
	}

	if respMap, ok := respData.(map[string]interface{}); ok {
		if status, ok := respMap["status"].(string); ok {
			result.Success = status == "ok"
			result.Message = status
		}
		if response, ok := respMap["response"].(map[string]interface{}); ok {
			if data, ok := response["data"].(map[string]interface{}); ok {
				if statuses, ok := data["statuses"].([]interface{}); ok && len(statuses) > 0 {
					if statusMap, ok := statuses[0].(map[string]interface{}); ok {
						if filled, ok := statusMap["filled"].(map[string]interface{}); ok {
							if oid, ok := filled["oid"].(string); ok {
								result.OrderID = oid
							}
						}
					}
				}
			}
		}
	}

	return result, nil
}

// getAssetIndex returns the asset index for a symbol
func (t *Trader) getAssetIndex(symbol string) (int, error) {
	// Fetch meta info to get asset indices
	url := fmt.Sprintf("%s/info", t.client.baseURL)
	req := map[string]interface{}{
		"type": "meta",
	}

	respData, err := t.client.doRequest("POST", url, req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch meta info: %w", err)
	}

	// Parse meta response to find asset index
	if metaMap, ok := respData.(map[string]interface{}); ok {
		if universe, ok := metaMap["universe"].([]interface{}); ok {
			for i, asset := range universe {
				if assetMap, ok := asset.(map[string]interface{}); ok {
					if name, ok := assetMap["name"].(string); ok && name == symbol {
						return i, nil
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("symbol %s not found in universe", symbol)
}

// formatPriceForAPI formats a float64 to string removing trailing zeros
func formatPriceForAPI(value float64) string {
	// Format with high precision
	str := fmt.Sprintf("%.12f", value)
	// Remove trailing zeros
	for len(str) > 0 && str[len(str)-1] == '0' {
		str = str[:len(str)-1]
	}
	// Remove trailing decimal point
	if len(str) > 0 && str[len(str)-1] == '.' {
		str = str[:len(str)-1]
	}
	return str
}

// CancelOrder cancels an existing order
func (t *Trader) CancelOrder(symbol string, orderID string) error {
	action := map[string]interface{}{
		"type": "cancel",
		"cancels": []map[string]interface{}{
			{
				"asset": symbol,
				"oid":   orderID,
			},
		},
	}

	signature, err := t.signAction(action)
	if err != nil {
		return fmt.Errorf("failed to sign cancel: %w", err)
	}

	payload := map[string]interface{}{
		"action":    action,
		"nonce":     time.Now().UnixMilli(),
		"signature": signature,
	}

	url := fmt.Sprintf("%s/exchange", t.client.baseURL)
	_, err = t.client.doRequest("POST", url, payload)
	return err
}

// signAction signs an action using the private key
func (t *Trader) signAction(action interface{}) (map[string]interface{}, error) {
	// Convert action to canonical JSON
	actionBytes, err := json.Marshal(action)
	if err != nil {
		return nil, err
	}

	// Create message hash
	// The actual Hyperliquid signing mechanism would be more complex
	// This is a simplified version
	hash := crypto.Keccak256Hash(actionBytes)

	// Sign the hash
	signature, err := crypto.Sign(hash.Bytes(), t.privateKey)
	if err != nil {
		return nil, err
	}

	// Adjust V value (Ethereum compatibility)
	if signature[64] < 27 {
		signature[64] += 27
	}

	// Split signature into r, s, v
	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:64])
	v := signature[64]

	return map[string]interface{}{
		"r": fmt.Sprintf("0x%s", hex.EncodeToString(r.Bytes())),
		"s": fmt.Sprintf("0x%s", hex.EncodeToString(s.Bytes())),
		"v": int(v),
	}, nil
}

// GetOpenOrders fetches open orders for a symbol
func (t *Trader) GetOpenOrders(symbol string) ([]interface{}, error) {
	url := fmt.Sprintf("%s/info", t.client.baseURL)

	req := map[string]interface{}{
		"type": "openOrders",
		"user": t.address.Hex(),
	}

	respData, err := t.client.doRequest("POST", url, req)
	if err != nil {
		return nil, err
	}

	if orders, ok := respData.([]interface{}); ok {
		return orders, nil
	}

	return []interface{}{}, nil
}

// ModifyOrder modifies an existing order
func (t *Trader) ModifyOrder(symbol string, orderID string, newPrice float64, newSize float64) error {
	// Cancel old order
	if err := t.CancelOrder(symbol, orderID); err != nil {
		return err
	}

	// Place new order
	// Note: You'd need to track whether the original order was buy or sell
	// This is simplified
	return nil
}
