# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [v1/api.proto](#v1_api-proto)
    - [AggregationProof](#api-proto-v1-AggregationProof)
    - [ChainEpochInfo](#api-proto-v1-ChainEpochInfo)
    - [ExtraData](#api-proto-v1-ExtraData)
    - [GetAggregationProofRequest](#api-proto-v1-GetAggregationProofRequest)
    - [GetAggregationProofResponse](#api-proto-v1-GetAggregationProofResponse)
    - [GetAggregationProofsByEpochRequest](#api-proto-v1-GetAggregationProofsByEpochRequest)
    - [GetAggregationProofsByEpochResponse](#api-proto-v1-GetAggregationProofsByEpochResponse)
    - [GetAggregationStatusRequest](#api-proto-v1-GetAggregationStatusRequest)
    - [GetAggregationStatusResponse](#api-proto-v1-GetAggregationStatusResponse)
    - [GetCurrentEpochRequest](#api-proto-v1-GetCurrentEpochRequest)
    - [GetCurrentEpochResponse](#api-proto-v1-GetCurrentEpochResponse)
    - [GetCustomScheduleNodeStatusRequest](#api-proto-v1-GetCustomScheduleNodeStatusRequest)
    - [GetCustomScheduleNodeStatusResponse](#api-proto-v1-GetCustomScheduleNodeStatusResponse)
    - [GetLastAllCommittedRequest](#api-proto-v1-GetLastAllCommittedRequest)
    - [GetLastAllCommittedResponse](#api-proto-v1-GetLastAllCommittedResponse)
    - [GetLastAllCommittedResponse.EpochInfosEntry](#api-proto-v1-GetLastAllCommittedResponse-EpochInfosEntry)
    - [GetLastCommittedRequest](#api-proto-v1-GetLastCommittedRequest)
    - [GetLastCommittedResponse](#api-proto-v1-GetLastCommittedResponse)
    - [GetLocalValidatorRequest](#api-proto-v1-GetLocalValidatorRequest)
    - [GetLocalValidatorResponse](#api-proto-v1-GetLocalValidatorResponse)
    - [GetSignatureRequestIDsByEpochRequest](#api-proto-v1-GetSignatureRequestIDsByEpochRequest)
    - [GetSignatureRequestIDsByEpochResponse](#api-proto-v1-GetSignatureRequestIDsByEpochResponse)
    - [GetSignatureRequestRequest](#api-proto-v1-GetSignatureRequestRequest)
    - [GetSignatureRequestResponse](#api-proto-v1-GetSignatureRequestResponse)
    - [GetSignatureRequestsByEpochRequest](#api-proto-v1-GetSignatureRequestsByEpochRequest)
    - [GetSignatureRequestsByEpochResponse](#api-proto-v1-GetSignatureRequestsByEpochResponse)
    - [GetSignaturesByEpochRequest](#api-proto-v1-GetSignaturesByEpochRequest)
    - [GetSignaturesByEpochResponse](#api-proto-v1-GetSignaturesByEpochResponse)
    - [GetSignaturesRequest](#api-proto-v1-GetSignaturesRequest)
    - [GetSignaturesResponse](#api-proto-v1-GetSignaturesResponse)
    - [GetValidatorByAddressRequest](#api-proto-v1-GetValidatorByAddressRequest)
    - [GetValidatorByAddressResponse](#api-proto-v1-GetValidatorByAddressResponse)
    - [GetValidatorByKeyRequest](#api-proto-v1-GetValidatorByKeyRequest)
    - [GetValidatorByKeyResponse](#api-proto-v1-GetValidatorByKeyResponse)
    - [GetValidatorSetHeaderRequest](#api-proto-v1-GetValidatorSetHeaderRequest)
    - [GetValidatorSetHeaderResponse](#api-proto-v1-GetValidatorSetHeaderResponse)
    - [GetValidatorSetMetadataRequest](#api-proto-v1-GetValidatorSetMetadataRequest)
    - [GetValidatorSetMetadataResponse](#api-proto-v1-GetValidatorSetMetadataResponse)
    - [GetValidatorSetRequest](#api-proto-v1-GetValidatorSetRequest)
    - [GetValidatorSetResponse](#api-proto-v1-GetValidatorSetResponse)
    - [Key](#api-proto-v1-Key)
    - [ListenProofsRequest](#api-proto-v1-ListenProofsRequest)
    - [ListenProofsResponse](#api-proto-v1-ListenProofsResponse)
    - [ListenSignaturesRequest](#api-proto-v1-ListenSignaturesRequest)
    - [ListenSignaturesResponse](#api-proto-v1-ListenSignaturesResponse)
    - [ListenValidatorSetRequest](#api-proto-v1-ListenValidatorSetRequest)
    - [ListenValidatorSetResponse](#api-proto-v1-ListenValidatorSetResponse)
    - [SignMessageRequest](#api-proto-v1-SignMessageRequest)
    - [SignMessageResponse](#api-proto-v1-SignMessageResponse)
    - [Signature](#api-proto-v1-Signature)
    - [SignatureRequest](#api-proto-v1-SignatureRequest)
    - [Validator](#api-proto-v1-Validator)
    - [ValidatorSet](#api-proto-v1-ValidatorSet)
    - [ValidatorVault](#api-proto-v1-ValidatorVault)
  
    - [ErrorCode](#api-proto-v1-ErrorCode)
    - [SigningStatus](#api-proto-v1-SigningStatus)
    - [ValidatorSetStatus](#api-proto-v1-ValidatorSetStatus)
  
    - [SymbioticAPIService](#api-proto-v1-SymbioticAPIService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="v1_api-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## v1/api.proto



<a name="api-proto-v1-AggregationProof"></a>

### AggregationProof
Response message for getting aggregation proof


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message_hash | [bytes](#bytes) |  | Message hash |
| proof | [bytes](#bytes) |  | Proof data |
| request_id | [string](#string) |  | Request ID |






<a name="api-proto-v1-ChainEpochInfo"></a>

### ChainEpochInfo
Settlement chain with its last committed epoch


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| last_committed_epoch | [uint64](#uint64) |  | Last committed epoch for this chain |
| start_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Epoch start time |






<a name="api-proto-v1-ExtraData"></a>

### ExtraData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [bytes](#bytes) |  |  |
| value | [bytes](#bytes) |  |  |






<a name="api-proto-v1-GetAggregationProofRequest"></a>

### GetAggregationProofRequest
Request message for getting aggregation proof


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request_id | [string](#string) |  |  |






<a name="api-proto-v1-GetAggregationProofResponse"></a>

### GetAggregationProofResponse
Response message for getting aggregation proof


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| aggregation_proof | [AggregationProof](#api-proto-v1-AggregationProof) |  |  |






<a name="api-proto-v1-GetAggregationProofsByEpochRequest"></a>

### GetAggregationProofsByEpochRequest
Request message for getting aggregation proof


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| epoch | [uint64](#uint64) |  | Epoch number |






<a name="api-proto-v1-GetAggregationProofsByEpochResponse"></a>

### GetAggregationProofsByEpochResponse
Response message for getting aggregation proof


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| aggregation_proofs | [AggregationProof](#api-proto-v1-AggregationProof) | repeated |  |






<a name="api-proto-v1-GetAggregationStatusRequest"></a>

### GetAggregationStatusRequest
Request message for getting aggregation status


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request_id | [string](#string) |  |  |






<a name="api-proto-v1-GetAggregationStatusResponse"></a>

### GetAggregationStatusResponse
Response message for getting aggregation status


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| current_voting_power | [string](#string) |  | Current voting power of the aggregator (big integer as string) |
| signer_operators | [string](#string) | repeated | List of operator addresses that signed the request |






<a name="api-proto-v1-GetCurrentEpochRequest"></a>

### GetCurrentEpochRequest
Request message for getting current epoch






<a name="api-proto-v1-GetCurrentEpochResponse"></a>

### GetCurrentEpochResponse
Response message for getting current epoch


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| epoch | [uint64](#uint64) |  | Epoch number |
| start_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Epoch start time |






<a name="api-proto-v1-GetCustomScheduleNodeStatusRequest"></a>

### GetCustomScheduleNodeStatusRequest
Request to check if the current node should be active in a custom schedule.
The validator set is divided into groups that rotate through time slots.
Use this to coordinate distributed tasks among multiple application instances.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| epoch | [uint64](#uint64) | optional | Epoch number to use for the validator set (optional, defaults to current epoch) |
| seed | [bytes](#bytes) | optional | Custom seed for randomizing the schedule (optional). Different seeds produce different schedules for the same epoch, allowing multiple independent scheduling schemes. If not provided, the schedule is deterministic based on epoch alone. |
| slot_duration_seconds | [uint64](#uint64) |  | Duration of each time slot in seconds. Determines how frequently active groups rotate. Example: 60 seconds means a new group becomes active every minute. |
| max_participants_per_slot | [uint32](#uint32) |  | Maximum validators per group. Controls redundancy: 1 for single-instance actions (less redundancy), 2&#43; for multi-instance actions (more reliability). All validators in a group are active simultaneously. |
| min_participants_per_slot | [uint32](#uint32) |  | Minimum validators required to form a remainder group. When dividing validators into groups, any remainder smaller than this is not scheduled. Set equal to max for strict group sizes. |






<a name="api-proto-v1-GetCustomScheduleNodeStatusResponse"></a>

### GetCustomScheduleNodeStatusResponse
Response indicating whether the current node should be active now.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| is_active | [bool](#bool) |  | True if this node is active in the current time slot and should perform the scheduled action. False if this node should wait (another group is active). When multiple validators share a slot, all return true simultaneously, enabling coordinated redundancy. |






<a name="api-proto-v1-GetLastAllCommittedRequest"></a>

### GetLastAllCommittedRequest
Request message for getting last committed epochs for all chains

No parameters needed






<a name="api-proto-v1-GetLastAllCommittedResponse"></a>

### GetLastAllCommittedResponse
Response message for getting all last committed epochs


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| epoch_infos | [GetLastAllCommittedResponse.EpochInfosEntry](#api-proto-v1-GetLastAllCommittedResponse-EpochInfosEntry) | repeated | List of settlement chains with their last committed epochs |
| suggested_epoch_info | [ChainEpochInfo](#api-proto-v1-ChainEpochInfo) |  | Suggested epoch info for signatures, it is the minimum commited epoch among all chains |






<a name="api-proto-v1-GetLastAllCommittedResponse-EpochInfosEntry"></a>

### GetLastAllCommittedResponse.EpochInfosEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [uint64](#uint64) |  |  |
| value | [ChainEpochInfo](#api-proto-v1-ChainEpochInfo) |  |  |






<a name="api-proto-v1-GetLastCommittedRequest"></a>

### GetLastCommittedRequest
Request message for getting last committed epoch for a specific settlement chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| settlement_chain_id | [uint64](#uint64) |  | Settlement chain ID |






<a name="api-proto-v1-GetLastCommittedResponse"></a>

### GetLastCommittedResponse
Response message for getting last committed epoch


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| settlement_chain_id | [uint64](#uint64) |  | Settlement chain ID |
| epoch_info | [ChainEpochInfo](#api-proto-v1-ChainEpochInfo) |  |  |






<a name="api-proto-v1-GetLocalValidatorRequest"></a>

### GetLocalValidatorRequest
Request message for getting local validator


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| epoch | [uint64](#uint64) | optional | Epoch number (optional, if not provided current epoch will be used) |






<a name="api-proto-v1-GetLocalValidatorResponse"></a>

### GetLocalValidatorResponse
Response message for getting local validator


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| validator | [Validator](#api-proto-v1-Validator) |  | The validator |






<a name="api-proto-v1-GetSignatureRequestIDsByEpochRequest"></a>

### GetSignatureRequestIDsByEpochRequest
Request message for getting all signature request IDs by epoch


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| epoch | [uint64](#uint64) |  | Epoch number |






<a name="api-proto-v1-GetSignatureRequestIDsByEpochResponse"></a>

### GetSignatureRequestIDsByEpochResponse
Response message for getting all signature request IDs by epoch


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request_ids | [string](#string) | repeated | List of all signature request IDs for the epoch |






<a name="api-proto-v1-GetSignatureRequestRequest"></a>

### GetSignatureRequestRequest
Request message for getting signature request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request_id | [string](#string) |  |  |






<a name="api-proto-v1-GetSignatureRequestResponse"></a>

### GetSignatureRequestResponse
Response message for getting signature request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| signature_request | [SignatureRequest](#api-proto-v1-SignatureRequest) |  |  |






<a name="api-proto-v1-GetSignatureRequestsByEpochRequest"></a>

### GetSignatureRequestsByEpochRequest
Request message for getting all signature requests by epoch


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| epoch | [uint64](#uint64) |  | Epoch number |






<a name="api-proto-v1-GetSignatureRequestsByEpochResponse"></a>

### GetSignatureRequestsByEpochResponse
Response message for getting all signature requests by epoch


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| signature_requests | [SignatureRequest](#api-proto-v1-SignatureRequest) | repeated | List of all signature requests for the epoch |






<a name="api-proto-v1-GetSignaturesByEpochRequest"></a>

### GetSignaturesByEpochRequest
Request message for getting signatures by epoch


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| epoch | [uint64](#uint64) |  | Epoch number |






<a name="api-proto-v1-GetSignaturesByEpochResponse"></a>

### GetSignaturesByEpochResponse
Response message for getting signatures by epoch


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| signatures | [Signature](#api-proto-v1-Signature) | repeated | List of signatures |






<a name="api-proto-v1-GetSignaturesRequest"></a>

### GetSignaturesRequest
Request message for getting signatures


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request_id | [string](#string) |  |  |






<a name="api-proto-v1-GetSignaturesResponse"></a>

### GetSignaturesResponse
Response message for getting signatures


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| signatures | [Signature](#api-proto-v1-Signature) | repeated | List of signatures |






<a name="api-proto-v1-GetValidatorByAddressRequest"></a>

### GetValidatorByAddressRequest
Request message for getting validator by address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| epoch | [uint64](#uint64) | optional | Epoch number (optional, if not provided current epoch will be used) |
| address | [string](#string) |  | Validator address (required) |






<a name="api-proto-v1-GetValidatorByAddressResponse"></a>

### GetValidatorByAddressResponse
Response message for getting validator by address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| validator | [Validator](#api-proto-v1-Validator) |  | The validator |






<a name="api-proto-v1-GetValidatorByKeyRequest"></a>

### GetValidatorByKeyRequest
Request message for getting validator by key


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| epoch | [uint64](#uint64) | optional | Epoch number (optional, if not provided current epoch will be used) |
| key_tag | [uint32](#uint32) |  | Validator key tag (required) |
| on_chain_key | [bytes](#bytes) |  | Validator on chain (public) key (required) |






<a name="api-proto-v1-GetValidatorByKeyResponse"></a>

### GetValidatorByKeyResponse
Response message for getting validator by key


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| validator | [Validator](#api-proto-v1-Validator) |  | The validator |






<a name="api-proto-v1-GetValidatorSetHeaderRequest"></a>

### GetValidatorSetHeaderRequest
Request message for getting validator set header


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| epoch | [uint64](#uint64) | optional | Epoch number (optional, if not provided current epoch will be used) |






<a name="api-proto-v1-GetValidatorSetHeaderResponse"></a>

### GetValidatorSetHeaderResponse
Response message for getting validator set header


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| version | [uint32](#uint32) |  | Version of the validator set |
| required_key_tag | [uint32](#uint32) |  | Key tag required to commit next validator set |
| epoch | [uint64](#uint64) |  | Validator set epoch |
| capture_timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Epoch capture timestamp |
| quorum_threshold | [string](#string) |  | Quorum threshold (big integer as string) |
| total_voting_power | [string](#string) |  | Total voting power (big integer as string) |
| validators_ssz_mroot | [string](#string) |  | Validators SSZ Merkle root (hex string) |






<a name="api-proto-v1-GetValidatorSetMetadataRequest"></a>

### GetValidatorSetMetadataRequest
Request message for getting validator set metadata


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| epoch | [uint64](#uint64) | optional | Epoch number (optional, if not provided current epoch will be used) |






<a name="api-proto-v1-GetValidatorSetMetadataResponse"></a>

### GetValidatorSetMetadataResponse
Response message for getting validator set header


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| extra_data | [ExtraData](#api-proto-v1-ExtraData) | repeated |  |
| commitment_data | [bytes](#bytes) |  |  |
| request_id | [string](#string) |  |  |






<a name="api-proto-v1-GetValidatorSetRequest"></a>

### GetValidatorSetRequest
Request message for getting validator set


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| epoch | [uint64](#uint64) | optional | Epoch number (optional, if not provided current epoch will be used) |






<a name="api-proto-v1-GetValidatorSetResponse"></a>

### GetValidatorSetResponse
Response message for getting validator set


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| validator_set | [ValidatorSet](#api-proto-v1-ValidatorSet) |  | The validator set |






<a name="api-proto-v1-Key"></a>

### Key
Cryptographic key


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tag | [uint32](#uint32) |  | Key tag identifier (0-127) |
| payload | [bytes](#bytes) |  | Key payload |






<a name="api-proto-v1-ListenProofsRequest"></a>

### ListenProofsRequest
Request message for listening to aggregation proofs stream


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| start_epoch | [uint64](#uint64) | optional | Optional: start epoch. If provided, stream will first send all historical proofs starting from this epoch, then continue with real-time updates If not provided, only proofs generated after stream creation will be sent |






<a name="api-proto-v1-ListenProofsResponse"></a>

### ListenProofsResponse
Response message for aggregation proofs stream


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request_id | [string](#string) |  | Id of the request |
| epoch | [uint64](#uint64) |  | Epoch number |
| aggregation_proof | [AggregationProof](#api-proto-v1-AggregationProof) |  | Final aggregation proof |






<a name="api-proto-v1-ListenSignaturesRequest"></a>

### ListenSignaturesRequest
Request message for listening to signatures stream


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| start_epoch | [uint64](#uint64) | optional | Optional: start epoch. If provided, stream will first send all historical signatures starting from this epoch, then continue with real-time updates If not provided, only signatures generated after stream creation will be sent |






<a name="api-proto-v1-ListenSignaturesResponse"></a>

### ListenSignaturesResponse
Response message for signatures stream


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request_id | [string](#string) |  | Id of the signature request |
| epoch | [uint64](#uint64) |  | Epoch number |
| signature | [Signature](#api-proto-v1-Signature) |  | Signature data |






<a name="api-proto-v1-ListenValidatorSetRequest"></a>

### ListenValidatorSetRequest
Request message for listening to validator set changes stream


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| start_epoch | [uint64](#uint64) | optional | Optional: start epoch. If provided, stream will first send all historical validator sets starting from this epoch, then continue with real-time updates If not provided, only validator sets generated after stream creation will be sent |






<a name="api-proto-v1-ListenValidatorSetResponse"></a>

### ListenValidatorSetResponse
Response message for validator set changes stream


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| validator_set | [ValidatorSet](#api-proto-v1-ValidatorSet) |  | The validator set |






<a name="api-proto-v1-SignMessageRequest"></a>

### SignMessageRequest
Request message for signing a message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key_tag | [uint32](#uint32) |  | Key tag identifier (0-127) |
| message | [bytes](#bytes) |  | Message to be signed |
| required_epoch | [uint64](#uint64) | optional | Required epoch (optional, if not provided latest committed epoch will be used) |






<a name="api-proto-v1-SignMessageResponse"></a>

### SignMessageResponse
Response message for sign message request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request_id | [string](#string) |  | Hash of the signature request |
| epoch | [uint64](#uint64) |  | Epoch number |






<a name="api-proto-v1-Signature"></a>

### Signature
Digital signature


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| signature | [bytes](#bytes) |  | Signature data |
| message_hash | [bytes](#bytes) |  | Message hash |
| public_key | [bytes](#bytes) |  | Public key |
| request_id | [string](#string) |  | Request ID |






<a name="api-proto-v1-SignatureRequest"></a>

### SignatureRequest
SignatureRequest represents a signature request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request_id | [string](#string) |  | Request ID |
| key_tag | [uint32](#uint32) |  | Key tag identifier (0-127) |
| message | [bytes](#bytes) |  | Message to be signed |
| required_epoch | [uint64](#uint64) |  | Required epoch |






<a name="api-proto-v1-Validator"></a>

### Validator
Validator information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| operator | [string](#string) |  | Operator address (hex string) |
| voting_power | [string](#string) |  | Voting power of the validator (big integer as string) |
| is_active | [bool](#bool) |  | Indicates if the validator is active |
| keys | [Key](#api-proto-v1-Key) | repeated | List of cryptographic keys |
| vaults | [ValidatorVault](#api-proto-v1-ValidatorVault) | repeated | List of validator vaults |






<a name="api-proto-v1-ValidatorSet"></a>

### ValidatorSet



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| version | [uint32](#uint32) |  | Version of the validator set |
| required_key_tag | [uint32](#uint32) |  | Key tag required to commit next validator set |
| epoch | [uint64](#uint64) |  | Validator set epoch |
| capture_timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Epoch capture timestamp |
| quorum_threshold | [string](#string) |  | Quorum threshold (big integer as string) |
| status | [ValidatorSetStatus](#api-proto-v1-ValidatorSetStatus) |  | Status of validator set header |
| validators | [Validator](#api-proto-v1-Validator) | repeated | List of validators |






<a name="api-proto-v1-ValidatorVault"></a>

### ValidatorVault
Validator vault information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| chain_id | [uint64](#uint64) |  | Chain identifier |
| vault | [string](#string) |  | Vault address |
| voting_power | [string](#string) |  | Voting power for this vault (big integer as string) |





 


<a name="api-proto-v1-ErrorCode"></a>

### ErrorCode
Error code enumeration

| Name | Number | Description |
| ---- | ------ | ----------- |
| ERROR_CODE_UNSPECIFIED | 0 | Default/unknown error |
| ERROR_CODE_NO_DATA | 1 | No data found |
| ERROR_CODE_INTERNAL | 2 | Internal server error |
| ERROR_CODE_NOT_AGGREGATOR | 3 | Not an aggregator node |



<a name="api-proto-v1-SigningStatus"></a>

### SigningStatus
Signing process status enumeration

| Name | Number | Description |
| ---- | ------ | ----------- |
| SIGNING_STATUS_UNSPECIFIED | 0 | Default/unknown status |
| SIGNING_STATUS_PENDING | 1 | Request has been created and is waiting for signatures |
| SIGNING_STATUS_COMPLETED | 2 | Signing process completed successfully with proof |
| SIGNING_STATUS_FAILED | 3 | Signing process failed |
| SIGNING_STATUS_TIMEOUT | 4 | Signing request timed out |



<a name="api-proto-v1-ValidatorSetStatus"></a>

### ValidatorSetStatus
Validator set status enumeration

| Name | Number | Description |
| ---- | ------ | ----------- |
| VALIDATOR_SET_STATUS_UNSPECIFIED | 0 | Default/unknown status |
| VALIDATOR_SET_STATUS_DERIVED | 1 | Derived status |
| VALIDATOR_SET_STATUS_AGGREGATED | 2 | Aggregated status |
| VALIDATOR_SET_STATUS_COMMITTED | 3 | Committed status |


 

 


<a name="api-proto-v1-SymbioticAPIService"></a>

### SymbioticAPIService
SymbioticAPI provides access to the Symbiotic relay functions

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| SignMessage | [SignMessageRequest](#api-proto-v1-SignMessageRequest) | [SignMessageResponse](#api-proto-v1-SignMessageResponse) | Sign a message |
| GetAggregationProof | [GetAggregationProofRequest](#api-proto-v1-GetAggregationProofRequest) | [GetAggregationProofResponse](#api-proto-v1-GetAggregationProofResponse) | Get aggregation proof |
| GetAggregationProofsByEpoch | [GetAggregationProofsByEpochRequest](#api-proto-v1-GetAggregationProofsByEpochRequest) | [GetAggregationProofsByEpochResponse](#api-proto-v1-GetAggregationProofsByEpochResponse) | Get aggregation proofs by epoch |
| GetCurrentEpoch | [GetCurrentEpochRequest](#api-proto-v1-GetCurrentEpochRequest) | [GetCurrentEpochResponse](#api-proto-v1-GetCurrentEpochResponse) | Get current epoch |
| GetSignatures | [GetSignaturesRequest](#api-proto-v1-GetSignaturesRequest) | [GetSignaturesResponse](#api-proto-v1-GetSignaturesResponse) | Get signature by request id |
| GetSignaturesByEpoch | [GetSignaturesByEpochRequest](#api-proto-v1-GetSignaturesByEpochRequest) | [GetSignaturesByEpochResponse](#api-proto-v1-GetSignaturesByEpochResponse) | Get signature by epoch |
| GetSignatureRequestIDsByEpoch | [GetSignatureRequestIDsByEpochRequest](#api-proto-v1-GetSignatureRequestIDsByEpochRequest) | [GetSignatureRequestIDsByEpochResponse](#api-proto-v1-GetSignatureRequestIDsByEpochResponse) | Get all signature request IDs by epoch |
| GetSignatureRequestsByEpoch | [GetSignatureRequestsByEpochRequest](#api-proto-v1-GetSignatureRequestsByEpochRequest) | [GetSignatureRequestsByEpochResponse](#api-proto-v1-GetSignatureRequestsByEpochResponse) | Get all signature requests by epoch |
| GetSignatureRequest | [GetSignatureRequestRequest](#api-proto-v1-GetSignatureRequestRequest) | [GetSignatureRequestResponse](#api-proto-v1-GetSignatureRequestResponse) | Get signature request by request id |
| GetAggregationStatus | [GetAggregationStatusRequest](#api-proto-v1-GetAggregationStatusRequest) | [GetAggregationStatusResponse](#api-proto-v1-GetAggregationStatusResponse) | Get aggregation status, can be sent only to aggregator nodes |
| GetValidatorSet | [GetValidatorSetRequest](#api-proto-v1-GetValidatorSetRequest) | [GetValidatorSetResponse](#api-proto-v1-GetValidatorSetResponse) | Get current validator set |
| GetValidatorByAddress | [GetValidatorByAddressRequest](#api-proto-v1-GetValidatorByAddressRequest) | [GetValidatorByAddressResponse](#api-proto-v1-GetValidatorByAddressResponse) | Get validator by address |
| GetValidatorByKey | [GetValidatorByKeyRequest](#api-proto-v1-GetValidatorByKeyRequest) | [GetValidatorByKeyResponse](#api-proto-v1-GetValidatorByKeyResponse) | Get validator by key |
| GetLocalValidator | [GetLocalValidatorRequest](#api-proto-v1-GetLocalValidatorRequest) | [GetLocalValidatorResponse](#api-proto-v1-GetLocalValidatorResponse) | Get local validator |
| GetValidatorSetHeader | [GetValidatorSetHeaderRequest](#api-proto-v1-GetValidatorSetHeaderRequest) | [GetValidatorSetHeaderResponse](#api-proto-v1-GetValidatorSetHeaderResponse) | Get validator set header |
| GetLastCommitted | [GetLastCommittedRequest](#api-proto-v1-GetLastCommittedRequest) | [GetLastCommittedResponse](#api-proto-v1-GetLastCommittedResponse) | Get last committed epoch for a specific settlement chain |
| GetLastAllCommitted | [GetLastAllCommittedRequest](#api-proto-v1-GetLastAllCommittedRequest) | [GetLastAllCommittedResponse](#api-proto-v1-GetLastAllCommittedResponse) | Get last committed epochs for all settlement chains |
| GetValidatorSetMetadata | [GetValidatorSetMetadataRequest](#api-proto-v1-GetValidatorSetMetadataRequest) | [GetValidatorSetMetadataResponse](#api-proto-v1-GetValidatorSetMetadataResponse) | Get validator set metadata like extra data and request id to fetch aggregation and signature requests |
| GetCustomScheduleNodeStatus | [GetCustomScheduleNodeStatusRequest](#api-proto-v1-GetCustomScheduleNodeStatusRequest) | [GetCustomScheduleNodeStatusResponse](#api-proto-v1-GetCustomScheduleNodeStatusResponse) | Checks if the current node should be active based on a custom schedule derived from the validator set. This enables external applications to use the relay&#39;s validator set for coordinating distributed tasks, such as deciding which application instances should commit data on-chain or perform other coordinated actions. The schedule ensures deterministic but randomized selection of active nodes at any given time. |
| ListenSignatures | [ListenSignaturesRequest](#api-proto-v1-ListenSignaturesRequest) | [ListenSignaturesResponse](#api-proto-v1-ListenSignaturesResponse) stream | Stream signatures in real-time. If start_epoch is provided, sends historical data first |
| ListenProofs | [ListenProofsRequest](#api-proto-v1-ListenProofsRequest) | [ListenProofsResponse](#api-proto-v1-ListenProofsResponse) stream | Stream aggregation proofs in real-time. If start_epoch is provided, sends historical data first |
| ListenValidatorSet | [ListenValidatorSetRequest](#api-proto-v1-ListenValidatorSetRequest) | [ListenValidatorSetResponse](#api-proto-v1-ListenValidatorSetResponse) stream | Stream validator set changes in real-time. If start_epoch is provided, sends historical data first |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

