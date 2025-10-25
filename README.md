# AI Trading System

一个基于AI的加密货币自动交易系统,支持Hyperliquid交易所。

## ✨ 核心特性

- 🤖 **AI决策引擎** - 支持DeepSeek和Qwen AI模型
- 📊 **多币种交易** - 同时监控ETH、BTC、DOGE等多个币种
- 📈 **技术指标分析** - SMA/EMA、MACD、RSI、布林带等
- ⚡ **实时监控** - 自动化交易周期,可配置时间间隔
- 🛡️ **风险控制** - 止损止盈、仓位限制、杠杆控制
- 💻 **CLI界面** - 紧凑型报告、颜色编码、完整信息展示
- 🎯 **模拟模式** - 安全测试,不执行真实交易

## 🚀 快速开始

### 1. 编译
```bash
./build.sh
```

### 2. 配置
编辑 `config.yaml`:
```yaml
trading:
  symbols: ["ETH", "BTC", "DOGE"]
  timeframe: "5m"
  trading_enabled: false  # 模拟模式

ai:
  provider: "deepseek"
  api_key: "your-api-key"
```

### 3. 运行
```bash
# 启动交易系统 (模拟模式)
./aitrading

# 查看当前仓位
./aitrading order

# 查看账户余额
./aitrading balance

# 查看帮助
./aitrading help
```

## 📚 完整文档

详细文档请查看 **[DOCS.md](DOCS.md)** - 文档索引

### 主要文档
- [QUICKSTART.md](QUICKSTART.md) - 快速开始指南
- [CLI_COMMANDS.md](CLI_COMMANDS.md) - CLI命令使用
- [ENABLE_TRADING.md](ENABLE_TRADING.md) - 启用真实交易
- [PRICE_PRECISION.md](PRICE_PRECISION.md) - 价格精度说明
- [KLINE_USAGE.md](KLINE_USAGE.md) - K线数据说明

## ⚙️ 系统架构

```
aitrading/
├── main.go              # 主程序
├── config.yaml          # 配置文件
├── ai/                  # AI决策模块
├── config/              # 配置模块
├── executor/            # 交易执行模块
├── hyperliquid/         # 交易所API
├── indicators/          # 技术指标
└── risk/               # 风险控制
```

## 🎯 技术指标

系统使用150根K线数据计算以下指标:

| 指标 | 周期 | 用途 |
|------|------|------|
| SMA | 10/60/120 | 趋势判断 |
| EMA | 10/60/120 | 趋势判断 |
| MACD | 12/26/9 | 动量分析 |
| RSI | 14 | 超买超卖 |
| 布林带 | 20 | 波动性 |
| VMA | 20 | 成交量分析 |

## 🛡️ 风险管理

- ✅ 自动止损止盈
- ✅ 最大仓位限制
- ✅ 杠杆上限控制
- ✅ 每日亏损限制
- ✅ 最大回撤保护
- ✅ 持仓数量限制

## ⚠️ 重要提示

**默认为模拟模式,不会执行真实交易!**

要启用真实交易:
1. 阅读 [ENABLE_TRADING.md](ENABLE_TRADING.md)
2. 在模拟模式下充分测试
3. 运行 `./enable_trading.sh`

## 📊 CLI报告示例

```
================================================================================
  📊 AI Trading Decision Report - DOGE @ 2025-10-25 12:30:00
================================================================================

📈 Market: DOGE $0.18452 | 24h: 🟢2.15% | Vol: $456.78M
📐 Indicators: Trend:🟢BULL | Momentum:🚀BULL | RSI:65.5(正常)
              MACD:🟢2.1111 | BB:MIDDLE | Vol:🟢BUY

💼 Position: None

🤖 Decision: 🟢 OPEN LONG | Conf:85% | Size:10.0% | Lev:5x | SL:$0.18145 | TP:$0.19203 | Risk:🟡MED

💭 Reasoning:
  市场显示明确的上涨信号,多个技术指标共振...
================================================================================
```

## 🔧 配置选项

### 交易配置
```yaml
trading:
  symbols: ["ETH", "BTC", "DOGE"]  # 交易币种
  timeframe: "5m"                   # K线周期
  interval: "5m"                    # 交易周期
  max_position_size: 0.1            # 最大仓位10%
  max_open_positions: 2             # 最多2个仓位
  max_leverage: 10                  # 最高10倍杠杆
  trading_enabled: false            # 模拟模式
```

### 风险配置
```yaml
risk:
  max_drawdown: 0.05               # 最大回撤5%
  daily_loss_limit: 0.02           # 每日亏损限制2%
  max_position_hold_time: "24h"    # 最长持仓时间
```

## 📈 支持的AI模型

- **DeepSeek** - deepseek-chat
- **Qwen** - qwen-max (通义千问)

## 🔗 相关链接

- [Hyperliquid](https://hyperliquid.xyz) - 交易所
- [DeepSeek](https://api.deepseek.com) - AI API
- [Qwen](https://dashscope.aliyuncs.com) - 通义千问API

## 📝 许可证

MIT License

## 🤝 贡献

欢迎提交Issue和Pull Request!

---

**版本**: v1.6
**最后更新**: 2025-10-25
