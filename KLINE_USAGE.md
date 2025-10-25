# K线数据使用说明

## 📊 K线数据获取

### 获取数量
系统每次获取 **150根K线** 进行技术指标计算:

```go
// main.go:188
candles, err := bot.hlClient.GetCandlestickData(symbol, bot.config.Trading.Timeframe, 150)
```

### 最低要求
系统要求至少 **120根K线** 才能进行计算:

```go
// main.go:193-195
if len(candles) < 120 {
    return fmt.Errorf("insufficient candle data: got %d, need at least 120", len(candles))
}
```

---

## 🔧 各技术指标使用的K线数量

### 1. **移动平均线 (MA)**

| 指标 | 周期 | 使用K线数 | 说明 |
|------|------|----------|------|
| SMA10 | 10 | 10 | 10周期简单移动平均 |
| SMA60 | 60 | 60 | 60周期简单移动平均 |
| SMA120 | 120 | 120 | 120周期简单移动平均 ⚠️ 最长 |
| EMA10 | 10 | 10+ | 10周期指数移动平均 |
| EMA60 | 60 | 60+ | 60周期指数移动平均 |
| EMA120 | 120 | 120+ | 120周期指数移动平均 ⚠️ 最长 |

**代码位置:**
```go
// indicators/calculator.go:81-88
indicators.SMA10 = c.SMA(closes, 10)
indicators.SMA60 = c.SMA(closes, 60)
indicators.SMA120 = c.SMA(closes, 120)

indicators.EMA10 = c.EMA(closes, 10)
indicators.EMA60 = c.EMA(closes, 60)
indicators.EMA120 = c.EMA(closes, 120)
```

---

### 2. **MACD (Moving Average Convergence Divergence)**

| 组件 | 周期 | 使用K线数 | 说明 |
|------|------|----------|------|
| EMA12 | 12 | 12+ | 快速EMA |
| EMA26 | 26 | 26+ | 慢速EMA |
| DIF | - | 26+ | EMA12 - EMA26 |
| DEA (信号线) | 9 | 26+9=35+ | DIF的9周期EMA |
| HIST (柱状图) | - | 35+ | DIF - DEA |

**代码位置:**
```go
// indicators/calculator.go:91
indicators.MACDDIF, indicators.MACDDEA, indicators.MACDHIST = c.MACD(closes)

// MACD计算细节 (calculator.go:142-177)
ema12 := c.EMASequence(data, 12)
ema26 := c.EMASequence(data, 26)
difSeq := ema12 - ema26
deaSeq := c.EMASequence(difSeq, 9)
```

**实际需要:** 约 **35根K线** (26 + 9)

---

### 3. **RSI (Relative Strength Index)**

| 指标 | 周期 | 使用K线数 | 说明 |
|------|------|----------|------|
| RSI14 | 14 | 15 | 14周期RSI (需要15根计算变化) |

**代码位置:**
```go
// indicators/calculator.go:94
indicators.RSI14 = c.RSI(closes, 14)

// RSI计算细节 (calculator.go:200-229)
// 需要period+1根K线来计算价格变化
if len(data) < period+1 {
    return 50
}
```

**实际需要:** **15根K线** (14周期 + 1根计算变化)

---

### 4. **布林带 (Bollinger Bands)**

| 指标 | 周期 | 使用K线数 | 说明 |
|------|------|----------|------|
| BB中轨 (Middle) | 20 | 20 | 20周期SMA |
| BB上轨 (Upper) | 20 | 20 | 中轨 + 2倍标准差 |
| BB下轨 (Lower) | 20 | 20 | 中轨 - 2倍标准差 |

**代码位置:**
```go
// indicators/calculator.go:97-98
indicators.BBUpper, indicators.BBMiddle, indicators.BBLower = c.BollingerBands(closes, 20, 2)

// 布林带计算 (calculator.go:232-250)
middle = c.SMA(data, 20)  // 20周期SMA
std = 计算标准差(20个数据点)
upper = middle + (2 * std)
lower = middle - (2 * std)
```

**实际需要:** **20根K线**

---

### 5. **成交量指标 (Volume)**

| 指标 | 周期 | 使用K线数 | 说明 |
|------|------|----------|------|
| VMA20 | 20 | 20 | 20周期成交量移动平均 |

**代码位置:**
```go
// indicators/calculator.go:101
indicators.VMA20 = c.SMA(volumes, 20)
```

**实际需要:** **20根K线**

---

## 📊 为什么获取150根K线?

### 计算需求分析

| 指标 | 最长周期 | 需要K线数 |
|------|---------|----------|
| SMA/EMA | 120 | 120 |
| MACD | 26+9 | ~35 |
| RSI | 14+1 | 15 |
| 布林带 | 20 | 20 |
| VMA | 20 | 20 |
| **最大需求** | **SMA120/EMA120** | **120** |

### 设置150根的原因

1. **满足最长周期需求**
   - SMA120/EMA120 需要 120根
   - 150 > 120 ✅ 满足

2. **提供安全余量**
   - 余量: 150 - 120 = 30根
   - 应对数据缺失或异常
   - 确保所有指标都能正确计算

3. **性能考虑**
   - 150根数据量适中
   - 不会过度占用内存
   - API调用效率合理

---

## 🔍 实际使用情况

### 代码执行流程

```go
// main.go 交易周期执行流程

// 步骤1: 获取150根K线
candles, err := bot.hlClient.GetCandlestickData(symbol, timeframe, 150)

// 步骤2: 验证数量
if len(candles) < 120 {
    return error  // 数据不足
}

// 步骤3: 计算指标 (使用所有150根)
indicators := bot.calculator.Calculate(candles)

// Calculate函数内部:
// - SMA10: 使用最后10根
// - SMA60: 使用最后60根
// - SMA120: 使用最后120根
// - MACD: 使用所有数据计算EMA序列
// - RSI14: 使用最后15根
// - BB20: 使用最后20根
// - VMA20: 使用最后20根成交量
```

---

## 📈 K线时间框架

K线周期由配置文件控制:

```yaml
# config.yaml
trading:
  timeframe: "5m"  # 可选: 1m, 5m, 15m, 1h, 4h, 1d
```

### 不同时间框架的含义

| 时间框架 | 150根K线代表 | 实际时间跨度 |
|---------|-------------|------------|
| 1m | 150分钟 | 2.5小时 |
| 5m | 750分钟 | 12.5小时 |
| 15m | 2250分钟 | 37.5小时 (1.5天) |
| 1h | 150小时 | 6.25天 |
| 4h | 600小时 | 25天 |
| 1d | 150天 | 5个月 |

**示例 (5分钟K线):**
- 获取150根K线 = 最近 750分钟 = 12.5小时的数据
- SMA120 = 最近120根K线 = 600分钟 = 10小时的移动平均

---

## ⚙️ 配置建议

### 当前配置 (默认)
```go
获取数量: 150根
最低要求: 120根
余量: 30根
```

### 如果要修改

**增加到200根:**
```go
// main.go:188
candles, err := bot.hlClient.GetCandlestickData(symbol, bot.config.Trading.Timeframe, 200)
```

**优点:**
- ✅ 更多历史数据
- ✅ EMA计算更平滑
- ✅ 更大的安全余量

**缺点:**
- ❌ API请求数据量增加
- ❌ 计算时间略微增加

---

## 📊 内存占用估算

### 单次计算的数据量

假设使用5分钟K线,150根:

```
数据结构:
- 每根K线: ~80字节 (6个float64 + 1个int64)
- 150根K线: 150 × 80 = 12KB

衍生数组:
- closes: 150 × 8 = 1.2KB
- highs: 150 × 8 = 1.2KB
- lows: 150 × 8 = 1.2KB
- volumes: 150 × 8 = 1.2KB
- 各种EMA序列: ~2KB

总计: 约 20KB / 每个币种
```

**3个币种 (ETH, BTC, DOGE):**
- 总内存: 20KB × 3 = 60KB
- 可忽略不计 ✅

---

## 🎯 总结

| 项目 | 数值 | 说明 |
|------|------|------|
| **获取数量** | **150根** | 每次API调用获取的K线数 |
| **最低要求** | **120根** | 系统运行的最低要求 |
| **安全余量** | **30根** | 150 - 120 |
| **最长指标** | **SMA120/EMA120** | 需要120根K线 |
| **实际使用** | **全部150根** | 所有指标计算都基于这150根 |

### 关键点

1. ✅ 获取150根K线数据
2. ✅ 最低需要120根 (因为SMA120/EMA120)
3. ✅ 每个指标使用其所需的周期长度
4. ✅ 提供30根余量确保稳定性
5. ✅ 适用于所有配置的时间框架 (1m-1d)

---

**文件位置:**
- K线获取: `main.go:188`
- 最低检查: `main.go:193-195`
- 指标计算: `indicators/calculator.go:59-110`

**配置文件:**
- `config.yaml` - 设置时间框架 (timeframe)
