# Validator Set Derivation

## Description

The valset derivation process is responsible for computing and storing validator sets (see [`ValidatorSet`](./types.md#validatorset)) when new epochs occur in the ValSetDriver contract. This process ensures that the validator set accurately reflects the voting power distribution across multiple chains at a specific point in time.

### Process Overview

1. **Epoch Detection**: The system continuously polls the ValSetDriver contract to detect when a new epoch occurs. All contract calls use finalized blocks to ensure data consistency and prevent reorgs.

2. **Configuration Retrieval**: When a new epoch is detected, the system retrieves:
   - The epoch start timestamp
   - Network configuration (see [`NetworkConfig`](./types.md#networkconfig)) including:
     - VotingPowerProvider contract addresses (deployed on multiple chains)
     - KeyRegistry contract address
     - Validator set formation rules (min/max voting power, validator limits, etc.)

3. **Cross-Chain Voting Power Aggregation**: The system queries voting powers from all VotingPowerProvider contracts in parallel. These providers may be deployed on different chains, allowing the validator set to aggregate voting power across multiple networks. All queries use finalized blocks to ensure deterministic Validator Set.

4. **Key Retrieval**: The system fetches operator keys from the KeyRegistry contract at the epoch timestamp, again using finalized blocks.

5. **Validator Set Formation**: The deriver combines voting powers and keys to form validators according to the network configuration rules:
   - Filters validators by minimum voting power thresholds
   - Applies maximum voting power caps if configured
   - Limits the total number of validators if specified
   - Sorts validators by voting power

6. **Quorum and Role Assignment**: The system calculates the quorum threshold and deterministically assigns aggregator and committer roles pseudo-randomly based on the validator set hash.


### Key Features

- **Multi-Chain Support**: VotingPowerProviders can be deployed on different chains, with voting powers aggregated across all chains
- **Deterministic Derivation**: The validator set is deterministically derived from on-chain data from finalized state, ensuring all nodes derive the same set
- **Configuration-Driven**: Validator set formation follows rules defined in the network configuration, allowing for flexible governance


### Diagram

```mermaid
sequenceDiagram
    participant ValSetDriver as ValSetDriver Contract<br/>(Driver Chain)
    participant Deriver as Valset Deriver
    participant VPP1 as VotingPowerProvider<br/>(Chain 1)
    participant VPP2 as VotingPowerProvider<br/>(Chain 2)
    participant KeyRegistry as KeyRegistry Contract
    participant DB as Local Database

    Note over ValSetDriver: New epoch occurs
    ValSetDriver->>ValSetDriver: Epoch transition event
    
    loop Polling for new epochs
        Deriver->>ValSetDriver: GetCurrentEpoch()<br/>(finalized block)
        ValSetDriver-->>Deriver: currentEpoch
        
        alt New epoch detected
            Deriver->>ValSetDriver: GetEpochStart(epoch)<br/>(finalized block)
            ValSetDriver-->>Deriver: epochStartTimestamp
            
            Deriver->>ValSetDriver: GetConfig(epochStartTimestamp, epoch)<br/>(finalized block)
            ValSetDriver-->>Deriver: NetworkConfig<br/>Includes:<br/>- VotingPowerProvider addresses<br/>- KeyRegistry address<br/>- Validator set formation rules
            
            Note over Deriver,VPP2: Query voting powers from all chains<br/>(waiting for finality)
            
            par Query Chain 1
                Deriver->>VPP1: GetVotingPowersAt(timestamp)<br/>(finalized block)
                VPP1-->>Deriver: OperatorVotingPower[]
            and Query Chain 2
                Deriver->>VPP2: GetVotingPowersAt(timestamp)<br/>(finalized block)
                VPP2-->>Deriver: OperatorVotingPower[]
            end
            
            Deriver->>KeyRegistry: GetKeys(timestamp)<br/>(finalized block)
            KeyRegistry-->>Deriver: OperatorWithKeys[]
            
            Note over Deriver: Form validators from<br/>voting powers + keys<br/>(Ruled by config)
            
            Deriver->>Deriver: Calculate quorum threshold
            Deriver->>Deriver: Assign aggregator/committer indices
            
            Deriver->>DB: SaveNetworkConfig(epoch, config)
            Deriver->>DB: SaveValidatorSet(epoch, valset)
            
            Note over DB: Valset stored in local DB
        end
    end
```