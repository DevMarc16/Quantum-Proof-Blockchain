#!/bin/bash

set -e

export PATH="$PATH:/home/dillondev/.foundry/bin"

echo "🚀 Deploying Quantum Blockchain Validator System Contracts"
echo "=========================================================="

RPC_URL="http://localhost:8545"
PRIVATE_KEY="0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
DEPLOYER="0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266"

# Create config file
CONFIG_FILE="build/deployment_config.json"
echo "{\"contracts\": {}, \"deployer\": \"$DEPLOYER\"}" > $CONFIG_FILE

echo ""
echo "🔵 1. Deploying QTM Token..."
QTM_OUTPUT=$(forge create contracts/QTMToken.sol:QTMToken \
    --rpc-url $RPC_URL \
    --private-key $PRIVATE_KEY \
    --legacy \
    --constructor-args "Quantum Token" "QTM" 1000000000000000000000000000 \
    --json 2>/dev/null || echo "")

if echo "$QTM_OUTPUT" | grep -q "deployedTo"; then
    QTM_ADDRESS=$(echo "$QTM_OUTPUT" | jq -r '.deployedTo')
    echo "✅ QTM Token deployed to: $QTM_ADDRESS"
    
    # Update config
    jq ".contracts.QTMToken = \"$QTM_ADDRESS\"" $CONFIG_FILE > tmp.$$ && mv tmp.$$ $CONFIG_FILE
else
    echo "❌ QTM Token deployment failed"
    exit 1
fi

echo ""
echo "🔵 2. Deploying TestnetFaucet..."
FAUCET_OUTPUT=$(forge create contracts/TestnetFaucet.sol:TestnetFaucet \
    --rpc-url $RPC_URL \
    --private-key $PRIVATE_KEY \
    --legacy \
    --constructor-args $QTM_ADDRESS \
    --json 2>/dev/null || echo "")

if echo "$FAUCET_OUTPUT" | grep -q "deployedTo"; then
    FAUCET_ADDRESS=$(echo "$FAUCET_OUTPUT" | jq -r '.deployedTo')
    echo "✅ TestnetFaucet deployed to: $FAUCET_ADDRESS"
    
    # Update config
    jq ".contracts.TestnetFaucet = \"$FAUCET_ADDRESS\"" $CONFIG_FILE > tmp.$$ && mv tmp.$$ $CONFIG_FILE
else
    echo "❌ TestnetFaucet deployment failed"
    exit 1
fi

echo ""
echo "🔵 3. Deploying ValidatorRegistry..."
REGISTRY_OUTPUT=$(forge create contracts/ValidatorRegistry.sol:ValidatorRegistry \
    --rpc-url $RPC_URL \
    --private-key $PRIVATE_KEY \
    --legacy \
    --constructor-args $QTM_TOKEN \
    --json 2>/dev/null || echo "")

if echo "$REGISTRY_OUTPUT" | grep -q "deployedTo"; then
    REGISTRY_ADDRESS=$(echo "$REGISTRY_OUTPUT" | jq -r '.deployedTo')
    echo "✅ ValidatorRegistry deployed to: $REGISTRY_ADDRESS"
    
    # Update config
    jq ".contracts.ValidatorRegistry = \"$REGISTRY_ADDRESS\"" $CONFIG_FILE > tmp.$$ && mv tmp.$$ $CONFIG_FILE
else
    echo "❌ ValidatorRegistry deployment failed"
    exit 1
fi

echo ""
echo "🔵 4. Deploying TokenDistribution..."
DISTRIBUTION_OUTPUT=$(forge create contracts/TokenDistribution.sol:TokenDistribution \
    --rpc-url $RPC_URL \
    --private-key $PRIVATE_KEY \
    --legacy \
    --constructor-args $QTM_ADDRESS 1735689600 \
    --json 2>/dev/null || echo "")

if echo "$DISTRIBUTION_OUTPUT" | grep -q "deployedTo"; then
    DISTRIBUTION_ADDRESS=$(echo "$DISTRIBUTION_OUTPUT" | jq -r '.deployedTo')
    echo "✅ TokenDistribution deployed to: $DISTRIBUTION_ADDRESS"
    
    # Update config
    jq ".contracts.TokenDistribution = \"$DISTRIBUTION_ADDRESS\"" $CONFIG_FILE > tmp.$$ && mv tmp.$$ $CONFIG_FILE
else
    echo "❌ TokenDistribution deployment failed"
    exit 1
fi

echo ""
echo "🔵 5. Setting up permissions..."

# Give faucet minting permissions
echo "Setting faucet as minter..."
cast send $QTM_ADDRESS "addMinter(address)" $FAUCET_ADDRESS \
    --rpc-url $RPC_URL \
    --private-key $PRIVATE_KEY \
    --legacy > /dev/null

# Give registry minting permissions  
echo "Setting registry as minter..."
cast send $QTM_ADDRESS "addMinter(address)" $REGISTRY_ADDRESS \
    --rpc-url $RPC_URL \
    --private-key $PRIVATE_KEY \
    --legacy > /dev/null

# Fund the faucet with initial tokens
echo "Funding faucet with 1M QTM..."
cast send $QTM_ADDRESS "transfer(address,uint256)" $FAUCET_ADDRESS 1000000000000000000000000 \
    --rpc-url $RPC_URL \
    --private-key $PRIVATE_KEY \
    --legacy > /dev/null

echo ""
echo "🎉 Deployment Complete!"
echo "======================"
echo ""
cat $CONFIG_FILE | jq .
echo ""
echo "🔗 Contract Addresses:"
echo "QTM Token:        $QTM_ADDRESS"  
echo "TestnetFaucet:    $FAUCET_ADDRESS"
echo "ValidatorRegistry: $REGISTRY_ADDRESS"
echo "TokenDistribution: $DISTRIBUTION_ADDRESS"
echo ""
echo "💰 Faucet funded with 1M QTM tokens"
echo "🔑 All permissions configured"
echo ""
echo "✅ Ready for validator registration!"