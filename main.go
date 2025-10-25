package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"aitrading/ai"
	"aitrading/config"
	"aitrading/executor"
	"aitrading/hyperliquid"
	"aitrading/indicators"
	"aitrading/risk"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// TradingBot represents the main trading bot
type TradingBot struct {
	config         *config.Config
	logger         *logrus.Logger
	hlClient       *hyperliquid.Client
	hlTrader       *hyperliquid.Trader
	aiDecision     *ai.DecisionMaker
	riskControl    *risk.Controller
	executor       *executor.Executor
	calculator     *indicators.Calculator
	scheduler      *cron.Cron
	lastStopLoss   float64
	lastTakeProfit float64
}

// NewTradingBot creates a new trading bot instance
func NewTradingBot(cfg *config.Config) (*TradingBot, error) {
	// Setup logger
	logger := logrus.New()
	logger.SetLevel(getLogLevel(cfg.Monitoring.LogLevel))
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Setup log file if specified
	if cfg.Monitoring.LogFile != "" {
		file, err := os.OpenFile(cfg.Monitoring.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			logger.SetOutput(file)
		} else {
			logger.Warnf("Failed to open log file: %v", err)
		}
	}

	// Initialize Hyperliquid client
	hlClient := hyperliquid.NewClient(cfg.Hyperliquid.APIURL)

	// Initialize Hyperliquid trader
	hlTrader, err := hyperliquid.NewTrader(hlClient, cfg.Hyperliquid.PrivateKey, cfg.Hyperliquid.AccountAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create trader: %w", err)
	}

	// Initialize AI decision maker
	var aiAPIKey, aiBaseURL, aiModel string
	switch cfg.AI.Provider {
	case "qwen":
		aiAPIKey = cfg.AI.Qwen.APIKey
		aiBaseURL = cfg.AI.Qwen.BaseURL
		aiModel = cfg.AI.Qwen.Model
	default: // "deepseek" or any other
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

	// Initialize risk controller
	riskControl := risk.NewController(&cfg.Risk, &cfg.Trading, logger)

	// Initialize executor
	exec := executor.NewExecutor(hlTrader, hlClient, cfg.Hyperliquid.AccountAddress, logger)

	// Initialize indicator calculator
	calc := indicators.NewCalculator()

	// Initialize scheduler
	scheduler := cron.New()

	bot := &TradingBot{
		config:      cfg,
		logger:      logger,
		hlClient:    hlClient,
		hlTrader:    hlTrader,
		aiDecision:  aiDecision,
		riskControl: riskControl,
		executor:    exec,
		calculator:  calc,
		scheduler:   scheduler,
	}

	return bot, nil
}

// Start starts the trading bot
func (bot *TradingBot) Start() error {
	bot.logger.Info("Starting AI Trading Bot...")

	// Add scheduled job based on interval
	cronExpr := bot.intervalToCron(bot.config.Trading.Interval)
	bot.logger.Infof("Scheduling trading cycle at: %s", cronExpr)

	_, err := bot.scheduler.AddFunc(cronExpr, func() {
		if err := bot.runTradingCycle(); err != nil {
			bot.logger.WithError(err).Error("Trading cycle failed")
		}
	})

	if err != nil {
		return fmt.Errorf("failed to schedule trading cycle: %w", err)
	}

	// Start scheduler
	bot.scheduler.Start()

	bot.logger.Info("Trading bot started successfully")

	// Run first cycle immediately
	go func() {
		time.Sleep(2 * time.Second)
		if err := bot.runTradingCycle(); err != nil {
			bot.logger.WithError(err).Error("Initial trading cycle failed")
		}
	}()

	return nil
}

// runTradingCycle executes one complete trading cycle
func (bot *TradingBot) runTradingCycle() error {
	bot.logger.Info("========== Starting Trading Cycle ==========")
	startTime := time.Now()

	// Iterate through all configured symbols
	for _, symbol := range bot.config.Trading.Symbols {
		if err := bot.runTradingCycleForSymbol(symbol); err != nil {
			bot.logger.WithError(err).WithField("symbol", symbol).Error("Trading cycle failed for symbol")
			// Continue with other symbols even if one fails
			continue
		}
	}

	elapsed := time.Since(startTime)
	bot.logger.WithField("elapsed", elapsed).Info("========== Trading Cycle Completed ==========")

	return nil
}

// runTradingCycleForSymbol executes trading cycle for a specific symbol
func (bot *TradingBot) runTradingCycleForSymbol(symbol string) error {
	bot.logger.WithField("symbol", symbol).Info("Processing symbol...")

	// Step 1: Fetch market data
	bot.logger.Info("Step 1: Fetching market data...")
	marketInfo, err := bot.hlClient.GetMarketData(symbol)
	if err != nil {
		return fmt.Errorf("failed to fetch market data: %w", err)
	}

	bot.logger.WithFields(logrus.Fields{
		"symbol": symbol,
		"price":  marketInfo.CurrentPrice,
		"volume": marketInfo.Volume24h,
	}).Info("Market data fetched")

	// Step 2: Fetch candlestick data for indicators
	bot.logger.Info("Step 2: Fetching candlestick data...")
	candles, err := bot.hlClient.GetCandlestickData(symbol, bot.config.Trading.Timeframe, 150)
	if err != nil {
		return fmt.Errorf("failed to fetch candlestick data: %w", err)
	}

	if len(candles) < 120 {
		return fmt.Errorf("insufficient candle data: got %d, need at least 120", len(candles))
	}

	bot.logger.Infof("Fetched %d candles", len(candles))

	// Step 3: Calculate technical indicators
	bot.logger.Info("Step 3: Calculating technical indicators...")
	indicators := bot.calculator.Calculate(candles)
	if indicators == nil {
		return fmt.Errorf("failed to calculate indicators")
	}

	bot.logger.WithFields(logrus.Fields{
		"trend":    indicators.TrendStrength,
		"momentum": indicators.MomentumStatus,
		"rsi":      indicators.RSI14,
	}).Info("Indicators calculated")

	// Step 4: Get current position
	bot.logger.Info("Step 4: Fetching current position...")
	position, err := bot.hlClient.GetPosition(symbol, bot.config.Hyperliquid.AccountAddress)
	if err != nil {
		return fmt.Errorf("failed to fetch position: %w", err)
	}

	bot.logger.WithFields(logrus.Fields{
		"side": position.Side,
		"size": position.Size,
		"pnl":  position.PnLPercent,
	}).Info("Position fetched")

	// Step 5: Check stop loss and take profit
	if position.Size > 0 {
		if bot.riskControl.CheckStopLoss(position, marketInfo.CurrentPrice, bot.lastStopLoss) {
			bot.logger.Warn("Stop loss triggered, closing position")
			decision := &ai.Decision{
				Action:     "CLOSE_POSITION",
				Confidence: 1.0,
				Reason:     "Stop loss triggered",
			}
			return bot.executeDecision(decision, marketInfo, position, symbol)
		}

		if bot.riskControl.CheckTakeProfit(position, marketInfo.CurrentPrice, bot.lastTakeProfit) {
			bot.logger.Info("Take profit triggered, closing position")
			decision := &ai.Decision{
				Action:     "CLOSE_POSITION",
				Confidence: 1.0,
				Reason:     "Take profit triggered",
			}
			return bot.executeDecision(decision, marketInfo, position, symbol)
		}
	}

	// Step 6: AI analysis
	bot.logger.Info("Step 6: Requesting AI decision...")
	analysis := &ai.MarketAnalysis{
		Symbol:     symbol,
		Timestamp:  time.Now(),
		Market:     marketInfo,
		Indicators: indicators,
		Position:   position,
	}

	decision, err := bot.aiDecision.Analyze(analysis)
	if err != nil {
		return fmt.Errorf("AI analysis failed: %w", err)
	}

	bot.logger.WithFields(logrus.Fields{
		"action":     decision.Action,
		"confidence": decision.Confidence,
		"reason":     decision.Reason,
	}).Info("AI decision received")

	// Print decision report to console
	bot.printDecisionReport(symbol, marketInfo, indicators, position, decision)

	// Step 7: Execute decision
	if err := bot.executeDecision(decision, marketInfo, position, symbol); err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	return nil
}

// executeDecision executes a trading decision with risk checks
func (bot *TradingBot) executeDecision(decision *ai.Decision, marketInfo *hyperliquid.MarketInfo, position *hyperliquid.Position, symbol string) error {
	// Get account balance
	balance, err := bot.hlClient.GetAccountBalance(bot.config.Hyperliquid.AccountAddress)
	if err != nil {
		return fmt.Errorf("failed to get account balance: %w", err)
	}

	bot.logger.WithField("balance", balance).Info("Account balance fetched")

	// Count open positions across all symbols
	openPositionCount := 0
	for _, sym := range bot.config.Trading.Symbols {
		pos, err := bot.hlClient.GetPosition(sym, bot.config.Hyperliquid.AccountAddress)
		if err == nil && pos.Size > 0 {
			openPositionCount++
		}
	}

	// Risk check
	bot.logger.Info("Performing risk checks...")
	riskCheck, err := bot.riskControl.CheckDecision(decision, marketInfo.CurrentPrice, balance, position, openPositionCount)
	if err != nil {
		return fmt.Errorf("risk check failed: %w", err)
	}

	if !riskCheck.Approved {
		bot.logger.WithField("reason", riskCheck.Reason).Warn("Decision rejected by risk control")
		return nil
	}

	// Adjust decision based on risk check
	decision.Size = riskCheck.AdjustedSize
	decision.Leverage = riskCheck.AdjustedLeverage

	// Check if trading is enabled
	if !bot.config.Trading.TradingEnabled {
		bot.logger.Warn("Trading is disabled - simulation mode")

		// Display simulated order details
		bot.printSimulatedOrder(decision, marketInfo, balance, symbol)

		bot.logger.WithFields(logrus.Fields{
			"action":   decision.Action,
			"size":     decision.Size,
			"leverage": decision.Leverage,
			"price":    marketInfo.CurrentPrice,
		}).Info("Simulated trade")
		return nil
	}

	// Execute trade
	result, err := bot.executor.Execute(symbol, decision, marketInfo.CurrentPrice, balance)
	if err != nil {
		bot.logger.WithError(err).Error("Trade execution failed")
		return err
	}

	// Update stop loss and take profit tracking
	if decision.Action == "OPEN_LONG" || decision.Action == "OPEN_SHORT" {
		bot.lastStopLoss = decision.StopLoss
		bot.lastTakeProfit = decision.TakeProfit
	}

	bot.logger.WithFields(logrus.Fields{
		"success":  result.Success,
		"action":   result.Action,
		"side":     result.Side,
		"size":     result.Size,
		"price":    result.Price,
		"order_id": result.OrderID,
		"message":  result.Message,
	}).Info("Trade executed")

	// Print execution result to console
	bot.printExecutionResult(result, symbol)

	return nil
}

// Stop stops the trading bot
func (bot *TradingBot) Stop() {
	bot.logger.Info("Stopping trading bot...")
	bot.scheduler.Stop()
	bot.logger.Info("Trading bot stopped")
}

// intervalToCron converts interval string to cron expression
func (bot *TradingBot) intervalToCron(interval string) string {
	switch interval {
	case "1m":
		return "* * * * *"
	case "5m":
		return "*/5 * * * *"
	case "15m":
		return "*/15 * * * *"
	case "1h":
		return "0 * * * *"
	case "4h":
		return "0 */4 * * *"
	default:
		return "*/5 * * * *" // default 5 minutes
	}
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

// printDecisionReport prints a formatted decision report to console
func (bot *TradingBot) printDecisionReport(symbol string, market *hyperliquid.MarketInfo, indicators *indicators.TechnicalIndicators, position *hyperliquid.Position, decision *ai.Decision) {
	// Determine color based on action
	actionColor := colorReset
	if decision.Action == "OPEN_LONG" || decision.Action == "OPEN_SHORT" || decision.Action == "ADD_POSITION" {
		actionColor = colorGreen
	} else if decision.Action == "CLOSE_POSITION" {
		actionColor = colorRed
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("%s  📊 AI Trading Decision Report - %s @ %s%s\n", actionColor, symbol, time.Now().Format("2006-01-02 15:04:05"), colorReset)
	fmt.Println(strings.Repeat("=", 80))

	// Market Overview - Compact to one line
	priceChangeEmoji := "🔴"
	if market.PriceChange > 0 {
		priceChangeEmoji = "🟢"
	}
	volumeStr := bot.formatLargeNumber(market.Volume24h)
	fmt.Printf("\n📈 Market: %s $%s | 24h: %s%.2f%% | Vol: $%s\n",
		symbol, bot.formatPrice(market.CurrentPrice), priceChangeEmoji, market.PriceChange, volumeStr)

	// Technical Indicators - Compress to 2-3 lines
	fmt.Printf("📐 Indicators: Trend:%s | Momentum:%s | RSI:%.1f%s\n",
		bot.formatTrendCompact(indicators.TrendStrength),
		bot.formatMomentumCompact(indicators.MomentumStatus),
		indicators.RSI14,
		bot.getRSIStatus(indicators.RSI14))

	macdSignal := "⚪"
	if indicators.MACDHIST > 0 {
		macdSignal = "🟢"
	} else if indicators.MACDHIST < 0 {
		macdSignal = "🔴"
	}
	fmt.Printf("              MACD:%s%.4f | BB:%s | Vol:%s\n",
		macdSignal,
		indicators.MACDHIST,
		indicators.BBPosition,
		bot.formatVolumeCompact(indicators.VolumePriceRelation))

	// Current Position - Compact
	if position.Size > 0 {
		pnlEmoji := "🔴"
		if position.PnLPercent > 0 {
			pnlEmoji = "🟢"
		}
		fmt.Printf("\n💼 Position: %s %.4f @ $%s | P&L:%s%.2f%% | Time:%s\n",
			position.Side, position.Size, bot.formatPrice(position.EntryPrice), pnlEmoji, position.PnLPercent, position.HoldingTime.String())
	} else {
		fmt.Println("\n💼 Position: None")
	}

	// AI Decision - Compact to one line for basic info
	actionStr := bot.formatAction(decision.Action)
	if decision.Action != "HOLD" && decision.Action != "CLOSE_POSITION" {
		fmt.Printf("\n%s🤖 Decision: %s | Conf:%.0f%% | Size:%.1f%% | Lev:%dx | SL:$%s | TP:$%s | Risk:%s%s\n",
			actionColor,
			actionStr,
			decision.Confidence*100,
			decision.Size*100,
			decision.Leverage,
			bot.formatPrice(decision.StopLoss),
			bot.formatPrice(decision.TakeProfit),
			bot.formatRiskLevelCompact(decision.RiskLevel),
			colorReset)
	} else {
		fmt.Printf("\n%s🤖 Decision: %s | Confidence:%.0f%%%s\n",
			actionColor,
			actionStr,
			decision.Confidence*100,
			colorReset)
	}

	// Reasoning - Keep but more compact
	fmt.Println("\n💭 Reasoning:")
	reasonLines := bot.wrapText(decision.Reason, 76)
	for _, line := range reasonLines {
		fmt.Printf("  %s\n", line)
	}

	fmt.Println(strings.Repeat("=", 80) + "\n")
}

// formatAction formats the action with emoji
func (bot *TradingBot) formatAction(action string) string {
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

// formatTrendIndicator formats trend with emoji
func (bot *TradingBot) formatTrendIndicator(trend string) string {
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

// formatMomentumIndicator formats momentum with emoji
func (bot *TradingBot) formatMomentumIndicator(momentum string) string {
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

// formatVolumeStatus formats volume status
func (bot *TradingBot) formatVolumeStatus(volumeStatus string) string {
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

// formatRiskLevel formats risk level with emoji
func (bot *TradingBot) formatRiskLevel(risk string) string {
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

// getRSIStatus returns RSI status indicator
func (bot *TradingBot) getRSIStatus(rsi float64) string {
	if rsi > 70 {
		return "(超买)"
	} else if rsi < 30 {
		return "(超卖)"
	}
	return "(正常)"
}

// getConfidenceBar returns a visual confidence bar
func (bot *TradingBot) getConfidenceBar(confidence float64) string {
	barLength := int(confidence * 20)
	bar := strings.Repeat("█", barLength) + strings.Repeat("░", 20-barLength)
	return "[" + bar + "]"
}

// wrapText wraps text to specified width
func (bot *TradingBot) wrapText(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	currentLine := ""

	for _, word := range words {
		// Count actual display width (Chinese chars count as 2)
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

// formatLargeNumber formats large numbers (e.g., 1330658854 -> 1.33B)
func (bot *TradingBot) formatLargeNumber(num float64) string {
	if num >= 1e9 {
		return fmt.Sprintf("%.2fB", num/1e9)
	} else if num >= 1e6 {
		return fmt.Sprintf("%.2fM", num/1e6)
	} else if num >= 1e3 {
		return fmt.Sprintf("%.2fK", num/1e3)
	}
	return fmt.Sprintf("%.2f", num)
}

// formatPrice formats price preserving original precision without trailing zeros
func (bot *TradingBot) formatPrice(price float64) string {
	// Use %g to preserve original precision and remove trailing zeros
	// But ensure we don't lose precision with scientific notation
	if price >= 1000 {
		// For large prices, use fixed notation with enough precision
		str := fmt.Sprintf("%.10f", price)
		return strings.TrimRight(strings.TrimRight(str, "0"), ".")
	} else if price >= 0.01 {
		// For medium/small prices, preserve up to 10 decimals
		str := fmt.Sprintf("%.10f", price)
		return strings.TrimRight(strings.TrimRight(str, "0"), ".")
	} else {
		// For very small prices, preserve up to 12 decimals
		str := fmt.Sprintf("%.12f", price)
		return strings.TrimRight(strings.TrimRight(str, "0"), ".")
	}
}

// formatTrendCompact returns compact trend indicator
func (bot *TradingBot) formatTrendCompact(trend string) string {
	switch trend {
	case "BULLISH":
		return "🟢BULL"
	case "BEARISH":
		return "🔴BEAR"
	case "NEUTRAL":
		return "⚪NEUT"
	default:
		return trend
	}
}

// formatMomentumCompact returns compact momentum indicator
func (bot *TradingBot) formatMomentumCompact(momentum string) string {
	switch momentum {
	case "BULLISH":
		return "🚀BULL"
	case "BEARISH":
		return "📉BEAR"
	case "NEUTRAL":
		return "➡️NEUT"
	default:
		return momentum
	}
}

// formatVolumeCompact returns compact volume status
func (bot *TradingBot) formatVolumeCompact(volumeStatus string) string {
	switch volumeStatus {
	case "STRONG_BUYING":
		return "🟢BUY"
	case "STRONG_SELLING":
		return "🔴SELL"
	case "NORMAL":
		return "⚪NORM"
	default:
		return volumeStatus
	}
}

// formatRiskLevelCompact returns compact risk level
func (bot *TradingBot) formatRiskLevelCompact(risk string) string {
	switch risk {
	case "LOW":
		return "🟢LOW"
	case "MEDIUM":
		return "🟡MED"
	case "HIGH":
		return "🔴HIGH"
	default:
		return risk
	}
}

// printSimulatedOrder displays simulated order information in CLI
func (bot *TradingBot) printSimulatedOrder(decision *ai.Decision, market *hyperliquid.MarketInfo, balance float64, symbol string) {
	// Only show for actual trading actions
	if decision.Action == "HOLD" || decision.Action == "CLOSE_POSITION" {
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("  💡 SIMULATED ORDER (模拟开仓)")
	fmt.Println(strings.Repeat("=", 80))

	// Determine side emoji
	sideEmoji := "🟢"
	sideText := "多单 (LONG)"
	if decision.Action == "OPEN_SHORT" {
		sideEmoji = "🔴"
		sideText = "空单 (SHORT)"
	}

	// Calculate position details
	positionValue := balance * decision.Size
	coinAmount := positionValue / market.CurrentPrice

	fmt.Printf("\n📊 币种: %s\n", symbol)
	fmt.Printf("   方向: %s %s\n", sideEmoji, sideText)
	fmt.Printf("   杠杆: %dx\n", decision.Leverage)
	fmt.Printf("   价格: $%s\n", bot.formatPrice(market.CurrentPrice))
	fmt.Println()
	fmt.Printf("💰 仓位信息:\n")
	fmt.Printf("   数量: %.4f %s\n", coinAmount, symbol)
	fmt.Printf("   价值: $%.2f USDT\n", positionValue)
	fmt.Printf("   占用资金: %.1f%% (账户余额: $%.2f)\n", decision.Size*100, balance)

	if decision.Leverage > 1 {
		actualExposure := positionValue * float64(decision.Leverage)
		fmt.Printf("   实际曝光: $%.2f USDT (%dx杠杆)\n", actualExposure, decision.Leverage)
	}

	fmt.Println()
	fmt.Printf("🎯 风险管理:\n")
	if decision.StopLoss > 0 {
		stopLossPercent := 0.0
		if decision.Action == "OPEN_LONG" {
			stopLossPercent = ((market.CurrentPrice - decision.StopLoss) / market.CurrentPrice) * 100
		} else {
			stopLossPercent = ((decision.StopLoss - market.CurrentPrice) / market.CurrentPrice) * 100
		}
		fmt.Printf("   止损: $%s (%.2f%%)\n", bot.formatPrice(decision.StopLoss), stopLossPercent)
	}

	if decision.TakeProfit > 0 {
		takeProfitPercent := 0.0
		if decision.Action == "OPEN_LONG" {
			takeProfitPercent = ((decision.TakeProfit - market.CurrentPrice) / market.CurrentPrice) * 100
		} else {
			takeProfitPercent = ((market.CurrentPrice - decision.TakeProfit) / market.CurrentPrice) * 100
		}
		fmt.Printf("   止盈: $%s (%.2f%%)\n", bot.formatPrice(decision.TakeProfit), takeProfitPercent)
	}

	fmt.Printf("   风险等级: %s\n", bot.formatRiskLevel(decision.RiskLevel))
	fmt.Printf("   预期持仓: %s\n", formatHoldingPeriod(decision.ExpectedHoldingPeriod))

	fmt.Println()
	fmt.Printf("📝 开仓理由:\n")
	reasonLines := bot.wrapText(decision.Reason, 76)
	for _, line := range reasonLines {
		fmt.Printf("   %s\n", line)
	}

	fmt.Println()
	fmt.Println("⚠️  注意: 这是模拟开仓,未执行真实交易")
	fmt.Println("   要启用真实交易,请设置 config.yaml 中的 trading_enabled: true")
	fmt.Println(strings.Repeat("=", 80) + "\n")
}

// formatHoldingPeriod formats expected holding period
func formatHoldingPeriod(period string) string {
	switch period {
	case "SHORT":
		return "短期 (几小时)"
	case "MEDIUM":
		return "中期 (1-7天)"
	case "LONG":
		return "长期 (>7天)"
	default:
		return period
	}
}

// formatPriceGlobal formats price preserving original precision (global helper)
func formatPriceGlobal(price float64) string {
	if price >= 1000 {
		str := fmt.Sprintf("%.10f", price)
		return strings.TrimRight(strings.TrimRight(str, "0"), ".")
	} else if price >= 0.01 {
		str := fmt.Sprintf("%.10f", price)
		return strings.TrimRight(strings.TrimRight(str, "0"), ".")
	} else {
		str := fmt.Sprintf("%.12f", price)
		return strings.TrimRight(strings.TrimRight(str, "0"), ".")
	}
}

// showPositions displays current positions for all symbols
func showPositions() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Printf("❌ Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	hlClient := hyperliquid.NewClient(cfg.Hyperliquid.APIURL)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("  💼 Current Positions - %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 80))

	// Get account balance first
	balance, err := hlClient.GetAccountBalance(cfg.Hyperliquid.AccountAddress)
	if err != nil {
		fmt.Printf("\n❌ Failed to get account balance: %v\n", err)
	} else {
		fmt.Printf("\n💰 Account Balance: $%.2f\n", balance)
	}

	hasOpenPosition := false
	totalPnL := 0.0
	totalExposure := 0.0

	fmt.Println("\n" + strings.Repeat("-", 80))

	// Check positions for all configured symbols
	for _, symbol := range cfg.Trading.Symbols {
		position, err := hlClient.GetPosition(symbol, cfg.Hyperliquid.AccountAddress)
		if err != nil {
			fmt.Printf("\n❌ %s: Failed to fetch position - %v\n", symbol, err)
			continue
		}

		if position.Size > 0 {
			hasOpenPosition = true

			// Get current market price
			marketInfo, err := hlClient.GetMarketData(symbol)
			currentPrice := 0.0
			if err == nil {
				currentPrice = marketInfo.CurrentPrice
			}

			// Calculate exposure
			exposure := position.Size * position.EntryPrice
			totalExposure += exposure
			totalPnL += position.CurrentPnL

			// Format side with emoji
			sideEmoji := "🟢"
			if position.Side == "SHORT" {
				sideEmoji = "🔴"
			}

			// Format PnL with emoji
			pnlEmoji := "🟢"
			if position.PnLPercent < 0 {
				pnlEmoji = "🔴"
			}

			fmt.Printf("\n📊 %s Position:\n", symbol)
			fmt.Printf("  Side:         %s %s\n", sideEmoji, position.Side)
			fmt.Printf("  Size:         %.4f\n", position.Size)
			fmt.Printf("  Entry Price:  $%s\n", formatPriceGlobal(position.EntryPrice))
			if currentPrice > 0 {
				fmt.Printf("  Current Price: $%s\n", formatPriceGlobal(currentPrice))
				priceChange := ((currentPrice - position.EntryPrice) / position.EntryPrice) * 100
				priceChangeEmoji := "🟢"
				if priceChange < 0 {
					priceChangeEmoji = "🔴"
				}
				fmt.Printf("  Price Change: %s %.2f%%\n", priceChangeEmoji, priceChange)
			}
			fmt.Printf("  Unrealized P&L: %s $%.2f (%.2f%%)\n", pnlEmoji, position.CurrentPnL, position.PnLPercent)
			fmt.Printf("  Exposure:     $%.2f\n", exposure)
			fmt.Printf("  Holding Time: %s\n", position.HoldingTime.String())

			fmt.Println(strings.Repeat("-", 80))
		}
	}

	if !hasOpenPosition {
		fmt.Println("\n📭 No open positions\n")
	} else {
		fmt.Println("\n📈 Summary:")
		fmt.Printf("  Total Exposure:  $%.2f\n", totalExposure)
		totalPnLEmoji := "🟢"
		if totalPnL < 0 {
			totalPnLEmoji = "🔴"
		}
		fmt.Printf("  Total P&L:       %s $%.2f\n", totalPnLEmoji, totalPnL)
		if balance > 0 {
			exposurePercent := (totalExposure / balance) * 100
			fmt.Printf("  Exposure Ratio:  %.2f%%\n", exposurePercent)
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", 80) + "\n")
}

// showBalance displays account balance
func showBalance() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Printf("❌ Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	hlClient := hyperliquid.NewClient(cfg.Hyperliquid.APIURL)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("  💰 Account Balance - %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 80))

	balance, err := hlClient.GetAccountBalance(cfg.Hyperliquid.AccountAddress)
	if err != nil {
		fmt.Printf("\n❌ Failed to get account balance: %v\n\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n💵 Available Balance: $%.2f\n", balance)

	// Calculate total exposure
	totalExposure := 0.0
	for _, symbol := range cfg.Trading.Symbols {
		position, err := hlClient.GetPosition(symbol, cfg.Hyperliquid.AccountAddress)
		if err == nil && position.Size > 0 {
			totalExposure += position.Size * position.EntryPrice
		}
	}

	if totalExposure > 0 {
		fmt.Printf("💼 Total Exposure:    $%.2f\n", totalExposure)
		fmt.Printf("📊 Exposure Ratio:    %.2f%%\n", (totalExposure/balance)*100)
		fmt.Printf("💵 Free Balance:      $%.2f\n", balance-totalExposure)
	}

	fmt.Println("\n" + strings.Repeat("=", 80) + "\n")
}

// showHelp displays usage information
func showHelp() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("  🤖 AI Trading System - Command Line Interface")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("\nUsage:")
	fmt.Println("  ./aitrading              Start the trading bot")
	fmt.Println("  ./aitrading order        Show current positions")
	fmt.Println("  ./aitrading orders       Show current positions (alias)")
	fmt.Println("  ./aitrading position     Show current positions (alias)")
	fmt.Println("  ./aitrading positions    Show current positions (alias)")
	fmt.Println("  ./aitrading balance      Show account balance")
	fmt.Println("  ./aitrading help         Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  # Start trading bot")
	fmt.Println("  ./aitrading")
	fmt.Println()
	fmt.Println("  # Check your positions")
	fmt.Println("  ./aitrading order")
	fmt.Println()
	fmt.Println("  # Check account balance")
	fmt.Println("  ./aitrading balance")
	fmt.Println("\nConfiguration:")
	fmt.Println("  Edit config.yaml to configure:")
	fmt.Println("  - Trading symbols (ETH, BTC, DOGE, etc.)")
	fmt.Println("  - Risk parameters (leverage, position limits)")
	fmt.Println("  - AI model (DeepSeek or Qwen)")
	fmt.Println("\nDocumentation:")
	fmt.Println("  QUICKSTART.md         Quick start guide")
	fmt.Println("  CLI_REPORT_GUIDE.md   CLI report documentation")
	fmt.Println("  ENHANCEMENT_REPORT.md System features overview")
	fmt.Println()
	fmt.Println(strings.Repeat("=", 80) + "\n")
}

// getLogLevel converts string to logrus level
func getLogLevel(level string) logrus.Level {
	switch level {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

// printExecutionResult prints trade execution result to console
func (bot *TradingBot) printExecutionResult(result *executor.ExecutionResult, symbol string) {
	fmt.Println("\n" + strings.Repeat("=", 80))

	// Determine color and emoji based on success
	var statusColor, statusEmoji, statusText string
	if result.Success {
		statusColor = colorGreen
		statusEmoji = "✅"
		statusText = "SUCCESS"
	} else {
		statusColor = colorRed
		statusEmoji = "❌"
		statusText = "FAILED"
	}

	fmt.Printf("%s  %s Order Execution %s - %s @ %s%s\n",
		statusColor, statusEmoji, statusText, symbol,
		time.Now().Format("2006-01-02 15:04:05"), colorReset)
	fmt.Println(strings.Repeat("=", 80))

	// Action info
	fmt.Printf("\n📋 Action: %s\n", bot.formatAction(result.Action))

	// Order details (if not HOLD)
	if result.Action != "HOLD" {
		if result.Side != "" {
			sideEmoji := "🟢"
			if result.Side == "SHORT" {
				sideEmoji = "🔴"
			}
			fmt.Printf("   Side:   %s %s\n", sideEmoji, result.Side)
		}

		if result.Size > 0 {
			fmt.Printf("   Size:   %.4f %s\n", result.Size, symbol)
		}

		if result.Price > 0 {
			fmt.Printf("   Price:  $%s\n", bot.formatPrice(result.Price))
		}

		if result.OrderID != "" {
			fmt.Printf("   Order ID: %s\n", result.OrderID)
		}
	}

	// Status message
	if result.Message != "" {
		fmt.Printf("\n💬 Message: %s\n", result.Message)
	}

	// Additional info for successful orders
	if result.Success && (result.Action == "OPEN_LONG" || result.Action == "OPEN_SHORT") {
		if result.StopLoss > 0 {
			fmt.Printf("\n🛡️  Stop Loss:   $%s\n", bot.formatPrice(result.StopLoss))
		}
		if result.TakeProfit > 0 {
			fmt.Printf("🎯 Take Profit: $%s\n", bot.formatPrice(result.TakeProfit))
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80) + "\n")
}

func main() {
	// Check command line arguments
	if len(os.Args) > 1 {
		command := os.Args[1]

		switch command {
		case "order", "orders", "position", "positions":
			showPositions()
			return
		case "balance":
			showBalance()
			return
		case "help", "-h", "--help":
			showHelp()
			return
		default:
			fmt.Printf("Unknown command: %s\n", command)
			fmt.Println("Run './aitrading help' for usage information")
			os.Exit(1)
		}
	}

	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Create trading bot
	bot, err := NewTradingBot(cfg)
	if err != nil {
		fmt.Printf("Failed to create trading bot: %v\n", err)
		os.Exit(1)
	}

	// Start bot
	if err := bot.Start(); err != nil {
		fmt.Printf("Failed to start trading bot: %v\n", err)
		os.Exit(1)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("AI Trading Bot is running. Press Ctrl+C to stop.")
	<-sigChan

	// Graceful shutdown
	bot.Stop()
	fmt.Println("Bot stopped gracefully")
}
