// Hash: 8c67e0de1079a95336f540c6c73e42f4633fd2cb9841d83ab84b052439e4f3c7
// Version: 0.1.3
package ssz

import (
	"fmt"
	"sort"

	ssz "github.com/ferranbt/fastssz"

	"github.com/ethereum/go-ethereum/common"
)

const (
	ValidatorsListLocalPosition       = 0
	OperatorLocalPosition             = 0
	ValidatorVotingPowerLocalPosition = 1
	IsActiveLocalPosition             = 2
	KeysListLocalPosition             = 3
	VaultsListLocalPosition           = 4
	TagLocalPosition                  = 0
	PayloadHashLocalPosition          = 1
	ChainIdLocalPosition              = 0
	VaultLocalPosition                = 1
	VaultVotingPowerLocalPosition     = 2
)

const (
	ValidatorSetElements      = 1
	ValidatorSetTreeHeight    = 0 // ceil(log2(ValidatorSetElements))
	ValidatorElements         = 5
	ValidatorTreeHeight       = 3 // ceil(log2(ValidatorElements))
	VaultElements             = 3
	VaultTreeHeight           = 2 // ceil(log2(VaultElements))
	KeyElements               = 2
	KeyTreeHeight             = 1 // ceil(log2(KeyElements))
	ValidatorsListMaxElements = 1048576
	ValidatorsListTreeHeight  = 20 // ceil(log2(ValidatorsListMaxElements))
	KeysListMaxElements       = 128
	KeysListTreeHeight        = 7 // ceil(log2(KeysListMaxElements))
	VaultsListMaxElements     = 32
	VaultsListTreeHeight      = 5 // ceil(log2(VaultsListMaxElements))
)

// MarshalSSZ ssz marshals the Key object
func (k *SszKey) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(k)
}

// MarshalSSZTo ssz marshals the Key object to a target array
func (k *SszKey) MarshalSSZTo(buf []byte) ([]byte, error) {
	dst := buf

	// Field (0) 'Tag'
	dst = ssz.MarshalUint8(dst, k.Tag)

	// Field (1) 'PayloadHash'
	dst = append(dst, k.PayloadHash[:]...)

	return dst, nil
}

// UnmarshalSSZ ssz unmarshals the Key object
func (k *SszKey) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size != 33 {
		return ssz.ErrSize
	}

	// Field (0) 'Tag'
	k.Tag = ssz.UnmarshallUint8(buf[0:1])

	// Field (1) 'PayloadHash'
	copy(k.PayloadHash[:], buf[1:33])

	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the Key object
func (k *SszKey) SizeSSZ() int {
	return 33
}

// HashTreeRoot ssz hashes the Key object
func (k *SszKey) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(k)
}

// HashTreeRootWith ssz hashes the Key object with a hasher
func (k *SszKey) HashTreeRootWith(hh ssz.HashWalker) error {
	indx := hh.Index()

	// Field (0) 'Tag'
	hh.PutUint8(k.Tag)

	// Field (1) 'PayloadHash'
	hh.PutBytes(k.PayloadHash[:])

	hh.Merkleize(indx)

	return nil
}

// GetTree ssz hashes the Key object
func (k *SszKey) GetTree() (*ssz.Node, error) {
	return ssz.ProofTree(k)
}

// MarshalSSZ ssz marshals the Vault object
func (v *SszVault) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(v)
}

// MarshalSSZTo ssz marshals the Vault object to a target array
func (v *SszVault) MarshalSSZTo(buf []byte) ([]byte, error) {
	dst := buf

	// Field (0) 'ChainId'
	dst = ssz.MarshalUint64(dst, v.ChainId)

	// Field (1) 'OperatorVault'
	dst = append(dst, v.Vault.Bytes()...)

	// Field (2) 'VotingPower'
	dst = append(dst, v.VotingPower.Bytes()...)

	return dst, nil
}

// UnmarshalSSZ ssz unmarshals the Vault object
func (v *SszVault) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size != 60 {
		return ssz.ErrSize
	}

	// Field (0) 'ChainId'
	v.ChainId = ssz.UnmarshallUint64(buf[0:8])

	// Field (1) 'OperatorVault'
	copy(v.Vault.Bytes(), buf[8:28])

	// Field (2) 'VotingPower'
	copy(v.VotingPower.Bytes(), buf[28:60])

	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the Vault object
func (v *SszVault) SizeSSZ() int {
	return 60
}

// HashTreeRoot ssz hashes the Vault object
func (v *SszVault) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(v)
}

// HashTreeRootWith ssz hashes the Vault object with a hasher
func (v *SszVault) HashTreeRootWith(hh ssz.HashWalker) error {
	indx := hh.Index()

	// Field (0) 'ChainId'
	hh.PutUint64(v.ChainId)

	// Field (1) 'OperatorVault'
	hh.PutBytes(v.Vault.Bytes())

	// Field (2) 'VotingPower'
	hh.PutBytes(v.VotingPower.Bytes())

	hh.Merkleize(indx)

	return nil
}

// GetTree ssz hashes the Vault object
func (v *SszVault) GetTree() (*ssz.Node, error) {
	return ssz.ProofTree(v)
}

// MarshalSSZ ssz marshals the Validator object
func (v *SszValidator) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(v)
}

// MarshalSSZTo ssz marshals the Validator object to a target array
func (v *SszValidator) MarshalSSZTo(buf []byte) ([]byte, error) {
	dst := buf
	offset := int(61)

	// Field (0) 'Operator'
	dst = append(dst, v.Operator.Bytes()...)

	// Field (1) 'VotingPower'
	dst = append(dst, v.VotingPower.Bytes()...)

	// Field (2) 'IsActive'
	dst = ssz.MarshalBool(dst, v.IsActive)

	// Offset (3) 'Keys'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(v.Keys) * 33

	// Offset (4) 'Vaults'
	dst = ssz.WriteOffset(dst, offset)

	// Field (3) 'Keys'
	if size := len(v.Keys); size > KeysListMaxElements {
		return nil, ssz.ErrListTooBigFn("Validator.Keys", size, KeysListMaxElements)
	}
	for ii := 0; ii < len(v.Keys); ii++ {
		var err error
		dst, err = v.Keys[ii].MarshalSSZTo(dst)
		if err != nil {
			return nil, err
		}
	}

	// Field (4) 'Vaults'
	if size := len(v.Vaults); size > VaultsListMaxElements {
		return nil, ssz.ErrListTooBigFn("Validator.Vaults", size, VaultsListMaxElements)
	}
	for ii := 0; ii < len(v.Vaults); ii++ {
		var err error
		dst, err = v.Vaults[ii].MarshalSSZTo(dst)
		if err != nil {
			return nil, err
		}
	}
	return dst, nil
}

// UnmarshalSSZ ssz unmarshals the Validator object
func (v *SszValidator) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 61 {
		return ssz.ErrSize
	}

	tail := buf
	var o3, o4 uint64

	// Field (0) 'Operator'
	copy(v.Operator.Bytes(), buf[0:20])

	// Field (1) 'VotingPower'
	copy(v.VotingPower.Bytes(), buf[20:52])

	// Field (2) 'IsActive'
	v.IsActive = ssz.UnmarshalBool(buf[52:53])

	// Offset (3) 'Keys'
	if o3 = ssz.ReadOffset(buf[53:57]); o3 > size {
		return ssz.ErrOffset
	}

	if o3 != 61 {
		return ssz.ErrInvalidVariableOffset
	}

	// Offset (4) 'Vaults'
	if o4 = ssz.ReadOffset(buf[57:61]); o4 > size || o3 > o4 {
		return ssz.ErrOffset
	}

	// Field (3) 'Keys'
	{
		buf = tail[o3:o4]
		num, err := ssz.DivideInt2(len(buf), 33, KeysListMaxElements)
		if err != nil {
			return err
		}
		v.Keys = make([]*SszKey, num)
		for ii := 0; ii < num; ii++ {
			if v.Keys[ii] == nil {
				v.Keys[ii] = new(SszKey)
			}
			if err = v.Keys[ii].UnmarshalSSZ(buf[ii*33 : (ii+1)*33]); err != nil {
				return err
			}
		}
	}

	// Field (4) 'Vaults'
	{
		buf = tail[o4:]
		num, err := ssz.DivideInt2(len(buf), 60, VaultsListMaxElements)
		if err != nil {
			return err
		}
		v.Vaults = make([]*SszVault, num)
		for ii := 0; ii < num; ii++ {
			if v.Vaults[ii] == nil {
				v.Vaults[ii] = new(SszVault)
			}
			if err = v.Vaults[ii].UnmarshalSSZ(buf[ii*60 : (ii+1)*60]); err != nil {
				return err
			}
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the Validator object
func (v *SszValidator) SizeSSZ() int {
	size := 61

	// Field (3) 'Keys'
	size += len(v.Keys) * 33

	// Field (4) 'Vaults'
	size += len(v.Vaults) * 60

	return size
}

// HashTreeRoot ssz hashes the Validator object
func (v *SszValidator) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(v)
}

// HashTreeRootWith ssz hashes the Validator object with a hasher
func (v *SszValidator) HashTreeRootWith(hh ssz.HashWalker) error {
	indx := hh.Index()

	// Field (0) 'Operator'
	hh.PutBytes(v.Operator.Bytes())

	// Field (1) 'VotingPower'
	hh.PutBytes(v.VotingPower.Bytes())

	// Field (2) 'IsActive'
	hh.PutBool(v.IsActive)

	// Field (3) 'Keys'
	{
		subIndx := hh.Index()
		num := uint64(len(v.Keys))
		if num > KeysListMaxElements {
			return ssz.ErrIncorrectListSize
		}
		for _, elem := range v.Keys {
			if err := elem.HashTreeRootWith(hh); err != nil {
				return err
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, KeysListMaxElements)
	}

	// Field (4) 'Vaults'
	{
		subIndx := hh.Index()
		num := uint64(len(v.Vaults))
		if num > VaultsListMaxElements {
			return ssz.ErrIncorrectListSize
		}
		for _, elem := range v.Vaults {
			if err := elem.HashTreeRootWith(hh); err != nil {
				return err
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, VaultsListMaxElements)
	}

	hh.Merkleize(indx)
	return nil
}

// GetTree ssz hashes the Validator object
func (v *SszValidator) GetTree() (*ssz.Node, error) {
	return ssz.ProofTree(v)
}

// MarshalSSZ ssz marshals the SszValidatorSet object
func (v *SszValidatorSet) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(v)
}

// MarshalSSZTo ssz marshals the SszValidatorSet object to a target array
func (v *SszValidatorSet) MarshalSSZTo(buf []byte) ([]byte, error) {
	dst := buf
	offset := int(4)

	// Offset (0) 'Validators'
	dst = ssz.WriteOffset(dst, offset)

	// Field (0) 'Validators'
	if size := len(v.Validators); size > ValidatorsListMaxElements {
		return nil, ssz.ErrListTooBigFn("SszValidatorSet.Validators", size, ValidatorsListMaxElements)
	}
	{
		offset = 4 * len(v.Validators)
		for ii := 0; ii < len(v.Validators); ii++ {
			dst = ssz.WriteOffset(dst, offset)
			offset += v.Validators[ii].SizeSSZ()
		}
	}
	for ii := 0; ii < len(v.Validators); ii++ {
		var err error
		dst, err = v.Validators[ii].MarshalSSZTo(dst)
		if err != nil {
			return nil, err
		}
	}

	return dst, nil
}

// UnmarshalSSZ ssz unmarshalls the SszValidatorSet object
func (v *SszValidatorSet) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 4 {
		return ssz.ErrSize
	}

	tail := buf
	var o0 uint64

	// Offset (0) 'Validators'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	if o0 != 4 {
		return ssz.ErrInvalidVariableOffset
	}

	// Field (0) 'Validators'
	{
		buf = tail[o0:]
		num, err := ssz.DecodeDynamicLength(buf, ValidatorsListMaxElements)
		if err != nil {
			return err
		}
		v.Validators = make([]*SszValidator, num)
		err = ssz.UnmarshalDynamic(buf, num, func(indx int, buf []byte) (err error) {
			if v.Validators[indx] == nil {
				v.Validators[indx] = new(SszValidator)
			}
			if err = v.Validators[indx].UnmarshalSSZ(buf); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the SszValidatorSet object
func (v *SszValidatorSet) SizeSSZ() int {
	size := 4

	// Field (0) 'Validators'
	for ii := 0; ii < len(v.Validators); ii++ {
		size += 4
		size += v.Validators[ii].SizeSSZ()
	}

	return size
}

// HashTreeRoot ssz hashes the SszValidatorSet object
func (v *SszValidatorSet) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(v)
}

// HashTreeRootWith ssz hashes the SszValidatorSet object with a hasher
func (v *SszValidatorSet) HashTreeRootWith(hh ssz.HashWalker) error {
	indx := hh.Index()

	// Field (0) 'Validators'
	{
		subIndx := hh.Index()
		num := uint64(len(v.Validators))
		if num > ValidatorsListMaxElements {
			return ssz.ErrIncorrectListSize
		}
		for _, elem := range v.Validators {
			if err := elem.HashTreeRootWith(hh); err != nil {
				return err
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, ValidatorsListMaxElements)
	}

	hh.Merkleize(indx)

	return nil
}

// GetTree ssz hashes the SszValidatorSet object
func (v *SszValidatorSet) GetTree() (*ssz.Node, error) {
	return ssz.ProofTree(v)
}

func (v *SszValidatorSet) ProveValidatorRoot(operator common.Address) (*SszValidator, int, *ssz.Proof, error) {
	validatorIndex := sort.Search(len(v.Validators), func(i int) bool {
		return v.Validators[i].Operator.Cmp(operator) >= 0
	})
	if validatorIndex >= len(v.Validators) || v.Validators[validatorIndex].Operator.Cmp(operator) != 0 {
		return nil, 0, nil, fmt.Errorf("validator %s not found", operator)
	}

	// go to SszValidatorSet.Validators
	validatorRootTreeIndex := 1<<ValidatorSetTreeHeight + ValidatorsListLocalPosition
	// consider List's length mix-in
	validatorRootTreeIndex <<= 1
	// go to SszValidatorSet.Validators[validatorIndex]
	validatorRootTreeIndex = validatorRootTreeIndex*ValidatorsListMaxElements + validatorIndex

	validatorRootTreeLocalIndex := ValidatorsListLocalPosition
	validatorRootTreeLocalIndex <<= 1
	validatorRootTreeLocalIndex = validatorRootTreeLocalIndex*ValidatorsListMaxElements + validatorIndex

	validatorSetRootNode, err := v.GetTree()
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to get validator set root node: %w", err)
	}

	validatorRootProof, err := validatorSetRootNode.Prove(validatorRootTreeIndex)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to get validator root proof: %w", err)
	}

	return v.Validators[validatorIndex], validatorRootTreeLocalIndex, validatorRootProof, nil
}

func (v *SszValidator) ProveOperator() (*ssz.Proof, error) {
	validatorRootNode, err := v.GetTree()
	if err != nil {
		return nil, fmt.Errorf("failed to get validator root node: %w", err)
	}

	// go to Validator.Operator
	operatorTreeIndex := 1<<ValidatorTreeHeight + OperatorLocalPosition

	operatorProof, err := validatorRootNode.Prove(operatorTreeIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator proof: %w", err)
	}

	return operatorProof, nil
}

func (v *SszValidator) ProveValidatorVotingPower() (*ssz.Proof, error) {
	validatorRootNode, err := v.GetTree()
	if err != nil {
		return nil, fmt.Errorf("failed to get validator root node: %w", err)
	}

	// go to Validator.VotingPower
	validatorVotingPowerTreeIndex := 1<<ValidatorTreeHeight + ValidatorVotingPowerLocalPosition

	validatorVotingPowerProof, err := validatorRootNode.Prove(validatorVotingPowerTreeIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get validator voting power proof: %w", err)
	}

	return validatorVotingPowerProof, nil
}

func (v *SszValidator) ProveIsActive() (*ssz.Proof, error) {
	validatorRootNode, err := v.GetTree()
	if err != nil {
		return nil, fmt.Errorf("failed to get validator root node: %w", err)
	}

	// go to Validator.IsActive
	isActiveTreeIndex := 1<<ValidatorTreeHeight + IsActiveLocalPosition

	isActiveProof, err := validatorRootNode.Prove(isActiveTreeIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get validator is active proof: %w", err)
	}

	return isActiveProof, nil
}

func (v *SszValidator) ProveKeyRoot(keyTag uint8) (*SszKey, int, *ssz.Proof, error) {
	keyIndex := sort.Search(len(v.Keys), func(i int) bool {
		return v.Keys[i].Tag >= uint8(keyTag)
	})
	if keyIndex >= len(v.Keys) || v.Keys[keyIndex].Tag != uint8(keyTag) {
		return nil, 0, nil, fmt.Errorf("SszKey %d not found", keyTag)
	}

	validatorRootNode, err := v.GetTree()
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to get validator root node: %w", err)
	}

	// go to Validator.Keys
	keyRootTreeIndex := 1<<ValidatorTreeHeight + KeysListLocalPosition
	// consider List's length mix-in
	keyRootTreeIndex <<= 1
	// go to Validator.Keys[keyIndex]
	keyRootTreeIndex = keyRootTreeIndex*KeysListMaxElements + keyIndex

	keyRootTreeLocalIndex := KeysListLocalPosition
	keyRootTreeLocalIndex <<= 1
	keyRootTreeLocalIndex = keyRootTreeLocalIndex*KeysListMaxElements + keyIndex

	keyRootProof, err := validatorRootNode.Prove(keyRootTreeIndex)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to get validator keys proof: %w", err)
	}

	return v.Keys[keyIndex], keyRootTreeLocalIndex, keyRootProof, nil
}

func (k *SszKey) ProveTag() (*ssz.Proof, error) {
	keyRootNode, err := k.GetTree()
	if err != nil {
		return nil, fmt.Errorf("failed to get SszKey root node: %w", err)
	}

	// go to Key.Tag
	tagTreeIndex := 1<<KeyTreeHeight + TagLocalPosition

	tagProof, err := keyRootNode.Prove(tagTreeIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get SszKey tag proof: %w", err)
	}

	return tagProof, nil
}

func (k *SszKey) ProvePayloadHash() (*ssz.Proof, error) {
	keyRootNode, err := k.GetTree()
	if err != nil {
		return nil, fmt.Errorf("failed to get SszKey root node: %w", err)
	}

	// go to Key.PayloadHash
	payloadHashTreeIndex := 1<<KeyTreeHeight + PayloadHashLocalPosition

	payloadHashProof, err := keyRootNode.Prove(payloadHashTreeIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get SszKey payload hash proof: %w", err)
	}

	return payloadHashProof, nil
}

func (v *SszValidator) ProveVaultRoot(vault common.Address) (*SszVault, int, *ssz.Proof, error) {
	vaultIndex := sort.Search(len(v.Vaults), func(i int) bool {
		return v.Vaults[i].Vault.Cmp(vault) >= 0
	})
	if vaultIndex >= len(v.Vaults) || v.Vaults[vaultIndex].Vault.Cmp(vault) != 0 {
		return nil, 0, nil, fmt.Errorf("vault %s not found", vault)
	}

	validatorRootNode, err := v.GetTree()
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to get validator root node: %w", err)
	}

	// go to Validator.Vaults
	vaultRootTreeIndex := 1<<ValidatorTreeHeight + VaultsListLocalPosition
	// consider List's length mix-in
	vaultRootTreeIndex <<= 1
	// go to Validator.Vaults[vaultIndex]
	vaultRootTreeIndex = vaultRootTreeIndex*VaultsListMaxElements + vaultIndex

	vaultRootTreeLocalIndex := VaultsListLocalPosition
	vaultRootTreeLocalIndex <<= 1
	vaultRootTreeLocalIndex = vaultRootTreeLocalIndex*VaultsListMaxElements + vaultIndex

	vaultRootProof, err := validatorRootNode.Prove(vaultRootTreeIndex)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to get validator vaults proof: %w", err)
	}

	return v.Vaults[vaultIndex], vaultRootTreeLocalIndex, vaultRootProof, nil
}

func (v *SszValidator) ProveChainId() (*ssz.Proof, error) {
	validatorRootNode, err := v.GetTree()
	if err != nil {
		return nil, fmt.Errorf("failed to get validator root node: %w", err)
	}

	// go to Validator.ChainId
	chainIdTreeIndex := 1<<ValidatorTreeHeight + ChainIdLocalPosition

	chainIdProof, err := validatorRootNode.Prove(chainIdTreeIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get validator chain id proof: %w", err)
	}

	return chainIdProof, nil
}

func (v *SszValidator) ProveVault() (*ssz.Proof, error) {
	vaultRootNode, err := v.GetTree()
	if err != nil {
		return nil, fmt.Errorf("failed to get vault root node: %w", err)
	}

	// go to OperatorVault.OperatorVault
	vaultTreeIndex := 1<<VaultTreeHeight + VaultLocalPosition

	vaultProof, err := vaultRootNode.Prove(vaultTreeIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get vault proof: %w", err)
	}

	return vaultProof, nil
}

func (v *SszValidator) ProveVaultVotingPower() (*ssz.Proof, error) {
	vaultRootNode, err := v.GetTree()
	if err != nil {
		return nil, fmt.Errorf("failed to get vault root node: %w", err)
	}

	// go to OperatorVault.VotingPower
	vaultVotingPowerTreeIndex := 1<<VaultTreeHeight + VaultVotingPowerLocalPosition

	vaultVotingPowerProof, err := vaultRootNode.Prove(vaultVotingPowerTreeIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get vault voting power proof: %w", err)
	}

	return vaultVotingPowerProof, nil
}
