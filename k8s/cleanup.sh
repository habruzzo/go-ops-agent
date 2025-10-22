#!/bin/bash

# Agent Framework Kubernetes Cleanup Script
# This script removes all resources created by the deployment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

NAMESPACE="agent-framework"

echo -e "${BLUE}🧹 Agent Framework Kubernetes Cleanup${NC}"
echo "================================================"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl is not installed or not in PATH${NC}"
    exit 1
fi

# Check if namespace exists
if ! kubectl get namespace ${NAMESPACE} &> /dev/null; then
    echo -e "${YELLOW}⚠️  Namespace ${NAMESPACE} does not exist${NC}"
    exit 0
fi

echo -e "${YELLOW}🗑️  Removing all resources from namespace ${NAMESPACE}...${NC}"

# Delete all resources in the namespace
kubectl delete all --all -n ${NAMESPACE}
kubectl delete configmap --all -n ${NAMESPACE}
kubectl delete secret --all -n ${NAMESPACE}
kubectl delete hpa --all -n ${NAMESPACE}

# Delete the namespace
echo -e "${YELLOW}🗑️  Deleting namespace...${NC}"
kubectl delete namespace ${NAMESPACE}

echo -e "${GREEN}✅ Cleanup completed successfully!${NC}"
echo ""
echo -e "${BLUE}📋 Verification:${NC}"
echo "================================================"
kubectl get namespaces | grep ${NAMESPACE} || echo "Namespace ${NAMESPACE} has been removed"

