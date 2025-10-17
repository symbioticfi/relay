# Symbiotic Relay - Development Guide

## Table of Contents

1. [Introduction](#introduction)
2. [Architecture Overview](#architecture-overview)
3. [Development Setup](#development-setup)
4. [Running Tests](#running-tests)
5. [Local Environment](#local-environment)
6. [Managing API/Configuration Changes](#managing-api-changes)
7. [Code Generation](#code-generation)
8. [Linting](#linting)
9. [Building Docker Images](#building-docker-images)
10. [Contributing](#contributing)
11. [Additional Resources](#additional-resources)

---

## Introduction

The Symbiotic Relay is a distributed middleware layer that facilitates cross-chain coordination, validator set management, and signature aggregation. This guide provides comprehensive information for developers working on the relay codebase, including setup, testing, and managing changes across the ecosystem.

## Architecture Overview

### Core Components

The Symbiotic Relay consists of several key components working together:

#### 1. **P2P Layer** (`internal/client/p2p/`)
The peer-to-peer networking layer handles decentralized communication between relay nodes:
- **GossipSub Protocol**: Uses libp2p's GossipSub for message broadcasting
- **Topic-based Communication**: 
  - `/relay/v1/signature/ready` - Signature ready notifications
  - `/relay/v1/proof/ready` - Aggregation proof ready notifications
- **gRPC over P2P**: Direct node-to-node sync requests using gRPC over libp2p streams
- **Discovery**: 
  - mDNS for local network discovery
  - Kademlia DHT for distributed peer discovery
  - Static peer support for known bootstrap nodes

#### 2. **Node Logic & Applications**
The relay operates in multiple modes, configurable per node:

- **Signer Nodes** (`internal/usecase/signer-app/`):
  - Listen for signature requests
  - Sign messages using BLS/Ecdsa keys
  - Broadcast signatures to the P2P network
  
- **Aggregator Nodes** (`internal/usecase/aggregator-app/`):
  - Collect signatures from multiple signers
  - Apply aggregation policies
  - Generate BLS aggregated signatures or ZK proofs
  - Broadcast aggregation proofs

- **Committer Nodes** (`internal/usecase/valset-listener/`):
  - Monitor validator set state on-chain
  - Submit aggregated proofs to settlement chains
  - Track epoch transitions

#### 3. **API Layer** (`internal/usecase/api-server/`)
Exposes gRPC API for external clients to create signature requests, query proofs/signatures and get validator/epoch info. Defined in `api/proto/v1/api.proto`

#### 4. **Storage Layer** (`internal/client/repository/badger/`)
- BadgerDB for persistent state storage
- Caches for high-performance reads
- Stores signatures, aggregation proofs, validator sets, and epochs

#### 5. **Symbiotic Integration** (`symbiotic/`)
- EVM client for on-chain interactions
- Validator set derivation logic
- Cryptographic primitives (BLS signatures, ZK proofs)

---

## Development Setup

### Prerequisites

- **Go 1.24.3 or later**
- **Node.js** (for contract compilation in E2E tests)
- **Foundry/Forge** (for Solidity contracts)
- **Docker & Docker Compose** (for E2E testing)
- **Buf CLI** (installed via `make install-tools`)

### Initial Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/symbioticfi/relay.git
   cd relay
   ```

2. **Install development tools**:
   ```bash
   make install-tools
   ```
   This installs:
   - `buf` - Protocol buffer compiler
   - `protoc-gen-go` and `protoc-gen-go-grpc` - Go protobuf generators
   - `mockgen` - Mock generation for testing
   - `protoc-gen-doc` - API documentation generator

3. **Generate code**:
   ```bash
   make generate
   ```
   This generates:
   - API types from proto files
   - P2P message types
   - BadgerDB repository types
   - Client types
   - Mock implementations
   - Contract ABI bindings

4. **Build the relay sidecar binary**:
   ```bash
   make build-relay-sidecar OS=linux ARCH=amd64
   ```
   or for macOS ARM:
   ```bash
   make build-relay-sidecar OS=darwin ARCH=arm64
   ```

5. **Build the relay utils CLI binary**:
   ```bash
   make build-relay-utils OS=linux ARCH=amd64
   ```

---

## Running Tests

### Unit Tests

Run all unit tests with coverage:

```bash
make unit-test
```

To run tests for a specific package:

```bash
go test -v ./internal/usecase/signer-app/
```

### End-to-End (E2E) Tests

E2E tests spin up a complete relay network with blockchain nodes.

#### Quick E2E Test Run

```bash
# Setup the E2E environment (only needed once or after cleanup)
cd e2e
./setup.sh

# Start the network
cd temp-network
docker compose up -d
cd ../..

# Run the tests
make e2e-test
```

#### Custom E2E Configuration

You can customize the E2E network topology using environment variables, see the setup.sh script for more envs:

```bash
# Configure network before setup
export OPERATORS=6          # Number of relay operator nodes (default: 4)
export COMMITERS=2          # Number of committer nodes (default: 1)
export AGGREGATORS=1        # Number of aggregator nodes (default: 1)
export VERIFICATION_TYPE=1  # 0=ZK proofs, 1=BLS simple (default: 1)
export EPOCH_TIME=30        # Epoch duration in seconds (default: 30)
export BLOCK_TIME=2         # Block time in seconds (default: 1)
export FINALITY_BLOCKS=2    # Blocks until finality (default: 2)

# Run setup with custom config
cd e2e
./setup.sh
```

#### Cleanup E2E Environment

```bash
# Stop and remove containers
cd e2e/temp-network
docker compose down
```

---

## Local Environment

### Setting Up Local Development Network

For local development and testing:

```bash
make local-setup
```

This command takes the same envs as the e2e setup for configuration.

This command:
1. Generates relay sidecar configurations
2. Sets up a local blockchain network (Anvil)
3. Deploys contracts
4. Starts relay nodes in Docker

---

## Managing API/Configuration Changes

### Overview

When you make changes to the API proto file (`api/proto/v1/api.proto`), those changes must be propagated across multiple repositories and examples to maintain consistency across the ecosystem.

### Step 1: Update Proto Definition

1. Make your changes to `api/proto/v1/api.proto`
2. Regenerate Go code and Docs:
   ```bash
   make generate
   ```

### Step 2: Update Local Examples and E2E Tests

**Update Go Client Examples:**

```bash
cd api/client/examples
# Update main.go to reflect API changes
vim main.go
# Test the example, follow the readme and test
go run main.go
```

**Update E2E Tests and Configuration:**

If API changes affect the relay sidecar configuration or client usage:

**E2E Test Files** (`e2e/tests/`):
- Update client instantiation in test setup files
- Update test cases that use the changed RPC methods
- Update `sidecar.yaml` if new configuration fields are required

**E2E Scripts** (`e2e/scripts/`):
- Update sidecar startup scripts if config template needs new fields
- Update genesis generation script if relay utils CLI commands changed
- Update network generation script if new environment variables are needed

**Common changes:**
- **New RPC methods** → Add corresponding test cases
- **Config field changes** → Update sidecar configuration files and startup templates
- **CLI command changes** → Update scripts that invoke relay utils commands
- **Client interface changes** → Update test setup and client instantiation code

### Step 3: Update External Client Repositories

#### TypeScript Client ([relay-client-ts](https://github.com/symbioticfi/relay-client-ts))

The TypeScript client has an automated workflow that regenerates the gRPC client when the proto file changes. However, you still need to:

1. **Wait for/Trigger the workflow** that updates the generated client code
2. **Update examples manually**:
   ```bash
   cd examples/
   # Update example files to use new API
   vim basic-usage.ts
   npm run build
   npm run basic-usage
   ```

#### Rust Client ([relay-client-rs](https://github.com/symbioticfi/relay-client-rs))

Similar to TypeScript, the Rust client auto-generates the gRPC implementation:

1. **Wait for/Trigger the workflow** that updates the generated client code
2. **Update examples manually**:
   ```bash
   cd examples/
   # Update example files to use new API
   vim basic_usage.rs
   cargo build
   cargo run --example basic_usage
   ```

### Step 4: Update Symbiotic Super Sum Example

The [symbiotic-super-sum](https://github.com/symbioticfi/symbiotic-super-sum) repository provides a comprehensive task-based example.

**What to update:**

1. **Go Client Package Version**:
   ```bash
   cd off-chain
   # Update go.mod to use latest relay client
   go get github.com/symbioticfi/relay@<commit/tag>
   go mod tidy
   ```

2. **Network Configuration**:
   - Check if any configuration changes are needed in network setup
   - The `generate_network.sh` script in symbiotic-super-sum is similar to the E2E setup
   - Update environment variables or config files if the relay now requires new settings

3. **Example Application Code**:
   - Update any client usage in the off-chain application
   - Test the example end-to-end

4. **Documentation**:
   - Update README if API usage patterns have changed

### Step 5: Update Cosmos Relay SDK

The [cosmos-relay-sdk](https://github.com/symbioticfi/cosmos-relay-sdk) integrates the relay with Cosmos SDK chains.

**What to update:**

1. **Go Client Package**:
   ```bash
   # Update the client package version
   go get github.com/symbioticfi/relay/api/client/v1@latest
   go mod tidy
   ```

2. **Mock Relay Client** (`x/symstaking/types/mock_relay.go`):
   
   The mock relay client must match the updated client interface:
   
   Example of what needs updating:
   - If you added a new RPC method, add it to the mock
   - If you changed method signatures, update them in the mock
   - If you changed request/response types, update the mock accordingly

3. **Documentation**:
   - Update SDK docs to reflect new API capabilities

---

## Code Generation

The project uses code generation extensively. Here's what each target generates:

### Generate All

```bash
make generate
```

This runs all generation targets that need codegen, including mocks, proto messages, etc.

---

## Linting

### Run All Linters

```bash
make lint
```

This runs:
- `buf lint` for proto files
- `golangci-lint` for Go code

### Auto-fix Linting Issues

```bash
make go-lint-fix
```

---

## Building Docker Images

Build the relay sidecar Docker image:

```bash
make image
```

To build and push multi-architecture images:

```bash
PUSH_IMAGE=true PUSH_LATEST=true make image
```

---

## Contributing

Please read [CONTRIBUTING.md](./CONTRIBUTING.md) for our branching strategy, PR process, and commit conventions.

### Key Points:

- **Target branch**: Always create PRs against `dev`, never `main`
- **Tests**: Ensure all tests pass before submitting PR
- **Linting**: Run `make lint` and fix all issues
- **Documentation**: Update docs when changing APIs or behavior

---

## Additional Resources

- **API Documentation**: [api/docs/v1/doc.md](api/docs/v1/doc.md)
- **Main README**: [README.md](README.md)
- **Contributing Guide**: [CONTRIBUTING.md](CONTRIBUTING.md)
- **Official Docs**: https://docs.symbiotic.fi/category/relay-sdk
- **Example Client Usage**: [api/client/examples/README.md](api/client/examples/README.md)

