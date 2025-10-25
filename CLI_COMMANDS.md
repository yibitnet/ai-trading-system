# CLI命令功能说明

## 新增功能

系统现在支持多个命令行参数,可以快速查询账户信息!

---

## 📋 可用命令

### 1. 启动交易机器人
```bash
./aitrading
```
启动自动交易系统,每5分钟执行一次交易决策

### 2. 查看当前仓位
```bash
./aitrading order
# 或
./aitrading orders
./aitrading position
./aitrading positions
```
显示所有币种的当前持仓情况

### 3. 查看账户余额
```bash
./aitrading balance
```
显示账户余额和资金使用情况

### 4. 查看帮助信息
```bash
./aitrading help
# 或
./aitrading -h
./aitrading --help
```
显示所有可用命令和使用说明

---

## 📊 命令输出示例

### `./aitrading order` - 查看仓位

**无持仓时:**
```
================================================================================
  💼 Current Positions - 2025-10-25 08:21:34
================================================================================

💰 Account Balance: $1,250.50

--------------------------------------------------------------------------------

📭 No open positions

================================================================================
```

**有持仓时:**
```
================================================================================
  💼 Current Positions - 2025-10-25 08:21:34
================================================================================

💰 Account Balance: $10,000.00

--------------------------------------------------------------------------------

📊 ETH Position:
  Side:         🟢 LONG
  Size:         0.5000
  Entry Price:  $3,900.00
  Current Price: $3,950.00
  Price Change: 🟢 1.28%
  Unrealized P&L: 🟢 $25.00 (1.28%)
  Exposure:     $1,950.00
  Holding Time: 2h15m30s
--------------------------------------------------------------------------------

📊 BTC Position:
  Side:         🔴 SHORT
  Size:         0.0200
  Entry Price:  $112,000.00
  Current Price: $111,500.00
  Price Change: 🟢 -0.45%
  Unrealized P&L: 🟢 $10.00 (0.45%)
  Exposure:     $2,240.00
  Holding Time: 45m20s
--------------------------------------------------------------------------------

📈 Summary:
  Total Exposure:  $4,190.00
  Total P&L:       🟢 $35.00
  Exposure Ratio:  41.90%

================================================================================
```

### `./aitrading balance` - 查看余额

**无持仓时:**
```
================================================================================
  💰 Account Balance - 2025-10-25 08:21:42
================================================================================

💵 Available Balance: $10,000.00

================================================================================
```

**有持仓时:**
```
================================================================================
  💰 Account Balance - 2025-10-25 08:21:42
================================================================================

💵 Available Balance: $10,000.00
💼 Total Exposure:    $4,190.00
📊 Exposure Ratio:    41.90%
💵 Free Balance:      $5,810.00

================================================================================
```

---

## 🎯 使用场景

### 场景1: 快速查看持仓
在不启动交易机器人的情况下,快速检查当前仓位:
```bash
./aitrading order
```

### 场景2: 监控账户状态
定期检查账户余额和资金使用率:
```bash
./aitrading balance
```

### 场景3: 组合使用
先查看持仓,再决定是否启动交易:
```bash
./aitrading order
./aitrading balance
# 如果满意,启动交易
./aitrading
```

### 场景4: 脚本自动化
在脚本中定期查询:
```bash
#!/bin/bash
while true; do
    echo "=== $(date) ==="
    ./aitrading order
    sleep 300  # 每5分钟查询一次
done
```

---

## 📝 详细字段说明

### 仓位信息
- **Side**: 持仓方向
  - 🟢 LONG = 多单(看涨)
  - 🔴 SHORT = 空单(看跌)
- **Size**: 持仓数量(以币种计)
- **Entry Price**: 开仓价格
- **Current Price**: 当前市场价格
- **Price Change**: 从开仓到现在的价格变化
- **Unrealized P&L**: 未实现盈亏
  - 🟢 = 盈利
  - 🔴 = 亏损
- **Exposure**: 仓位占用的资金(Size × Entry Price)
- **Holding Time**: 持仓时长

### 账户余额
- **Available Balance**: 账户总余额
- **Total Exposure**: 所有持仓占用的总资金
- **Exposure Ratio**: 资金使用率(Exposure / Balance × 100%)
- **Free Balance**: 可用余额(Balance - Exposure)

### 汇总信息
- **Total Exposure**: 所有币种的持仓资金总和
- **Total P&L**: 所有仓位的盈亏总和
- **Exposure Ratio**: 总体资金使用率

---

## 🔧 命令别名

为了使用方便,提供了多个别名:

| 主命令 | 别名 | 说明 |
|--------|------|------|
| order | orders, position, positions | 查看仓位 |
| help | -h, --help | 帮助信息 |

示例:
```bash
./aitrading order      # ✓
./aitrading orders     # ✓ 相同效果
./aitrading position   # ✓ 相同效果
./aitrading positions  # ✓ 相同效果
```

---

## ⚡ 快速命令参考

```bash
# 查看所有命令
./aitrading help

# 查看当前持仓
./aitrading order

# 查看账户余额
./aitrading balance

# 启动交易机器人
./aitrading
```

---

## 🎨 视觉元素说明

### Emoji含义
- 💼 = 仓位/持仓
- 💰 = 账户余额
- 💵 = 可用资金
- 📊 = 统计数据
- 📈 = 汇总信息
- 📭 = 无持仓
- 🟢 = 多单/盈利/上涨
- 🔴 = 空单/亏损/下跌
- 🤖 = 系统/AI
- ✅ = 成功
- ❌ = 错误

---

## 💡 实用技巧

### 技巧1: 创建快捷别名
在`.bashrc`或`.zshrc`中添加:
```bash
alias pos='cd /root/go/src/aitrading && ./aitrading order'
alias bal='cd /root/go/src/aitrading && ./aitrading balance'
alias trade='cd /root/go/src/aitrading && ./aitrading'
```

然后可以在任何目录执行:
```bash
pos   # 查看仓位
bal   # 查看余额
trade # 启动交易
```

### 技巧2: 实时监控仓位
```bash
watch -n 60 './aitrading order'
```
每60秒自动刷新仓位信息

### 技巧3: 保存仓位历史
```bash
./aitrading order >> position_history.log
```
将仓位信息追加到日志文件

### 技巧4: 条件告警
```bash
#!/bin/bash
pnl=$(./aitrading order | grep "Total P&L" | awk '{print $4}')
if [ "$pnl" -lt "-100" ]; then
    echo "Warning: P&L below -$100!" | mail -s "Trading Alert" your@email.com
fi
```

---

## 🔍 故障排除

### 问题1: 命令不存在
**错误**: `bash: ./aitrading: No such file or directory`
**解决**:
```bash
cd /root/go/src/aitrading
go build -o aitrading
```

### 问题2: 权限不足
**错误**: `Permission denied`
**解决**:
```bash
chmod +x aitrading
```

### 问题3: 配置文件错误
**错误**: `Failed to load configuration`
**解决**: 确保`config.yaml`存在且格式正确

### 问题4: API连接失败
**错误**: `Failed to fetch position`
**解决**: 检查网络连接和API地址配置

---

## 🚀 高级用法

### 组合命令查看完整信息
```bash
{
    echo "=== Trading Status ==="
    echo ""
    ./aitrading balance
    echo ""
    ./aitrading order
} | less
```

### 导出为JSON格式
虽然当前输出是文本格式,你可以创建自定义脚本解析输出:
```bash
./aitrading order | grep -E "(Balance|Side|P&L)" > status.txt
```

### Cron定时查询
添加到crontab:
```bash
# 每小时查询一次仓位
0 * * * * cd /root/go/src/aitrading && ./aitrading order >> ~/trading_log.txt
```

---

## 📊 与交易机器人的区别

| 特性 | `./aitrading` | `./aitrading order` |
|------|---------------|---------------------|
| 功能 | 启动交易机器人 | 查询当前仓位 |
| 执行时间 | 持续运行 | 立即返回 |
| 自动交易 | ✅ 是 | ❌ 否 |
| 显示决策 | ✅ 是 | ❌ 否 |
| 适用场景 | 自动化交易 | 快速查询 |

---

## 📦 实现细节

### 代码位置
- 文件: `main.go`
- 函数:
  - `showPositions()` - 显示仓位
  - `showBalance()` - 显示余额
  - `showHelp()` - 显示帮助

### 命令解析
```go
if len(os.Args) > 1 {
    command := os.Args[1]
    switch command {
        case "order", "orders", "position", "positions":
            showPositions()
        case "balance":
            showBalance()
        case "help", "-h", "--help":
            showHelp()
    }
}
```

---

## 🎉 总结

新增的CLI命令功能让您可以:
- ✅ 快速查看当前持仓
- ✅ 实时监控账户余额
- ✅ 无需启动机器人即可获取信息
- ✅ 美观的格式化输出
- ✅ 支持多个命令别名
- ✅ 完善的帮助系统

**立即尝试:**
```bash
./aitrading order
```

---

**更新时间**: 2025-10-25 08:22
**版本**: v1.3
**功能状态**: ✅ 已完成并测试通过
