# Sign Message API

## Description

The Sign Message API allows client applications to request cryptographic signatures from the validator network. Each node has its own client application that calls the SignMessage API on that node. When all client applications call their respective nodes with the same message, key tag (see [`KeyTag`](./types.md#keytag)), and epoch, the nodes coordinate to produce an aggregation proof (see [`AggregationProof`](./types.md#aggregationproof)) that can be verified on-chain using settlement contracts.

### Process Overview

1. **API Request**: Each client application calls the SignMessage API endpoint on its respective node with:
   - The message to be signed
   - The key tag (see [`KeyTag`](./types.md#keytag)) specifying which validator keys should sign
   - The required epoch (or current epoch if not specified)

2. **Request Distribution**: Each node receives a SignMessage API call from its own client application. For the aggregation to work correctly, all client applications must call their nodes with exactly the same message, key tag, and epoch. This ensures all nodes generate the same request ID and coordinate on the same signature request.

3. **Signature Request Creation**: Each node creates a signature request (see [`SignatureRequest`](./types.md#signaturerequest)) internally. The request ID is deterministically calculated from the message hash, key tag, and epoch, ensuring all nodes generate the same request ID for identical inputs.

4. **Signature Aggregation**: The nodes follow the signature aggregation process (see [Signature Aggregation](./signature_aggregation.md)):
   - Signers sign the message
   - Signatures are collected via P2P
   - When quorum is reached, aggregators generate an aggregation proof

5. **Proof Retrieval**: Each client application can retrieve the aggregation proof from its own node using the GetAggregationProof API endpoint, providing the request ID returned from the SignMessage call.

6. **On-Chain Verification**: Any client application can verify the aggregation proof on-chain using settlement contracts. The settlement contract's `VerifyQuorumSigAt` function verifies:
   - The aggregation proof is valid (BN254 Simple pairing check)
   - The signers meet the quorum threshold
   - The proof corresponds to the specified epoch and message

### Key Features

- **Deterministic Request IDs**: Identical messages, key tags, and epochs produce the same request ID across all nodes, enabling coordination
- **Quorum-Based Signing**: Signatures are only aggregated when sufficient validator voting power has signed
- **On-Chain Verifiable**: Aggregation proofs can be verified on-chain using settlement contracts, enabling trustless verification
- **Multi-Node Coordination**: All nodes process the same request, ensuring consistent aggregation across the network
- **BN254 Simple**: Aggregation proofs use BN254 Simple aggregation for efficient on-chain verification

### Diagram

```mermaid
sequenceDiagram
    participant Client1 as Client App 1
    participant Node1 as Node 1
    participant Client2 as Client App 2
    participant Node2 as Node 2
    participant Client3 as Client App 3
    participant Node3 as Node 3
    participant Client4 as Client App 4
    participant Node4 as Node 4
    participant Settlement as Settlement Contract

    Note over Client1,Client4: SignMessage API call<br/>(same message, keyTag, epoch)

    par Client 1 -> Node 1
        Client1->>Node1: SignMessage(message, keyTag, epoch)
        Node1->>Node1: Create signature request<br/>(same requestID for same data)
        Node1-->>Client1: requestID, epoch
    and Client 2 -> Node 2
        Client2->>Node2: SignMessage(message, keyTag, epoch)
        Node2->>Node2: Create signature request<br/>(same requestID for same data)
        Node2-->>Client2: requestID, epoch
    and Client 3 -> Node 3
        Client3->>Node3: SignMessage(message, keyTag, epoch)
        Node3->>Node3: Create signature request<br/>(same requestID for same data)
        Node3-->>Client3: requestID, epoch
    and Client 4 -> Node 4
        Client4->>Node4: SignMessage(message, keyTag, epoch)
        Node4->>Node4: Create signature request<br/>(same requestID for same data)
        Node4-->>Client4: requestID, epoch
    end

    Note over Node1,Node4: Signature aggregation process<br/>(see Signature Aggregation flow)

    par Client 1 polls Node 1
        loop Poll until proof available
            Client1->>Node1: GetAggregationProof(requestID)
            alt Proof not ready
                Node1-->>Client1: Not found / Not ready
            else Proof ready
                Node1-->>Client1: AggregationProof
            end
        end
    and Client 2 polls Node 2
        loop Poll until proof available
            Client2->>Node2: GetAggregationProof(requestID)
            alt Proof not ready
                Node2-->>Client2: Not found / Not ready
            else Proof ready
                Node2-->>Client2: AggregationProof
            end
        end
    and Client 3 polls Node 3
        loop Poll until proof available
            Client3->>Node3: GetAggregationProof(requestID)
            alt Proof not ready
                Node3-->>Client3: Not found / Not ready
            else Proof ready
                Node3-->>Client3: AggregationProof
            end
        end
    and Client 4 polls Node 4
        loop Poll until proof available
            Client4->>Node4: GetAggregationProof(requestID)
            alt Proof not ready
                Node4-->>Client4: Not found / Not ready
            else Proof ready
                Node4-->>Client4: AggregationProof
            end
        end
    end

    Note over Client1,Client4: Proof received

    Client1->>Settlement: VerifyQuorumSigAt(<br/>message, keyTag, epoch,<br/>threshold, proof)
    Note over Settlement: Verify aggregation proof<br/>(BN254 Simple pairing check)
    Settlement->>Settlement: Check quorum threshold
    Settlement->>Settlement: Verify proof validity
    Settlement-->>Client1: Verification result

    Note over Client1: Proof verified on-chain
```

