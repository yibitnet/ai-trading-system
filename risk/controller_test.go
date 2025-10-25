package risk

import (
	"testing"
	"time"

	"aitrading/ai"
	"aitrading/config"
	"aitrading/hyperliquid"
	"github.com/sirupsen/logrus"
)

func TestNewController(t *testing.T) {
	cfg := &config.RiskConfig{
		MaxDrawdown:         0.05,
		DailyLossLimit:      0.02,
		PositionRiskPerTrade: 0.01,
		MaxTotalExposure:    0.25,
		MinRiskRewardRatio:  2.0,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Quiet for tests

	rc := NewController(cfg, logger)
	if rc == nil {
		t.Fatal("NewController should not return nil")
	}
}

func TestCheckDecision_LowConfidence(t *testing.T) {
	cfg := &config.RiskConfig{
		MaxDrawdown:         0.05,
		DailyLossLimit:      0.02,
		PositionRiskPerTrade: 0.01,
		MaxTotalExposure:    0.25,
		MinRiskRewardRatio:  2.0,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	rc := NewController(cfg, logger)

	decision := &ai.Decision{
		Action:     "OPEN_LONG",
		Confidence: 0.5, // Too low
		Size:       0.1,
		StopLoss:   1800,
		TakeProfit: 2200,
	}

	position := &hyperliquid.Position{
		Symbol: "ETH",
		Side:   "NONE",
		Size:   0,
	}

	result, err := rc.CheckDecision(decision, 2000.0, 10000.0, position)
	if err != nil {
		t.Fatalf("CheckDecision should not error: %v", err)
	}

	if result.Approved {
		t.Error("Decision should be rejected due to low confidence")
	}
}

func TestCheckDecision_ValidLongPosition(t *testing.T) {
	cfg := &config.RiskConfig{
		MaxDrawdown:         0.05,
		DailyLossLimit:      0.02,
		PositionRiskPerTrade: 0.01,
		MaxTotalExposure:    0.25,
		MinRiskRewardRatio:  2.0,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	rc := NewController(cfg, logger)

	decision := &ai.Decision{
		Action:     "OPEN_LONG",
		Confidence: 0.8,
		Size:       0.05,
		StopLoss:   1900, // Risk: 100
		TakeProfit: 2200, // Reward: 200, R/R = 2:1
	}

	position := &hyperliquid.Position{
		Symbol: "ETH",
		Side:   "NONE",
		Size:   0,
	}

	result, err := rc.CheckDecision(decision, 2000.0, 10000.0, position)
	if err != nil {
		t.Fatalf("CheckDecision should not error: %v", err)
	}

	if !result.Approved {
		t.Errorf("Valid decision should be approved: %s", result.Reason)
	}
}

func TestCheckDecision_InvalidStopLoss(t *testing.T) {
	cfg := &config.RiskConfig{
		MaxDrawdown:         0.05,
		DailyLossLimit:      0.02,
		PositionRiskPerTrade: 0.01,
		MaxTotalExposure:    0.25,
		MinRiskRewardRatio:  2.0,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	rc := NewController(cfg, logger)

	decision := &ai.Decision{
		Action:     "OPEN_LONG",
		Confidence: 0.8,
		Size:       0.05,
		StopLoss:   2100, // Above current price - invalid!
		TakeProfit: 2200,
	}

	position := &hyperliquid.Position{
		Symbol: "ETH",
		Side:   "NONE",
		Size:   0,
	}

	result, err := rc.CheckDecision(decision, 2000.0, 10000.0, position)
	if err != nil {
		t.Fatalf("CheckDecision should not error: %v", err)
	}

	if result.Approved {
		t.Error("Decision should be rejected due to invalid stop loss")
	}
}

func TestCheckStopLoss(t *testing.T) {
	cfg := &config.RiskConfig{
		MaxDrawdown:         0.05,
		DailyLossLimit:      0.02,
		PositionRiskPerTrade: 0.01,
		MaxTotalExposure:    0.25,
		MinRiskRewardRatio:  2.0,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	rc := NewController(cfg, logger)

	// Test long position stop loss
	position := &hyperliquid.Position{
		Symbol: "ETH",
		Side:   "LONG",
		Size:   1.0,
	}

	// Price below stop loss - should trigger
	if !rc.CheckStopLoss(position, 1800, 1900) {
		t.Error("Stop loss should trigger when price drops below stop loss")
	}

	// Price above stop loss - should not trigger
	if rc.CheckStopLoss(position, 2000, 1900) {
		t.Error("Stop loss should not trigger when price is above stop loss")
	}

	// Test short position stop loss
	position.Side = "SHORT"

	// Price above stop loss - should trigger
	if !rc.CheckStopLoss(position, 2100, 2000) {
		t.Error("Stop loss should trigger when price rises above stop loss")
	}

	// Price below stop loss - should not trigger
	if rc.CheckStopLoss(position, 1900, 2000) {
		t.Error("Stop loss should not trigger when price is below stop loss")
	}
}

func TestCheckTakeProfit(t *testing.T) {
	cfg := &config.RiskConfig{
		MaxDrawdown:         0.05,
		DailyLossLimit:      0.02,
		PositionRiskPerTrade: 0.01,
		MaxTotalExposure:    0.25,
		MinRiskRewardRatio:  2.0,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	rc := NewController(cfg, logger)

	// Test long position take profit
	position := &hyperliquid.Position{
		Symbol: "ETH",
		Side:   "LONG",
		Size:   1.0,
	}

	// Price above take profit - should trigger
	if !rc.CheckTakeProfit(position, 2200, 2100) {
		t.Error("Take profit should trigger when price rises above target")
	}

	// Price below take profit - should not trigger
	if rc.CheckTakeProfit(position, 2000, 2100) {
		t.Error("Take profit should not trigger when price is below target")
	}

	// Test short position take profit
	position.Side = "SHORT"

	// Price below take profit - should trigger
	if !rc.CheckTakeProfit(position, 1800, 1900) {
		t.Error("Take profit should trigger when price drops below target")
	}

	// Price above take profit - should not trigger
	if rc.CheckTakeProfit(position, 2000, 1900) {
		t.Error("Take profit should not trigger when price is above target")
	}
}

func TestUpdatePnL(t *testing.T) {
	cfg := &config.RiskConfig{
		MaxDrawdown:         0.05,
		DailyLossLimit:      0.02,
		PositionRiskPerTrade: 0.01,
		MaxTotalExposure:    0.25,
		MinRiskRewardRatio:  2.0,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	rc := NewController(cfg, logger)

	// Update with profit
	rc.UpdatePnL(100)
	if rc.GetDailyPnL() != 100 {
		t.Errorf("Daily PnL should be 100, got %f", rc.GetDailyPnL())
	}

	// Update with loss
	rc.UpdatePnL(-50)
	if rc.GetDailyPnL() != 50 {
		t.Errorf("Daily PnL should be 50, got %f", rc.GetDailyPnL())
	}
}

func TestDailyPnLReset(t *testing.T) {
	cfg := &config.RiskConfig{
		MaxDrawdown:         0.05,
		DailyLossLimit:      0.02,
		PositionRiskPerTrade: 0.01,
		MaxTotalExposure:    0.25,
		MinRiskRewardRatio:  2.0,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	rc := NewController(cfg, logger)

	// Set some PnL
	rc.UpdatePnL(100)

	// Force reset by setting reset time to yesterday
	rc.dailyPnLReset = time.Now().AddDate(0, 0, -1)

	// Get daily PnL should trigger reset
	pnl := rc.GetDailyPnL()
	if pnl != 0 {
		t.Errorf("Daily PnL should be reset to 0, got %f", pnl)
	}
}
