package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test loading the actual config file
	cfg, err := Load("../config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify trading config
	if cfg.Trading.Symbol == "" {
		t.Error("Symbol should not be empty")
	}
	if cfg.Trading.Interval == "" {
		t.Error("Interval should not be empty")
	}

	// Verify risk config
	if cfg.Risk.MaxDrawdown <= 0 || cfg.Risk.MaxDrawdown > 1 {
		t.Errorf("MaxDrawdown should be between 0-1, got %f", cfg.Risk.MaxDrawdown)
	}

	// Verify AI config
	if cfg.AI.Model == "" {
		t.Error("AI model should not be empty")
	}
	if cfg.AI.Temperature < 0 || cfg.AI.Temperature > 1 {
		t.Errorf("Temperature should be between 0-1, got %f", cfg.AI.Temperature)
	}

	// Verify Hyperliquid config
	if cfg.Hyperliquid.APIURL == "" {
		t.Error("Hyperliquid API URL should not be empty")
	}
}

func TestLoadTestConfig(t *testing.T) {
	cfg, err := Load("../config.test.yaml")
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// Test config should have trading disabled
	if cfg.Trading.TradingEnabled {
		t.Error("Test config should have trading disabled")
	}

	// Test config should be more conservative
	if cfg.Trading.MaxPositionSize > 0.1 {
		t.Errorf("Test config should have conservative position size, got %f", cfg.Trading.MaxPositionSize)
	}
}

func TestEnvironmentVariableExpansion(t *testing.T) {
	// Set test environment variables
	os.Setenv("TEST_API_KEY", "test_key_123")
	defer os.Unsetenv("TEST_API_KEY")

	// Create a temporary config file
	configContent := `
trading:
  symbol: "BTC"
  interval: "5m"
  timeframe: "5m"
  max_position_size: 0.1
  min_confidence: 0.7
  trading_enabled: false

risk:
  max_drawdown: 0.05
  daily_loss_limit: 0.02
  position_risk_per_trade: 0.01
  max_total_exposure: 0.25
  correlation_limit: 0.8
  min_risk_reward_ratio: 2.0

ai:
  provider: "deepseek"
  api_key: "${TEST_API_KEY}"
  base_url: "https://api.test.com"
  model: "test-model"
  temperature: 0.1
  max_tokens: 1000
  timeout: 30

hyperliquid:
  api_url: "https://api.test.com"
  private_key: "test_private_key"
  account_address: "test_address"
  testnet: true

monitoring:
  log_level: "debug"
  log_file: "test.log"
  performance_tracking: true
  alert_on_error: true

system:
  max_retries: 3
  retry_delay: 5
  health_check_interval: 60
`

	tmpfile, err := os.CreateTemp("", "config_test_*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.AI.APIKey != "test_key_123" {
		t.Errorf("Environment variable expansion failed: got %s, expected test_key_123", cfg.AI.APIKey)
	}
}
