#!/bin/bash

# AI Trading System - Quick Start Script

echo "================================"
echo "AI Trading System - Quick Start"
echo "================================"
echo ""

# Step 1: Check environment variables
echo "Step 1: Checking environment variables..."
MISSING_VARS=0

if [ -z "$DEEPSEEK_API_KEY" ]; then
    echo "  ❌ DEEPSEEK_API_KEY is not set"
    MISSING_VARS=1
else
    echo "  ✓ DEEPSEEK_API_KEY is set"
fi

if [ -z "$HYPERLIQUID_PRIVATE_KEY" ]; then
    echo "  ❌ HYPERLIQUID_PRIVATE_KEY is not set"
    MISSING_VARS=1
else
    echo "  ✓ HYPERLIQUID_PRIVATE_KEY is set"
fi

if [ -z "$HYPERLIQUID_ADDRESS" ]; then
    echo "  ❌ HYPERLIQUID_ADDRESS is not set"
    MISSING_VARS=1
else
    echo "  ✓ HYPERLIQUID_ADDRESS is set"
fi

if [ $MISSING_VARS -eq 1 ]; then
    echo ""
    echo "Please set the missing environment variables:"
    echo "  export DEEPSEEK_API_KEY=\"your_api_key\""
    echo "  export HYPERLIQUID_PRIVATE_KEY=\"your_private_key\""
    echo "  export HYPERLIQUID_ADDRESS=\"your_address\""
    echo ""
    echo "Or copy .env.example to .env and source it:"
    echo "  cp .env.example .env"
    echo "  # Edit .env with your values"
    echo "  source .env"
    exit 1
fi

echo ""

# Step 2: Create logs directory
echo "Step 2: Creating logs directory..."
mkdir -p logs
echo "  ✓ Logs directory created"
echo ""

# Step 3: Download dependencies
echo "Step 3: Downloading Go dependencies..."
go mod download
if [ $? -ne 0 ]; then
    echo "  ❌ Failed to download dependencies"
    exit 1
fi
echo "  ✓ Dependencies downloaded"
echo ""

# Step 4: Build
echo "Step 4: Building application..."
go build -o aitrading main.go
if [ $? -ne 0 ]; then
    echo "  ❌ Build failed"
    exit 1
fi
echo "  ✓ Build successful"
echo ""

# Step 5: Configuration check
echo "Step 5: Checking configuration..."
if [ ! -f "config.yaml" ]; then
    echo "  ❌ config.yaml not found"
    exit 1
fi
echo "  ✓ Configuration file found"
echo ""

# Step 6: Ask user confirmation
echo "================================"
echo "Configuration Summary"
echo "================================"
grep "symbol:" config.yaml
grep "interval:" config.yaml
grep "trading_enabled:" config.yaml
echo ""

read -p "Ready to start trading bot? (y/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled by user"
    exit 0
fi

echo ""
echo "Starting AI Trading Bot..."
echo "Press Ctrl+C to stop"
echo ""

# Step 7: Run
./aitrading
