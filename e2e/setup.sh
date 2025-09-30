#!/bin/bash

set -eo pipefail

# Configuration

# Contracts commit hash to use
CONTRACTS_COMMIT="24bcb351c8b6125b0412d5bf7916da405c548000"

# Circuits commit
CIRCUITS_COMMIT="2f3eada1f03aa4aaee26f9bb67bcf21f66d5de89"

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

# If verification type is 0, clone the circuit keys repository
if [ "${VERIFICATION_TYPE}" = "0" ]; then
    echo "Verification type is 0, cloning circuit keys repository..."
    if [ ! -d "temp-network/circuits" ]; then
        echo "Cloning circuit keys repository (shallow clone of specific commit)..."
        # Create circuits directory in temp-network
        mkdir -p temp-network/circuits
        cd temp-network/circuits
        
        # Initialize empty git repo and add remote
        git init
        git remote add origin https://github.com/symbioticfi/relay-bn254-example-circuit-keys
        
        # Fetch only the specific commit (shallow) with parallel jobs for speed
        git fetch --depth 1 --jobs=4 origin $CIRCUITS_COMMIT
        git checkout FETCH_HEAD
        
        # Remove git metadata to keep only the files
        rm -rf .git
        
        cd ../..
        echo "Circuit keys cloned successfully to temp-network/circuits/"
    else
        echo "Circuits directory already exists in temp-network, skipping clone..."
    fi

    echo "Copying circuits to contracts directory and building..."
    cp -r temp-network/circuits contracts/circuits
    cd contracts
    echo "Building circuits with Forge..."
    forge build circuits/
else
    echo "Verification type is not 0, skipping circuit keys clone..."
fi



echo "Setup complete! Network configuration generated in temp-network/ directory."