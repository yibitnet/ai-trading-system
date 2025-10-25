# AI Trading System

[中文](README.md) | English

An AI-powered cryptocurrency automated trading system supporting Hyperliquid exchange.

## ✨ Core Features

- 🤖 **AI Decision Engine** - Support for DeepSeek and Qwen AI models
- 📊 **Multi-Symbol Trading** - Monitor multiple cryptocurrencies (ETH, BTC, DOGE, etc.)
- 📈 **Technical Analysis** - SMA/EMA, MACD, RSI, Bollinger Bands, and more
- ⚡ **Real-time Monitoring** - Automated trading cycles with configurable intervals
- 🛡️ **Risk Control** - Stop-loss/take-profit, position limits, leverage control
- 💻 **CLI Interface** - Compact reports, color-coded, comprehensive information display
- 🎯 **Simulation Mode** - Safe testing without executing real trades

## 🚀 Quick Start

### 1. Build
```bash
./build.sh
```

### 2. Configuration
Edit `config.yaml`:
```yaml
trading:
  symbols: ["ETH", "BTC", "DOGE"]
  timeframe: "5m"
  trading_enabled: false  # Simulation mode

ai:
  provider: "deepseek"
  api_key: "your-api-key"
```

### 3. Run
```bash
# Start trading system (simulation mode)
./aitrading

# View current positions
./aitrading order

# View account balance
./aitrading balance

# Show help
./aitrading help
```

## 📚 Documentation

See **[DOCS.md](DOCS.md)** for complete documentation index.

### Main Documentation
- [QUICKSTART.md](QUICKSTART.md) - Quick start guide
- [CLI_COMMANDS.md](CLI_COMMANDS.md) - CLI commands usage
- [ENABLE_TRADING.md](ENABLE_TRADING.md) - Enable live trading
- [PRICE_PRECISION.md](PRICE_PRECISION.md) - Price precision details
- [KLINE_USAGE.md](KLINE_USAGE.md) - K-line data usage

## ⚙️ System Architecture

```
aitrading/
├── main.go              # Main program
├── config.yaml          # Configuration file
├── ai/                  # AI decision module
├── config/              # Configuration module
├── executor/            # Trade execution module
├── hyperliquid/         # Exchange API
├── indicators/          # Technical indicators
└── risk/               # Risk control
```

## 🎯 Technical Indicators

The system uses 150 K-lines to calculate the following indicators:

| Indicator | Period | Purpose |
|-----------|--------|---------|
| SMA | 10/60/120 | Trend identification |
| EMA | 10/60/120 | Trend identification |
| MACD | 12/26/9 | Momentum analysis |
| RSI | 14 | Overbought/oversold |
| Bollinger Bands | 20 | Volatility |
| VMA | 20 | Volume analysis |

## 🛡️ Risk Management

- ✅ Automatic stop-loss and take-profit
- ✅ Maximum position size limits
- ✅ Leverage ceiling control
- ✅ Daily loss limits
- ✅ Maximum drawdown protection
- ✅ Position quantity limits

## ⚠️ Important Notice

**Simulation mode is enabled by default - no real trades will be executed!**

To enable live trading:
1. Read [ENABLE_TRADING.md](ENABLE_TRADING.md)
2. Test thoroughly in simulation mode
3. Run `./enable_trading.sh`

## 📊 CLI Report Example

```
================================================================================
  📊 AI Trading Decision Report - DOGE @ 2025-10-25 12:30:00
================================================================================

📈 Market: DOGE $0.18452 | 24h: 🟢2.15% | Vol: $456.78M
📐 Indicators: Trend:🟢BULL | Momentum:🚀BULL | RSI:65.5(Normal)
              MACD:🟢2.1111 | BB:MIDDLE | Vol:🟢BUY

💼 Position: None

🤖 Decision: 🟢 OPEN LONG | Conf:85% | Size:10.0% | Lev:5x | SL:$0.18145 | TP:$0.19203 | Risk:🟡MED

💭 Reasoning:
  Market shows clear uptrend signals, multiple technical indicators aligned...
================================================================================
```

## 🔧 Configuration Options

### Trading Configuration
```yaml
trading:
  symbols: ["ETH", "BTC", "DOGE"]  # Trading symbols
  timeframe: "5m"                   # K-line period
  interval: "5m"                    # Trading cycle
  max_position_size: 0.1            # Max position 10%
  max_open_positions: 2             # Max 2 positions
  max_leverage: 10                  # Max 10x leverage
  trading_enabled: false            # Simulation mode
```

### Risk Configuration
```yaml
risk:
  max_drawdown: 0.05               # Max drawdown 5%
  daily_loss_limit: 0.02           # Daily loss limit 2%
  max_position_hold_time: "24h"    # Max holding time
```

## 📈 Supported AI Models

- **DeepSeek** - deepseek-chat
- **Qwen** - qwen-max (Tongyi Qianwen)

## 🔗 Related Links

- [Hyperliquid](https://hyperliquid.xyz) - Exchange
- [DeepSeek](https://api.deepseek.com) - AI API
- [Qwen](https://dashscope.aliyuncs.com) - Qwen API

## 📝 License

MIT License

## 🤝 Contributing

Issues and Pull Requests are welcome!

---

**Version**: v1.6
**Last Updated**: 2025-10-25
