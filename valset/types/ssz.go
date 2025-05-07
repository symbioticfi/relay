// Hash: 8c67e0de1079a95336f540c6c73e42f4633fd2cb9841d83ab84b052439e4f3c7
// Version: 0.1.3
package types

import (
	"encoding/hex"
	"fmt"

	ssz "github.com/ferranbt/fastssz"
	"github.com/samber/lo"

	"github.com/ethereum/go-ethereum/common"
)

const MaxValidators = 1048576
const MaxVaults = 32
const MaxKeys = 128


// MarshalSSZ ssz marshals the Key object
func (k *Key) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(k)
}

// MarshalSSZTo ssz marshals the Key object to a target array
func (k *Key) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf

	// Field (0) 'Tag'
	dst = ssz.MarshalUint8(dst, k.Tag)

	// Field (1) 'PayloadHash'
	dst = append(dst, k.PayloadHash[:]...)

	return
}

// UnmarshalSSZ ssz unmarshals the Key object
func (k *Key) UnmarshalSSZ(buf []byte) error {
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
func (k *Key) SizeSSZ() (size int) {
	size = 33
	return
}

// HashTreeRoot ssz hashes the Key object
func (k *Key) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(k)
}

// HashTreeRootWith ssz hashes the Key object with a hasher
func (k *Key) HashTreeRootWith(hh ssz.HashWalker) (err error) {
	indx := hh.Index()

	// Field (0) 'Tag'
	hh.PutUint8(k.Tag)

	// Field (1) 'PayloadHash'
	hh.PutBytes(k.PayloadHash[:])

	hh.Merkleize(indx)
	return
}

// GetTree ssz hashes the Key object
func (k *Key) GetTree() (*ssz.Node, error) {
	return ssz.ProofTree(k)
}

// MarshalSSZ ssz marshals the Vault object
func (v *Vault) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(v)
}

// MarshalSSZTo ssz marshals the Vault object to a target array
func (v *Vault) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf

	// Field (0) 'ChainId'
	dst = ssz.MarshalUint64(dst, v.ChainId)

	// Field (1) 'Vault'
	dst = append(dst, v.Vault.Bytes()...)

	// Field (2) 'VotingPower'
	dst = append(dst, v.VotingPower.Bytes()...)

	return
}

// UnmarshalSSZ ssz unmarshals the Vault object
func (v *Vault) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size != 60 {
		return ssz.ErrSize
	}

	// Field (0) 'ChainId'
	v.ChainId = ssz.UnmarshallUint64(buf[0:8])

	// Field (1) 'Vault'
	copy(v.Vault.Bytes(), buf[8:28])

	// Field (2) 'VotingPower'
	copy(v.VotingPower.Bytes(), buf[28:60])

	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the Vault object
func (v *Vault) SizeSSZ() (size int) {
	size = 60
	return
}

// HashTreeRoot ssz hashes the Vault object
func (v *Vault) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(v)
}

// HashTreeRootWith ssz hashes the Vault object with a hasher
func (v *Vault) HashTreeRootWith(hh ssz.HashWalker) (err error) {
	indx := hh.Index()

	// Field (0) 'ChainId'
	hh.PutUint64(v.ChainId)

	// Field (1) 'Vault'
	hh.PutBytes(v.Vault.Bytes())

	// Field (2) 'VotingPower'
	hh.PutBytes(v.VotingPower.Bytes())

	hh.Merkleize(indx)
	return
}

// GetTree ssz hashes the Vault object
func (v *Vault) GetTree() (*ssz.Node, error) {
	return ssz.ProofTree(v)
}

// MarshalSSZ ssz marshals the Validator object
func (v *Validator) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(v)
}

// MarshalSSZTo ssz marshals the Validator object to a target array
func (v *Validator) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
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
	if size := len(v.Keys); size > 128 {
		err = ssz.ErrListTooBigFn("Validator.Keys", size, 128)
		return
	}
	for ii := 0; ii < len(v.Keys); ii++ {
		if dst, err = v.Keys[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (4) 'Vaults'
	if size := len(v.Vaults); size > 32 {
		err = ssz.ErrListTooBigFn("Validator.Vaults", size, 32)
		return
	}
	for ii := 0; ii < len(v.Vaults); ii++ {
		if dst, err = v.Vaults[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	return
}

// UnmarshalSSZ ssz unmarshals the Validator object
func (v *Validator) UnmarshalSSZ(buf []byte) error {
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
		num, err := ssz.DivideInt2(len(buf), 33, 128)
		if err != nil {
			return err
		}
		v.Keys = make([]*Key, num)
		for ii := 0; ii < num; ii++ {
			if v.Keys[ii] == nil {
				v.Keys[ii] = new(Key)
			}
			if err = v.Keys[ii].UnmarshalSSZ(buf[ii*33 : (ii+1)*33]); err != nil {
				return err
			}
		}
	}

	// Field (4) 'Vaults'
	{
		buf = tail[o4:]
		num, err := ssz.DivideInt2(len(buf), 60, 32)
		if err != nil {
			return err
		}
		v.Vaults = make([]*Vault, num)
		for ii := 0; ii < num; ii++ {
			if v.Vaults[ii] == nil {
				v.Vaults[ii] = new(Vault)
			}
			if err = v.Vaults[ii].UnmarshalSSZ(buf[ii*60 : (ii+1)*60]); err != nil {
				return err
			}
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the Validator object
func (v *Validator) SizeSSZ() (size int) {
	size = 61

	// Field (3) 'Keys'
	size += len(v.Keys) * 33

	// Field (4) 'Vaults'
	size += len(v.Vaults) * 60

	return
}

// HashTreeRoot ssz hashes the Validator object
func (v *Validator) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(v)
}

// HashTreeRootWith ssz hashes the Validator object with a hasher
func (v *Validator) HashTreeRootWith(hh ssz.HashWalker) (err error) {
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
		if num > 128 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for _, elem := range v.Keys {
			if err = elem.HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 128)
	}

	// Field (4) 'Vaults'
	{
		subIndx := hh.Index()
		num := uint64(len(v.Vaults))
		if num > 32 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for _, elem := range v.Vaults {
			if err = elem.HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 32)
	}

	hh.Merkleize(indx)
	return
}

// GetTree ssz hashes the Validator object
func (v *Validator) GetTree() (*ssz.Node, error) {
	return ssz.ProofTree(v)
}

// MarshalSSZ ssz marshals the ValidatorSet object
func (v *ValidatorSet) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(v)
}

// MarshalSSZTo ssz marshals the ValidatorSet object to a target array
func (v *ValidatorSet) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(4)

	// Offset (0) 'Validators'
	dst = ssz.WriteOffset(dst, offset)

	// Field (0) 'Validators'
	if size := len(v.Validators); size > 1048576 {
		err = ssz.ErrListTooBigFn("ValidatorSet.Validators", size, 1048576)
		return
	}
	{
		offset = 4 * len(v.Validators)
		for ii := 0; ii < len(v.Validators); ii++ {
			dst = ssz.WriteOffset(dst, offset)
			offset += v.Validators[ii].SizeSSZ()
		}
	}
	for ii := 0; ii < len(v.Validators); ii++ {
		if dst, err = v.Validators[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	return
}

// UnmarshalSSZ ssz unmarshals the ValidatorSet object
func (v *ValidatorSet) UnmarshalSSZ(buf []byte) error {
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
		num, err := ssz.DecodeDynamicLength(buf, 1048576)
		if err != nil {
			return err
		}
		v.Validators = make([]*Validator, num)
		err = ssz.UnmarshalDynamic(buf, num, func(indx int, buf []byte) (err error) {
			if v.Validators[indx] == nil {
				v.Validators[indx] = new(Validator)
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

// SizeSSZ returns the ssz encoded size in bytes for the ValidatorSet object
func (v *ValidatorSet) SizeSSZ() (size int) {
	size = 4

	// Field (0) 'Validators'
	for ii := 0; ii < len(v.Validators); ii++ {
		size += 4
		size += v.Validators[ii].SizeSSZ()
	}

	return
}

// HashTreeRoot ssz hashes the ValidatorSet object
func (v *ValidatorSet) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(v)
}

// HashTreeRootWith ssz hashes the ValidatorSet object with a hasher
func (v *ValidatorSet) HashTreeRootWith(hh ssz.HashWalker) (err error) {
	indx := hh.Index()

	// Field (0) 'Validators'
	{
		subIndx := hh.Index()
		num := uint64(len(v.Validators))
		if num > 1048576 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for _, elem := range v.Validators {
			if err = elem.HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 1048576)
	}

	hh.Merkleize(indx)
	return
}

// GetTree ssz hashes the ValidatorSet object
func (v *ValidatorSet) GetTree() (*ssz.Node, error) {
	return ssz.ProofTree(v)
}

func (v *ValidatorSet) GetValidatorRootIndex(operator common.Address) (int) {
	for i, validator := range v.Validators {
		if validator.Operator.Cmp(operator) == 0 {
			validatorRootIndex := 1 << 0 + 0
			validatorRootIndex <<= 1
			validatorRootIndex = validatorRootIndex * MaxValidators + i

			localValidatorRootIndex := 0
			localValidatorRootIndex <<= 1
			localValidatorRootIndex = localValidatorRootIndex * MaxValidators + i

			node, err := v.GetTree()
			if err != nil {
				return -1
			}

			fmt.Printf("operator: %x\n", operator)
			fmt.Printf("root: %x\n", node.Hash())
			
			first, _ := node.Get(validatorRootIndex)
			
			fmt.Printf("first: %x\n", first.Hash())
			fmt.Printf("validatorRootIndex: %d\n", validatorRootIndex)
			fmt.Printf("localValidatorRootIndex: %d\n", localValidatorRootIndex)
			validatorRootProof, _ := node.Prove(validatorRootIndex)
			fmt.Printf("validatorRootProof leaf: %x\n", validatorRootProof.Leaf)
			fmt.Printf("validatorRootProof hashes: %s\n", lo.Map(validatorRootProof.Hashes, func(hash []byte, _ int) string {
				return fmt.Sprintf("0x%s,", hex.EncodeToString(hash))
			}))

			operatorIndex := 1 << 3 + 0 

			second, _ := first.Get(operatorIndex)

			fmt.Printf("second: %x\n", second.Hash())
			fmt.Printf("operatorIndex: %d\n", operatorIndex)
			
			operatorProof, _ := first.Prove(operatorIndex)
			fmt.Printf("operatorProof leaf: %x\n", operatorProof.Leaf)
			fmt.Printf("operatorProof hashes: %s\n", lo.Map(operatorProof.Hashes, func(hash []byte, _ int) string {
				return fmt.Sprintf("0x%s,", hex.EncodeToString(hash))
			}))

			// validatorRootIndexNode, _ := node.Get(validatorRootIndex)
			// fmt.Printf("node.Get(validatorRootIndex): %x\n", validatorRootIndexNode.Hash())

			return validatorRootIndex
		}
	}
	
	return -1
}