#!/bin/bash

set -e

echo "🚀 Workflow Builder - Quick Start"
echo "=================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed"
    echo "   Install with: brew install go"
    exit 1
fi

echo "✅ Go installed: $(go version)"

# Check if Docker is running
if ! docker info &> /dev/null; then
    echo "❌ Docker is not running"
    echo "   Please start Docker Desktop"
    exit 1
fi

echo "✅ Docker is running"
echo ""

# Start Temporal server
echo "🔧 Starting Temporal server..."
docker-compose up -d

echo ""
echo "⏳ Waiting for Temporal to be ready (15 seconds)..."
sleep 15

# Check if Temporal is healthy
if docker-compose ps | grep -q "temporal.*Up"; then
    echo "✅ Temporal server started"
else
    echo "❌ Temporal server failed to start"
    docker-compose logs temporal
    exit 1
fi

echo ""
echo "📦 Installing Go dependencies..."
cd temporal-worker
go mod tidy
cd ..

echo ""
echo "✅ Setup complete!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Next Steps:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "1. Start the Temporal worker:"
echo "   cd temporal-worker"
echo "   go run main.go"
echo ""
echo "2. In another terminal, trigger a workflow:"
echo "   cd temporal-worker"
echo "   go run trigger/main.go --title \"Test Ticket\" --agent YOUR_AGENT_ID --rating 5"
echo ""
echo "3. View workflows in Temporal UI:"
echo "   http://localhost:8080"
echo ""
echo "4. Get agent IDs:"
echo "   curl -s http://localhost:3000/api/agents | jq '.[] | {id, name}'"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
