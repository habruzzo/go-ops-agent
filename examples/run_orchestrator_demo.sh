#!/bin/bash

# Orchestrator Demo Runner

set -e

echo "Agent Orchestration Demo"
echo "======================="

# Check if API key is set
if [ -z "$OPENAI_API_KEY" ]; then
    echo "ERROR: OPENAI_API_KEY environment variable is required"
    echo "Get your API key from: https://platform.openai.com/api-keys"
    echo "Then run: export OPENAI_API_KEY=your_key_here"
    exit 1
fi

echo "OpenAI API key is set"

# Check if we're in the right directory
if [ ! -f "orchestrator_demo.go" ]; then
    echo "ERROR: orchestrator_demo.go not found. Please run this script from the examples directory."
    exit 1
fi

echo "Building orchestrator demo..."
go build -o orchestrator_demo orchestrator_demo.go

echo "Running orchestrator demo..."
echo "This demonstrates:"
echo "- Multi-agent coordination"
echo "- Workflow orchestration"
echo "- Agent communication"
echo "- Agent monitoring"
echo ""

./orchestrator_demo

echo ""
echo "Demo completed!"
echo ""
echo "What you just saw:"
echo "1. Multiple agents working together"
echo "2. Workflow orchestration in action"
echo "3. Agent communication and coordination"
echo "4. Real-time monitoring and metrics"
echo ""
echo "Next steps:"
echo "1. Read the ReAct paper: https://arxiv.org/abs/2210.03629"
echo "2. Study the orchestrator code"
echo "3. Build your own multi-agent system"
echo "4. Add more sophisticated workflows"


