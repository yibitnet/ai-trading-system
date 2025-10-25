# AI Trading System - 文档索引

## 📚 核心文档

### 快速开始
- **[README.md](README.md)** - 项目总览和系统介绍
- **[QUICKSTART.md](QUICKSTART.md)** - 快速开始指南

### 功能文档
- **[CLI_COMMANDS.md](CLI_COMMANDS.md)** - CLI命令使用指南
  - `./aitrading` - 启动交易系统
  - `./aitrading order` - 查看当前仓位
  - `./aitrading balance` - 查看账户余额

- **[CLI_REPORT_GUIDE.md](CLI_REPORT_GUIDE.md)** - CLI报告显示说明
  - 紧凑型报告格式
  - 颜色编码 (开仓绿色, 平仓红色)

- **[COMPACT_CLI_REPORT.md](COMPACT_CLI_REPORT.md)** - 紧凑CLI报告详细说明
  - 空间节省78%
  - 所有信息压缩显示

- **[SIMULATED_ORDER_DISPLAY.md](SIMULATED_ORDER_DISPLAY.md)** - 模拟开仓信息显示
  - 模拟模式下的开仓信息展示

- **[ENABLE_TRADING.md](ENABLE_TRADING.md)** - 真实交易启用指南
  - 如何从模拟模式切换到真实交易
  - 安全检查清单

- **[PRICE_PRECISION.md](PRICE_PRECISION.md)** - 价格精度显示说明
  - 保留原始价格精度
  - 自动去除尾部0

- **[KLINE_USAGE.md](KLINE_USAGE.md)** - K线数据使用说明
  - 使用150根K线
  - 各指标的周期说明

## 🔧 实用脚本

- **[build.sh](build.sh)** - 编译脚本
- **[start.sh](start.sh)** - 启动脚本
- **[enable_trading.sh](enable_trading.sh)** - 启用真实交易脚本
- **[demo_original_precision.sh](demo_original_precision.sh)** - 价格精度演示

## 📁 目录结构

```
aitrading/
├── main.go                          # 主程序
├── config.yaml                      # 配置文件
├── go.mod / go.sum                  # Go依赖管理
│
├── ai/                              # AI决策模块
│   └── decision_maker.go
├── config/                          # 配置模块
│   └── config.go
├── executor/                        # 交易执行模块
│   └── executor.go
├── hyperliquid/                     # Hyperliquid API
│   ├── client.go
│   └── trader.go
├── indicators/                      # 技术指标模块
│   └── calculator.go
├── risk/                           # 风险控制模块
│   └── controller.go
│
├── README.md                        # 项目说明
├── QUICKSTART.md                    # 快速开始
├── CLI_COMMANDS.md                  # CLI命令文档
├── CLI_REPORT_GUIDE.md             # CLI报告文档
├── COMPACT_CLI_REPORT.md           # 紧凑报告文档
├── SIMULATED_ORDER_DISPLAY.md      # 模拟开仓文档
├── ENABLE_TRADING.md               # 真实交易文档
├── PRICE_PRECISION.md              # 价格精度文档
├── KLINE_USAGE.md                  # K线使用文档
│
├── build.sh                         # 编译脚本
├── start.sh                         # 启动脚本
├── enable_trading.sh               # 启用交易脚本
└── demo_original_precision.sh      # 价格精度演示
```

## 🎯 常用命令

```bash
# 编译
./build.sh

# 启动交易系统
./aitrading

# 查看仓位
./aitrading order

# 查看余额
./aitrading balance

# 查看帮助
./aitrading help

# 启用真实交易
./enable_trading.sh
```

## 📖 阅读顺序建议

### 新用户
1. README.md - 了解系统
2. QUICKSTART.md - 快速开始
3. CLI_COMMANDS.md - 学习命令
4. SIMULATED_ORDER_DISPLAY.md - 了解模拟模式

### 准备启用真实交易
1. ENABLE_TRADING.md - 启用真实交易前必读
2. 在模拟模式下充分测试
3. 使用 `./enable_trading.sh` 启用

### 技术细节
1. KLINE_USAGE.md - K线数据说明
2. PRICE_PRECISION.md - 价格精度说明
3. CLI_REPORT_GUIDE.md - CLI显示说明

## ⚙️ 配置文件

主要配置在 `config.yaml`:

```yaml
trading:
  symbols: ["ETH", "BTC", "DOGE"]
  timeframe: "5m"
  trading_enabled: false  # 模拟模式

ai:
  provider: "deepseek"

risk:
  max_drawdown: 0.05
  daily_loss_limit: 0.02
```

## 🔗 相关链接

- Hyperliquid: https://hyperliquid.xyz
- DeepSeek API: https://api.deepseek.com
- Qwen API: https://dashscope.aliyuncs.com

---

**最后更新**: 2025-10-25
**版本**: v1.6
