package indicators

import (
	"math"
)

// MarketData represents a single candlestick
type MarketData struct {
	Timestamp int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

// TechnicalIndicators contains all calculated indicators
type TechnicalIndicators struct {
	// Trend Indicators
	SMA10  float64
	SMA60  float64
	SMA120 float64
	EMA10  float64
	EMA60  float64
	EMA120 float64

	// Momentum Indicators
	MACDDIF  float64
	MACDDEA  float64
	MACDHIST float64
	RSI14    float64

	// Volatility Indicators
	BBUpper  float64
	BBMiddle float64
	BBLower  float64
	BBWidth  float64

	// Volume Indicators
	VMA20         float64
	CurrentVolume float64

	// Derived Analysis
	TrendStrength       string
	MomentumStatus      string
	BBPosition          string
	VolumePriceRelation string
}

// Calculator provides technical indicator calculations
type Calculator struct{}

// NewCalculator creates a new indicator calculator
func NewCalculator() *Calculator {
	return &Calculator{}
}

// Calculate computes all technical indicators from market data
func (c *Calculator) Calculate(data []MarketData) *TechnicalIndicators {
	if len(data) < 120 {
		return nil // Need at least 120 periods for all indicators
	}

	indicators := &TechnicalIndicators{
		CurrentVolume: data[len(data)-1].Volume,
	}

	closes := make([]float64, len(data))
	highs := make([]float64, len(data))
	lows := make([]float64, len(data))
	volumes := make([]float64, len(data))

	for i, candle := range data {
		closes[i] = candle.Close
		highs[i] = candle.High
		lows[i] = candle.Low
		volumes[i] = candle.Volume
	}

	// Calculate SMAs
	indicators.SMA10 = c.SMA(closes, 10)
	indicators.SMA60 = c.SMA(closes, 60)
	indicators.SMA120 = c.SMA(closes, 120)

	// Calculate EMAs
	indicators.EMA10 = c.EMA(closes, 10)
	indicators.EMA60 = c.EMA(closes, 60)
	indicators.EMA120 = c.EMA(closes, 120)

	// Calculate MACD
	indicators.MACDDIF, indicators.MACDDEA, indicators.MACDHIST = c.MACD(closes)

	// Calculate RSI
	indicators.RSI14 = c.RSI(closes, 14)

	// Calculate Bollinger Bands
	indicators.BBUpper, indicators.BBMiddle, indicators.BBLower = c.BollingerBands(closes, 20, 2)
	indicators.BBWidth = (indicators.BBUpper - indicators.BBLower) / indicators.BBMiddle

	// Calculate Volume MA
	indicators.VMA20 = c.SMA(volumes, 20)

	// Derived analysis
	indicators.TrendStrength = c.analyzeTrend(indicators, closes[len(closes)-1])
	indicators.MomentumStatus = c.analyzeMomentum(indicators)
	indicators.BBPosition = c.analyzeBBPosition(indicators, closes[len(closes)-1])
	indicators.VolumePriceRelation = c.analyzeVolumePriceRelation(indicators, data)

	return indicators
}

// SMA calculates Simple Moving Average
func (c *Calculator) SMA(data []float64, period int) float64 {
	if len(data) < period {
		return 0
	}

	sum := 0.0
	for i := len(data) - period; i < len(data); i++ {
		sum += data[i]
	}
	return sum / float64(period)
}

// EMA calculates Exponential Moving Average
func (c *Calculator) EMA(data []float64, period int) float64 {
	if len(data) < period {
		return 0
	}

	multiplier := 2.0 / float64(period+1)
	ema := c.SMA(data[:period], period)

	for i := period; i < len(data); i++ {
		ema = (data[i] * multiplier) + (ema * (1 - multiplier))
	}

	return ema
}

// MACD calculates Moving Average Convergence Divergence
func (c *Calculator) MACD(data []float64) (dif, dea, hist float64) {
	ema12 := c.EMASequence(data, 12)
	ema26 := c.EMASequence(data, 26)

	if len(ema12) == 0 || len(ema26) == 0 {
		return 0, 0, 0
	}

	// DIF = EMA12 - EMA26
	// Need to align the arrays since they have different lengths
	// ema26 will be shorter, so we use its length
	minLen := len(ema26)
	if len(ema12) < minLen {
		minLen = len(ema12)
	}

	// Skip the beginning of ema12 to align with ema26
	offset := len(ema12) - minLen

	difSeq := make([]float64, minLen)
	for i := 0; i < minLen; i++ {
		difSeq[i] = ema12[i+offset] - ema26[i]
	}

	// DEA = EMA of DIF (9 periods)
	deaSeq := c.EMASequence(difSeq, 9)
	if len(deaSeq) == 0 {
		return 0, 0, 0
	}

	dif = difSeq[len(difSeq)-1]
	dea = deaSeq[len(deaSeq)-1]
	hist = dif - dea

	return dif, dea, hist
}

// EMASequence calculates EMA for entire sequence
func (c *Calculator) EMASequence(data []float64, period int) []float64 {
	if len(data) < period {
		return nil
	}

	result := make([]float64, len(data)-period+1)
	multiplier := 2.0 / float64(period+1)

	// First EMA is SMA
	result[0] = c.SMA(data[:period], period)

	// Calculate rest
	for i := period; i < len(data); i++ {
		result[i-period+1] = (data[i] * multiplier) + (result[i-period] * (1 - multiplier))
	}

	return result
}

// RSI calculates Relative Strength Index
func (c *Calculator) RSI(data []float64, period int) float64 {
	if len(data) < period+1 {
		return 50
	}

	gains := 0.0
	losses := 0.0

	// Calculate initial average gain/loss
	for i := len(data) - period; i < len(data); i++ {
		change := data[i] - data[i-1]
		if change > 0 {
			gains += change
		} else {
			losses -= change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	if avgLoss == 0 {
		return 100
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi
}

// BollingerBands calculates Bollinger Bands
func (c *Calculator) BollingerBands(data []float64, period int, stdDev float64) (upper, middle, lower float64) {
	if len(data) < period {
		return 0, 0, 0
	}

	middle = c.SMA(data, period)

	// Calculate standard deviation
	variance := 0.0
	for i := len(data) - period; i < len(data); i++ {
		variance += math.Pow(data[i]-middle, 2)
	}
	std := math.Sqrt(variance / float64(period))

	upper = middle + (stdDev * std)
	lower = middle - (stdDev * std)

	return upper, middle, lower
}

// analyzeTrend determines trend strength
func (c *Calculator) analyzeTrend(ind *TechnicalIndicators, currentPrice float64) string {
	bullishSignals := 0
	bearishSignals := 0

	// Price vs EMAs
	if currentPrice > ind.EMA10 {
		bullishSignals++
	} else {
		bearishSignals++
	}

	if currentPrice > ind.EMA60 {
		bullishSignals++
	} else {
		bearishSignals++
	}

	// EMA alignment
	if ind.EMA10 > ind.EMA60 && ind.EMA60 > ind.EMA120 {
		bullishSignals += 2
	} else if ind.EMA10 < ind.EMA60 && ind.EMA60 < ind.EMA120 {
		bearishSignals += 2
	}

	if bullishSignals >= 3 {
		return "STRONG_BULLISH"
	} else if bullishSignals > bearishSignals {
		return "BULLISH"
	} else if bearishSignals >= 3 {
		return "STRONG_BEARISH"
	} else if bearishSignals > bullishSignals {
		return "BEARISH"
	}

	return "NEUTRAL"
}

// analyzeMomentum determines momentum status
func (c *Calculator) analyzeMomentum(ind *TechnicalIndicators) string {
	signals := 0

	// MACD analysis
	if ind.MACDHIST > 0 {
		signals++
	} else {
		signals--
	}

	// RSI analysis
	if ind.RSI14 > 70 {
		return "OVERBOUGHT"
	} else if ind.RSI14 < 30 {
		return "OVERSOLD"
	} else if ind.RSI14 > 50 {
		signals++
	} else {
		signals--
	}

	if signals > 0 {
		return "BULLISH"
	} else if signals < 0 {
		return "BEARISH"
	}

	return "NEUTRAL"
}

// analyzeBBPosition determines price position in Bollinger Bands
func (c *Calculator) analyzeBBPosition(ind *TechnicalIndicators, currentPrice float64) string {
	bandRange := ind.BBUpper - ind.BBLower
	if bandRange == 0 {
		return "MIDDLE"
	}

	position := (currentPrice - ind.BBLower) / bandRange

	if position >= 0.8 {
		return "NEAR_UPPER"
	} else if position >= 0.6 {
		return "UPPER_HALF"
	} else if position >= 0.4 {
		return "MIDDLE"
	} else if position >= 0.2 {
		return "LOWER_HALF"
	}

	return "NEAR_LOWER"
}

// analyzeVolumePriceRelation analyzes volume-price relationship
func (c *Calculator) analyzeVolumePriceRelation(ind *TechnicalIndicators, data []MarketData) string {
	if len(data) < 2 {
		return "NEUTRAL"
	}

	currentVolume := data[len(data)-1].Volume
	priceChange := data[len(data)-1].Close - data[len(data)-2].Close

	volumeRatio := currentVolume / ind.VMA20

	if volumeRatio > 1.5 && priceChange > 0 {
		return "STRONG_BUYING"
	} else if volumeRatio > 1.5 && priceChange < 0 {
		return "STRONG_SELLING"
	} else if volumeRatio > 1.0 && priceChange > 0 {
		return "BUYING"
	} else if volumeRatio > 1.0 && priceChange < 0 {
		return "SELLING"
	}

	return "NEUTRAL"
}
