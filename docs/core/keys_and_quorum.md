# Keys and Quorum

## Description

The Symbiotic Relay supports a flexible key management system where each validator can have multiple cryptographic keys, each serving different purposes. This enables fine-grained control over signing operations, aggregation capabilities, and quorum requirements for different use cases.

**See also**:
- [`ValidatorKey`](./types.md#validatorkey) for key structure
- [`KeyTag`](./types.md#keytag) for key tag encoding
- [`NetworkConfig`](./types.md#networkconfig) for quorum threshold configuration

## Multiple Keys per Validator

Each validator in the system can possess multiple cryptographic keys, where each key is identified by a unique [`KeyTag`](./types.md#keytag). This design enables:

- **Separation of Concerns**: Different keys can be used for different purposes (e.g., consensus signing, light operations, heavy operations)
- **Flexible Quorum Configuration**: Each key tag can have its own quorum threshold, allowing different security levels for different operations
- **Key Mapping**: External keys (e.g., from other consensus systems) can be mapped to aggregation-enabled keys, enabling interoperability

### Key Structure

A validator's keys are stored in the `Keys` field of the [`Validator`](./types.md#validator) structure:

```go
type Validator struct {
    Operator    common.Address
    VotingPower VotingPower
    IsActive    bool
    Keys        []ValidatorKey  // Multiple keys per validator
    Vaults      []ValidatorVault
}

type ValidatorKey struct {
    Tag     KeyTag              // Unique identifier (KeyType + Key ID)
    Payload CompactPublicKey    // Public key bytes
}
```

Each key is uniquely identified by its [`KeyTag`](./types.md#keytag), which encodes both the key type (upper 4 bits) and a key ID (lower 4 bits), allowing up to 16 keys per key type per validator.

## Supported Key Types

The system supports multiple key types, each with different capabilities:

| Key Type | Value | Listing Enabled | Signing Enabled | Aggregation Enabled | Description |
|----------|-------|-----------------|-----------------|---------------------|-------------|
| `KeyTypeBlsBn254` | 0 | ✅ Yes | ✅ Yes | ✅ Yes | BLS signatures on BN254 curve. Supports efficient signature aggregation. |
| `KeyTypeEcdsaSecp256k1` | 1 | ✅ Yes | ✅ Yes | ❌ No | ECDSA signatures on secp256k1 curve. Standard Ethereum signing. |
| `KeyTypeBls12381Bn254` | 2 | ✅ Yes | ✅ Yes | ❌ No | BLS signatures on BLS12-381/BN254. Signing only, no aggregation. |
| `KeyTypeBls12381Bn254` | any | ✅ Yes | ❌ No | ❌ No | Any unknown key type can be just listed. |
| `KeyTypeInvalid` | 255 | ❌ No | ❌ No | ❌ No | Invalid key type. |

### Key Type Properties

**Listing Enabled**: All valid key types (except `KeyTypeInvalid`) can be registered and listed in the validator set. This means validators can have keys of these types associated with them.

**Signing Enabled**: Keys with signing enabled can be used to sign messages.

**Aggregation Enabled**: Currently, only `KeyTypeBlsBn254` supports signature aggregation. This means:
- Individual signatures from BLS BN254 keys can be aggregated into a single proof
- Aggregated proofs can be verified efficiently on-chain using a single pairing check
- Non-signers are explicitly tracked in the aggregation proof

## Quorum Thresholds

The system supports multiple quorum thresholds, where each threshold is associated with a specific [`KeyTag`](./types.md#keytag). This allows different security requirements for different operations.

### Configuration

Quorum thresholds are configured in [`NetworkConfig`](./types.md#networkconfig) via the `QuorumThresholds` field:

```go
type QuorumThreshold struct {
    KeyTag          KeyTag              // Key tag this threshold applies to
    QuorumThreshold QuorumThresholdPct  // Threshold as percentage (0-10^18)
}

type NetworkConfig struct {
    // ... other fields ...
    QuorumThresholds []QuorumThreshold  // Multiple thresholds per key tag
    RequiredKeyTags  []KeyTag           // Key tags validators must have
    RequiredHeaderKeyTag KeyTag          // Key tag for valset header commitments
}
```

### Threshold Calculation

Quorum thresholds are specified as percentages using a scale where `10^18` represents 100%. The absolute quorum threshold for a given key tag is calculated as:

```
absoluteThreshold = (totalVotingPower × thresholdPercentage) / 10^18 + 1
```

The `+1` ensures rounding up, so the threshold is always at least the specified percentage.

### Multiple Thresholds per Key Tag

For aggregation-enabled keys (BLS BN254), you can configure multiple quorum thresholds by using different key IDs within the same key type. For example:

- `KeyTag 0x00` (BLS BN254, ID 0): 66% threshold for heavy operations
- `KeyTag 0x01` (BLS BN254, ID 1): 51% threshold for light operations
- `KeyTag 0x02` (BLS BN254, ID 2): 80% threshold for critical operations

Each threshold is independent and applies only to signature requests using that specific key tag.

## Use Cases

### Multiple Use Cases with Different Thresholds

A common pattern is to configure multiple keys for the same validator with different quorum thresholds, each optimized for different use cases:

**Example: Heavy vs. Light Thresholds**

```yaml
NetworkConfig:
  QuorumThresholds:
    - KeyTag: 0x00  # BLS BN254, ID 0 - Heavy operations
      QuorumThreshold: 800000000000000000  # 80% threshold
    - KeyTag: 0x01  # BLS BN254, ID 1 - Light operations  
      QuorumThreshold: 510000000000000000  # 51% threshold
  RequiredKeyTags: [0x00, 0x01]
```

**Use Cases:**
- **Heavy Threshold (80%)**: Used for critical operations requiring high security, such as:
  - Large value transfers
  - Protocol upgrades
  - Security-sensitive operations
  
- **Light Threshold (51%)**: Used for routine operations requiring lower latency, such as:
  - Frequent state updates
  - Low-value transactions
  - High-throughput operations

This allows the same validator set to support both high-security and high-throughput use cases simultaneously.

### Mapping External Keys to Aggregation Keys

Another powerful use case is mapping external keys (e.g., from other blockchain networks or consensus systems) to aggregation-enabled keys in the relay system.

**Example: External Consensus Key Mapping**

```yaml
Validator:
  Operator: 0x1234...
  Keys:
    - Tag: 0x10  # ECDSA secp256k1, ID 0 - External consensus key
      Payload: <external_consensus_public_key>
    - Tag: 0x00  # BLS BN254, ID 0 - Aggregation key (mapped)
      Payload: <bls_bn254_public_key>
```

**How It Works:**
1. The validator registers their external consensus key (e.g., ECDSA secp256k1) with ID 0
2. The same validator also registers a BLS BN254 aggregation key with ID 0
3. The system treats both keys as belonging to the same validator (same operator address)
4. When a signature request is made with the BLS BN254 key tag, the validator uses their aggregation key
5. The aggregation key enables the validator to participate in efficient signature aggregation

**Benefits:**
- **Interoperability**: Existing consensus systems can integrate with the relay without changing their key infrastructure
- **Efficiency**: External keys can be mapped to aggregation-enabled keys, enabling efficient on-chain verification
- **Flexibility**: Validators can maintain their existing key management while gaining aggregation capabilities

**Real-World Scenario:**
A validator running a PoS blockchain with ECDSA consensus keys wants to participate in the relay network. They:
1. Keep their existing ECDSA keys for their primary blockchain
2. Generate additional BLS BN254 keys for relay participation
3. Register both key types under the same operator address
4. Use ECDSA keys for their blockchain operations
5. Use BLS BN254 keys for relay aggregation operations

This mapping allows arbitrary key sets (that may not natively support aggregation) to be "turned into" aggregatable sets by associating them with BLS BN254 keys.

## Key Tag Structure

A [`KeyTag`](./types.md#keytag) is an 8-bit value encoding both the key type and key ID:

```
KeyTag = (KeyType << 4) | KeyID
```

- **Upper 4 bits (bits 4-7)**: Key type (0-15, currently 0-2 used)
- **Lower 4 bits (bits 0-3)**: Key ID (0-15, allows 16 keys per type)

**Examples:**
- `0x00` = BLS BN254, ID 0
- `0x01` = BLS BN254, ID 1
- `0x10` = ECDSA secp256k1, ID 0
- `0x11` = ECDSA secp256k1, ID 1
- `0x20` = BLS12381 BN254, ID 0

## Reserved Key Tags

| Key Tag | Value | Description |
|---------|-------|-------------|
| `ValsetHeaderKeyTag` | 15 (0x0F) | Reserved for validator set header commitments. |

## References

- [`Types`](./types.md) for complete type definitions
- [`Signature Aggregation`](./signature_aggregation.md) for how aggregation works
- [`Valset Commitment`](./valset_commitment.md) for how keys are used in commitments

