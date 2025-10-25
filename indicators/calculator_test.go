package indicators

import (
	"math"
	"testing"
)

func TestSMA(t *testing.T) {
	calc := NewCalculator()
	data := []float64{10, 20, 30, 40, 50}

	result := calc.SMA(data, 3)
	expected := 40.0 // (30+40+50)/3

	if math.Abs(result-expected) > 0.01 {
		t.Errorf("SMA calculation failed: got %f, expected %f", result, expected)
	}
}

func TestEMA(t *testing.T) {
	calc := NewCalculator()
	data := []float64{22.27, 22.19, 22.08, 22.17, 22.18, 22.13, 22.23, 22.43, 22.24, 22.29}

	result := calc.EMA(data, 5)

	if result <= 0 {
		t.Errorf("EMA calculation failed: got %f, should be positive", result)
	}
}

func TestRSI(t *testing.T) {
	calc := NewCalculator()
	// Create test data with upward trend
	data := make([]float64, 20)
	for i := 0; i < 20; i++ {
		data[i] = float64(100 + i*2)
	}

	result := calc.RSI(data, 14)

	if result < 0 || result > 100 {
		t.Errorf("RSI out of range: got %f, should be between 0-100", result)
	}

	// RSI should be high for upward trend
	if result < 50 {
		t.Errorf("RSI should be high for upward trend: got %f", result)
	}
}

func TestBollingerBands(t *testing.T) {
	calc := NewCalculator()
	data := []float64{10, 12, 15, 14, 13, 16, 18, 20, 19, 21, 23, 22, 24, 25, 26, 27, 28, 29, 30, 31}

	upper, middle, lower := calc.BollingerBands(data, 20, 2)

	if upper <= middle {
		t.Errorf("Upper band should be above middle: upper=%f, middle=%f", upper, middle)
	}

	if middle <= lower {
		t.Errorf("Middle should be above lower band: middle=%f, lower=%f", middle, lower)
	}

	if lower <= 0 {
		t.Errorf("Lower band should be positive: got %f", lower)
	}
}

func TestCalculateWithInsufficientData(t *testing.T) {
	calc := NewCalculator()
	// Only 50 candles, need 120
	data := make([]MarketData, 50)
	for i := range data {
		data[i] = MarketData{
			Close:  100.0 + float64(i),
			High:   105.0 + float64(i),
			Low:    95.0 + float64(i),
			Volume: 1000.0,
		}
	}

	result := calc.Calculate(data)
	if result != nil {
		t.Error("Calculate should return nil with insufficient data")
	}
}

func TestCalculateWithValidData(t *testing.T) {
	calc := NewCalculator()
	// Create 150 candles
	data := make([]MarketData, 150)
	for i := range data {
		data[i] = MarketData{
			Close:  100.0 + float64(i)*0.5,
			High:   105.0 + float64(i)*0.5,
			Low:    95.0 + float64(i)*0.5,
			Open:   100.0 + float64(i)*0.5,
			Volume: 1000.0 + float64(i)*10,
		}
	}

	result := calc.Calculate(data)
	if result == nil {
		t.Fatal("Calculate should return indicators with valid data")
	}

	// Verify all indicators are calculated
	if result.SMA10 <= 0 {
		t.Error("SMA10 should be positive")
	}
	if result.EMA10 <= 0 {
		t.Error("EMA10 should be positive")
	}
	if result.RSI14 < 0 || result.RSI14 > 100 {
		t.Errorf("RSI14 out of range: %f", result.RSI14)
	}
	if result.BBUpper <= result.BBMiddle || result.BBMiddle <= result.BBLower {
		t.Error("Bollinger Bands order is incorrect")
	}
	if result.TrendStrength == "" {
		t.Error("Trend strength should be calculated")
	}
}
