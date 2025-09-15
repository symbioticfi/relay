#!/bin/bash

set -eo pipefail

# Configuration

# Contracts commit hash to use
CONTRACTS_COMMIT="f15b7ac1d8f1ec0b434178b64823aab67f8afc4e"

# -----------------------------------------

echo "Building Relay Docker image for e2e tests..."
# go to root of repo
cd ..
make image TAG=dev
# get back into e2e
cd e2e

# Check if temp-network directory exists and clean up any running containers
if [ -d "temp-network" ]; then
    echo "Found existing temp-network directory. Attempting to clean up running containers..."
    cd temp-network
    if ! docker compose down; then
        echo "WARNING: Failed to run 'docker compose down' in temp-network directory."
        echo "Please manually clean up the temp-network directory and any running containers before proceeding."
        echo "You can try: cd temp-network && docker compose down --remove-orphans && cd .. && rm -rf temp-network"
        exit 1
    fi
    cd ..
    echo "Successfully cleaned up existing containers."
fi

echo "Setting up Symbiotic contracts..."
if [ ! -d "contracts" ]; then
    echo "Cloning Symbiotic contracts repository..."
    git clone https://github.com/symbioticfi/symbiotic-super-sum contracts
else
    echo "Contracts directory already exists, skipping clone..."
fi
cd contracts
git fetch origin
git checkout $CONTRACTS_COMMIT

echo "Installing dependencies..."
npm install
echo "Building contracts..."
forge build

cd ..

# Pass through all environment variables to generate_network.sh with defaults
export OPERATORS=${OPERATORS}
export COMMITERS=${COMMITERS}
export AGGREGATORS=${AGGREGATORS}
export VERIFICATION_TYPE=${VERIFICATION_TYPE}
export EPOCH_TIME=${EPOCH_TIME}
export BLOCK_TIME=${BLOCK_TIME}
export FINALITY_BLOCKS=${FINALITY_BLOCKS}
export GENERATE_SIDECARS=${GENERATE_SIDECARS}

# Call the generate network script
./scripts/generate_network.sh

echo "Setup complete! Network configuration generated in temp-network/ directory."