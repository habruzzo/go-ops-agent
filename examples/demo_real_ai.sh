#!/bin/bash

# Demo script for Real AI Agent

set -e

echo "Real AI Agent Demo"
echo "=================="

# Check if API key is set
if [ -z "$OPENAI_API_KEY" ]; then
    echo "ERROR: OPENAI_API_KEY environment variable is required"
    echo "Get your API key from: https://platform.openai.com/api-keys"
    echo "Then run: export OPENAI_API_KEY=your_key_here"
    exit 1
fi

# Check if the agent is built
if [ ! -f "real_ai_agent" ]; then
    echo "Building AI agent..."
    go build -o real_ai_agent real_ai_agent.go
fi

echo "Starting AI agent demo..."
echo "This will run for 60 seconds and show example AI queries."
echo ""

# Run the agent in the background
./real_ai_agent &
AGENT_PID=$!

# Wait for it to start
sleep 5

echo "AI Agent is running! Check the output above for AI responses."
echo "Press Ctrl+C to stop the demo."

# Wait for user to stop
trap 'echo ""; echo "Stopping demo..."; kill $AGENT_PID 2>/dev/null; exit 0' INT

# Keep running until interrupted
while true; do
    sleep 1
done
