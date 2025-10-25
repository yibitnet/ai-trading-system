package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"aitrading/hyperliquid"
	"aitrading/indicators"
)

// DecisionMaker handles AI-based trading decisions
type DecisionMaker struct {
	provider    string
	apiKey      string
	baseURL     string
	model       string
	temperature float64
	maxTokens   int
	timeout     time.Duration
	httpClient  *http.Client
}

// NewDecisionMaker creates a new AI decision maker
func NewDecisionMaker(provider, apiKey, baseURL, model string, temperature float64, maxTokens, timeout int) *DecisionMaker {
	return &DecisionMaker{
		provider:    provider,
		apiKey:      apiKey,
		baseURL:     baseURL,
		model:       model,
		temperature: temperature,
		maxTokens:   maxTokens,
		timeout:     time.Duration(timeout) * time.Second,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// Decision represents a trading decision from AI
type Decision struct {
	Action                string  `json:"action"`
	Confidence            float64 `json:"confidence"`
	Size                  float64 `json:"size"`
	Leverage              int     `json:"leverage"`
	Reason                string  `json:"reason"`
	StopLoss              float64 `json:"stop_loss"`
	TakeProfit            float64 `json:"take_profit"`
	RiskLevel             string  `json:"risk_level"`
	ExpectedHoldingPeriod string  `json:"expected_holding_period"`
}

// MarketAnalysis contains all data for AI analysis
type MarketAnalysis struct {
	Symbol     string
	Timestamp  time.Time
	Market     *hyperliquid.MarketInfo
	Indicators *indicators.TechnicalIndicators
	Position   *hyperliquid.Position
}

// Analyze sends market data to AI and gets trading decision
func (dm *DecisionMaker) Analyze(analysis *MarketAnalysis) (*Decision, error) {
	// Build the prompt with all market data
	prompt := dm.buildPrompt(analysis)

	// Call AI API based on provider
	response, err := dm.callAI(prompt)
	if err != nil {
		return nil, fmt.Errorf("AI API call failed: %w", err)
	}

	// Parse the JSON decision from response
	decision, err := dm.parseDecision(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI decision: %w", err)
	}

	return decision, nil
}

// buildPrompt creates the prompt for AI analysis
func (dm *DecisionMaker) buildPrompt(analysis *MarketAnalysis) string {
	ind := analysis.Indicators
	pos := analysis.Position
	mkt := analysis.Market

	prompt := fmt.Sprintf(`你是一位专业的量化交易员,拥有20年加密货币交易经验。现在作为自动化交易系统的核心决策引擎,你需要基于实时市场数据做出冷静、理性的交易决策。

**你的任务:**
分析当前市场状况,给出明确的可执行交易指令,帮助自动化系统执行交易。

## 市场数据 - %s @ %s
- 当前价格: %.2f
- 24小时变化: %.2f%%
- 交易量: %.2f

## 技术指标状态
**趋势指标:**
- SMA系列: SMA10=%.2f, SMA60=%.2f, SMA120=%.2f
- EMA系列: EMA10=%.2f, EMA60=%.2f, EMA120=%.2f
- 趋势判断: %s

**动量指标:**
- MACD: DIF=%.4f, DEA=%.4f, HIST=%.4f
- RSI(14): %.2f
- 动量状态: %s

**波动性指标:**
- 布林带: 上轨=%.2f, 中轨=%.2f, 下轨=%.2f
- 价格位置: %s
- 带宽: %.4f

**成交量指标:**
- 当前成交量: %.2f
- VMA20: %.2f
- 量价关系: %s

## 当前持仓状态
- 持仓方向: %s
- 持仓数量: %.4f
- 开仓价格: %.2f
- 当前盈亏: %.2f%%
- 持仓时间: %s

## 决策规则指导

**开仓条件(需满足至少3个条件):**

多头开仓信号:
- 价格突破关键均线(如EMA20)
- MACD金叉且柱状图转正
- RSI从超卖区域回升(<30向上)
- 布林带下轨支撑有效
- 成交量放大确认

空头开仓信号:
- 价格跌破关键均线支撑
- MACD死叉且柱状图转负
- RSI从超买区域回落(>70向下)
- 布林带上轨阻力有效
- 成交量配合下跌

**加仓条件:**
- 已有仓位处于盈利状态
- 趋势确认延续
- 关键技术位突破
- 风险敞口在可控范围内

**平仓条件:**
- 达到目标止盈位
- 触及止损位
- 技术指标出现反转信号
- 持仓时间超过最大限制

## 风险管理要求

**仓位管理原则:**
- 单次开仓不超过总资金的10%%
- 总风险敞口不超过25%%
- 风险回报比至少1:2
- 根据市场波动性调整仓位大小

**置信度标准:**
- 高置信度(>0.8): 多个指标强烈共振,趋势明确
- 中置信度(0.6-0.8): 主要指标一致,但有轻微分歧
- 低置信度(<0.6): 指标分歧较大,趋势不明确

**杠杆倍数建议:**
- 根据市场波动性和趋势强度建议杠杆倍数
- 低风险/强趋势: 可以使用5-10倍杠杆
- 中风险/中等趋势: 使用3-5倍杠杆
- 高风险/弱趋势: 使用1-3倍杠杆
- 系统会自动限制不超过配置的最大杠杆倍数

请严格按照以下JSON格式返回决策,不要包含任何其他文字:

{
  "action": "OPEN_LONG|OPEN_SHORT|ADD_POSITION|CLOSE_POSITION|HOLD",
  "confidence": 0.0-1.0,
  "size": 0.0-1.0,
  "leverage": 1-20,
  "reason": "详细的技术分析理由,引用具体��标",
  "stop_loss": 具体价格,
  "take_profit": 具体价格,
  "risk_level": "LOW|MEDIUM|HIGH",
  "expected_holding_period": "SHORT|MEDIUM|LONG"
}`,
		analysis.Symbol,
		analysis.Timestamp.Format("2006-01-02 15:04:05"),
		mkt.CurrentPrice,
		mkt.PriceChange,
		mkt.Volume24h,
		ind.SMA10, ind.SMA60, ind.SMA120,
		ind.EMA10, ind.EMA60, ind.EMA120,
		ind.TrendStrength,
		ind.MACDDIF, ind.MACDDEA, ind.MACDHIST,
		ind.RSI14,
		ind.MomentumStatus,
		ind.BBUpper, ind.BBMiddle, ind.BBLower,
		ind.BBPosition,
		ind.BBWidth,
		ind.CurrentVolume,
		ind.VMA20,
		ind.VolumePriceRelation,
		pos.Side,
		pos.Size,
		pos.EntryPrice,
		pos.PnLPercent,
		pos.HoldingTime.String(),
	)

	return prompt
}

// callAI makes API call to the configured AI provider
func (dm *DecisionMaker) callAI(prompt string) (string, error) {
	url := fmt.Sprintf("%s/chat/completions", dm.baseURL)

	reqBody := map[string]interface{}{
		"model": dm.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": dm.temperature,
		"max_tokens":  dm.maxTokens,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", dm.apiKey))

	resp, err := dm.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Extract the response text
	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					return content, nil
				}
			}
		}
	}

	return "", fmt.Errorf("unexpected response format")
}

// parseDecision extracts Decision struct from AI response
func (dm *DecisionMaker) parseDecision(response string) (*Decision, error) {
	// Try to find JSON in the response
	// Sometimes AI might wrap it in markdown code blocks
	start := -1
	end := -1

	for i := 0; i < len(response); i++ {
		if response[i] == '{' && start == -1 {
			start = i
		}
		if response[i] == '}' {
			end = i + 1
		}
	}

	if start == -1 || end == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}

	jsonStr := response[start:end]

	var decision Decision
	if err := json.Unmarshal([]byte(jsonStr), &decision); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w\nResponse: %s", err, jsonStr)
	}

	// Validate decision
	if err := dm.validateDecision(&decision); err != nil {
		return nil, err
	}

	return &decision, nil
}

// validateDecision validates the AI decision
func (dm *DecisionMaker) validateDecision(decision *Decision) error {
	validActions := map[string]bool{
		"OPEN_LONG":      true,
		"OPEN_SHORT":     true,
		"ADD_POSITION":   true,
		"CLOSE_POSITION": true,
		"HOLD":           true,
	}

	if !validActions[decision.Action] {
		return fmt.Errorf("invalid action: %s", decision.Action)
	}

	if decision.Confidence < 0 || decision.Confidence > 1 {
		return fmt.Errorf("invalid confidence: %f", decision.Confidence)
	}

	if decision.Size < 0 || decision.Size > 1 {
		return fmt.Errorf("invalid size: %f", decision.Size)
	}

	if decision.Leverage < 1 || decision.Leverage > 20 {
		return fmt.Errorf("invalid leverage: %d", decision.Leverage)
	}

	validRiskLevels := map[string]bool{
		"LOW":    true,
		"MEDIUM": true,
		"HIGH":   true,
	}

	if !validRiskLevels[decision.RiskLevel] {
		return fmt.Errorf("invalid risk level: %s", decision.RiskLevel)
	}

	return nil
}
