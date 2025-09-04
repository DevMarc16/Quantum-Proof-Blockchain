#!/bin/bash
# Convenience script to deploy multi-validator quantum network
# This script calls the actual deployment script from the scripts/ directory

echo "🚀 Starting Quantum Network Deployment"
echo "======================================="

# Change to the scripts directory and run the deployment
cd "$(dirname "$0")/scripts" && ./deploy_multi_validators.sh