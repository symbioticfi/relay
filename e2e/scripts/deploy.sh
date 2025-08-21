#!/bin/sh
set -e

echo "Waiting for anvil to be ready..."
until cast client --rpc-url http://anvil:8545 > /dev/null 2>&1; do sleep 1; done
until cast client --rpc-url http://anvil-settlement:8546 > /dev/null 2>&1; do sleep 1; done

echo "Deploying contracts..."
forge script script/LocalDeploy.s.sol:LocalDeploy --rpc-url http://anvil:8545 -vv --broadcast --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 | tee /deploy-data/deployment.log
forge script script/SettlementChainDeploy.s.sol:SettlementChainDeploy --rpc-url http://anvil-settlement:8546 -vv --broadcast --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 | tee /deploy-data/settlement-deployment.log
forge script script/ValSetDriverDeploy.s.sol:ValSetDriverDeploy --rpc-url http://anvil:8545 -vv --broadcast --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 | tee /deploy-data/valsetdriver-deployment.log

echo 'Waiting for relay contracts file...'
until [ -f /deploy-data/relay_contracts.json ]; do sleep 2; done

echo "Setting interval mining..."
BLOCK_TIME=${BLOCK_TIME:-1}
cast rpc --rpc-url http://anvil:8545 evm_setIntervalMining $BLOCK_TIME
cast rpc --rpc-url http://anvil-settlement:8546 evm_setIntervalMining $BLOCK_TIME

echo "Mine a single block to finalize the deployment..."
cast rpc --rpc-url http://anvil:8545 evm_mine
cast rpc --rpc-url http://anvil-settlement:8546 evm_mine

echo "Deployment completed successfully!"

# Create deployment completion marker
echo "$(date): Deployment completed successfully" > /deploy-data/deployment-complete.marker
echo "Deployment completion marker created" 
