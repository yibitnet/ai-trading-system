package risk

import (
	"fmt"
	"time"

	"aitrading/ai"
	"aitrading/config"
	"aitrading/hyperliquid"
	"github.com/sirupsen/logrus"
)

// Controller handles risk management
type Controller struct {
	config        *config.RiskConfig
	tradingConfig *config.TradingConfig
	logger        *logrus.Logger
	dailyPnL      float64
	dailyPnLReset time.Time
	maxDrawdown   float64
	peakBalance   float64
}

// NewController creates a new risk controller
func NewController(riskCfg *config.RiskConfig, tradingCfg *config.TradingConfig, logger *logrus.Logger) *Controller {
	return &Controller{
		config:        riskCfg,
		tradingConfig: tradingCfg,
		logger:        logger,
		dailyPnLReset: time.Now(),
	}
}

// RiskCheckResult represents the result of risk check
type RiskCheckResult struct {
	Approved       bool
	Reason         string
	AdjustedSize   float64
	AdjustedLeverage int
	StopLoss       float64
	TakeProfit     float64
}

// CheckDecision validates a trading decision against risk rules
func (rc *Controller) CheckDecision(
	decision *ai.Decision,
	currentPrice float64,
	accountBalance float64,
	position *hyperliquid.Position,
	openPositionCount int,
) (*RiskCheckResult, error) {

	result := &RiskCheckResult{
		Approved:         true,
		AdjustedSize:     decision.Size,
		AdjustedLeverage: decision.Leverage,
		StopLoss:         decision.StopLoss,
		TakeProfit:       decision.TakeProfit,
	}

	// Reset daily PnL if needed
	rc.resetDailyPnLIfNeeded()

	// Check 1: Maximum open positions
	if decision.Action == "OPEN_LONG" || decision.Action == "OPEN_SHORT" {
		if rc.tradingConfig.MaxOpenPositions > 0 && openPositionCount >= rc.tradingConfig.MaxOpenPositions {
			result.Approved = false
			result.Reason = fmt.Sprintf("Maximum open positions reached: %d >= %d",
				openPositionCount, rc.tradingConfig.MaxOpenPositions)
			rc.logger.Warn(result.Reason)
			return result, nil
		}
	}

	// Check 2: Leverage limit
	if decision.Action == "OPEN_LONG" || decision.Action == "OPEN_SHORT" || decision.Action == "ADD_POSITION" {
		if rc.tradingConfig.MaxLeverage > 0 && decision.Leverage > rc.tradingConfig.MaxLeverage {
			result.AdjustedLeverage = rc.tradingConfig.MaxLeverage
			rc.logger.WithFields(logrus.Fields{
				"original_leverage": decision.Leverage,
				"adjusted_leverage": result.AdjustedLeverage,
			}).Warn("Leverage adjusted down to max limit")
		}
	}

	// Check 3: Minimum confidence
	if decision.Confidence < 0.6 {
		result.Approved = false
		result.Reason = fmt.Sprintf("Confidence too low: %.2f < 0.60", decision.Confidence)
		rc.logger.Warn(result.Reason)
		return result, nil
	}

	// Check 4: Daily loss limit
	if rc.dailyPnL < 0 && -rc.dailyPnL/accountBalance > rc.config.DailyLossLimit {
		result.Approved = false
		result.Reason = fmt.Sprintf("Daily loss limit exceeded: %.2f%% > %.2f%%",
			-rc.dailyPnL/accountBalance*100, rc.config.DailyLossLimit*100)
		rc.logger.Warn(result.Reason)
		return result, nil
	}

	// Check 5: Maximum drawdown
	if rc.peakBalance > 0 {
		currentDrawdown := (rc.peakBalance - accountBalance) / rc.peakBalance
		if currentDrawdown > rc.config.MaxDrawdown {
			result.Approved = false
			result.Reason = fmt.Sprintf("Max drawdown exceeded: %.2f%% > %.2f%%",
				currentDrawdown*100, rc.config.MaxDrawdown*100)
			rc.logger.Warn(result.Reason)
			return result, nil
		}
	}

	// Update peak balance
	if accountBalance > rc.peakBalance {
		rc.peakBalance = accountBalance
	}

	// Check 6: Position size limits
	if decision.Action == "OPEN_LONG" || decision.Action == "OPEN_SHORT" || decision.Action == "ADD_POSITION" {
		// Check against max position size
		if decision.Size > rc.config.PositionRiskPerTrade*10 { // Max 10x the per-trade risk
			result.AdjustedSize = rc.config.PositionRiskPerTrade * 10
			rc.logger.WithFields(logrus.Fields{
				"original_size": decision.Size,
				"adjusted_size": result.AdjustedSize,
			}).Warn("Position size adjusted down")
		}

		// Check total exposure
		currentExposure := 0.0
		if position.Size > 0 {
			currentExposure = (position.Size * currentPrice) / accountBalance
		}

		newExposure := currentExposure + result.AdjustedSize
		if newExposure > rc.config.MaxTotalExposure {
			result.Approved = false
			result.Reason = fmt.Sprintf("Total exposure would exceed limit: %.2f%% > %.2f%%",
				newExposure*100, rc.config.MaxTotalExposure*100)
			rc.logger.Warn(result.Reason)
			return result, nil
		}
	}

	// Check 7: Risk-reward ratio
	if decision.Action == "OPEN_LONG" || decision.Action == "OPEN_SHORT" {
		if decision.StopLoss > 0 && decision.TakeProfit > 0 {
			var riskReward float64

			if decision.Action == "OPEN_LONG" {
				risk := currentPrice - decision.StopLoss
				reward := decision.TakeProfit - currentPrice
				if risk > 0 {
					riskReward = reward / risk
				}
			} else {
				risk := decision.StopLoss - currentPrice
				reward := currentPrice - decision.TakeProfit
				if risk > 0 {
					riskReward = reward / risk
				}
			}

			if riskReward < rc.config.MinRiskRewardRatio {
				result.Approved = false
				result.Reason = fmt.Sprintf("Risk-reward ratio too low: %.2f < %.2f",
					riskReward, rc.config.MinRiskRewardRatio)
				rc.logger.Warn(result.Reason)
				return result, nil
			}
		}
	}

	// Check 8: Stop loss validation
	if decision.Action == "OPEN_LONG" || decision.Action == "OPEN_SHORT" {
		if decision.StopLoss <= 0 || decision.TakeProfit <= 0 {
			result.Approved = false
			result.Reason = "Stop loss and take profit must be set"
			rc.logger.Warn(result.Reason)
			return result, nil
		}

		// Validate stop loss is on correct side
		if decision.Action == "OPEN_LONG" && decision.StopLoss >= currentPrice {
			result.Approved = false
			result.Reason = "Stop loss for long position must be below current price"
			rc.logger.Warn(result.Reason)
			return result, nil
		}

		if decision.Action == "OPEN_SHORT" && decision.StopLoss <= currentPrice {
			result.Approved = false
			result.Reason = "Stop loss for short position must be above current price"
			rc.logger.Warn(result.Reason)
			return result, nil
		}
	}

	// Check 9: Position holding time limits
	if position != nil && position.Size > 0 {
		maxHoldingTime := rc.getMaxHoldingTime(decision.ExpectedHoldingPeriod)
		if position.HoldingTime > maxHoldingTime {
			// Force close if holding too long
			if decision.Action != "CLOSE_POSITION" {
				rc.logger.WithFields(logrus.Fields{
					"holding_time": position.HoldingTime,
					"max_time":     maxHoldingTime,
				}).Warn("Position held too long, should consider closing")
			}
		}
	}

	rc.logger.WithFields(logrus.Fields{
		"approved":      result.Approved,
		"adjusted_size": result.AdjustedSize,
		"reason":        result.Reason,
	}).Info("Risk check completed")

	return result, nil
}

// UpdatePnL updates the daily PnL tracking
func (rc *Controller) UpdatePnL(pnl float64) {
	rc.resetDailyPnLIfNeeded()
	rc.dailyPnL += pnl

	rc.logger.WithFields(logrus.Fields{
		"trade_pnl": pnl,
		"daily_pnl": rc.dailyPnL,
	}).Info("PnL updated")
}

// GetDailyPnL returns current daily PnL
func (rc *Controller) GetDailyPnL() float64 {
	rc.resetDailyPnLIfNeeded()
	return rc.dailyPnL
}

// resetDailyPnLIfNeeded resets daily PnL if a new day has started
func (rc *Controller) resetDailyPnLIfNeeded() {
	now := time.Now()
	if now.Day() != rc.dailyPnLReset.Day() || now.Month() != rc.dailyPnLReset.Month() || now.Year() != rc.dailyPnLReset.Year() {
		rc.logger.WithField("previous_daily_pnl", rc.dailyPnL).Info("Resetting daily PnL")
		rc.dailyPnL = 0
		rc.dailyPnLReset = now
	}
}

// getMaxHoldingTime returns maximum holding time based on expected period
func (rc *Controller) getMaxHoldingTime(expectedPeriod string) time.Duration {
	switch expectedPeriod {
	case "SHORT":
		return 4 * time.Hour
	case "MEDIUM":
		return 24 * time.Hour
	case "LONG":
		return 7 * 24 * time.Hour
	default:
		return 24 * time.Hour
	}
}

// CheckStopLoss checks if position should be closed due to stop loss
func (rc *Controller) CheckStopLoss(position *hyperliquid.Position, currentPrice float64, stopLoss float64) bool {
	if position.Size == 0 || stopLoss <= 0 {
		return false
	}

	if position.Side == "LONG" && currentPrice <= stopLoss {
		rc.logger.WithFields(logrus.Fields{
			"current_price": currentPrice,
			"stop_loss":     stopLoss,
			"side":          "LONG",
		}).Warn("Stop loss triggered")
		return true
	}

	if position.Side == "SHORT" && currentPrice >= stopLoss {
		rc.logger.WithFields(logrus.Fields{
			"current_price": currentPrice,
			"stop_loss":     stopLoss,
			"side":          "SHORT",
		}).Warn("Stop loss triggered")
		return true
	}

	return false
}

// CheckTakeProfit checks if position should be closed due to take profit
func (rc *Controller) CheckTakeProfit(position *hyperliquid.Position, currentPrice float64, takeProfit float64) bool {
	if position.Size == 0 || takeProfit <= 0 {
		return false
	}

	if position.Side == "LONG" && currentPrice >= takeProfit {
		rc.logger.WithFields(logrus.Fields{
			"current_price": currentPrice,
			"take_profit":   takeProfit,
			"side":          "LONG",
		}).Info("Take profit triggered")
		return true
	}

	if position.Side == "SHORT" && currentPrice <= takeProfit {
		rc.logger.WithFields(logrus.Fields{
			"current_price": currentPrice,
			"take_profit":   takeProfit,
			"side":          "SHORT",
		}).Info("Take profit triggered")
		return true
	}

	return false
}
