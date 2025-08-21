# E2E Testing Guide

This directory contains end-to-end tests for the Symbiotic Relay project. The tests run against a local blockchain network with configurable relay operators, commiters, and aggregators.

## Prerequisites

- Docker and Docker Compose
- Go 1.24+
- Node.js and npm (for smart contract compilation)
- Foundry (forge) for contract building

## Quick Start

1. **Setup the test environment:**
   ```bash
   ./setup.sh
   ```

2. **Start the network:**
   ```bash
   cd temp-network
   docker compose up -d
   cd ..
   ```

3. **Run the tests:**
   ```bash
   cd tests
   go test -v
   ```

## Configuration

You can customize the test environment by setting environment variables before running `setup.sh`. All variables have sensible defaults.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OPERATORS` | `4` | Number of relay operators (max: 999) |
| `COMMITERS` | `1` | Number of commiter nodes |
| `AGGREGATORS` | `1` | Number of aggregator nodes |
| `VERIFICATION_TYPE` | `1` | Verification type: `0`=BLS-BN254-ZK, `1`=BLS-BN254-SIMPLE |
| `EPOCH_TIME` | `30` | Time for new epochs in relay network (seconds) |
| `BLOCK_TIME` | `1` | Block time in seconds for anvil interval mining |
| `FINALITY_BLOCKS` | `2` | Number of blocks for finality |

### Example with Custom Configuration

```bash
# Set custom configuration
export OPERATORS=6
export COMMITERS=2
export AGGREGATORS=1
export VERIFICATION_TYPE=0
export EPOCH_TIME=32
export BLOCK_TIME=2

# Run setup
./setup.sh

# Start network
cd temp-network
docker compose up -d
cd ..

# Run tests
cd tests
go test -v
```

## Contract Information

The tests use smart contracts from the Symbiotic protocol:
- **Repository**: https://github.com/symbioticfi/symbiotic-super-sum

The commit hash can be updated in `setup.sh` if needed for testing against different contract versions.