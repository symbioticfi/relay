# Aggregation

## Description

The aggregation process generates aggregation proofs (see [`AggregationProof`](./types.md#aggregationproof)) from individual validator signatures (see [`Signature`](./types.md#signature)) when a quorum threshold is reached. This process enables the network to produce a single cryptographic proof representing the collective agreement of validators, which can be efficiently verified on-chain.

### Process Overview

1. **Signature Request Reception**: Each node receives a signature request (see [`SignatureRequest`](./types.md#signaturerequest)) (e.g., when a new validator set is derived). The request contains the message to be signed, the required key tag (see [`KeyTag`](./types.md#keytag)), and the epoch.

2. **Individual Signing**: Each node that is a signer for the requested key tag signs the message using their private key. The signature is then broadcast to other nodes via P2P network.

3. **Signature Collection**: Nodes receive signatures from other validators via P2P. Each signature is validated (verifying the signature against the validator's public key) and stored in the local database. The system tracks which validators have signed using a signature map.

4. **Quorum Check**: Aggregator nodes continuously monitor the signature map. When a new signature is processed, aggregators check if the total voting power of signers has reached the quorum threshold defined in the validator set.

5. **Aggregation Proof Generation**: Once quorum is reached, aggregator nodes generate an aggregation proof. For BN254 Simple aggregation:
   - All individual signatures are aggregated into a single G1 point
   - All signer public keys are aggregated into a single G2 point
   - Non-signer validators are identified and encoded
   - Validator data (keys, voting powers) is encoded
   - The proof is assembled containing: aggregated signature (G1), aggregated public key (G2), validators data, and non-signer indices

6. **Proof Broadcast**: The generated aggregation proof is broadcast via P2P network to all nodes, allowing them to verify and use the proof.

### Key Features

- **Quorum-Based**: Aggregation only occurs when sufficient voting power has signed, ensuring security and consensus
- **BLS Signature Aggregation**: Uses BLS (Boneh-Lynn-Shacham) signatures on BN254 curve, allowing efficient aggregation of multiple signatures into a single proof
- **Deterministic**: All aggregators produce the same proof from the same set of signatures, ensuring consistency across the network
- **Efficient Verification**: The aggregated proof can be verified on-chain with a single pairing check, regardless of the number of signers
- **Non-Signer Tracking**: The proof explicitly tracks which validators did not sign, allowing the verifier to calculate the effective signing voting power

### Diagram (BN254 Simple)

```mermaid
sequenceDiagram
    participant Node1 as Node 1<br/>(Signer)
    participant Node2 as Node 2<br/>(Signer)
    participant Node3 as Node 3<br/>(Signer + Aggregator)
    participant Node4 as Node 4<br/>(Signer)
    participant P2P as P2P Network

    Note over Node1,Node4: Signature request received

    par Node 1 signs
        Node1->>Node1: Sign message with<br/>private key
        Node1->>P2P: Broadcast signature
    and Node 2 signs
        Node2->>Node2: Sign message with<br/>private key
        Node2->>P2P: Broadcast signature
    and Node 3 signs
        Node3->>Node3: Sign message with<br/>private key
        Node3->>P2P: Broadcast signature
    and Node 4 signs
        Node4->>Node4: Sign message with<br/>private key
        Node4->>P2P: Broadcast signature
    end

    Note over P2P: Signatures distributed via P2P

    par Node 1 receives signatures
        P2P->>Node1: Receive signatures<br/>from other nodes
        Node1->>Node1: Verify signatures
    and Node 2 receives signatures
        P2P->>Node2: Receive signatures<br/>from other nodes
        Node2->>Node2: Verify signatures
    and Node 3 receives signatures
        P2P->>Node3: Receive signatures<br/>from other nodes
        Node3->>Node3: Verify signatures
        Node3->>Node3: Check quorum threshold
    and Node 4 receives signatures
        P2P->>Node4: Receive signatures<br/>from other nodes
        Node4->>Node4: Verify signatures
    end

    Note over Node3: Quorum reached
    Note over Node3: Aggregate signatures<br/>(BN254 Simple)
    Node3->>Node3: Aggregate G1 signatures
    Node3->>Node3: Aggregate G2 public keys
    Node3->>Node3: Encode validators data
    Node3->>Node3: Identify non-signers
    Node3->>Node3: Assemble aggregation proof
    Node3->>P2P: Broadcast aggregation proof

    P2P->>Node1: Receive aggregation proof
    P2P->>Node2: Receive aggregation proof
    P2P->>Node3: Receive aggregation proof
    P2P->>Node4: Receive aggregation proof

    Note over Node1,Node4: All nodes have aggregation proof
```

