#!/bin/bash

# Agent Framework Kubernetes Demo Script
# This script demonstrates the complete agent framework deployment and testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

NAMESPACE="agent-framework"

echo -e "${BLUE}🎬 Agent Framework Kubernetes Demo${NC}"
echo "================================================"
echo ""
echo -e "${CYAN}This demo will:${NC}"
echo "1. Deploy the agent framework to Kubernetes"
echo "2. Set up monitoring with Prometheus and Grafana"
echo "3. Deploy a traffic generator"
echo "4. Run various traffic patterns"
echo "5. Show you how to monitor and scale"
echo ""

# Check prerequisites
echo -e "${YELLOW}🔍 Checking prerequisites...${NC}"

if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl is not installed${NC}"
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ docker is not installed${NC}"
    exit 1
fi

if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}❌ Cannot connect to Kubernetes cluster${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Prerequisites check passed${NC}"
echo ""

# Ask for confirmation
echo -e "${PURPLE}🚀 Ready to start the demo?${NC}"
echo "This will deploy the agent framework to your Kubernetes cluster."
echo "Press Enter to continue or Ctrl+C to cancel..."
read

# Step 1: Deploy the framework
echo -e "${BLUE}📦 Step 1: Deploying Agent Framework${NC}"
echo "================================================"
./deploy.sh

echo ""
echo -e "${GREEN}✅ Deployment completed!${NC}"
echo ""

# Step 2: Show current status
echo -e "${BLUE}📊 Step 2: Current Status${NC}"
echo "================================================"
kubectl get pods -n ${NAMESPACE}
echo ""
kubectl get services -n ${NAMESPACE}
echo ""

# Step 3: Demonstrate traffic patterns
echo -e "${BLUE}🌊 Step 3: Traffic Pattern Demonstrations${NC}"
echo "================================================"

# Function to run a quick demo pattern
run_demo_pattern() {
    local pattern=$1
    local duration=$2
    local base_load=$3
    local peak_load=$4
    local description=$5
    
    echo -e "${YELLOW}🎯 Demo: ${pattern}${NC}"
    echo "Description: ${description}"
    echo "Duration: ${duration}"
    echo ""
    
    # Run the pattern in background
    kubectl exec -it deployment/traffic-generator -n ${NAMESPACE} -- ./traffic-generator \
        -pattern=${pattern} \
        -duration=${duration} \
        -base-load=${base_load} \
        -peak-load=${peak_load} \
        -target=http://agent-framework:8080 &
    
    local pid=$!
    
    # Show live metrics for 30 seconds
    echo -e "${CYAN}📈 Live Metrics (30 seconds):${NC}"
    for i in {1..6}; do
        echo "--- Sample $i ---"
        kubectl top pods -n ${NAMESPACE} 2>/dev/null || echo "Metrics server not available"
        kubectl get hpa -n ${NAMESPACE}
        echo ""
        sleep 5
    done
    
    # Wait for pattern to complete
    wait $pid
    
    echo -e "${GREEN}✅ ${pattern} demo completed${NC}"
    echo ""
    echo "Waiting 10 seconds before next demo..."
    sleep 10
    echo ""
}

# Run quick demos
run_demo_pattern "gradual" "2m" "5" "50" "Gradual traffic increase"
run_demo_pattern "spike" "1m" "5" "100" "Sudden traffic spike"

# Step 4: Show monitoring
echo -e "${BLUE}📊 Step 4: Monitoring Dashboard${NC}"
echo "================================================"
echo ""
echo -e "${CYAN}Access the monitoring dashboards:${NC}"
echo ""
echo -e "${YELLOW}Grafana Dashboard:${NC}"
echo "  kubectl port-forward -n ${NAMESPACE} service/grafana 3000:3000"
echo "  Then open: http://localhost:3000 (admin/admin)"
echo ""
echo -e "${YELLOW}Prometheus:${NC}"
echo "  kubectl port-forward -n ${NAMESPACE} service/prometheus 9090:9090"
echo "  Then open: http://localhost:9090"
echo ""
echo -e "${YELLOW}Agent Framework Status:${NC}"
echo "  kubectl port-forward -n ${NAMESPACE} service/agent-framework 8080:8080"
echo "  Then open: http://localhost:8080/status"
echo ""

# Step 5: Interactive scaling demo
echo -e "${BLUE}⚖️  Step 5: Auto-scaling Demo${NC}"
echo "================================================"
echo ""
echo -e "${CYAN}Let's demonstrate auto-scaling:${NC}"
echo ""

# Show current HPA status
echo "Current HPA status:"
kubectl get hpa -n ${NAMESPACE}
echo ""

# Run a sustained load
echo -e "${YELLOW}🚀 Running sustained load to trigger scaling...${NC}"
kubectl exec -it deployment/traffic-generator -n ${NAMESPACE} -- ./traffic-generator \
    -pattern=gradual \
    -duration=5m \
    -base-load=20 \
    -peak-load=200 \
    -target=http://agent-framework:8080 &

load_pid=$!

# Monitor scaling
echo -e "${CYAN}📈 Monitoring auto-scaling (watch for 2 minutes):${NC}"
for i in {1..24}; do
    echo "--- Check $i ---"
    kubectl get pods -n ${NAMESPACE} | grep agent-framework
    kubectl get hpa -n ${NAMESPACE}
    echo ""
    sleep 5
done

# Wait for load to complete
wait $load_pid

echo -e "${GREEN}✅ Auto-scaling demo completed${NC}"
echo ""

# Step 6: Show final status
echo -e "${BLUE}📊 Step 6: Final Status${NC}"
echo "================================================"
kubectl get pods -n ${NAMESPACE}
echo ""
kubectl get hpa -n ${NAMESPACE}
echo ""

# Step 7: Cleanup option
echo -e "${BLUE}🧹 Step 7: Cleanup${NC}"
echo "================================================"
echo ""
echo -e "${PURPLE}Demo completed! 🎉${NC}"
echo ""
echo -e "${CYAN}What would you like to do next?${NC}"
echo "1. Keep the deployment running for further testing"
echo "2. Clean up all resources"
echo "3. Run more traffic patterns"
echo ""
echo -e "${YELLOW}To clean up later, run:${NC}"
echo "  ./cleanup.sh"
echo ""
echo -e "${YELLOW}To run more traffic tests:${NC}"
echo "  ./test-traffic.sh"
echo ""
echo -e "${YELLOW}To access services:${NC}"
echo "  kubectl port-forward -n ${NAMESPACE} service/grafana 3000:3000"
echo "  kubectl port-forward -n ${NAMESPACE} service/prometheus 9090:9090"
echo "  kubectl port-forward -n ${NAMESPACE} service/agent-framework 8080:8080"
echo ""

echo -e "${GREEN}🎯 Demo completed successfully!${NC}"
echo "You now have a fully functional agent framework running in Kubernetes"
echo "with monitoring, auto-scaling, and traffic testing capabilities!"

