# 真实交易启用指南

## ⚠️ 重要警告

**当前系统处于模拟模式,不会执行真实交易!**

虽然AI会做出交易决策(OPEN_LONG/OPEN_SHORT等),但由于`trading_enabled: false`,系统只会**模拟**这些操作,不会向交易所发送真实订单。

---

## 📊 当前状态

```yaml
# config.yaml
trading:
  trading_enabled: false  # ⬅️ 模拟模式
```

**日志中会显示:**
```
level=warning msg="Trading is disabled - simulation mode"
level=info msg="Simulated trade" action=OPEN_LONG
```

---

## 🔧 如何启用真实交易

### 方法1: 使用脚本(推荐)

```bash
./enable_trading.sh
```

脚本会:
1. 显示当前设置
2. 要求确认
3. 备份配置文件
4. 修改trading_enabled为true
5. 显示警告和建议

### 方法2: 手动编辑

```bash
# 1. 备份配置
cp config.yaml config.yaml.backup

# 2. 编辑配置文件
nano config.yaml
# 或
vi config.yaml

# 3. 修改这一行:
#   trading_enabled: false
# 改为:
#   trading_enabled: true

# 4. 保存并退出

# 5. 验证修改
grep "trading_enabled:" config.yaml
```

### 方法3: 使用sed命令

```bash
# 备份
cp config.yaml config.yaml.backup

# 一键修改
sed -i 's/trading_enabled: false/trading_enabled: true/' config.yaml

# 验证
grep "trading_enabled:" config.yaml
```

---

## ✅ 启用后的变化

### 模拟模式 (trading_enabled: false)
```
AI决策: OPEN_LONG, Confidence: 85%
↓
风险检查: 通过
↓
⚠️  模拟交易 - 不执行真实订单
↓
日志: "Simulated trade"
```

### 真实交易 (trading_enabled: true)
```
AI决策: OPEN_LONG, Confidence: 85%
↓
风险检查: 通过
↓
✅ 执行真实交易
↓
下单到Hyperliquid交易所
↓
日志: "Trade executed" order_id=xxx
```

---

## 📋 启用前检查清单

在启用真实交易前,请确保:

- [ ] **账户余额充足**
  ```bash
  ./aitrading balance
  ```

- [ ] **API配置正确**
  ```yaml
  hyperliquid:
    api_url: "https://api.hyperliquid.xyz"
    private_key: "0x..." # ⬅️ 检查是否正确
    account_address: "0x..." # ⬅️ 检查是否正确
    testnet: false # 或 true(测试网)
  ```

- [ ] **理解风险参数**
  ```yaml
  trading:
    max_position_size: 0.1  # 单笔最大10%
    max_open_positions: 2   # 最多2个仓位
    max_leverage: 10        # 最高10倍杠杆

  risk:
    max_drawdown: 0.05      # 最大回撤5%
    daily_loss_limit: 0.02  # 每日亏损限制2%
  ```

- [ ] **AI置信度阈值**
  ```yaml
  trading:
    min_confidence: 0.7  # AI置信度至少70%才交易
  ```

- [ ] **已充分测试模拟模式**
  ```bash
  # 查看历史模拟交易日志
  grep "Simulated trade" logs/trading.log
  ```

---

## 🧪 推荐的启用流程

### 第1步: 使用测试网(如果支持)
```yaml
hyperliquid:
  testnet: true  # 先在测试网试运行
  trading_enabled: true
```

### 第2步: 小额真实测试
```yaml
trading:
  trading_enabled: true
  max_position_size: 0.01  # 只用1%资金
  max_open_positions: 1    # 只开1个仓位
  max_leverage: 1          # 不使用杠杆
```

### 第3步: 密切监控第一笔交易
```bash
# 终端1: 运行交易系统
./aitrading

# 终端2: 实时查看日志
tail -f logs/trading.log

# 终端3: 定期检查仓位
watch -n 60 './aitrading order'
```

### 第4步: 逐步增加规模
确认系统稳定运行后,逐步调整:
```yaml
max_position_size: 0.05  # 5%
max_open_positions: 2
max_leverage: 3
```

---

## 📊 真实交易日志示例

启用后,日志会显示:

```
level=info msg="Risk check completed" approved=true
level=info msg="Executing trading decision" action=OPEN_LONG confidence=0.85
level=info msg="Opening long position" symbol=ETH size=0.1 price=3950.00
level=info msg="Long position opened" order_id=abc123 success=true
level=info msg="Trade executed"
  action=OPEN_LONG
  side=LONG
  size=0.1
  price=3950.00
  order_id=abc123
```

---

## ⚠️ 关键安全提示

### 1. 私钥安全
```bash
# 确保config.yaml权限正确
chmod 600 config.yaml

# 不要在公开仓库提交私钥
echo "config.yaml" >> .gitignore
```

### 2. 使用环境变量(更安全)
```yaml
hyperliquid:
  private_key: "${HYPERLIQUID_PRIVATE_KEY}"
  account_address: "${HYPERLIQUID_ADDRESS}"
```

```bash
export HYPERLIQUID_PRIVATE_KEY="0x..."
export HYPERLIQUID_ADDRESS="0x..."
./aitrading
```

### 3. 止损机制
系统已内置多重保护:
- AI设置的止损价
- 每日亏损限制
- 最大回撤限制
- 持仓时间限制

### 4. 监控告警
建议设置:
```bash
# 定期检查异常交易
grep "error\|failed" logs/trading.log

# 监控大额亏损
./aitrading order | grep "P&L.*-"
```

---

## 🔄 如何恢复模拟模式

如需关闭真实交易:

```bash
# 方法1: 编辑配置
nano config.yaml
# 改回 trading_enabled: false

# 方法2: 使用sed
sed -i 's/trading_enabled: true/trading_enabled: false/' config.yaml

# 方法3: 恢复备份
cp config.yaml.backup config.yaml
```

---

## 📞 常见问题

### Q1: 为什么AI决策了但没有交易?
**A:** 检查`trading_enabled: false`,这是模拟模式

### Q2: 如何确认交易已执行?
**A:** 查看日志中的`order_id`和`success=true`,或使用:
```bash
./aitrading order  # 查看实际持仓
```

### Q3: 交易失败怎么办?
**A:** 检查:
- 账户余额是否充足
- API密钥和私钥是否正确
- 网络连接是否正常
- 日志中的具体错误信息

### Q4: 可以手动停止自动交易吗?
**A:** 可以,按`Ctrl+C`停止程序,或:
```bash
sed -i 's/trading_enabled: true/trading_enabled: false/' config.yaml
```
然后重启

---

## 📚 相关文档

- `README.md` - 系统整体说明
- `QUICKSTART.md` - 快速开始
- `CLI_COMMANDS.md` - 命令使用指南
- `logs/trading.log` - 交易日志

---

## 🎯 总结

**当前状态:**
- ❌ 真实交易: **禁用**
- ✅ 模拟模式: **激活**

**启用步骤:**
1. 运行 `./enable_trading.sh`
2. 或手动修改 `trading_enabled: true`
3. 重启 `./aitrading`
4. 监控第一笔交易

**建议:**
- ✅ 先在模拟模式充分测试
- ✅ 使用小额资金开始
- ✅ 密切监控初期交易
- ✅ 逐步增加规模

---

**完成时间**: 2025-10-25
**状态**: 交易功能已实现,默认安全模拟模式
