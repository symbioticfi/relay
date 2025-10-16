# Symbiotic Relay

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/symbioticfi/relay)

> [!WARNING]  
> The code is a work in progress and not production ready yet.
> Breaking changes may occur in the code updates as well as backward compatibility is not guaranteed.
> Use with caution.

## Overview

The Symbiotic Relay operates as a distributed middleware layer that facilitates:

- **Validator Set Management**: Derives and maintains validator sets across different epochs based on on-chain state
- **Signature Aggregation**: Collects individual validator signatures and aggregates them using BLS signatures or zero-knowledge proofs
- **Cross-Chain Coordination**: Manages validator sets across multiple EVM-compatible blockchains

## Architecture

The relay consists of several key components:

- **P2P Layer**: Uses libp2p with GossipSub for decentralized communication
- **Signer Nodes**: Sign messages using BLS/ECDSA keys
- **Aggregator Nodes**: Collect and aggregate signatures with configurable policies
- **Committer Nodes**: Submit aggregated proofs to settlement chains
- **API Server**: Exposes gRPC API for external clients

For detailed architecture information, see [DEVELOPMENT.md](DEVELOPMENT.md).

## Documentation

- **[Development Guide](DEVELOPMENT.md)** - Comprehensive guide for developers including testing, API changes, and code generation
- **[Relay Docs](https://docs.symbiotic.fi/category/relay-sdk)** - Official documentation
- **[Contributing](CONTRIBUTING.md)** - Contribution guidelines and workflow

## Running Examples

For a complete end-to-end example application using the relay, see:

**[Symbiotic Super Sum](https://github.com/symbioticfi/symbiotic-super-sum)** - A task-based network example demonstrating relay integration

## API

The relay exposes a gRPC API for interacting with the network. See:

- **API Documentation**: [api/docs/v1/doc.md](api/docs/v1/doc.md)
- **Proto Definitions**: [api/proto/v1/api.proto](api/proto/v1/api.proto)
- **Go Client**: [api/client/v1/](api/client/v1/)
- **Client Examples**: [api/client/examples/](api/client/examples/)

### Client Libraries

- **Go**: Included in this repository at `github.com/symbioticfi/relay/api/client/v1`
- **TypeScript**: [relay-client-ts](https://github.com/symbioticfi/relay-client-ts)
- **Rust**: [relay-client-rs](https://github.com/symbioticfi/relay-client-rs)


## Quick Start

### Use Pre-built Releases

Instead of building from source, you can download pre-built binaries from GitHub releases:

```bash
# Download the latest release for your platform
# Linux AMD64
wget https://github.com/symbioticfi/relay/releases/latest/download/relay_sidecar_linux_amd64
wget https://github.com/symbioticfi/relay/releases/latest/download/relay_utils_linux_amd64
chmod +x relay_sidecar_linux_amd64 relay_utils_linux_amd64

# macOS ARM64
wget https://github.com/symbioticfi/relay/releases/latest/download/relay_sidecar_darwin_arm64
wget https://github.com/symbioticfi/relay/releases/latest/download/relay_utils_darwin_arm64
chmod +x relay_sidecar_darwin_arm64 relay_utils_darwin_arm64

# Run the binaries
./relay_sidecar_linux_amd64 --config config.yaml
```

Browse all releases at: https://github.com/symbioticfi/relay/releases

### Use Docker Images

Pre-built Docker images are available from Docker Hub:

```bash
# Pull the latest image
docker pull symbioticfi/relay:latest

# Or pull a specific version
docker pull symbioticfi/relay:<tag>

# Run the relay sidecar
docker run -v $(pwd)/config.yaml:/config.yaml \
  symbioticfi/relay:latest \
  --config /config.yaml
```

Docker Hub: https://hub.docker.com/r/symbioticfi/relay


## Build localy

### Dependencies

- **Go 1.24.3+**
- **Docker & Docker Compose** (for local setup and E2E tests)
- **Node.js & Foundry** (for contract compilation in E2E)


### Build Binaries

Build the relay sidecar and utils binaries:

```bash
# For Linux
make build-relay-sidecar OS=linux ARCH=amd64
make build-relay-utils OS=linux ARCH=amd64

# For macOS ARM
make build-relay-sidecar OS=darwin ARCH=arm64
make build-relay-utils OS=darwin ARCH=arm64
```

### Build Docker Image

```bash
make image TAG=dev
```

## Local Setup

### Automated Local Network

Set up a complete local relay network with blockchain nodes and multiple relay sidecars:

```bash
make local-setup
```

This command:
1. Builds the relay Docker image
2. Sets up local blockchain nodes (Anvil)
3. Deploys contracts
4. Generates sidecar configurations
5. Starts relay nodes in Docker

**Customize the network** using environment variables (see [DEVELOPMENT.md](DEVELOPMENT.md) for details):

```bash
OPERATORS=6 COMMITERS=2 AGGREGATORS=1 make local-setup
```

## Configuration File Structure

Create a `config.yaml` file with the following structure:

```yaml
# Logging
log-level: "debug"                    # Options: debug, info, warn, error
log-mode: "pretty"                    # Options: json, text, pretty

# Storage
storage-dir: ".data"                  # Directory for persistent data
circuits-dir: ""                      # Path to ZK circuits (optional, empty disables ZK proofs)

# API Server
server:
  listen: ":8080"                     # API server address
  verbose-logging: false              # Enable verbose API logging
  pprof: false                        # Enable pprof debug endpoints

# Metrics (optional)
metrics:
  listen: ":9090"                     # Metrics endpoint address

# Driver Contract
driver:
  chain-id: 31337                     # Chain ID where driver contract is deployed
  address: "0x..."                    # Driver contract address

# Secret Keys
secret-keys:
  - namespace: "symb"                 # Namespace for the key
    key-type: 0                       # 0=BLS-BN254, 1=ECDSA
    key-id: 15                        # Key identifier
    secret: "0x..."                   # Private key hex

  - namespace: "evm"
    key-type: 1
    key-id: 31337
    secret: "0x..."

  - namespace: "p2p"
    key-type: 1
    key-id: 1
    secret: "0x..."

# Alternatively, use keystore
# keystore:
#   path: "/path/to/keystore.json"
#   password: "your-password"

# Signal Configuration, used for internal messages and event queues
signal:
  worker-count: 10                    # Number of signal workers
  buffer-size: 20                     # Signal buffer size

# Cache Configuration, used for in memorylookups for db queries
cache:
  network-config-size: 10             # Network config cache size
  validator-set-size: 10              # Validator set cache size

# Sync Configuration, sync signatures and proofs over p2p to recover missing information
sync:
  enabled: true                       # Enable P2P sync
  period: 5s                          # Sync period
  timeout: 1m                         # Sync timeout
  epochs: 5                           # Number of epochs to sync

# Key Cache, used for fast public key lookups
key-cache:
  size: 100                           # Key cache size
  enabled: true                       # Enable key caching

# P2P Configuration
p2p:
  listen: "/ip4/0.0.0.0/tcp/8880"    # P2P listen address
  bootnodes:                          # List of bootstrap nodes (optional)
    - /dns4/node1/tcp/8880/p2p/...
  dht-mode: "server"                  # Options: auto, server, client, disabled, default: server (ideally should not change)
  mdns: true                         # Enable mDNS local discovery (useful for local networks)

# EVM Configuration
evm:
  chains:                             # List of settlement chain RPC endpoints
    - "http://localhost:8545"
    - "http://localhost:8546"
  max-calls: 30                       # Max calls in multicall batches

# Aggregation Policy
aggregation-policy-max-unsigners: 50  # Max unsigners for low-cost policy
```

#### Configuration via Command-Line Flags

You can override config file values with command-line flags:

```bash
./relay_sidecar \
  --config config.yaml \
  --log-level debug \
  --storage-dir /var/lib/relay \
  --server.listen ":8080" \
  --p2p.listen "/ip4/0.0.0.0/tcp/8880" \
  --driver.chain-id 1 \
  --driver.address "0x..." \
  --secret-keys "symb/0/15/0x...,evm/1/31337/0x..." \
  --evm.chains "http://localhost:8545"
```

#### Configuration via Environment Variables

Environment variables use the `SYMB_` prefix with underscores instead of dashes and dots:

```bash
export SYMB_LOG_LEVEL=debug
export SYMB_STORAGE_DIR=/var/lib/relay
export SYMB_SERVER_LISTEN=":8080"
export SYMB_P2P_LISTEN="/ip4/0.0.0.0/tcp/8880"
export SYMB_DRIVER_CHAIN_ID=1
export SYMB_DRIVER_ADDRESS="0x..."

./relay_sidecar --config config.yaml
```

#### Configuration Priority

Configuration is loaded in the following order (highest priority first):
1. Command-line flags
2. Environment variables (with `SYMB_` prefix)
3. Configuration file (specified by `--config`)

#### Example Configuration Generation

For reference, see how configurations are generated in the E2E setup:

```bash
# See the template in e2e/scripts/sidecar-start.sh (lines 11-27)
cat e2e/scripts/sidecar-start.sh
```

To customize the local setup configuration, modify the template in `e2e/scripts/sidecar-start.sh` and run:

```bash
make local-setup
```

### Running the Relay Sidecar

Once you have your configuration file ready:

```bash
./relay_sidecar --config config.yaml
```

Or with Docker:

```bash
docker run -v $(pwd)/config.yaml:/config.yaml symbioticfi/relay:latest --config /config.yaml
```


## Contributing

We welcome contributions! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Branching strategy and PR process
- Code style and linting requirements
- Testing requirements

For development workflows, API changes, and testing procedures, see [DEVELOPMENT.md](DEVELOPMENT.md).

## License

See [LICENSE](LICENSE) file for details.
