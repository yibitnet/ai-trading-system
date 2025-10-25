# 快速开始指南

## 🎉 新功能已全部实现!

### ✅ 已完成功能

1. **Qwen3 Max模型支持** - 可随时切换AI模型
2. **多币种交易** - 同时监控ETH/BTC/DOGE等多个币种
3. **仓位数量限制** - 防止过度开仓
4. **AI杠杆建议** - AI会根据市场情况推荐杠杆倍数
5. **杠杆上限保护** - 自动限制杠杆不超过设定值
6. **紧凑型CLI报告** - 空间节省78%,颜色编码(开仓绿色,平仓红色)
7. **模拟开仓显示** - 详细的模拟交易信息展示
8. **CLI查询命令** - 快速查看仓位和余额

---

## 🚀 快速使用

### 1. 启动交易系统
```bash
./aitrading
```
系统每5分钟自动显示所有配置币种(ETH/BTC/DOGE)的决策报告

### 2. 查看当前仓位
```bash
./aitrading order
```
快速查看所有币种的持仓情况、盈亏和持仓时长

### 3. 查看账户余额
```bash
./aitrading balance
```
显示账户余额、资金使用率和可用余额

### 4. 查看帮助信息
```bash
./aitrading help
```
显示所有可用命令和使用说明

---

## 📋 命令速查

| 命令 | 说明 | 示例 |
|------|------|------|
| `./aitrading` | 启动交易机器人 | 持续运行,自动交易 |
| `./aitrading order` | 查看当前仓位 | 立即显示所有持仓 |
| `./aitrading balance` | 查看账户余额 | 显示余额和使用率 |
| `./aitrading help` | 显示帮助信息 | 查看所有命令 |

---

## 🔧 常见配置

### 切换到Qwen模型

**步骤1**: 设置环境变量
```bash
export QWEN_API_KEY="sk-your-qwen-api-key"
```

**步骤2**: 修改config.yaml
```yaml
ai:
  provider: "qwen"  # 改这里
```

**步骤3**: 重启系统
```bash
./aitrading
```

### 修改交易币种

编辑config.yaml:
```yaml
trading:
  symbols: ["ETH", "BTC", "SOL", "AVAX"]  # 添加或删除币种
```

### 调整风险参数

```yaml
trading:
  max_open_positions: 3  # 最多3个仓位
  max_leverage: 5        # 最高5倍杠杆(保守)
```

---

## 📊 测试结果

### 实际运行的市场数据 (2025-10-25 08:00)
```
ETH:  $3,928.95  |  -0.52%  |  Volume: $1.34B
BTC:  $111,551.50 |  +0.43%  |  Volume: $2.69B
DOGE: $0.20      |  +0.60%  |  Volume: $16.06M
```

### 风险控制验证
```
✅ 杠杆限制: 15倍 → 10倍 (自动调整)
✅ 仓位限制: 2个仓位已满,阻止第3个开仓
✅ AI决策: 包含leverage字段
```

---

## 📁 重要文件

- `config.yaml` - 主配置文件
- `ENHANCEMENT_REPORT.md` - 详细技术报告
- `test_features.go` - 功能测试脚本
- `logs/trading.log` - 运行日志

---

## 🎯 下一步

1. **启用实盘交易**:
   ```yaml
   trading:
     trading_enabled: true  # 改为true
   ```

2. **监控系统运行**:
   ```bash
   tail -f logs/trading.log
   ```

3. **根据需要调整配置**:
   - 添加/删除币种
   - 调整仓位和杠杆限制
   - 切换AI模型

---

## ⚠️ 注意事项

1. **切换Qwen前先测试**: 确保API key有效
2. **逐步增加币种**: 建议从2-3个币种开始
3. **杠杆要谨慎**: 建议从3-5倍开始,不要直接用10倍
4. **定期查看日志**: 确保所有币种都能正常获取数据

---

## 💡 示例配置

### 保守配置
```yaml
trading:
  symbols: ["ETH", "BTC"]  # 只交易主流币
  max_open_positions: 1     # 只开1个仓位
  max_leverage: 3           # 最高3倍杠杆
  trading_enabled: true
```

### 激进配置
```yaml
trading:
  symbols: ["ETH", "BTC", "SOL", "AVAX", "DOGE"]
  max_open_positions: 3     # 最多3个仓位
  max_leverage: 10          # 最高10倍杠杆
  trading_enabled: true
```

---

**祝交易顺利! 🎉**

如有问题,请查看:
- `logs/trading.log` - 运行日志
- `ENHANCEMENT_REPORT.md` - 技术文档
- `API_FIX_REPORT.md` - API修复文档
