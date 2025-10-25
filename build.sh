#!/bin/bash

# Create logs directory
mkdir -p logs

# Check if environment variables are set
if [ -z "$DEEPSEEK_API_KEY" ]; then
    echo "Warning: DEEPSEEK_API_KEY environment variable is not set"
    echo "Please set it with: export DEEPSEEK_API_KEY=your_api_key"
fi

if [ -z "$HYPERLIQUID_PRIVATE_KEY" ]; then
    echo "Warning: HYPERLIQUID_PRIVATE_KEY environment variable is not set"
    echo "Please set it with: export HYPERLIQUID_PRIVATE_KEY=your_private_key"
fi

if [ -z "$HYPERLIQUID_ADDRESS" ]; then
    echo "Warning: HYPERLIQUID_ADDRESS environment variable is not set"
    echo "Please set it with: export HYPERLIQUID_ADDRESS=your_address"
fi

# Install dependencies
echo "Installing dependencies..."
go mod download
go mod tidy

# Build the application
echo "Building AI Trading Bot..."
go build -o aitrading main.go

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo ""
    echo "To run the bot:"
    echo "  ./aitrading"
    echo ""
    echo "Make sure to configure config.yaml and set environment variables first!"
else
    echo "Build failed!"
    exit 1
fi
