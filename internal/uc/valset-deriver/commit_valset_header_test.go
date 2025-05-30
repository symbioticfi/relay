package valsetDeriver

//
//import (
//	"encoding/hex"
//	"fmt"
//	"math/big"
//	"testing"
//
//	"github.com/ethereum/go-ethereum/accounts/abi"
//	"github.com/stretchr/testify/require"
//	"go.uber.org/mock/gomock"
//
//	"middleware-offchain/internal/client/valset/mocks"
//	"middleware-offchain/internal/entity"
//	"middleware-offchain/pkg/bls"
//	"middleware-offchain/pkg/proof"
//)
//
//func Test_CommitValsetHeaderUnit(t *testing.T) {
//	t.Skip("need fix")
//	pk1 := "87191036493798670866484781455694320176667203290824056510541300741498740913410"
//	pk2 := "26972876870930381973856869753776124637336739336929668162870464864826929175089"
//	pk3 := "11008377096554045051122023680185802911050337017631086444859313200352654461863"
//
//	keyPair1 := bls.ComputeKeyPair(bytesFromPK(t, pk1))
//	keyPair2 := bls.ComputeKeyPair(bytesFromPK(t, pk2))
//	keyPair3 := bls.ComputeKeyPair(bytesFromPK(t, pk3))
//
//	validatorSet := entity.ValidatorSet{
//		//Version: 1,
//		Validators: []entity.Validator{
//			{
//				IsActive: true,
//				Keys:     nil,
//			},
//			{
//				IsActive: true,
//				Keys:     nil,
//			},
//			{
//				IsActive: true,
//				Keys:     nil,
//			},
//		},
//		//TotalActiveVotingPower: new(big.Int).SetInt64(30000000000000),
//	}
//	_ = validatorSet
//
//	valsetHeader1 := entity.ValidatorSetHeader{
//		Version: 1,
//		ActiveAggregatedKeys: []entity.Key{{
//			Tag:     15,
//			Payload: decodeHex(t, "264621561abeb4dac9a497cb21f305b8f41b56389734832656d7c7adde2247081ffa73b25b82c16096babd6a15d259a24a8304cd96ee6c27e790ff27d8744a5b"),
//		}},
//		TotalActiveVotingPower: new(big.Int).SetInt64(30000000000000),
//		ValidatorsSszMRoot:     [32]byte(decodeHex(t, "d9354a3cf52fba5126422c86d35db53d566d46f9208faa86c7b9155d7dcf3926")),
//		ExtraData:              decodeHex(t, "2695ed079545bb906f5868716071ab237e36d04fdc1aa07b06bd98c81185067d"),
//		Epoch:                  new(big.Int).SetInt64(1),
//		DomainEip712: entity.Eip712Domain{
//			Name:    "Middleware",
//			Version: "1",
//		},
//		Subnetwork: decodeHex(t, "f39fd6e51aad88f6f4ce6ab8827279cfffb92266000000000000000000000000"),
//	}
//
//	ctrl := gomock.NewController(t)
//	der := mocks.NewMockderiver(ctrl)
//	eth := mocks.NewMockethClient(ctrl)
//
//	generator, err := NewGenerator(der, eth)
//	require.NoError(t, err)
//
//	headerHash1, err := generator.GenerateValidatorSetHeaderHash(valsetHeader1)
//	require.NoError(t, err)
//	headerSignature1, err := keyPair1.Sign(headerHash1)
//	require.NoError(t, err)
//
//	headerHash2, err := generator.GenerateValidatorSetHeaderHash(valsetHeader1)
//	require.NoError(t, err)
//	headerSignature2, err := keyPair2.Sign(headerHash2)
//	require.NoError(t, err)
//
//	headerHash3, err := generator.GenerateValidatorSetHeaderHash(valsetHeader1)
//	require.NoError(t, err)
//	headerSignature3, err := keyPair3.Sign(headerHash3)
//	require.NoError(t, err)
//
//	aggSignature := bls.ZeroG1().
//		Add(headerSignature1).
//		Add(headerSignature2).
//		Add(headerSignature3)
//	//aggPublicKeyG1 := bls.ZeroG1().
//	//	Add(&svc.keyPair1.PublicKeyG1).
//	//	Add(&svc.keyPair2.PublicKeyG1).
//	//	Add(&svc.keyPair3.PublicKeyG1)
//	aggPublicKeyG2 := bls.ZeroG2().
//		Add(&keyPair1.PublicKeyG2).
//		Add(&keyPair2.PublicKeyG2).
//		Add(&keyPair3.PublicKeyG2)
//
//	proofData, err := proof.DoProve(proof.RawProveInput{
//		AllValidators:    validatorSet.Validators,
//		SignerValidators: validatorSet.Validators,
//		RequiredKeyTag:   15,
//		Message:          headerHash1,
//		Signature:        *aggSignature,
//		SignersAggKeyG2:  *aggPublicKeyG2,
//	})
//	require.NoError(t, err)
//
//	// proofData := decodeHex(t, "01c70454a912bf226d9b0a7b38dfef9319f92b893115fd5b168f0061c56a11e30d8c66ed7585aafd81e6c20cdbe81d385ee13871ef1da2b041e218076fbfd88e0cf9e7f5e7b25f241973a4a4ae6a7f29d430af5c243cd254d5035e7ad1883d9d1f0c0f76011867ebb6115185f5b9fe538de1181a39cd9e5efa03046b031c64df0b86f9a8fcb28738e82eabe0237a57bc47a02158841039f0d12ec3abb3d9ee2d12185fed2304764a978ba5405c684093479e18d934c7c8ba9e031981c836d4ff028f842b327dd18be5ba410bc423ce989f6807f1766acdae5669dea546d8cd591f98ba029d5e2a77520b2639234354c2e3983ce9590efbee7b293a8ee32bfbf80000000124997c0ef7b3e53580aaa97c84ae4682a7a7ec617110c5790ce06ca6bf837600114ec9b6c4503e96f11bcdb0e4601fecd83b5b8e4c7d9df6204aea2a4b7617471e9bada9e6dd91bf84b89967925bf1a90aa162f5f2883c4713522263f983f5ec101464ab309ff2a609396c898689eb0e5e4f703d350adc6ed69d6dfdc1a5bbbd")
//
//	fmt.Println("aggPublicKeyG2>>>", hex.EncodeToString(aggPublicKeyG2.Marshal()))
//	fmt.Println("aggSigG1>>>", hex.EncodeToString(aggSignature.Marshal()))
//
//	fmt.Println("proof_>>>", hex.EncodeToString(proofData.Proof))
//	fmt.Println("Commitments>>>", hex.EncodeToString(proofData.Commitments))
//	fmt.Println("commitmentPok>>>", hex.EncodeToString(proofData.CommitmentPok))
//	fmt.Println("inputs>>>", hex.EncodeToString(inputs(t)))
//
//	fmt.Println("fullProof", hex.EncodeToString(proofData.Marshall()))
//
//	require.NoError(t, err)
//}
//
//func inputs(t *testing.T) []byte {
//	t.Helper()
//	arguments := abi.Arguments{
//		{
//			Name: "activeAggregatedKeys",
//			Type: abi.Type{
//				T: abi.SliceTy,
//				Elem: &abi.Type{
//					T: abi.UintTy, Size: 256,
//				},
//			},
//		},
//	}
//
//	t.Helper()
//	in := []string{"0", "0", "0", "0", "0", "0", "0", "0", "17452784377140135873242247846499243451530443834097508626974155003329264289405", "0"}
//	result := make([]*big.Int, 0, len(in))
//	for _, s := range in {
//		b, ok := new(big.Int).SetString(s, 10)
//		require.True(t, ok)
//		result = append(result, b)
//	}
//	pack, err := arguments.Pack(result)
//	require.NoError(t, err)
//
//	return pack[64:] // remove first 64 bytes of dynamic array prefix, we need only bytes of inputs
//}
//
//func bytesFromPK(t *testing.T, pk1 string) []byte {
//	t.Helper()
//	b, ok := new(big.Int).SetString(pk1, 10)
//	require.True(t, ok)
//	return b.Bytes()
//}
