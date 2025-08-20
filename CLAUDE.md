# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build
```bash
# Build relay utilities
make uild-relay-utils-linux     # linux/amd64
make build-relay-sidecar-darwin # darwin/arm64
```

### Lint and Test
```bash
make lint       # Runs both buf-lint and go-lint
make buf-lint   # Lint protobuf files
make go-lint    # Lint Go code with extensive ruleset
make unit-test  # Run unit tests with coverage
```

### Code Generation
```bash
make generate                # Generate all: mocks, API types, client types, ABI
make generate-api-types      # Generate protobuf types (buf generate)
make generate-mocks          # Generate test mocks (go generate ./...)
make generate-client-types   # Generate client types
make gen-abi                 # Generate Ethereum contract ABI bindings
```

**Important**: Always run `make generate-api-types` after modifying protobuf files in `api/proto/v1/`.

### Single Test Execution
```bash
go test ./path/to/specific/package -v
go test ./path/to/package -run TestSpecificFunction
```

## Architecture Overview

The Symbiotic Relay is a distributed middleware layer that facilitates validator set management, signature aggregation, and cross-chain coordination for Ethereum-based networks.

### High-Level Structure

```
core/           - Business logic and domain entities
├── client/     - External service clients (EVM)
├── entity/     - Core domain entities and types
├── usecase/    - Business use cases and strategies
└── symbiotic.go - Core service orchestration

internal/       - Implementation-specific services
├── client/     - Internal clients (P2P, repository)
├── gen/        - Generated code (protobuf, mocks)
├── usecase/    - Application services and apps
└── entity/     - Internal entities

api/            - External API definitions
├── proto/v1/   - gRPC protobuf definitions
├── client/v1/  - API client implementations
└── docs/v1/    - API documentation

cmd/            - Command-line applications
├── relay/      - Main relay sidecar binary
└── utils/      - Network and operator utilities
```

### Key Architectural Patterns

#### Hexagonal Architecture
The codebase follows hexagonal architecture principles whenever possible:
- **Core Domain** (`core/` and `/interval/usecase/`) contains business logic isolated from external dependencies
- **Ports** are defined as interfaces (e.g., `evmClient`, `repo`, `signer` interfaces)
- **Adapters** implement these interfaces in `internal/client/*` (e.g., EVM client, Badger repository)
- **Dependency Injection** is used to wire adapters to core use cases

#### Use Case Architecture
Business logic is organized in `core/usecase/` with pluggable strategies:
- **Aggregator**: Multiple signature aggregation strategies (BLS Simple, BLS ZK)
- **Crypto**: Key management for different cryptographic schemes
- **Growth Strategy**: Async vs Sync validator set growth patterns
- **Valset Deriver**: Derives validator sets from on-chain state

#### Entity-Driven Design
Core business objects are defined in `core/entity/entity.go`:
- `ValidatorSet`: Complete validator set with metadata
- `ValidatorSetHeader`: Cryptographic header for validator sets
- `Validator`: Individual validator with keys and voting power
- `NetworkConfig`: Cross-chain network configuration

#### Repository Pattern
Storage abstraction with multiple implementations:
- **Badger**: Persistent key-value storage (`internal/client/repository/badger/`)
- **Memory**: In-memory storage for testing (`internal/client/repository/memory/`)

### Cross-Chain Architecture

The relay manages validator sets across multiple EVM-compatible chains:

1. **Driver Contract**: Central coordination contract that defines network configuration
2. **Settlement Contracts**: Per-chain contracts that store committed validator set headers
3. **Voting Power Providers**: Cross-chain voting power calculation
4. **Key Registry**: Cross-chain validator key management

### P2P Communication

Uses libp2p for distributed signature collection:
- **Discovery**: DHT-based peer discovery
- **Message Broadcasting**: Signature request/response propagation
- **Aggregation**: Distributed signature aggregation with Byzantine fault tolerance

## Important Development Notes

### Error Handling
- Use `errors.Errorf()` instead of `fmt.Errorf()` (enforced by linter)
- Wrap errors with context using `errors.Errorf("context: %w", err)`
- **Error messages start with lowercase letter**: `"failed to parse config"`, not `"Failed to parse config"`

### Logging
- **Log messages start with capital letter**: `"Starting API server"`, not `"starting API server"`
- Use structured logging with `slog.InfoContext()`, `slog.ErrorContext()`, etc.
- Prefer to use `slog.InfoContext` instead of `slog.Info` for better context propagation

### Protobuf Development
- Always use generated getter methods: `req.GetField()` instead of `req.Field`
- Regenerate protobuf after schema changes: `make generate-api-types`
- Generated files in `internal/gen/api/v1/` should never be manually edited

### Code Generation
Files in these directories are auto-generated and should not be manually edited:
- `internal/gen/api/v1/` - Protobuf generated code
- `core/client/evm/gen/` - Contract ABI bindings
- `**/mocks/` - Test mocks

### Shared Patterns

#### Validator Set Retrieval
Use the shared `getValidatorSetForEpoch()` function in API server handlers rather than duplicating the logic of checking repository first, then deriving if not found.

#### EVM Client Optimization
When making multiple blockchain calls, consider:
- Batch RPC calls for multiple requests to same block
- Parallel calls to different replicas
- Circuit breaking for failed replicas

### Testing
- **Use testify require library**: `require.NoError(t, err)`, `require.Equal(t, expected, actual)`
- **Prefer table-driven tests**: Use `[]struct{}` pattern for multiple test cases whenever possible
- Mock interfaces are generated using `go:generate` directives
- Integration tests use build tag `integration`
- Repository implementations have comprehensive test coverage

## Configuration

The relay uses YAML configuration with these key sections:
- **Driver**: Central contract address and chain
- **Chains**: RPC URLs for supported chains  
- **Keys**: Cryptographic key management
- **P2P**: Network discovery and communication
- **Storage**: Persistence configuration

Example command-line usage:
```bash
relay_sidecar --driver.address 0x... --driver.chain-id 111 \
  --chains 111@http://127.0.0.1:8545 \
  --secret-keys symb/0/15/1000000000000000000 \
  --signer true --aggregator true --committer true
```