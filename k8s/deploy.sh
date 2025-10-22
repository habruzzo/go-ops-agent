#!/bin/bash

# Agent Framework Kubernetes Deployment Script
# This script deploys the agent framework to Kubernetes with monitoring

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="agent-framework"
REGISTRY="localhost:5000"  # Change this to your registry
IMAGE_TAG="latest"

echo -e "${BLUE} Agent Framework Kubernetes Deployment${NC}"
echo "================================================"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED} kubectl is not installed or not in PATH${NC}"
    exit 1
fi

# Check if we can connect to cluster
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED} Cannot connect to Kubernetes cluster${NC}"
    echo "Make sure your kubeconfig is set up correctly"
    exit 1
fi

echo -e "${GREEN} Connected to Kubernetes cluster${NC}"

# Build and push images
echo -e "${YELLOW} Building Docker images...${NC}"

# Build agent framework image
echo "Building agent-framework image..."
docker build -t ${REGISTRY}/agent-framework:${IMAGE_TAG} .
docker push ${REGISTRY}/agent-framework:${IMAGE_TAG}

# Build traffic generator image
echo "Building traffic-generator image..."
cd traffic-generator
docker build -t ${REGISTRY}/traffic-generator:${IMAGE_TAG} .
docker push ${REGISTRY}/traffic-generator:${IMAGE_TAG}
cd ..

echo -e "${GREEN} Images built and pushed${NC}"

# Create namespace
echo -e "${YELLOW}  Creating namespace...${NC}"
kubectl apply -f namespace.yaml

# Apply secrets (you'll need to update these with real values)
echo -e "${YELLOW} Setting up secrets...${NC}"
echo "Remember to update the secrets with real API keys!"
kubectl apply -f secret.yaml

# Apply configmaps
echo -e "${YELLOW}  Setting up configuration...${NC}"
kubectl apply -f configmap.yaml

# Deploy monitoring stack
echo -e "${YELLOW} Deploying monitoring stack...${NC}"
kubectl apply -f monitoring/prometheus.yaml
kubectl apply -f monitoring/grafana.yaml

# Wait for monitoring to be ready
echo "Waiting for monitoring stack to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/prometheus -n ${NAMESPACE}
kubectl wait --for=condition=available --timeout=300s deployment/grafana -n ${NAMESPACE}

# Deploy agent framework
echo -e "${YELLOW} Deploying agent framework...${NC}"
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f hpa.yaml

# Wait for agent framework to be ready
echo "Waiting for agent framework to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/agent-framework -n ${NAMESPACE}

# Deploy traffic generator
echo -e "${YELLOW} Deploying traffic generator...${NC}"
kubectl apply -f traffic-generator.yaml

# Wait for traffic generator to be ready
echo "Waiting for traffic generator to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/traffic-generator -n ${NAMESPACE}

# Get service URLs
echo -e "${GREEN} Deployment completed successfully!${NC}"
echo ""
echo -e "${BLUE} Service URLs:${NC}"
echo "================================================"

# Get cluster IP or use port-forward
echo "To access services, you can use port-forward:"
echo ""
echo -e "${YELLOW}Agent Framework:${NC}"
echo "  kubectl port-forward -n ${NAMESPACE} service/agent-framework 8080:8080"
echo "  Health: http://localhost:8080/health"
echo "  Metrics: http://localhost:8080/metrics"
echo "  Status: http://localhost:8080/status"
echo ""
echo -e "${YELLOW}Prometheus:${NC}"
echo "  kubectl port-forward -n ${NAMESPACE} service/prometheus 9090:9090"
echo "  Web UI: http://localhost:9090"
echo ""
echo -e "${YELLOW}Grafana:${NC}"
echo "  kubectl port-forward -n ${NAMESPACE} service/grafana 3000:3000"
echo "  Web UI: http://localhost:3000 (admin/admin)"
echo ""
echo -e "${YELLOW}Traffic Generator:${NC}"
echo "  kubectl port-forward -n ${NAMESPACE} service/traffic-generator 8081:8081"
echo "  Metrics: http://localhost:8081/metrics"
echo ""

# Show current status
echo -e "${BLUE} Current Status:${NC}"
echo "================================================"
kubectl get pods -n ${NAMESPACE}
echo ""
kubectl get services -n ${NAMESPACE}
echo ""
kubectl get hpa -n ${NAMESPACE}

echo ""
echo -e "${GREEN} Next Steps:${NC}"
echo "1. Update API keys in the secret: kubectl edit secret agent-secrets -n ${NAMESPACE}"
echo "2. Access Grafana and import the agent framework dashboard"
echo "3. Run traffic patterns: kubectl exec -it deployment/traffic-generator -n ${NAMESPACE} -- ./traffic-generator -pattern spike"
echo "4. Monitor performance in Grafana and Prometheus"
echo ""
echo -e "${BLUE} Useful Commands:${NC}"
echo "kubectl logs -f deployment/agent-framework -n ${NAMESPACE}"
echo "kubectl logs -f deployment/traffic-generator -n ${NAMESPACE}"
echo "kubectl scale deployment agent-framework --replicas=5 -n ${NAMESPACE}"
echo "kubectl describe hpa agent-framework-hpa -n ${NAMESPACE}"

