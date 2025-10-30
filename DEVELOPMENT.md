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
9. [Logging Conventions](#logging-conventions)
10. [Building Docker Images](#building-docker-images)
11. [Contributing](#contributing)
12. [Additional Resources](#additional-resources)

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

## Logging Conventions

### Overview

The Symbiotic Relay uses Go's standard `log/slog` (structured logging) with a custom wrapper located in `pkg/log/`. This provides:

- **Structured logging** with type-safe fields
- **Context propagation** for automatic field inheritance
- **Component-based tagging** for filtering and debugging
- **Multiple output modes**: pretty (colored), text, and JSON

**Configuration:**
Set log level and mode in `sidecar.yaml`:
```yaml
log:
  level: debug  # debug, info, warn, error
  mode: pretty  # pretty, text, json
```

### Log Levels - When to Use

| Level                   | When to Use                                                           | Examples                                                                                     |
|-------------------------|-----------------------------------------------------------------------|----------------------------------------------------------------------------------------------|
| **Debug**               | Flow tracing, detailed decisions, skipped operations, variable values | "Skipped signing (not in validator set)", "Checked for missing epochs"                       |
| **Info** (prod default) | Key state changes, service lifecycle events, completed operations     | "Signature aggregation completed", "Started P2P listener", "Submitted proof to chain"        |
| **Warn**                | Recoverable issues, unusual but handled situations, security concerns | "Message signing disabled", "Retrying after temporary failure", "Invalid signature received" |
| **Error**               | Unrecoverable failures, critical problems requiring attention         | "Failed to load validator set", "Database connection lost", "Contract call failed"           |

**Important:** Log errors only at high-level error handlers (API handlers, main event loops, service entry points), not at every location where an error occurs. Internal functions should return errors without logging them, allowing the top-level handler to log once with full context.


### Context Propagation Pattern

**Always use context-aware logging variants** (`slog.InfoContext`, `slog.DebugContext`, etc.) when context is available.

```go
func (s *Service) HandleRequest(ctx context.Context, req *Request) error {
    // 1. Add component tag (required)
    ctx = log.WithComponent(ctx, "aggregator")

    // 2. Add request-specific context (if applicable)
    ctx = log.WithAttrs(ctx,
        slog.Uint64("epoch", uint64(req.Epoch)),
        slog.String("requestId", req.RequestID.Hex()),
    )

    // 3. All subsequent logs automatically include these fields
    slog.InfoContext(ctx, "Started processing request")

    // ... rest of implementation
    return nil
}

```
### Standard Field Names

All field names **must** use `camelCase` notation. Use these standard field names consistently:

#### Request Context
- `requestId` - Unique request identifier (use for cross-service correlation)
- `epoch` - Epoch number (uint64)

#### Identifiers
- `keyTag` - Key tag identifier
- `operator` - Operator address/ID
- `publicKey` - Public key (formatted as hex)
- `address` - Ethereum address
- `validatorIndex` - Validator index in the set

#### Operation Tracking
- `error` - Error message/object (always use "error", not "err")
- `duration` - Operation duration (use `time.Since()`)
- `attempt` - Current retry attempt number
- `maxRetries` - Maximum retry attempts

#### Network/P2P
- `topic` - P2P topic name
- `sender` - Message sender identifier
- `peer` - Peer identifier
- `peerId` - Libp2p peer ID

#### Component/Method
- `component` - Component name (auto-added via `log.WithComponent()`)
- `method` - gRPC method name or function identifier

#### Blockchain
- `chainId` - Chain identifier
- `blockNumber` - Block number
- `txHash` - Transaction hash
- `contractAddress` - Smart contract address

#### Custom Fields
When adding custom fields not in this list:
- Use descriptive `camelCase` names
- Prefer specific names over generic ones (`validatorCount` not `count`)
- Document new commonly-used fields by updating this list

### Error Logging Standards

**1. Error Wrapping (Internal Functions):**

Always wrap errors using `github.com/go-errors/errors` to capture stacktrace in the place of error:

```go
import "github.com/go-errors/errors"

func (s *Service) loadData(id string) error {
    data, err := s.repo.Get(id)
    if err != nil {
        // Wrap with context, preserve stack trace
        return errors.Errorf("failed to load data for id=%s: %w", id, err)
    }
    return nil
}
```

**2. Error Logging (Boundaries Only):**

Log errors at **boundaries** (API handlers, main loops, service entry points):

**3. Always use `"error"` as the field name** (not "err"):

```go
// ❌ Incorrect: Using "err" as field name
slog.ErrorContext(ctx, "Failed to process request", "err", err)
slog.WarnContext(ctx, "Retry attempt failed", "err", retryErr)

// ✅ Correct: Always use "error" for consistency
slog.ErrorContext(ctx, "Failed to process request", "error", err)
slog.WarnContext(ctx, "Retry attempt failed", "error", retryErr)
```

### Duration Tracking

Always track and log operation durations for performance monitoring:

```go
func (s *Service) ProcessSignature(ctx context.Context, sig *Signature) error {
    start := time.Now()

    // ... perform operation ...

    slog.InfoContext(ctx, "Signature processed successfully",
        "duration", time.Since(start),
        "requestId", sig.RequestID,
    )
    return nil
}
```

### Log Message Format

Log messages **must** follow these conventions:

**1. Start with past tense verb:**

```go
// ✅ Correct: Past tense verbs indicating completed actions
slog.InfoContext(ctx, "Signature received")
slog.InfoContext(ctx, "Validator set loaded")
slog.InfoContext(ctx, "Proof submitted to chain")
slog.DebugContext(ctx, "Checked for missing epochs")

// ❌ Incorrect
slog.InfoContext(ctx, "Receiving signature")      // present continuous
slog.InfoContext(ctx, "Load validator set")       // imperative
slog.InfoContext(ctx, "Signature receive")        // noun only
```

**Note:** Present continuous tense ("Processing...", "Aggregating signatures...") may be valid for long-running operations where you want to communicate progress and show that the process is active, not stuck. However, always pair these with a past tense completion log:

```go
// ✅ Acceptable for long-running operations
slog.DebugContext(ctx, "Aggregating signatures from validators", "count", validatorCount)
// ... long operation ...
slog.InfoContext(ctx, "Aggregation completed", "duration", time.Since(start))
```

**2. Use consistent terminology:**

- "Started" / "Completed" for long operations
- "Received" / "Sent" for messages
- "Loaded" / "Stored" for data operations
- "Failed" for errors (not "Error:", the log level indicates it's an error)

### Component Naming Conventions

Use these standard component names with `log.WithComponent()`:

| Component         | Usage                      |
|-------------------|----------------------------|
| `"grpc"`          | gRPC handlers              |
| `"signer"`        | Signer application         |
| `"aggregator"`    | Aggregator application     |
| `"sign_listener"` | Signature listener service |
| `"listener"`      | Validator set listener     |
| `"p2p"`           | P2P network layer          |
| `"evm"`           | EVM client interactions    |

Keep component names:
- Lowercase
- Short and recognizable
- Consistent across the codebase

### *Prefer context-aware logging variants whenever possible

```go
// ✅ Preferred (context-aware)
slog.InfoContext(ctx, "Message processed", "count", count)
slog.ErrorContext(ctx, "Failed to process", "error", err)

// ⚠️ Avoid when context is available (legacy pattern)
slog.Info("Message processed", "count", count)
slog.Error("Failed to process", "error", err)
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

- **API Documentation**: [docs/api/v1/doc.md](docs/api/v1/doc.md)
- **Main README**: [README.md](README.md)
- **Contributing Guide**: [CONTRIBUTING.md](CONTRIBUTING.md)
- **Official Docs**: https://docs.symbiotic.fi/category/relay-sdk
- **Example Client Usage**: [api/client/examples/README.md](api/client/examples/README.md)

