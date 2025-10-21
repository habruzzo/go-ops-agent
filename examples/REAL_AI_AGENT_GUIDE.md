# Real AI Agent Setup Guide

This guide shows you how to create and run a real AI agent for observability using your own API keys.

## Prerequisites

### 1. Get an OpenAI API Key
1. Go to [OpenAI Platform](https://platform.openai.com/api-keys)
2. Sign up or log in
3. Create a new API key
4. Copy the key (starts with `sk-`)

### 2. Set Environment Variable
```bash
export OPENAI_API_KEY=sk-your-actual-key-here
```

### 3. Install Dependencies
```bash
cd /path/to/your/agent
go mod tidy
```

## Quick Start

### 1. Run the Real AI Agent
```bash
cd examples
go run real_ai_agent.go
```

### 2. What You'll See
```
Starting Real AI Agent for Observability
========================================
Found Prometheus, adding collector...
Starting framework...
Framework started successfully!

AI Agent is ready! Try these queries:
  - 'What's the current system status?'
  - 'Are there any anomalies?'
  - 'Give me recommendations'
  - 'Analyze the performance'

--- AI Query 1 ---
Query: What's the current system status?
AI Response: Based on the current metrics, I can see that your system is running normally. The Prometheus server is up and collecting metrics successfully. I notice some standard Go runtime metrics including memory allocation and garbage collection cycles, which are within normal ranges. No critical anomalies detected at this time.

--- AI Query 2 ---
Query: Are there any performance issues?
AI Response: Currently, I don't see any significant performance issues. The system appears to be operating within normal parameters. However, I recommend monitoring memory usage trends and garbage collection frequency to ensure optimal performance over time.
```

## Configuration Options

### AI Model Selection
Edit `real-config.yaml` to change the AI model:

```yaml
plugins:
  - name: "real-ai"
    type: "ai"
    config:
      model: "gpt-4"        # Options: gpt-4, gpt-3.5-turbo, gpt-4-turbo
      max_tokens: 500       # Response length (100-4000)
      temperature: 0.7      # Creativity level (0.0-2.0)
```

### Prometheus Integration
If you have Prometheus running:

```yaml
plugins:
  - name: "real-prometheus"
    type: "prometheus"
    config:
      url: "http://localhost:9090"
      interval: "10s"
      queries: 
        - "up"
        - "cpu_usage_percent"
        - "memory_usage_percent"
        - "http_requests_total"
```

## Custom Queries

### System Health Queries
```bash
# Ask about system status
"What's the current system health?"

# Get performance analysis
"Analyze the system performance"

# Check for issues
"Are there any problems I should know about?"
```

### Anomaly Analysis
```bash
# Get anomaly summary
"What anomalies have been detected?"

# Get recommendations
"What should I investigate first?"

# Get detailed analysis
"Explain the CPU spike I'm seeing"
```

### Optimization Queries
```bash
# Get optimization suggestions
"How can I improve system performance?"

# Resource recommendations
"What resources should I monitor?"

# Scaling advice
"Should I scale up or down?"
```

## Advanced Setup

### 1. Custom Prometheus Queries
Add your own metrics to monitor:

```yaml
queries: 
  - "your_custom_metric"
  - "rate(http_requests_total[5m])"
  - "histogram_quantile(0.95, http_request_duration_seconds)"
```

### 2. Multiple AI Agents
You can run multiple AI agents with different configurations:

```yaml
plugins:
  - name: "performance-ai"
    type: "ai"
    config:
      model: "gpt-4"
      max_tokens: 300
      temperature: 0.3  # More focused responses
  
  - name: "security-ai"
    type: "ai"
    config:
      model: "gpt-4"
      max_tokens: 400
      temperature: 0.5  # Balanced responses
```

### 3. Custom Responders
Add custom responders for different outputs:

```yaml
plugins:
  - name: "slack-responder"
    type: "slack"
    config:
      webhook_url: "https://hooks.slack.com/your-webhook"
      channel: "#alerts"
  
  - name: "email-responder"
    type: "email"
    config:
      smtp_server: "smtp.gmail.com"
      smtp_port: 587
      username: "your-email@gmail.com"
      password: "your-app-password"
```

## Cost Management

### OpenAI Pricing (as of 2024)
- **GPT-4**: $0.03/1K input tokens, $0.06/1K output tokens
- **GPT-3.5-turbo**: $0.001/1K input tokens, $0.002/1K output tokens

### Cost Optimization Tips
1. **Use GPT-3.5-turbo** for simple queries
2. **Limit max_tokens** to control response length
3. **Set up usage alerts** in OpenAI dashboard
4. **Use caching** for repeated queries

### Example Cost Calculation
```
Query: "What's the system status?" (20 tokens)
Response: "System is healthy..." (100 tokens)
Cost: (20 * $0.001 + 100 * $0.002) / 1000 = $0.00022 per query
```

## Troubleshooting

### Common Issues

#### 1. API Key Not Found
```
ERROR: OPENAI_API_KEY environment variable is required
```
**Solution**: Set the environment variable:
```bash
export OPENAI_API_KEY=sk-your-key-here
```

#### 2. Rate Limit Exceeded
```
Error: rate limit exceeded
```
**Solution**: 
- Wait a few minutes
- Upgrade your OpenAI plan
- Reduce query frequency

#### 3. Invalid Model
```
Error: model not found
```
**Solution**: Check model name in config:
```yaml
model: "gpt-4"  # Not "gpt4" or "GPT-4"
```

#### 4. Prometheus Connection Failed
```
Error: connection refused
```
**Solution**: 
- Start Prometheus: `docker run -p 9090:9090 prom/prometheus`
- Check URL in config
- Verify Prometheus is accessible

### Debug Mode
Enable debug logging:
```yaml
logging:
  level: debug
  format: text
  output: stdout
```

## Production Deployment

### 1. Environment Variables
```bash
export OPENAI_API_KEY=sk-your-production-key
export LOG_LEVEL=info
export CONFIG_FILE=/etc/agent/config.yaml
```

### 2. Systemd Service
Create `/etc/systemd/system/ai-agent.service`:
```ini
[Unit]
Description=AI Observability Agent
After=network.target

[Service]
Type=simple
User=agent
WorkingDirectory=/opt/agent
ExecStart=/opt/agent/ai-agent
Environment=OPENAI_API_KEY=sk-your-key
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### 3. Docker Deployment
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o ai-agent examples/real_ai_agent.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/ai-agent .
COPY --from=builder /app/examples/real-config.yaml .
CMD ["./ai-agent"]
```

## Security Considerations

### 1. API Key Security
- Never commit API keys to version control
- Use environment variables or secret management
- Rotate keys regularly
- Monitor usage for anomalies

### 2. Data Privacy
- Be aware that queries are sent to OpenAI
- Don't send sensitive data in queries
- Consider on-premises alternatives for sensitive environments

### 3. Network Security
- Use HTTPS for all API calls
- Implement proper firewall rules
- Monitor network traffic

## Next Steps

1. **Start Simple**: Begin with basic queries and gradually add complexity
2. **Monitor Costs**: Set up usage alerts and monitor spending
3. **Customize**: Add your own metrics and queries
4. **Scale**: Deploy to production with proper monitoring
5. **Integrate**: Connect with your existing monitoring stack

## Support

- **OpenAI Documentation**: https://platform.openai.com/docs
- **Framework Issues**: Check the main README.md
- **Community**: Join discussions in the project repository

Your AI agent is now ready to provide intelligent insights about your system!
