package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Trading     TradingConfig     `yaml:"trading"`
	Risk        RiskConfig        `yaml:"risk"`
	AI          AIConfig          `yaml:"ai"`
	Hyperliquid HyperliquidConfig `yaml:"hyperliquid"`
	Monitoring  MonitoringConfig  `yaml:"monitoring"`
	System      SystemConfig      `yaml:"system"`
}

type TradingConfig struct {
	Symbols           []string `yaml:"symbols"`
	Timeframe         string   `yaml:"timeframe"`
	Interval          string   `yaml:"interval"`
	MaxPositionSize   float64  `yaml:"max_position_size"`
	MinConfidence     float64  `yaml:"min_confidence"`
	TradingEnabled    bool     `yaml:"trading_enabled"`
	MaxOpenPositions  int      `yaml:"max_open_positions"`
	MaxLeverage       int      `yaml:"max_leverage"`
}

type RiskConfig struct {
	MaxDrawdown         float64 `yaml:"max_drawdown"`
	DailyLossLimit      float64 `yaml:"daily_loss_limit"`
	PositionRiskPerTrade float64 `yaml:"position_risk_per_trade"`
	MaxTotalExposure    float64 `yaml:"max_total_exposure"`
	CorrelationLimit    float64 `yaml:"correlation_limit"`
	MinRiskRewardRatio  float64 `yaml:"min_risk_reward_ratio"`
}

type AIConfig struct {
	Provider    string     `yaml:"provider"`
	APIKey      string     `yaml:"api_key"`
	BaseURL     string     `yaml:"base_url"`
	Model       string     `yaml:"model"`
	Temperature float64    `yaml:"temperature"`
	MaxTokens   int        `yaml:"max_tokens"`
	Timeout     int        `yaml:"timeout"`
	Qwen        QwenConfig `yaml:"qwen"`
}

type QwenConfig struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
}

type HyperliquidConfig struct {
	APIURL         string `yaml:"api_url"`
	PrivateKey     string `yaml:"private_key"`
	AccountAddress string `yaml:"account_address"`
	Testnet        bool   `yaml:"testnet"`
}

type MonitoringConfig struct {
	LogLevel            string `yaml:"log_level"`
	LogFile             string `yaml:"log_file"`
	PerformanceTracking bool   `yaml:"performance_tracking"`
	AlertOnError        bool   `yaml:"alert_on_error"`
}

type SystemConfig struct {
	MaxRetries          int `yaml:"max_retries"`
	RetryDelay          int `yaml:"retry_delay"`
	HealthCheckInterval int `yaml:"health_check_interval"`
}

func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	content := os.ExpandEnv(string(data))

	var config Config
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Expand environment variables in fields that might contain them
	config.AI.APIKey = expandEnv(config.AI.APIKey)
	config.Hyperliquid.PrivateKey = expandEnv(config.Hyperliquid.PrivateKey)
	config.Hyperliquid.AccountAddress = expandEnv(config.Hyperliquid.AccountAddress)

	return &config, nil
}

func expandEnv(s string) string {
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		envVar := s[2 : len(s)-1]
		return os.Getenv(envVar)
	}
	return s
}
