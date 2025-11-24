# Core Types Reference

The document contains abstract core types definitions with fields size and their underlying types.

*For golang specific types definition visit: https://github.com/symbioticfi/relay/tree/dev/symbiotic/entity*

## Structures

### NetworkConfig

| Field | Type | Size(b) | Description |
|-------|------|------|-------------|
| `VotingPowerProviders` | `[]`[`CrossChainAddress`](#crosschainaddress) | variable | List of VotingPowerProvider contract addresses across different chains |
| `KeysProvider` | [`CrossChainAddress`](#crosschainaddress) | 28 | KeyRegistry contract address for retrieving validator keys |
| `Settlements` | `[]`[`CrossChainAddress`](#crosschainaddress) | variable | List of Settlement contract addresses where valsets are committed |
| `VerificationType` | [`VerificationType`](#verificationtype) | 4 | Type of verification (BN254 Simple or BN254 ZK) |
| `MaxVotingPower` | `uint256` | 32 | Maximum voting power cap per validator (0 = no cap) |
| `MinInclusionVotingPower` | `uint256` | 32 | Minimum voting power required for validator inclusion |
| `MaxValidatorsCount` | `uint256` | 32 | Maximum number of validators (0 = no limit) |
| `RequiredKeyTags` | `[]`[`KeyTag`](#keytag) | variable | List of key tags required for validators |
| `RequiredHeaderKeyTag` | [`KeyTag`](#keytag) | 1 | Key tag required for signing valset header commitments |
| `QuorumThresholds` | `[]uint256` | variable | Quorum threshold configurations per key tag |
| `EpochDuration` | `uint64` | 8 | Duration of an epoch in seconds |
| `NumAggregators` | `uint64` | 8 | Number of aggregators assigned per epoch |
| `NumCommitters` | `uint64` | 8 | Number of committers assigned per epoch |
| `CommitterSlotDuration` | `uint64` | 8 | Duration of each committer's time slot in seconds |

### CrossChainAddress

| Field | Type | Size(b) | Description |
|-------|------|------|-------------|
| `ChainId` | `uint64` | 8 | Chain ID where the contract is deployed |
| `Address` | `address` | 20 | Contract address on the specified chain |

### ValidatorSet

| Field | Type | Size(b) | Description |
|-------|------|------|-------------|
| `Version` | `uint8` | 1 | Validator set version |
| `RequiredKeyTag` | [`KeyTag`](#keytag) | 1 | Key tag required to commit the next valset |
| `Epoch` | `uint64` | 8 | Epoch number for this validator set |
| `CaptureTimestamp` | `uint64` | 8 | Timestamp when the validator set was captured |
| `QuorumThreshold` | `uint256` | 32 | Absolute quorum threshold (not percentage) |
| `Validators` | `[]`[`Validator`](#validator) | variable | List of validators in the set |
| `Status` | [`ValidatorSetStatus`](#validatorsetstatus) | 1 | Current status (Derived, Aggregated, Committed, Missed) |
| `AggregatorIndices` | `[]uint32` | variable | Indices of validators assigned as aggregators (off-chain) |
| `CommitterIndices` | `[]uint32` | variable | Indices of validators assigned as committers (off-chain) |

**Note: Check [`ValidatorsSszMRoot`](#validatorssszmroot) to see limitations applied to validators size**

### ValidatorSetHeader

| Field | Type | Size(b) | Description |
|-------|------|------|-------------|
| `Version` | `uint8` | 1 | Validator set version |
| `RequiredKeyTag` | [`KeyTag`](#keytag) | 1 | Key tag required for signing |
| `Epoch` | `uint64` | 8 | Epoch number |
| `CaptureTimestamp` | `uint64` | 8 | Timestamp when validator set was captured |
| `QuorumThreshold` | `uint256` | 32 | Absolute quorum threshold |
| `TotalVotingPower` | `uint256` | 32 | Total voting power of active validators |
| `ValidatorsSszMRoot` | `bytes32` | 32 | Merkle root of validators tree (SSZ encoding, see details: [`ValidatorsSszMRoot`](#validatorssszmroot)) |


### Validator

| Field | Type | Size(b) | Description |
|-------|------|------|-------------|
| `Operator` | `address` | 20 | Operator address |
| `VotingPower` | `uint256` | 32 | Total voting power (sum of all vaults) |
| `IsActive` | `bool` | 1 | Whether the validator is active |
| `Keys` | `[]`[`ValidatorKey`](#validatorkey) | variable | List of cryptographic keys associated with this validator |
| `Vaults` | `[]`[`ValidatorVault`](#validatorvault) | variable | List of vaults contributing voting power to this validator |

**Note: Check [`ValidatorsSszMRoot`](#validatorssszmroot) to see limitations applied to vaults, keys size**

#### ValidatorKey

| Field | Type | Size(b) | Description |
|-------|------|------|-------------|
| `Tag` | [`KeyTag`](#keytag) | 1 | Key tag identifying the key type and ID |
| `Payload` | `[]byte` | variable | Compact public key representation (on-chain format) |

#### ValidatorVault

| Field | Type | Size(b) | Description |
|-------|------|------|-------------|
| `ChainID` | `uint64` | 8 | Chain ID where the vault is deployed |
| `Vault` | `address` | 20 | Vault contract address |
| `VotingPower` | `uint256` | 32 | Voting power contributed by this vault |

### SignatureRequest

| Field | Type | Size(b) | Description |
|-------|------|------|-------------|
| `KeyTag` | [`KeyTag`](#keytag) | 1 | Key tag specifying which validator keys should sign |
| `RequiredEpoch` | `uint64` | 8 | Epoch in which the signature is required |
| `Message` | `[]byte` | variable | Raw message bytes to be signed |

### Signature

| Field | Type | Size(b) | Description |
|-------|------|------|-------------|
| `MessageHash` | `[]byte` | variable | Hash of the message (scheme depends on [`KeyTag`](#keytag)) |
| `KeyTag` | [`KeyTag`](#keytag) | 1 | Key tag used for validation |
| `Epoch` | `uint64` | 8 | Epoch for validation |
| `PublicKey` | `[]byte` | variable | Public key of the signer (format depends on [`KeyTag`](#keytag)) |
| `Signature` | `[]byte` | variable | Raw signature bytes (format depends on [`KeyTag`](#keytag)) |

### AggregationProof

| Field | Type | Size(b) | Description |
|-------|------|------|-------------|
| `MessageHash` | `[]byte` | variable | Hash of the message (scheme depends on [`KeyTag`](#keytag)) |
| `KeyTag` | [`KeyTag`](#keytag) | 1 | Key tag used for validation |
| `Epoch` | `uint64` | 8 | Epoch for validation |
| `Proof` | `[]byte` | variable | Raw aggregation proof bytes (format depends on [`VerificationType`](#verificationtype)) |


## Enums / unions

### KeyType

| Name | Value | Description |
| --- | --- | --- |
| `KeyTypeBlsBn254` | 0 | BLS signatures on BN254 curve |
| `KeyTypeEcdsaSecp256k1` | 1 | ECDSA signatures on secp256k1 curve |
| `KeyTypeBls12381Bn254` | 2 | BLS signatures on BLS12-381/BN254 |
| `KeyTypeInvalid` | 255 | Invalid key type |

Underlying type: `uint8` (in fact used only 4 bits, except `KeyTypeInvalid`)

### KeyTag

| Field | Bits | Description |
| --- | --- | --- |
| *`KeyType`* | `[0..4]` | [`KeyType`](#keytype) (upper 4 bits) |
| *`Key ID`* | `[5..8]` | Key ID (lower 4 bits) |

Underlying type: `uint8`

Reserved values:

| Name | Value | Description |
| --- | --- | --- |
| `ValsetHeaderKeyTag` | 15 | Reserved key tag for validator set header commitments |

### VerificationType

| Name | Value | Description |
| --- | --- | --- |
| `VerificationTypeBn254Simple` | 0 | BLS signature aggregation/verification on the BN254 curve (supports fast aggregation, single pairing verification). |
| `VerificationTypeBn254ZK` | 1 | Zero-knowledge proof based verification for BLS signatures on BN254 (used for privacy-preserving or batched proofs). |
| `VerificationTypeUnknown` | 255 | Unknown or unsupported verification type |

Underlying type: `uint32`

### ValidatorSetStatus

| Name | Value | Description |
| --- | --- | --- |
| `ValidatorSetStatusDerived` | 0 | The validator set has been derived from on-chain data but not yet aggregated or committed. |
| `ValidatorSetStatusAggregated` | 1 | The aggregation proof for the validator set has been created but not yet committed. |
| `ValidatorSetStatusCommitted` | 2 | The validator set has been successfully committed on-chain. |
| `ValidatorSetStatusMissed` | 3 | The validator set commitment was missed (e.g., not committed in time). |

Underlying type: `uint8`


## ValidatorsSszMRoot

The `ValidatorsSszMRoot` is a 32-byte Merkle root hash computed from the SSZ (Simple Serialize) encoding of the validator set. It provides a compact cryptographic commitment to the entire validator set structure.

**Construction:**

The root is computed by:
1. Converting the [`ValidatorSet`](#validatorset) to SSZ format (`SszValidatorSet`)
2. Computing the SSZ hash tree root using Merkleization

**SSZ Structure:**

The hash tree root is computed from the `SszValidators` structure, which contains only the validators list (Version is stored separately in [`ValidatorSetHeader`](#validatorsetheader) and is not included in the SSZ root):

```
SszValidators
└── Validators: []SszValidator (max 1,048,576)
    └── SszValidator
        ├── Operator: address (20 bytes)
        ├── VotingPower: uint256 (32 bytes)
        ├── IsActive: bool (1 byte)
        ├── Keys: []SszKey (max 128)
        │   └── SszKey
        │       ├── Tag: uint8 (1 byte)
        │       └── PayloadHash: bytes32 (32 bytes) - Keccak256 hash of Payload
        └── Vaults: []SszVault (max 1024)
            └── SszVault
                ├── ChainId: uint64 (8 bytes)
                ├── Vault: address (20 bytes)
                └── VotingPower: uint256 (32 bytes)
```

**Key Points:**
- Validator keys store `PayloadHash` (Keccak256 hash of the key payload) rather than the full payload in the SSZ structure
- Validators must be sorted by operator address (ascending) before encoding
- The SSZ encoding uses Merkleization with mix-in for variable-length lists

**Limitations:**

| Element | Maximum Count | Description |
|---------|---------------|-------------|
| Validators | 1,048,576 | Maximum number of validators in the set |
| Keys per Validator | 128 | Maximum number of keys per validator |
| Vaults per Validator | 1,024 | Maximum number of vaults per validator |

**Merkleization:**
- Uses SSZ Merkleization algorithm with mix-in for list lengths
- Tree heights: Validators list (20 levels), Keys list (7 levels), Vaults list (10 levels)
- The final root is a 32-byte hash representing the entire validator set structure

**Reference Implementation:**
 - Golang: [`relay/symbiotic/usecase/ssz/ssz.go`](https://github.com/symbioticfi/relay/tree/dev/symbiotic/usecase/ssz/ssz.go)
 - Typescript: [`relay-stats-ts/src/encoding.ts`](https://github.com/symbioticfi/relay-stats-ts/blob/main/src/encoding.ts)