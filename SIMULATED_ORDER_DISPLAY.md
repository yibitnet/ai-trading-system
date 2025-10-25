# 模拟开仓信息显示功能

## ✅ 功能已完成

现在在模拟模式下,当AI决定开仓时,系统会显示详细的开仓信息!

---

## 🎯 显示内容

### 完整信息包括:

1. **📊 基本信息**
   - 币种 (ETH/BTC/DOGE等)
   - 方向 (🟢多单 / 🔴空单)
   - 杠杆倍数
   - 开仓价格

2. **💰 仓位信息**
   - 开仓数量 (例如: 5000.0000 DOGE)
   - 仓位价值 (例如: $1000.00 USDT)
   - 占用资金比例 (例如: 10.0%)
   - 账户余额
   - 实际曝光 (如使用杠杆)

3. **🎯 风险管理**
   - 止损价格和百分比
   - 止盈价格和百分比
   - 风险等级 (🟢低 / 🟡中 / 🔴高)
   - 预期持仓时间

4. **📝 开仓理由**
   - AI的完整技术分析说明

---

## 📊 显示示例

### 示例1: DOGE多单

```
================================================================================
  💡 SIMULATED ORDER (模拟开仓)
================================================================================

📊 币种: DOGE
   方向: 🟢 多单 (LONG)
   杠杆: 5x
   价格: $0.20

💰 仓位信息:
   数量: 5000.0000 DOGE
   价值: $1000.00 USDT
   占用资金: 10.0% (账户余额: $10000.00)
   实际曝光: $5000.00 USDT (5x杠杆)

🎯 风险管理:
   止损: $0.19 (5.00%)
   止盈: $0.22 (10.00%)
   风险等级: 🟡 MEDIUM (中风险)
   预期持仓: 中期 (1-7天)

📝 开仓理由:
   市场显示明确的上涨信号,多个技术指标共振。RSI从超卖区回升,
   MACD金叉且柱状图转正...

⚠️  注意: 这是模拟开仓,未执行真实交易
   要启用真实交易,请设置 config.yaml 中的 trading_enabled: true
================================================================================
```

### 示例2: ETH空单

```
================================================================================
  💡 SIMULATED ORDER (模拟开仓)
================================================================================

📊 币种: ETH
   方向: 🔴 空单 (SHORT)
   杠杆: 3x
   价格: $3950.00

💰 仓位信息:
   数量: 0.2532 ETH
   价值: $1000.00 USDT
   占用资金: 10.0% (账户余额: $10000.00)
   实际曝光: $3000.00 USDT (3x杠杆)

🎯 风险管理:
   止损: $4050.00 (2.53%)
   止盈: $3750.00 (5.06%)
   风险等级: 🟢 LOW (低风险)
   预期持仓: 短期 (几小时)

📝 开仓理由:
   价格触及布林带上轨,RSI超买(>70),MACD死叉确认...

⚠️  注意: 这是模拟开仓,未执行真实交易
   要启用真实交易,请设置 config.yaml 中的 trading_enabled: true
================================================================================
```

---

## 🔧 显示逻辑

### 何时显示

- ✅ AI决定 OPEN_LONG (开多单)
- ✅ AI决定 OPEN_SHORT (开空单)
- ✅ AI决定 ADD_POSITION (加仓)
- ❌ AI决定 HOLD (持有) - 不显示
- ❌ AI决定 CLOSE_POSITION (平仓) - 不显示

### 显示顺序

```
1. AI决策报告 (包含所有技术指标分析)
   ↓
2. 风险控制检查
   ↓
3. 模拟开仓信息 ⬅️ 新增
   ↓
4. 日志记录
```

### 计算说明

**数量计算:**
```
positionValue = accountBalance × decision.Size
coinAmount = positionValue ÷ currentPrice

例如:
账户余额 = $10,000
决策仓位 = 10% (0.1)
DOGE价格 = $0.20

positionValue = 10000 × 0.1 = $1000
coinAmount = 1000 ÷ 0.20 = 5000 DOGE
```

**杠杆曝光:**
```
actualExposure = positionValue × leverage

例如:
仓位价值 = $1000
杠杆 = 5x

actualExposure = 1000 × 5 = $5000
```

---

## 🎨 视觉元素

### Emoji说明

| Emoji | 含义 |
|-------|------|
| 💡 | 模拟订单 |
| 📊 | 交易信息 |
| 🟢 | 多单/盈利 |
| 🔴 | 空单/亏损 |
| 💰 | 资金信息 |
| 🎯 | 风险管理 |
| 📝 | 说明文本 |
| ⚠️ | 警告提示 |

### 风险等级颜色

- 🟢 LOW (低风险)
- 🟡 MEDIUM (中风险)
- 🔴 HIGH (高风险)

### 持仓周期

- **SHORT**: 短期 (几小时)
- **MEDIUM**: 中期 (1-7天)
- **LONG**: 长期 (>7天)

---

## 💡 使用场景

### 场景1: 学习AI决策逻辑

在模拟模式下观察AI如何:
- 选择开仓时机
- 确定仓位大小
- 设置杠杆倍数
- 计算止损止盈

### 场景2: 验证风险控制

查看系统如何:
- 限制单笔仓位比例
- 调整杠杆倍数
- 计算风险暴露
- 设置保护价格

### 场景3: 模拟交易日志

记录模拟交易以便:
- 回测AI表现
- 分析决策质量
- 优化参数设置
- 评估收益风险比

### 场景4: 教育培训

用于:
- 理解加密货币交易
- 学习技术分析
- 掌握风险管理
- 熟悉交易流程

---

## 🔄 与真实交易的对比

### 模拟模式 (trading_enabled: false)

```
AI决策: OPEN_LONG
↓
风险检查: 通过
↓
💡 显示模拟开仓信息
↓
日志: "Simulated trade"
↓
❌ 不执行真实订单
```

### 真实交易 (trading_enabled: true)

```
AI决策: OPEN_LONG
↓
风险检查: 通过
↓
✅ 执行真实交易
↓
下单到交易所
↓
日志: "Trade executed" + order_id
```

---

## 📝 代码实现

### 新增函数

**printSimulatedOrder()** - 显示模拟开仓信息
```go
func (bot *TradingBot) printSimulatedOrder(
    decision *ai.Decision,
    market *hyperliquid.MarketInfo,
    balance float64,
    symbol string,
) {
    // 计算仓位详情
    positionValue := balance * decision.Size
    coinAmount := positionValue / market.CurrentPrice

    // 格式化输出
    fmt.Printf("数量: %.4f %s\n", coinAmount, symbol)
    fmt.Printf("价值: $%.2f USDT\n", positionValue)
    // ...
}
```

**formatHoldingPeriod()** - 格式化持仓周期
```go
func formatHoldingPeriod(period string) string {
    switch period {
    case "SHORT":
        return "短期 (几小时)"
    case "MEDIUM":
        return "中期 (1-7天)"
    case "LONG":
        return "长期 (>7天)"
    }
}
```

### 调用位置

在 `main.go` 的 `executeDecision()` 函数中:

```go
if !bot.config.Trading.TradingEnabled {
    bot.logger.Warn("Trading is disabled - simulation mode")

    // 显示模拟开仓信息
    bot.printSimulatedOrder(decision, marketInfo, balance, symbol)

    bot.logger.Info("Simulated trade")
    return nil
}
```

---

## 🧪 测试

### 查看演示

```bash
./demo_simulated_order.sh
```

### 实际运行

```bash
# 确保模拟模式
grep "trading_enabled:" config.yaml
# 应显示: trading_enabled: false

# 启动系统
./aitrading

# 等待AI做出开仓决策
# 系统会自动显示模拟开仓信息
```

### 保存模拟记录

```bash
# 重定向输出
./aitrading | tee simulated_trades.log

# 查看所有模拟开仓
grep -A 20 "SIMULATED ORDER" simulated_trades.log
```

---

## 📊 信息完整度对比

| 信息类型 | 之前 | 现在 |
|---------|------|------|
| 操作类型 | ✅ | ✅ |
| 币种 | ✅ | ✅ |
| 方向 | ❌ | ✅ 明确显示 |
| 杠杆 | ✅ 日志 | ✅ 显著显示 |
| 数量(币) | ❌ | ✅ 计算显示 |
| 价值(USDT) | ❌ | ✅ 计算显示 |
| 实际曝光 | ❌ | ✅ 杠杆曝光 |
| 止损止盈 | ✅ | ✅ 百分比 |
| 风险等级 | ✅ | ✅ emoji |
| 开仓理由 | ✅ | ✅ 格式化 |

---

## 💡 实用技巧

### 技巧1: 保存模拟交易记录

```bash
# 创建交易日志目录
mkdir -p simulation_logs

# 运行并保存
./aitrading | tee "simulation_logs/sim_$(date +%Y%m%d_%H%M%S).log"
```

### 技巧2: 分析模拟表现

```bash
# 统计开仓次数
grep -c "SIMULATED ORDER" simulation_logs/*.log

# 查看所有多单
grep -A 20 "多单 (LONG)" simulation_logs/*.log

# 查看高杠杆交易
grep -A 20 "杠杆: [5-9]x\|杠杆: 10x" simulation_logs/*.log
```

### 技巧3: 提取关键信息

```bash
# 提取所有模拟开仓的币种和方向
grep -E "币种:|方向:" simulation_logs/*.log

# 提取风险等级分布
grep "风险等级:" simulation_logs/*.log | sort | uniq -c
```

---

## 🎯 下一步

### 完成模拟测试后

1. **分析结果**
   - 开仓频率是否合理
   - 杠杆使用是否过高
   - 风险等级分布如何

2. **调整参数**
   ```yaml
   trading:
     min_confidence: 0.7  # 提高阈值
     max_leverage: 3      # 降低杠杆
   ```

3. **启用真实交易**
   ```bash
   ./enable_trading.sh
   ```

4. **小额试运行**
   先用少量资金验证

---

## 📚 相关文档

- `ENABLE_TRADING.md` - 启用真实交易指南
- `CLI_REPORT_GUIDE.md` - CLI报告使用指南
- `QUICKSTART.md` - 快速开始
- `config.yaml` - 配置文件

---

## 🎉 总结

**新增功能:**
- ✅ 模拟开仓详细信息显示
- ✅ 包含币种、方向、杠杆、数量
- ✅ 显示USDT价值和实际曝光
- ✅ 完整的风险管理信息
- ✅ 美观的格式化输出

**使用方式:**
```bash
# 确保模拟模式
grep "trading_enabled: false" config.yaml

# 启动系统
./aitrading

# 观察模拟开仓信息
```

**演示:**
```bash
./demo_simulated_order.sh
```

---

**更新时间**: 2025-10-25
**版本**: v1.4
**功能状态**: ✅ 已完成并测试通过
