package executor

import (
	"fmt"
	"time"

	"aitrading/ai"
	"aitrading/hyperliquid"
	"github.com/sirupsen/logrus"
)

// Executor handles trade execution based on AI decisions
type Executor struct {
	trader         *hyperliquid.Trader
	client         *hyperliquid.Client
	accountAddress string
	logger         *logrus.Logger
}

// NewExecutor creates a new trade executor
func NewExecutor(trader *hyperliquid.Trader, client *hyperliquid.Client, accountAddress string, logger *logrus.Logger) *Executor {
	return &Executor{
		trader:         trader,
		client:         client,
		accountAddress: accountAddress,
		logger:         logger,
	}
}

// ExecutionResult represents the result of trade execution
type ExecutionResult struct {
	Success     bool
	Action      string
	Symbol      string
	Side        string
	Size        float64
	Price       float64
	OrderID     string
	Message     string
	Timestamp   time.Time
	Confidence  float64
	Reason      string
	StopLoss    float64
	TakeProfit  float64
}

// Execute executes a trading decision
func (e *Executor) Execute(symbol string, decision *ai.Decision, currentPrice float64, accountBalance float64) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Action:     decision.Action,
		Symbol:     symbol,
		Timestamp:  time.Now(),
		Confidence: decision.Confidence,
		Reason:     decision.Reason,
		StopLoss:   decision.StopLoss,
		TakeProfit: decision.TakeProfit,
	}

	e.logger.WithFields(logrus.Fields{
		"action":     decision.Action,
		"confidence": decision.Confidence,
		"size":       decision.Size,
		"reason":     decision.Reason,
	}).Info("Executing trading decision")

	switch decision.Action {
	case "OPEN_LONG":
		return e.executeOpenLong(symbol, decision, currentPrice, accountBalance, result)

	case "OPEN_SHORT":
		return e.executeOpenShort(symbol, decision, currentPrice, accountBalance, result)

	case "ADD_POSITION":
		return e.executeAddPosition(symbol, decision, currentPrice, accountBalance, result)

	case "CLOSE_POSITION":
		return e.executeClosePosition(symbol, decision, currentPrice, result)

	case "HOLD":
		result.Success = true
		result.Message = "Holding current position"
		e.logger.Info("Decision: HOLD - No action taken")
		return result, nil

	default:
		result.Success = false
		result.Message = fmt.Sprintf("Unknown action: %s", decision.Action)
		return result, fmt.Errorf("unknown action: %s", decision.Action)
	}
}

// executeOpenLong opens a long position
func (e *Executor) executeOpenLong(symbol string, decision *ai.Decision, currentPrice float64, accountBalance float64, result *ExecutionResult) (*ExecutionResult, error) {
	// Calculate position size based on account balance and decision size
	positionValue := accountBalance * decision.Size
	size := positionValue / currentPrice

	e.logger.WithFields(logrus.Fields{
		"symbol":   symbol,
		"size":     size,
		"price":    currentPrice,
		"stop_loss": decision.StopLoss,
		"take_profit": decision.TakeProfit,
	}).Info("Opening long position")

	orderResult, err := e.trader.OpenLongPosition(symbol, size, currentPrice)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to open long position: %v", err)
		e.logger.WithError(err).Error("Failed to open long position")
		return result, err
	}

	result.Success = orderResult.Success
	result.Side = "LONG"
	result.Size = size
	result.Price = currentPrice
	result.OrderID = orderResult.OrderID
	result.Message = orderResult.Message

	e.logger.WithFields(logrus.Fields{
		"order_id": orderResult.OrderID,
		"success":  orderResult.Success,
	}).Info("Long position opened")

	return result, nil
}

// executeOpenShort opens a short position
func (e *Executor) executeOpenShort(symbol string, decision *ai.Decision, currentPrice float64, accountBalance float64, result *ExecutionResult) (*ExecutionResult, error) {
	// Calculate position size
	positionValue := accountBalance * decision.Size
	size := positionValue / currentPrice

	e.logger.WithFields(logrus.Fields{
		"symbol":      symbol,
		"size":        size,
		"price":       currentPrice,
		"stop_loss":   decision.StopLoss,
		"take_profit": decision.TakeProfit,
	}).Info("Opening short position")

	orderResult, err := e.trader.OpenShortPosition(symbol, size, currentPrice)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to open short position: %v", err)
		e.logger.WithError(err).Error("Failed to open short position")
		return result, err
	}

	result.Success = orderResult.Success
	result.Side = "SHORT"
	result.Size = size
	result.Price = currentPrice
	result.OrderID = orderResult.OrderID
	result.Message = orderResult.Message

	e.logger.WithFields(logrus.Fields{
		"order_id": orderResult.OrderID,
		"success":  orderResult.Success,
	}).Info("Short position opened")

	return result, nil
}

// executeAddPosition adds to existing position
func (e *Executor) executeAddPosition(symbol string, decision *ai.Decision, currentPrice float64, accountBalance float64, result *ExecutionResult) (*ExecutionResult, error) {
	// Get current position
	position, err := e.client.GetPosition(symbol, e.accountAddress)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to get position: %v", err)
		return result, err
	}

	if position.Side == "NONE" {
		result.Success = false
		result.Message = "No existing position to add to"
		e.logger.Warn("Cannot add to position: no existing position")
		return result, fmt.Errorf("no existing position")
	}

	// Calculate additional size
	positionValue := accountBalance * decision.Size
	additionalSize := positionValue / currentPrice

	e.logger.WithFields(logrus.Fields{
		"symbol":          symbol,
		"existing_side":   position.Side,
		"existing_size":   position.Size,
		"additional_size": additionalSize,
		"price":           currentPrice,
	}).Info("Adding to position")

	var orderResult *hyperliquid.OrderResult
	if position.Side == "LONG" {
		orderResult, err = e.trader.OpenLongPosition(symbol, additionalSize, currentPrice)
		result.Side = "LONG"
	} else {
		orderResult, err = e.trader.OpenShortPosition(symbol, additionalSize, currentPrice)
		result.Side = "SHORT"
	}

	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to add to position: %v", err)
		e.logger.WithError(err).Error("Failed to add to position")
		return result, err
	}

	result.Success = orderResult.Success
	result.Size = additionalSize
	result.Price = currentPrice
	result.OrderID = orderResult.OrderID
	result.Message = orderResult.Message

	e.logger.WithFields(logrus.Fields{
		"order_id": orderResult.OrderID,
		"success":  orderResult.Success,
	}).Info("Position added")

	return result, nil
}

// executeClosePosition closes existing position
func (e *Executor) executeClosePosition(symbol string, decision *ai.Decision, currentPrice float64, result *ExecutionResult) (*ExecutionResult, error) {
	// Get current position
	position, err := e.client.GetPosition(symbol, e.accountAddress)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to get position: %v", err)
		return result, err
	}

	if position.Side == "NONE" || position.Size == 0 {
		result.Success = true
		result.Message = "No position to close"
		e.logger.Info("No position to close")
		return result, nil
	}

	e.logger.WithFields(logrus.Fields{
		"symbol": symbol,
		"side":   position.Side,
		"size":   position.Size,
		"price":  currentPrice,
		"pnl":    position.CurrentPnL,
	}).Info("Closing position")

	orderResult, err := e.trader.ClosePosition(symbol, position.Side, position.Size, currentPrice)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to close position: %v", err)
		e.logger.WithError(err).Error("Failed to close position")
		return result, err
	}

	result.Success = orderResult.Success
	result.Side = position.Side
	result.Size = position.Size
	result.Price = currentPrice
	result.OrderID = orderResult.OrderID
	result.Message = fmt.Sprintf("Position closed. PnL: %.2f%%", position.PnLPercent)

	e.logger.WithFields(logrus.Fields{
		"order_id": orderResult.OrderID,
		"success":  orderResult.Success,
		"pnl":      position.PnLPercent,
	}).Info("Position closed")

	return result, nil
}
