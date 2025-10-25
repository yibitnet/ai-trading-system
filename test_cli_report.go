// +build ignore

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"aitrading/ai"
	"aitrading/config"
	"aitrading/hyperliquid"
	"aitrading/indicators"

	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("=== Testing CLI Decision Report Display ===\n")

	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize components
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetOutput(os.Stderr) // Send logs to stderr to keep stdout clean

	hlClient := hyperliquid.NewClient(cfg.Hyperliquid.APIURL)

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

	calc := indicators.NewCalculator()

	// Test with first symbol
	symbol := cfg.Trading.Symbols[0]
	fmt.Printf("Fetching data for %s...\n\n", symbol)

	// Fetch market data
	marketInfo, err := hlClient.GetMarketData(symbol)
	if err != nil {
		fmt.Printf("Failed to fetch market data: %v\n", err)
		os.Exit(1)
	}

	// Fetch candlestick data
	candles, err := hlClient.GetCandlestickData(symbol, cfg.Trading.Timeframe, 150)
	if err != nil {
		fmt.Printf("Failed to fetch candle data: %v\n", err)
		os.Exit(1)
	}

	// Calculate indicators
	inds := calc.Calculate(candles)

	// Get position
	position, err := hlClient.GetPosition(symbol, cfg.Hyperliquid.AccountAddress)
	if err != nil {
		fmt.Printf("Failed to fetch position: %v\n", err)
		os.Exit(1)
	}

	// Get AI decision
	analysis := &ai.MarketAnalysis{
		Symbol:     symbol,
		Timestamp:  time.Now(),
		Market:     marketInfo,
		Indicators: inds,
		Position:   position,
	}

	decision, err := aiDecision.Analyze(analysis)
	if err != nil {
		fmt.Printf("AI decision failed: %v\n", err)
		os.Exit(1)
	}

	// Now print the report (this simulates what the main bot does)
	printDecisionReport(symbol, marketInfo, inds, position, decision)

	fmt.Println("✅ CLI Report display test completed!")
}

// This is a copy of the printDecisionReport function from main.go
func printDecisionReport(symbol string, market *hyperliquid.MarketInfo, indicators *indicators.TechnicalIndicators, position *hyperliquid.Position, decision *ai.Decision) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("  📊 AI Trading Decision Report - %s @ %s\n", symbol, time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 80))

	// Market Overview
	fmt.Println("\n📈 Market Overview:")
	fmt.Printf("  Price:        $%.2f\n", market.CurrentPrice)
	priceChangeColor := "🔴"
	if market.PriceChange > 0 {
		priceChangeColor = "🟢"
	}
	fmt.Printf("  24h Change:   %s %.2f%%\n", priceChangeColor, market.PriceChange)
	fmt.Printf("  24h Volume:   $%.2f\n", market.Volume24h)

	// Technical Indicators
	fmt.Println("\n📐 Technical Indicators:")
	fmt.Printf("  Trend:        %s\n", formatTrendIndicator(indicators.TrendStrength))
	fmt.Printf("  Momentum:     %s\n", formatMomentumIndicator(indicators.MomentumStatus))
	fmt.Printf("  RSI(14):      %.2f %s\n", indicators.RSI14, getRSIStatus(indicators.RSI14))
	fmt.Printf("  MACD:         DIF=%.4f, DEA=%.4f, HIST=%.4f\n", indicators.MACDDIF, indicators.MACDDEA, indicators.MACDHIST)
	fmt.Printf("  Bollinger:    %.2f / %.2f / %.2f (%s)\n",
		indicators.BBUpper, indicators.BBMiddle, indicators.BBLower, indicators.BBPosition)
	fmt.Printf("  Volume:       %s\n", formatVolumeStatus(indicators.VolumePriceRelation))

	// Current Position
	fmt.Println("\n💼 Current Position:")
	if position.Size > 0 {
		pnlColor := "🔴"
		if position.PnLPercent > 0 {
			pnlColor = "🟢"
		}
		fmt.Printf("  Side:         %s\n", position.Side)
		fmt.Printf("  Size:         %.4f\n", position.Size)
		fmt.Printf("  Entry Price:  $%.2f\n", position.EntryPrice)
		fmt.Printf("  Current P&L:  %s %.2f%%\n", pnlColor, position.PnLPercent)
		fmt.Printf("  Holding Time: %s\n", position.HoldingTime.String())
	} else {
		fmt.Println("  Status:       No open position")
	}

	// AI Decision
	fmt.Println("\n🤖 AI Decision:")
	fmt.Printf("  Action:       %s\n", formatAction(decision.Action))
	fmt.Printf("  Confidence:   %.0f%% %s\n", decision.Confidence*100, getConfidenceBar(decision.Confidence))
	if decision.Action != "HOLD" && decision.Action != "CLOSE_POSITION" {
		fmt.Printf("  Position Size: %.1f%%\n", decision.Size*100)
		fmt.Printf("  Leverage:     %dx\n", decision.Leverage)
		fmt.Printf("  Stop Loss:    $%.2f (%.2f%%)\n", decision.StopLoss, (decision.StopLoss-market.CurrentPrice)/market.CurrentPrice*100)
		fmt.Printf("  Take Profit:  $%.2f (%.2f%%)\n", decision.TakeProfit, (decision.TakeProfit-market.CurrentPrice)/market.CurrentPrice*100)
		fmt.Printf("  Risk Level:   %s\n", formatRiskLevel(decision.RiskLevel))
		fmt.Printf("  Hold Period:  %s\n", decision.ExpectedHoldingPeriod)
	}

	// Reasoning
	fmt.Println("\n💭 Analysis Reasoning:")
	reasonLines := wrapText(decision.Reason, 76)
	for _, line := range reasonLines {
		fmt.Printf("  %s\n", line)
	}

	fmt.Println(strings.Repeat("=", 80) + "\n")
}

func formatAction(action string) string {
	switch action {
	case "OPEN_LONG":
		return "🟢 OPEN LONG"
	case "OPEN_SHORT":
		return "🔴 OPEN SHORT"
	case "ADD_POSITION":
		return "➕ ADD POSITION"
	case "CLOSE_POSITION":
		return "❌ CLOSE POSITION"
	case "HOLD":
		return "⏸️  HOLD"
	default:
		return action
	}
}

func formatTrendIndicator(trend string) string {
	switch trend {
	case "BULLISH":
		return "🟢 BULLISH (上涨趋势)"
	case "BEARISH":
		return "🔴 BEARISH (下跌趋势)"
	case "NEUTRAL":
		return "⚪ NEUTRAL (震荡)"
	default:
		return trend
	}
}

func formatMomentumIndicator(momentum string) string {
	switch momentum {
	case "BULLISH":
		return "🚀 BULLISH (强势上涨)"
	case "BEARISH":
		return "📉 BEARISH (弱势下跌)"
	case "NEUTRAL":
		return "➡️  NEUTRAL (动能中性)"
	default:
		return momentum
	}
}

func formatVolumeStatus(volumeStatus string) string {
	switch volumeStatus {
	case "STRONG_BUYING":
		return "🟢 STRONG BUYING (强力买入)"
	case "STRONG_SELLING":
		return "🔴 STRONG SELLING (强力卖出)"
	case "NORMAL":
		return "⚪ NORMAL (正常)"
	default:
		return volumeStatus
	}
}

func formatRiskLevel(risk string) string {
	switch risk {
	case "LOW":
		return "🟢 LOW (低风险)"
	case "MEDIUM":
		return "🟡 MEDIUM (中风险)"
	case "HIGH":
		return "🔴 HIGH (高风险)"
	default:
		return risk
	}
}

func getRSIStatus(rsi float64) string {
	if rsi > 70 {
		return "(超买)"
	} else if rsi < 30 {
		return "(超卖)"
	}
	return "(正常)"
}

func getConfidenceBar(confidence float64) string {
	barLength := int(confidence * 20)
	bar := strings.Repeat("█", barLength) + strings.Repeat("░", 20-barLength)
	return "[" + bar + "]"
}

func wrapText(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	currentLine := ""

	for _, word := range words {
		wordWidth := 0
		for _, r := range word {
			if r > 127 {
				wordWidth += 2
			} else {
				wordWidth += 1
			}
		}

		lineWidth := 0
		for _, r := range currentLine {
			if r > 127 {
				lineWidth += 2
			} else {
				lineWidth += 1
			}
		}

		if currentLine == "" {
			currentLine = word
		} else if lineWidth+wordWidth+1 <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}
