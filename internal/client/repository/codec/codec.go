package codec

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"

	pb "github.com/symbioticfi/relay/internal/client/repository/badger/proto/v1"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

func MarshalProto(msg proto.Message) ([]byte, error) {
	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("failed to marshal proto: %v", err)
	}
	return data, nil
}

func UnmarshalProto(data []byte, msg proto.Message) error {
	if err := proto.Unmarshal(data, msg); err != nil {
		return errors.Errorf("failed to unmarshal proto: %v", err)
	}
	return nil
}

// Signature

func SignatureToBytes(sig symbiotic.Signature) ([]byte, error) {
	return MarshalProto(&pb.Signature{
		MessageHash:  sig.MessageHash,
		KeyTag:       uint32(sig.KeyTag),
		Epoch:        uint64(sig.Epoch),
		Signature:    sig.Signature,
		RawPublicKey: sig.PublicKey.Raw(),
	})
}

func BytesToSignature(value []byte) (symbiotic.Signature, error) {
	signaturePB := &pb.Signature{}
	if err := UnmarshalProto(value, signaturePB); err != nil {
		return symbiotic.Signature{}, errors.Errorf("failed to unmarshal signature: %w", err)
	}

	signature := symbiotic.Signature{
		MessageHash: signaturePB.GetMessageHash(),
		KeyTag:      symbiotic.KeyTag(signaturePB.GetKeyTag()),
		Epoch:       symbiotic.Epoch(signaturePB.GetEpoch()),
		Signature:   signaturePB.GetSignature(),
	}

	publicKey, err := crypto.NewPublicKey(signature.KeyTag.Type(), signaturePB.GetRawPublicKey())
	if err != nil {
		return symbiotic.Signature{}, errors.Errorf("failed to get public key from raw: %w", err)
	}

	signature.PublicKey = publicKey

	return signature, nil
}

// SignatureMap

func SignatureMapToBytes(vm entity.SignatureMap) ([]byte, error) {
	bitmapBytes, err := vm.SignedValidatorsBitmap.ToBytes()
	if err != nil {
		return nil, errors.Errorf("failed to serialize roaring bitmap: %w", err)
	}

	return MarshalProto(&pb.SignatureMap{
		RequestId:              vm.RequestID.Bytes(),
		Epoch:                  uint64(vm.Epoch),
		SignedValidatorsBitmap: bitmapBytes,
		CurrentVotingPower:     vm.CurrentVotingPower.String(),
		TotalValidators:        vm.TotalValidators,
	})
}

func BytesToSignatureMap(data []byte) (entity.SignatureMap, error) {
	signatureMap := &pb.SignatureMap{}
	if err := UnmarshalProto(data, signatureMap); err != nil {
		return entity.SignatureMap{}, errors.Errorf("failed to unmarshal signature map: %w", err)
	}

	requestId := common.BytesToHash(signatureMap.GetRequestId())

	bitmap, err := entity.BitmapFromBytes(signatureMap.GetSignedValidatorsBitmap())
	if err != nil {
		return entity.SignatureMap{}, errors.Errorf("failed to deserialize bitmap: %w", err)
	}

	currentVotingPower, ok := new(big.Int).SetString(signatureMap.GetCurrentVotingPower(), 10)
	if !ok {
		return entity.SignatureMap{}, errors.Errorf("failed to parse current voting power: %s", signatureMap.GetCurrentVotingPower())
	}

	return entity.SignatureMap{
		RequestID:              requestId,
		Epoch:                  symbiotic.Epoch(signatureMap.GetEpoch()),
		SignedValidatorsBitmap: bitmap,
		CurrentVotingPower:     symbiotic.ToVotingPower(currentVotingPower),
		TotalValidators:        signatureMap.GetTotalValidators(),
	}, nil
}

// SignatureRequest

func SignatureRequestToBytes(req symbiotic.SignatureRequest) ([]byte, error) {
	return MarshalProto(&pb.SignatureRequest{
		KeyTag:        uint32(req.KeyTag),
		RequiredEpoch: uint64(req.RequiredEpoch),
		Message:       req.Message,
	})
}

func BytesToSignatureRequest(data []byte) (symbiotic.SignatureRequest, error) {
	signatureRequest := &pb.SignatureRequest{}
	if err := UnmarshalProto(data, signatureRequest); err != nil {
		return symbiotic.SignatureRequest{}, errors.Errorf("failed to unmarshal signature request: %w", err)
	}

	return symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(signatureRequest.GetKeyTag()),
		RequiredEpoch: symbiotic.Epoch(signatureRequest.GetRequiredEpoch()),
		Message:       signatureRequest.GetMessage(),
	}, nil
}

// AggregationProof

func AggregationProofToBytes(ap symbiotic.AggregationProof) ([]byte, error) {
	return MarshalProto(&pb.AggregationProof{
		MessageHash: ap.MessageHash,
		KeyTag:      uint32(ap.KeyTag),
		Epoch:       uint64(ap.Epoch),
		Proof:       ap.Proof,
	})
}

func BytesToAggregationProof(value []byte) (symbiotic.AggregationProof, error) {
	aggregationProof := &pb.AggregationProof{}
	if err := UnmarshalProto(value, aggregationProof); err != nil {
		return symbiotic.AggregationProof{}, errors.Errorf("failed to unmarshal aggregation proof: %w", err)
	}

	return symbiotic.AggregationProof{
		MessageHash: aggregationProof.GetMessageHash(),
		KeyTag:      symbiotic.KeyTag(aggregationProof.GetKeyTag()),
		Epoch:       symbiotic.Epoch(aggregationProof.GetEpoch()),
		Proof:       aggregationProof.GetProof(),
	}, nil
}

// Validator

func ValidatorToBytes(validator symbiotic.Validator, activeIndex uint32) ([]byte, error) {
	return MarshalProto(&pb.Validator{
		Operator:    validator.Operator.Bytes(),
		VotingPower: validator.VotingPower.String(),
		IsActive:    validator.IsActive,
		ActiveIndex: activeIndex,
		Keys: lo.Map(validator.Keys, func(k symbiotic.ValidatorKey, _ int) *pb.ValidatorKey {
			return &pb.ValidatorKey{
				Tag:     uint32(k.Tag),
				Payload: k.Payload,
			}
		}),
		Vaults: lo.Map(validator.Vaults, func(v symbiotic.ValidatorVault, _ int) *pb.ValidatorVault {
			return &pb.ValidatorVault{
				ChainId:     v.ChainID,
				Vault:       v.Vault.Bytes(),
				VotingPower: v.VotingPower.String(),
			}
		}),
	})
}

func BytesToValidator(data []byte) (symbiotic.Validator, uint32, error) {
	validator := &pb.Validator{}
	if err := UnmarshalProto(data, validator); err != nil {
		return symbiotic.Validator{}, 0, errors.Errorf("failed to unmarshal validator: %w", err)
	}

	operator := common.BytesToAddress(validator.GetOperator())

	votingPower, ok := new(big.Int).SetString(validator.GetVotingPower(), 10)
	if !ok {
		return symbiotic.Validator{}, 0, errors.Errorf("failed to parse voting power: %s", validator.GetVotingPower())
	}

	keys := lo.Map(validator.GetKeys(), func(k *pb.ValidatorKey, _ int) symbiotic.ValidatorKey {
		return symbiotic.ValidatorKey{
			Tag:     symbiotic.KeyTag(k.GetTag()),
			Payload: k.GetPayload(),
		}
	})

	vaults := make([]symbiotic.ValidatorVault, 0, len(validator.GetVaults()))
	for _, v := range validator.GetVaults() {
		votingPowerVault, parseOk := new(big.Int).SetString(v.GetVotingPower(), 10)
		if !parseOk {
			return symbiotic.Validator{}, 0, errors.Errorf("failed to parse vault voting power for operator %s: %s", operator.Hex(), v.GetVotingPower())
		}
		vaults = append(vaults, symbiotic.ValidatorVault{
			ChainID:     v.GetChainId(),
			Vault:       common.BytesToAddress(v.GetVault()),
			VotingPower: symbiotic.ToVotingPower(votingPowerVault),
		})
	}

	return symbiotic.Validator{
		Operator:    operator,
		VotingPower: symbiotic.ToVotingPower(votingPower),
		IsActive:    validator.GetIsActive(),
		Keys:        keys,
		Vaults:      vaults,
	}, validator.GetActiveIndex(), nil
}

// ValidatorSetHeader

func ValidatorSetHeaderToBytes(valset symbiotic.ValidatorSet) ([]byte, error) {
	header, err := valset.GetHeader()
	if err != nil {
		return nil, errors.Errorf("failed to get validator set header: %w", err)
	}

	var aggIndices, commIndices []byte
	if len(valset.AggregatorIndices) > 0 {
		aggBitmap := entity.NewBitmapOf(valset.AggregatorIndices...)
		aggIndices, err = aggBitmap.ToBytes()
		if err != nil {
			return nil, errors.Errorf("failed to serialize aggregator indices: %w", err)
		}
	}

	if len(valset.CommitterIndices) > 0 {
		commBitmap := entity.NewBitmapOf(valset.CommitterIndices...)
		commIndices, err = commBitmap.ToBytes()
		if err != nil {
			return nil, errors.Errorf("failed to serialize committer indices: %w", err)
		}
	}

	return MarshalProto(&pb.ValidatorSetHeader{
		Version:            uint32(header.Version),
		RequiredKeyTag:     uint32(header.RequiredKeyTag),
		Epoch:              uint64(header.Epoch),
		CaptureTimestamp:   uint64(header.CaptureTimestamp),
		QuorumThreshold:    header.QuorumThreshold.String(),
		TotalVotingPower:   header.TotalVotingPower.String(),
		ValidatorsSszMroot: header.ValidatorsSszMRoot.Bytes(),
		AggregatorIndices:  aggIndices,
		CommitterIndices:   commIndices,
	})
}

func BytesToValidatorSetHeader(data []byte) (symbiotic.ValidatorSetHeader, error) {
	validatorSetHeader := &pb.ValidatorSetHeader{}
	if err := UnmarshalProto(data, validatorSetHeader); err != nil {
		return symbiotic.ValidatorSetHeader{}, errors.Errorf("failed to unmarshal validator set header: %w", err)
	}

	quorumThreshold, ok := new(big.Int).SetString(validatorSetHeader.GetQuorumThreshold(), 10)
	if !ok {
		return symbiotic.ValidatorSetHeader{}, errors.Errorf("failed to parse quorum threshold: %s", validatorSetHeader.GetQuorumThreshold())
	}

	totalVotingPower, ok := new(big.Int).SetString(validatorSetHeader.GetTotalVotingPower(), 10)
	if !ok {
		return symbiotic.ValidatorSetHeader{}, errors.Errorf("failed to parse total voting power: %s", validatorSetHeader.GetTotalVotingPower())
	}

	return symbiotic.ValidatorSetHeader{
		Version:            uint8(validatorSetHeader.GetVersion()),
		RequiredKeyTag:     symbiotic.KeyTag(validatorSetHeader.GetRequiredKeyTag()),
		Epoch:              symbiotic.Epoch(validatorSetHeader.GetEpoch()),
		CaptureTimestamp:   symbiotic.Timestamp(validatorSetHeader.GetCaptureTimestamp()),
		QuorumThreshold:    symbiotic.ToVotingPower(quorumThreshold),
		TotalVotingPower:   symbiotic.ToVotingPower(totalVotingPower),
		ValidatorsSszMRoot: common.BytesToHash(validatorSetHeader.GetValidatorsSszMroot()),
	}, nil
}

func ExtractAdditionalInfoFromHeaderData(data []byte) (aggIndices []uint32, commIndices []uint32, err error) {
	validatorSetHeader := &pb.ValidatorSetHeader{}
	if err := UnmarshalProto(data, validatorSetHeader); err != nil {
		return nil, nil, errors.Errorf("failed to unmarshal validator set header: %w", err)
	}

	if len(validatorSetHeader.GetAggregatorIndices()) > 0 {
		aggBitmap, err := entity.BitmapFromBytes(validatorSetHeader.GetAggregatorIndices())
		if err != nil {
			return nil, nil, errors.Errorf("failed to deserialize aggregator indices: %w", err)
		}
		aggIndices = aggBitmap.ToArray()
	} else {
		aggIndices = []uint32{}
	}

	if len(validatorSetHeader.GetCommitterIndices()) > 0 {
		commBitmap, err := entity.BitmapFromBytes(validatorSetHeader.GetCommitterIndices())
		if err != nil {
			return nil, nil, errors.Errorf("failed to deserialize committer indices: %w", err)
		}
		commIndices = commBitmap.ToArray()
	} else {
		commIndices = []uint32{}
	}

	return aggIndices, commIndices, nil
}

// ValidatorSetMetadata

func ValidatorSetMetadataToBytes(data symbiotic.ValidatorSetMetadata) ([]byte, error) {
	return MarshalProto(&pb.ValidatorSetMetadata{
		RequestId: data.RequestID.Bytes(),
		Epoch:     uint64(data.Epoch),
		ExtraData: lo.Map(data.ExtraData, func(ed symbiotic.ExtraData, _ int) *pb.ExtraData {
			return &pb.ExtraData{
				Key:   ed.Key.Bytes(),
				Value: ed.Value.Bytes(),
			}
		}),
		CommitmentData: data.CommitmentData,
	})
}

func BytesToValidatorSetMetadata(data []byte) (symbiotic.ValidatorSetMetadata, error) {
	validatorSetMetadata := &pb.ValidatorSetMetadata{}
	if err := UnmarshalProto(data, validatorSetMetadata); err != nil {
		return symbiotic.ValidatorSetMetadata{}, errors.Errorf("failed to unmarshal validator set metadata: %w", err)
	}

	return symbiotic.ValidatorSetMetadata{
		RequestID: common.BytesToHash(validatorSetMetadata.GetRequestId()),
		ExtraData: lo.Map(validatorSetMetadata.GetExtraData(), func(ed *pb.ExtraData, _ int) symbiotic.ExtraData {
			return symbiotic.ExtraData{
				Key:   common.BytesToHash(ed.GetKey()),
				Value: common.BytesToHash(ed.GetValue()),
			}
		}),
		Epoch:          symbiotic.Epoch(validatorSetMetadata.GetEpoch()),
		CommitmentData: validatorSetMetadata.GetCommitmentData(),
	}, nil
}

// NetworkConfig

func NetworkConfigToBytes(config symbiotic.NetworkConfig) ([]byte, error) {
	return MarshalProto(&pb.NetworkConfig{
		VotingPowerProviders: lo.Map(config.VotingPowerProviders, func(addr symbiotic.CrossChainAddress, _ int) *pb.CrossChainAddress {
			return &pb.CrossChainAddress{
				ChainId: addr.ChainId,
				Address: addr.Address.Bytes(),
			}
		}),
		KeysProvider: &pb.CrossChainAddress{
			Address: config.KeysProvider.Address.Bytes(),
			ChainId: config.KeysProvider.ChainId,
		},
		Settlements: lo.Map(config.Settlements, func(addr symbiotic.CrossChainAddress, _ int) *pb.CrossChainAddress {
			return &pb.CrossChainAddress{
				ChainId: addr.ChainId,
				Address: addr.Address.Bytes(),
			}
		}),
		VerificationType:        uint32(config.VerificationType),
		MaxVotingPower:          config.MaxVotingPower.String(),
		MinInclusionVotingPower: config.MinInclusionVotingPower.String(),
		MaxValidatorsCount:      config.MaxValidatorsCount.String(),
		RequiredKeyTags:         lo.Map(config.RequiredKeyTags, func(tag symbiotic.KeyTag, _ int) uint32 { return uint32(tag) }),
		RequiredHeaderKeyTag:    uint32(config.RequiredHeaderKeyTag),
		QuorumThresholds: lo.Map(config.QuorumThresholds, func(qt symbiotic.QuorumThreshold, _ int) *pb.QuorumThreshold {
			return &pb.QuorumThreshold{
				KeyTag:          uint32(qt.KeyTag),
				QuorumThreshold: qt.QuorumThreshold.String(),
			}
		}),
		NumCommitters:         config.NumCommitters,
		NumAggregators:        config.NumAggregators,
		CommitterSlotDuration: config.CommitterSlotDuration,
		EpochDuration:         config.EpochDuration,
	})
}

func BytesToNetworkConfig(data []byte) (symbiotic.NetworkConfig, error) {
	networkConfig := &pb.NetworkConfig{}
	if err := UnmarshalProto(data, networkConfig); err != nil {
		return symbiotic.NetworkConfig{}, errors.Errorf("failed to unmarshal network config: %w", err)
	}

	maxVotingPower, ok := new(big.Int).SetString(networkConfig.GetMaxVotingPower(), 10)
	if !ok {
		return symbiotic.NetworkConfig{}, errors.Errorf("failed to parse max voting power: %s", networkConfig.GetMaxVotingPower())
	}

	minInclusionVotingPower, ok := new(big.Int).SetString(networkConfig.GetMinInclusionVotingPower(), 10)
	if !ok {
		return symbiotic.NetworkConfig{}, errors.Errorf("failed to parse min inclusion voting power: %s", networkConfig.GetMinInclusionVotingPower())
	}

	maxValidatorsCount, ok := new(big.Int).SetString(networkConfig.GetMaxValidatorsCount(), 10)
	if !ok {
		return symbiotic.NetworkConfig{}, errors.Errorf("failed to parse max validators count: %s", networkConfig.GetMaxValidatorsCount())
	}

	quorumThresholds := make([]symbiotic.QuorumThreshold, 0, len(networkConfig.GetQuorumThresholds()))

	for _, qt := range networkConfig.GetQuorumThresholds() {
		threshold, parseOk := new(big.Int).SetString(qt.GetQuorumThreshold(), 10)
		if !parseOk {
			return symbiotic.NetworkConfig{}, errors.Errorf("failed to parse quorum threshold: %s", qt.GetQuorumThreshold())
		}

		quorumThresholds = append(quorumThresholds, symbiotic.QuorumThreshold{
			KeyTag:          symbiotic.KeyTag(qt.GetKeyTag()),
			QuorumThreshold: symbiotic.ToQuorumThresholdPct(threshold),
		})
	}

	return symbiotic.NetworkConfig{
		VotingPowerProviders: lo.Map(networkConfig.GetVotingPowerProviders(), func(addr *pb.CrossChainAddress, _ int) symbiotic.CrossChainAddress {
			return symbiotic.CrossChainAddress{
				ChainId: addr.GetChainId(),
				Address: common.BytesToAddress(addr.GetAddress()),
			}
		}),
		KeysProvider: symbiotic.CrossChainAddress{
			ChainId: networkConfig.GetKeysProvider().GetChainId(),
			Address: common.BytesToAddress(networkConfig.GetKeysProvider().GetAddress()),
		},
		Settlements: lo.Map(networkConfig.GetSettlements(), func(addr *pb.CrossChainAddress, _ int) symbiotic.CrossChainAddress {
			return symbiotic.CrossChainAddress{
				ChainId: addr.GetChainId(),
				Address: common.BytesToAddress(addr.GetAddress()),
			}
		}),
		VerificationType:        symbiotic.VerificationType(networkConfig.GetVerificationType()),
		MaxVotingPower:          symbiotic.ToVotingPower(maxVotingPower),
		MinInclusionVotingPower: symbiotic.ToVotingPower(minInclusionVotingPower),
		MaxValidatorsCount:      symbiotic.ToVotingPower(maxValidatorsCount),
		RequiredKeyTags:         lo.Map(networkConfig.GetRequiredKeyTags(), func(tag uint32, _ int) symbiotic.KeyTag { return symbiotic.KeyTag(tag) }),
		RequiredHeaderKeyTag:    symbiotic.KeyTag(networkConfig.GetRequiredHeaderKeyTag()),
		QuorumThresholds:        quorumThresholds,
		NumAggregators:          networkConfig.GetNumAggregators(),
		NumCommitters:           networkConfig.GetNumCommitters(),
		CommitterSlotDuration:   networkConfig.GetCommitterSlotDuration(),
		EpochDuration:           networkConfig.GetEpochDuration(),
	}, nil
}

// ValidatorKeyHash computes the keccak256 hash of a public key for key lookup indexing.
func ValidatorKeyHash(publicKey []byte) common.Hash {
	return ethcrypto.Keccak256Hash(publicKey)
}
