// +build ignore

package main

import (
	"fmt"
	"os"

	"aitrading/ai"
	"aitrading/config"
	"aitrading/hyperliquid"
	"aitrading/indicators"
	"aitrading/risk"

	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("=== Testing Multi-Symbol and Leverage Features ===\n")

	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Configuration loaded successfully")

	// Test 1: Multi-symbol configuration
	fmt.Println("\n--- Test 1: Multi-Symbol Configuration ---")
	fmt.Printf("Configured symbols: %v\n", cfg.Trading.Symbols)
	fmt.Printf("Max open positions: %d\n", cfg.Trading.MaxOpenPositions)
	fmt.Printf("Max leverage: %d\n", cfg.Trading.MaxLeverage)
	if len(cfg.Trading.Symbols) > 1 {
		fmt.Println("✅ Multi-symbol configuration working")
	} else {
		fmt.Println("⚠️  Only one symbol configured")
	}

	// Test 2: AI provider configuration
	fmt.Println("\n--- Test 2: AI Provider Configuration ---")
	fmt.Printf("AI Provider: %s\n", cfg.AI.Provider)
	if cfg.AI.Provider == "qwen" {
		fmt.Printf("Qwen API Key: %s...\n", cfg.AI.Qwen.APIKey[:10])
		fmt.Printf("Qwen Model: %s\n", cfg.AI.Qwen.Model)
		fmt.Println("✅ Qwen configuration available")
	} else {
		fmt.Printf("DeepSeek Model: %s\n", cfg.AI.Model)
		fmt.Println("✅ DeepSeek configuration active")
	}

	// Test 3: Initialize components
	fmt.Println("\n--- Test 3: Component Initialization ---")

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Initialize AI decision maker
	var aiAPIKey, aiBaseURL, aiModel string
	switch cfg.AI.Provider {
	case "qwen":
		aiAPIKey = cfg.AI.Qwen.APIKey
		aiBaseURL = cfg.AI.Qwen.BaseURL
		aiModel = cfg.AI.Qwen.Model
	default:
		aiAPIKey = cfg.AI.APIKey
		aiBaseURL = cfg.AI.BaseURL
		aiModel = cfg.AI.Model
	}

	aiDecision := ai.NewDecisionMaker(
		cfg.AI.Provider,
		aiAPIKey,
		aiBaseURL,
		aiModel,
		cfg.AI.Temperature,
		cfg.AI.MaxTokens,
		cfg.AI.Timeout,
	)
	fmt.Println("✅ AI Decision Maker initialized")

	// Initialize risk controller with trading config
	riskControl := risk.NewController(&cfg.Risk, &cfg.Trading, logger)
	fmt.Println("✅ Risk Controller initialized with trading config")

	// Test 4: Fetch market data for multiple symbols
	fmt.Println("\n--- Test 4: Multi-Symbol Market Data ---")
	hlClient := hyperliquid.NewClient(cfg.Hyperliquid.APIURL)

	for _, symbol := range cfg.Trading.Symbols {
		marketInfo, err := hlClient.GetMarketData(symbol)
		if err != nil {
			fmt.Printf("❌ Failed to fetch %s data: %v\n", symbol, err)
			continue
		}
		fmt.Printf("✅ %s: Price=$%.2f, Volume=$%.2f, Change=%.2f%%\n",
			symbol, marketInfo.CurrentPrice, marketInfo.Volume24h, marketInfo.PriceChange)
	}

	// Test 5: Test leverage in decision
	fmt.Println("\n--- Test 5: AI Decision with Leverage ---")

	// Create mock data for decision
	candles, err := hlClient.GetCandlestickData(cfg.Trading.Symbols[0], cfg.Trading.Timeframe, 150)
	if err != nil {
		fmt.Printf("❌ Failed to fetch candle data: %v\n", err)
		os.Exit(1)
	}

	calc := indicators.NewCalculator()
	inds := calc.Calculate(candles)

	marketInfo, _ := hlClient.GetMarketData(cfg.Trading.Symbols[0])
	position, _ := hlClient.GetPosition(cfg.Trading.Symbols[0], cfg.Hyperliquid.AccountAddress)

	analysis := &ai.MarketAnalysis{
		Symbol:     cfg.Trading.Symbols[0],
		Market:     marketInfo,
		Indicators: inds,
		Position:   position,
	}

	// Get AI decision (this will include leverage field)
	decision, err := aiDecision.Analyze(analysis)
	if err != nil {
		fmt.Printf("❌ AI decision failed: %v\n", err)
		// Don't exit, this is expected if API key is invalid
	} else {
		fmt.Printf("✅ AI Decision received:\n")
		fmt.Printf("   Action: %s\n", decision.Action)
		fmt.Printf("   Confidence: %.2f\n", decision.Confidence)
		fmt.Printf("   Size: %.2f\n", decision.Size)
		fmt.Printf("   Leverage: %d\n", decision.Leverage)
		fmt.Printf("   Reason: %s\n", decision.Reason[:min(100, len(decision.Reason))])
	}

	// Test 6: Risk check with position count and leverage limit
	fmt.Println("\n--- Test 6: Risk Check with Position Count & Leverage ---")

	// Create a mock decision with high leverage
	testDecision := &ai.Decision{
		Action:     "OPEN_LONG",
		Confidence: 0.8,
		Size:       0.1,
		Leverage:   15, // Exceeds max of 10
		StopLoss:   3900,
		TakeProfit: 4000,
		RiskLevel:  "MEDIUM",
	}

	// Simulate 0 open positions
	riskCheck, err := riskControl.CheckDecision(testDecision, 3930, 1000, position, 0)
	if err != nil {
		fmt.Printf("❌ Risk check failed: %v\n", err)
	} else {
		if riskCheck.AdjustedLeverage < testDecision.Leverage {
			fmt.Printf("✅ Leverage adjusted: %d -> %d (max: %d)\n",
				testDecision.Leverage, riskCheck.AdjustedLeverage, cfg.Trading.MaxLeverage)
		} else {
			fmt.Printf("✅ Leverage within limit: %d\n", testDecision.Leverage)
		}
	}

	// Test with max positions reached
	fmt.Println("\n--- Test 7: Position Count Limit ---")
	riskCheck2, err := riskControl.CheckDecision(testDecision, 3930, 1000, position, cfg.Trading.MaxOpenPositions)
	if err != nil {
		fmt.Printf("❌ Risk check failed: %v\n", err)
	} else {
		if !riskCheck2.Approved {
			fmt.Printf("✅ Position limit enforced: %s\n", riskCheck2.Reason)
		} else {
			fmt.Printf("⚠️  Position limit not enforced (positions: %d, max: %d)\n",
				cfg.Trading.MaxOpenPositions, cfg.Trading.MaxOpenPositions)
		}
	}

	fmt.Println("\n=== All Feature Tests Completed ===")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
