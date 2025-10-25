package main

import (
	"testing"
	"time"

	"aitrading/config"
	"aitrading/indicators"
)

func TestMainConfigLoad(t *testing.T) {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Trading.Symbol == "" {
		t.Error("Trading symbol should not be empty")
	}
}

func TestTradingBotCreation(t *testing.T) {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	bot, err := NewTradingBot(cfg)
	if err != nil {
		t.Fatalf("Failed to create trading bot: %v", err)
	}

	if bot == nil {
		t.Fatal("Trading bot should not be nil")
	}

	if bot.config == nil {
		t.Error("Bot config should not be nil")
	}
	if bot.logger == nil {
		t.Error("Bot logger should not be nil")
	}
	if bot.calculator == nil {
		t.Error("Bot calculator should not be nil")
	}
}

func TestIntervalToCron(t *testing.T) {
	cfg, _ := config.Load("config.yaml")
	bot, _ := NewTradingBot(cfg)

	tests := []struct {
		interval string
		expected string
	}{
		{"1m", "* * * * *"},
		{"5m", "*/5 * * * *"},
		{"15m", "*/15 * * * *"},
		{"1h", "0 * * * *"},
		{"unknown", "*/5 * * * *"}, // default
	}

	for _, tt := range tests {
		result := bot.intervalToCron(tt.interval)
		if result != tt.expected {
			t.Errorf("intervalToCron(%s) = %s, expected %s", tt.interval, result, tt.expected)
		}
	}
}

func TestIndicatorCalculationPipeline(t *testing.T) {
	// Create realistic test data
	data := make([]indicators.MarketData, 150)
	basePrice := 2000.0

	for i := range data {
		// Simulate price movement with some volatility
		variation := float64(i%10 - 5) * 2
		data[i] = indicators.MarketData{
			Timestamp: time.Now().Add(-time.Duration(150-i) * 5 * time.Minute).Unix(),
			Open:      basePrice + variation,
			High:      basePrice + variation + 10,
			Low:       basePrice + variation - 10,
			Close:     basePrice + variation + float64(i)*0.1,
			Volume:    1000.0 + float64(i)*10,
		}
	}

	calc := indicators.NewCalculator()
	result := calc.Calculate(data)

	if result == nil {
		t.Fatal("Indicator calculation should not return nil")
	}

	// Verify all indicators are calculated
	checks := []struct {
		name  string
		value float64
		min   float64
		max   float64
	}{
		{"SMA10", result.SMA10, 100, 5000},
		{"SMA60", result.SMA60, 100, 5000},
		{"EMA10", result.EMA10, 100, 5000},
		{"RSI14", result.RSI14, 0, 100},
		{"BBUpper", result.BBUpper, 100, 5000},
		{"BBMiddle", result.BBMiddle, 100, 5000},
		{"BBLower", result.BBLower, 100, 5000},
		{"VMA20", result.VMA20, 0, 100000},
	}

	for _, check := range checks {
		if check.value < check.min || check.value > check.max {
			t.Errorf("%s out of range: %f (expected %f-%f)", check.name, check.value, check.min, check.max)
		}
	}

	// Check derived analysis
	validTrends := map[string]bool{
		"STRONG_BULLISH": true,
		"BULLISH":        true,
		"NEUTRAL":        true,
		"BEARISH":        true,
		"STRONG_BEARISH": true,
	}

	if !validTrends[result.TrendStrength] {
		t.Errorf("Invalid trend strength: %s", result.TrendStrength)
	}

	validMomentum := map[string]bool{
		"OVERBOUGHT": true,
		"OVERSOLD":   true,
		"BULLISH":    true,
		"BEARISH":    true,
		"NEUTRAL":    true,
	}

	if !validMomentum[result.MomentumStatus] {
		t.Errorf("Invalid momentum status: %s", result.MomentumStatus)
	}
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warning"},
		{"error", "error"},
		{"unknown", "info"}, // default
	}

	for _, tt := range tests {
		level := getLogLevel(tt.input)
		if level.String() != tt.expected {
			t.Errorf("getLogLevel(%s) = %s, expected %s", tt.input, level.String(), tt.expected)
		}
	}
}
