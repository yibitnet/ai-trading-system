#!/bin/bash

echo "=========================================="
echo "  启用真实交易功能"
echo "=========================================="
echo ""

echo "⚠️  警告: 此操作将启用真实交易!"
echo ""
echo "当前设置:"
grep "trading_enabled:" config.yaml
echo ""

read -p "确认启用真实交易? (输入 YES 继续): " confirm

if [ "$confirm" != "YES" ]; then
    echo "❌ 操作已取消"
    exit 1
fi

echo ""
echo "📝 备份当前配置..."
cp config.yaml config.yaml.backup
echo "✅ 已备份到 config.yaml.backup"

echo ""
echo "🔧 修改配置..."
sed -i 's/trading_enabled: false/trading_enabled: true/' config.yaml

echo ""
echo "新设置:"
grep "trading_enabled:" config.yaml

echo ""
echo "✅ 真实交易已启用!"
echo ""
echo "⚠️  重要提示:"
echo "  1. 确保账户有足够余额"
echo "  2. 检查API密钥和私钥配置正确"
echo "  3. 建议先用小资金测试"
echo "  4. 密切监控第一笔交易"
echo ""
echo "启动交易:"
echo "  ./aitrading"
echo ""
