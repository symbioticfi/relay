# Symbiotic Relay

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/symbioticfi/relay)

## Overview

The Symbiotic Relay operates as a distributed middleware layer that facilitates:

- **Validator Set Management**: Derives and maintains validator sets across different epochs based on on-chain state
- **Signature Aggregation**: Collects individual validator signatures and aggregates them using BLS signatures or zero-knowledge proofs
- **Cross-Chain Coordination**: Manages validator sets across multiple EVM-compatible blockchains

## Documentation

- [Relay Docs](https://docs.symbiotic.fi/category/relay-sdk)

## Dependencies

- Go 1.24.3

## Build

The generic build targets require `OS` and `ARCH` parameters:

```bash
make build-relay-utils OS=linux ARCH=amd64
make build-relay-sidecar OS=darwin ARCH=arm64
```

## Lint & Test

```bash
make lint
make unit-test
```

## Run example

To run example please use this repo: https://github.com/symbioticfi/symbiotic-super-sum

## Contribution

You can find contribution guide here: [CONTRIBUTING.md](./CONTRIBUTING.md)
