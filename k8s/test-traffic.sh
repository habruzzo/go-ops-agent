#!/bin/bash

# Traffic Testing Script for Agent Framework
# This script runs various traffic patterns to test the framework

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

NAMESPACE="agent-framework"

echo -e "${BLUE}🌊 Agent Framework Traffic Testing${NC}"
echo "================================================"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl is not installed or not in PATH${NC}"
    exit 1
fi

# Check if traffic generator pod exists
if ! kubectl get pods -n ${NAMESPACE} -l app=traffic-generator | grep -q Running; then
    echo -e "${RED}❌ Traffic generator pod is not running${NC}"
    echo "Make sure the deployment is running: kubectl get pods -n ${NAMESPACE}"
    exit 1
fi

# Function to run a traffic pattern
run_traffic_pattern() {
    local pattern=$1
    local duration=$2
    local base_load=$3
    local peak_load=$4
    local description=$5
    
    echo -e "${YELLOW}🚀 Running pattern: ${pattern}${NC}"
    echo "Description: ${description}"
    echo "Duration: ${duration}"
    echo "Base Load: ${base_load} RPS"
    echo "Peak Load: ${peak_load} RPS"
    echo ""
    
    kubectl exec -it deployment/traffic-generator -n ${NAMESPACE} -- ./traffic-generator \
        -pattern=${pattern} \
        -duration=${duration} \
        -base-load=${base_load} \
        -peak-load=${peak_load} \
        -target=http://agent-framework:8080
    
    echo -e "${GREEN}✅ Pattern ${pattern} completed${NC}"
    echo ""
    echo "Waiting 30 seconds before next test..."
    sleep 30
    echo ""
}

# Function to show current metrics
show_metrics() {
    echo -e "${BLUE}📊 Current Metrics:${NC}"
    echo "================================================"
    
    # Get pod status
    echo "Pod Status:"
    kubectl get pods -n ${NAMESPACE}
    echo ""
    
    # Get HPA status
    echo "HPA Status:"
    kubectl get hpa -n ${NAMESPACE}
    echo ""
    
    # Get resource usage
    echo "Resource Usage:"
    kubectl top pods -n ${NAMESPACE} 2>/dev/null || echo "Metrics server not available"
    echo ""
}

# Main testing sequence
echo -e "${GREEN}🎯 Starting traffic testing sequence...${NC}"
echo ""

# Show initial status
show_metrics

# Test 1: Gradual ramp-up
run_traffic_pattern "gradual" "5m" "10" "100" "Gradual traffic increase to test auto-scaling"

# Test 2: Sudden spike
run_traffic_pattern "spike" "2m" "10" "500" "Sudden traffic spike to test burst capacity"

# Test 3: Periodic bursts
run_traffic_pattern "burst" "3m" "5" "200" "Periodic bursts to test recovery"

# Test 4: Chaos pattern
run_traffic_pattern "chaos" "2m" "5" "300" "Chaotic traffic to test resilience"

# Test 5: Black Friday simulation
run_traffic_pattern "black-friday" "10m" "50" "1000" "Black Friday style sustained high load"

# Test 6: DDoS simulation
run_traffic_pattern "ddos" "1m" "10" "2000" "DDoS attack simulation"

# Show final status
echo -e "${GREEN}🎉 Traffic testing completed!${NC}"
echo ""
show_metrics

echo -e "${BLUE}📋 Analysis Recommendations:${NC}"
echo "================================================"
echo "1. Check Grafana dashboard for performance metrics"
echo "2. Review Prometheus alerts for any triggered alerts"
echo "3. Analyze HPA scaling behavior"
echo "4. Check pod logs for any errors or issues"
echo "5. Review resource utilization patterns"
echo ""
echo -e "${YELLOW}🔍 Useful Commands for Analysis:${NC}"
echo "kubectl logs -f deployment/agent-framework -n ${NAMESPACE}"
echo "kubectl describe hpa agent-framework-hpa -n ${NAMESPACE}"
echo "kubectl top pods -n ${NAMESPACE}"
echo "kubectl get events -n ${NAMESPACE} --sort-by='.lastTimestamp'"

