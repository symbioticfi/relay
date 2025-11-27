# Architecture Overview

## Introduction

The Symbiotic Relay is a distributed system that enables cryptographic signature aggregation and validator set management across multiple blockchain networks. The system coordinates validator nodes to produce aggregation proofs that can be verified on-chain, enabling trustless verification of collective validator consensus.

**See**: [Core Types Reference](./types.md) for detailed type definitions

## System Architecture

The system operates through four main interconnected processes:

1. **Valset Derivation**: Derives validator sets from on-chain data across multiple chains
2. **Signature Aggregation**: Aggregates individual validator signatures into cryptographic proofs
3. **Valset Commitment**: Commits validator sets to settlement contracts with cryptographic proofs
4. **Sign Message API**: Provides an API for client applications to request signatures from the network

## Core Components

### Validator Set Management

The system maintains a chain of validator sets (see [`ValidatorSet`](./types.md#validatorset)), where each epoch has a derived validator set that represents the voting power distribution at that point in time. Validator sets are deterministically derived from on-chain data, ensuring all nodes compute the same set.

**See**: [Epoch Progression](./epoch_progression.md) for how epochs progress from creation through derivation to commitment

### Signature Aggregation

The system uses BLS (Boneh-Lynn-Shacham) signatures on the BN254 curve, which allows efficient aggregation of multiple signatures into a single proof. This enables quorum-based signing where a threshold of validator voting power must sign for a proof to be generated.

### Cross-Chain Coordination

The system operates across multiple blockchain networks:
- **Driver Chain**: Contains the ValSetDriver contract that defines epochs and network configuration
- **Settlement Chains**: Multiple chains where validator sets are committed for cross-chain verification
- **Voting Power Provider Chains**: Chains where voting power providers are deployed, allowing voting power aggregation across networks

### On-Chain Verification

All aggregation proofs (see [`AggregationProof`](./types.md#aggregationproof)) can be verified on-chain using settlement contracts, enabling trustless verification without requiring trust in the relay nodes themselves.

## Process Key Flows

### 1. Valset Derivation

When a new epoch occurs in the ValSetDriver contract, the system derives a new validator set by:
- Querying voting powers from VotingPowerProvider contracts across multiple chains
- Retrieving validator keys from the KeyRegistry
- Forming validators according to network configuration rules
- Calculating quorum thresholds and assigning roles

**See**: [Valset Derivation](./valset_derivation.md) for detailed flow

### 2. Signature Aggregation

When a signature request is created (either from valset commitment or API calls), the system:
- Distributes the request to all validator nodes
- Collects individual signatures via P2P network
- Aggregates signatures into a single proof when quorum is reached
- Broadcasts the proof to all nodes

**See**: [Signature Aggregation](./signature_aggregation.md) for detailed flow

### 3. Valset Commitment

After a valset is derived and an aggregation proof is generated, the system:
- Generates valset header and extra data (aggregated public keys)
- Creates commitment data for signing
- Commits the valset to settlement contracts on multiple chains
- Verifies each commitment against the previous committed valset

**See**: [Valset Commitment](./valset_commitment.md) for detailed flow

### 4. Sign Message API

Client applications can request signatures from the network by:
- Calling the SignMessage API on their respective nodes with the same data
- Waiting for the aggregation process to complete
- Retrieving the aggregation proof from their node
- Verifying the proof on-chain using settlement contracts

**See**: [Sign Message API](./sign_message.md) for detailed flow

## Key Design Principles

### Determinism

All processes are deterministic, ensuring that all nodes produce the same results from the same inputs. This includes:
- Validator set (see [`ValidatorSet`](./types.md#validatorset)) derivation from on-chain data
- Request ID calculation from message, key tag (see [`KeyTag`](./types.md#keytag)), and epoch
- Aggregation proof (see [`AggregationProof`](./types.md#aggregationproof)) generation from the same set of signatures

### Finality Guarantees

All on-chain queries use finalized blocks to ensure data consistency and prevent reorgs. This ensures that validator sets are derived from stable, irreversible blockchain state.

### Quorum-Based Security

All critical operations require a quorum of validator voting power:
- Valset commitments must be signed by sufficient validators
- API signature requests require quorum before aggregation
- Proofs explicitly track which validators signed and which did not

### Multi-Chain Support

The system is designed to operate across multiple blockchain networks:
- Voting power can be aggregated from providers on different chains
- Validator sets are committed to multiple settlement chains
- Proofs can be verified on any settlement chain

## Cryptographic Foundation

The system uses **BN254 Simple** aggregation for:
- **Extra Data Generation**: Aggregating validator public keys
- **Signature Aggregation**: Combining individual signatures into proofs
- **On-Chain Verification**: Efficient pairing checks on settlement contracts

BN254 Simple provides:
- Efficient aggregation of multiple signatures
- Single pairing check for verification regardless of signer count
- Explicit tracking of non-signers for quorum calculation

## System Lifecycle

1. **Genesis**: The first valset header and extra data are set through trusted genesis functionality on settlement contracts
2. **Epoch Transitions**: New epochs trigger valset derivation (see [Epoch Progression](./epoch_progression.md) for detailed lifecycle)
3. **Commitment Cycles**: Each derived valset is committed to settlement contracts with aggregation proofs
4. **API Requests**: Client applications can request signatures at any time, which are aggregated and can be verified on-chain

## Inter-Process Dependencies

The processes are interconnected:
- **Valset Derivation** → Creates validator sets that enable signing
- **Valset Commitment** → Requires aggregation proof from **Signature Aggregation**
- **Signature Aggregation** → Can be triggered by **Valset Commitment** or **Sign Message API**
- **Sign Message API** → Uses the same aggregation infrastructure as valset commitments

All processes rely on the same validator set structure (see [`ValidatorSet`](./types.md#validatorset)) and aggregation mechanisms, ensuring consistency across the system.

**See**: [Core Types Reference](./types.md) for complete type definitions


##  Related specifications

- [Epoch Progression](./epoch_progression.md)
- [Validator Set Derivation](./valset_commitment.md)
- [Validator Set Commitment](./valset_commitment.md)
- [Keys and quorum](./keys_and_quorum.md)
- [Signature Aggregation](./signature_aggregation.md)
- [Sign Message API](./sign_message.md)
- [Core Types Reference](./types.md)

