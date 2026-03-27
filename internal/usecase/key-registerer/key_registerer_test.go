package key_registerer

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	symbioticCrypto "github.com/symbioticfi/relay/symbiotic/usecase/crypto"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto/blsBn254"
)

func TestBuildArtifactECDSANormalizesSignatureAndLeavesEmptyExtraData(t *testing.T) {
	t.Parallel()

	keyTag, err := symbiotic.KeyTagFromTypeAndId(symbiotic.KeyTypeEcdsaSecp256k1, 1)
	require.NoError(t, err)

	privateKeyBytes := make([]byte, 32)
	privateKeyBytes[len(privateKeyBytes)-1] = 0x7b

	privateKey, err := symbioticCrypto.NewPrivateKey(symbiotic.KeyTypeEcdsaSecp256k1, privateKeyBytes)
	require.NoError(t, err)

	registerer, evmClient := newTestRegisterer(t)
	operator := common.HexToAddress("0x00000000000000000000000000000000000000a1")

	artifact, err := registerer.BuildArtifact(context.Background(), privateKey, keyTag, operator)
	require.NoError(t, err)

	commitmentData, err := keyCommitmentData(testEip712Domain, operator, privateKey.PublicKey().OnChain())
	require.NoError(t, err)
	require.Len(t, commitmentData, 66)
	expectedSignature, _, err := privateKey.Sign(commitmentData)
	require.NoError(t, err)
	expectedSignature, err = normalizeSignature(keyTag, expectedSignature)
	require.NoError(t, err)

	require.Equal(t, keyTag, artifact.KeyTag)
	require.Equal(t, privateKey.PublicKey().OnChain(), artifact.Key)
	require.Equal(t, expectedSignature, artifact.Signature)
	require.Empty(t, artifact.ExtraData)

	jsonBytes, err := artifact.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t,
		`{"keyTag":17,"keyHex":"`+hexutil.Encode(artifact.Key)+`","signatureHex":"`+hexutil.Encode(artifact.Signature)+`","extraDataHex":"0x"}`,
		string(jsonBytes),
	)

	require.Equal(t, testNetworkConfig.KeysProvider, evmClient.eip712RequestedFor)
}

func TestBuildArtifactBLS12381FormatsSignedEnvelopeForContract(t *testing.T) {
	t.Parallel()

	keyTag, err := symbiotic.KeyTagFromTypeAndId(symbiotic.KeyTypeBls12381, 1)
	require.NoError(t, err)

	privateKey, err := symbioticCrypto.NewPrivateKey(symbiotic.KeyTypeBls12381, []byte("bls12381-private-key-material"))
	require.NoError(t, err)

	registerer, _ := newTestRegisterer(t)
	operator := common.HexToAddress("0x00000000000000000000000000000000000000a5")

	artifact, err := registerer.BuildArtifact(context.Background(), privateKey, keyTag, operator)
	require.NoError(t, err)

	commitmentData, err := keyCommitmentData(testEip712Domain, operator, artifact.Key)
	require.NoError(t, err)
	require.Len(t, commitmentData, 66)

	rawSignature, _, err := privateKey.Sign(commitmentData)
	require.NoError(t, err)
	expectedSignature, err := normalizeSignature(keyTag, rawSignature)
	require.NoError(t, err)

	require.Len(t, artifact.Signature, 128)
	require.Equal(t, expectedSignature, artifact.Signature)
}

func TestBuildArtifactBLSDerivesExtraData(t *testing.T) {
	t.Parallel()

	keyTag, err := symbiotic.KeyTagFromTypeAndId(symbiotic.KeyTypeBlsBn254, 1)
	require.NoError(t, err)

	privateKey, err := symbioticCrypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, []byte("bls-private-key-material"))
	require.NoError(t, err)

	registerer, _ := newTestRegisterer(t)
	operator := common.HexToAddress("0x00000000000000000000000000000000000000a2")

	artifact, err := registerer.BuildArtifact(context.Background(), privateKey, keyTag, operator)
	require.NoError(t, err)

	blsKey, err := blsBn254.FromRaw(privateKey.PublicKey().Raw())
	require.NoError(t, err)
	expectedExtraData := blsKey.G2().RawBytes()

	require.Equal(t, expectedExtraData[:], artifact.ExtraData)
	require.NotEmpty(t, artifact.ExtraData)
}

func TestBuildArtifactBLS12381DerivesExtraData(t *testing.T) {
	t.Parallel()

	keyTag, err := symbiotic.KeyTagFromTypeAndId(symbiotic.KeyTypeBls12381, 1)
	require.NoError(t, err)

	privateKey, err := symbioticCrypto.NewPrivateKey(symbiotic.KeyTypeBls12381, []byte("bls12381-private-key-material"))
	require.NoError(t, err)

	registerer, _ := newTestRegisterer(t)
	operator := common.HexToAddress("0x00000000000000000000000000000000000000a3")

	artifact, err := registerer.BuildArtifact(context.Background(), privateKey, keyTag, operator)
	require.NoError(t, err)
	require.Len(t, artifact.Key, 128)
	require.Len(t, artifact.Signature, 128)
	require.Len(t, artifact.ExtraData, 256)
	require.NotEmpty(t, artifact.ExtraData)

	commitmentData, err := keyCommitmentData(testEip712Domain, operator, artifact.Key)
	require.NoError(t, err)
	rawSignature, _, err := privateKey.Sign(commitmentData)
	require.NoError(t, err)
	expectedSignature, err := normalizeSignature(keyTag, rawSignature)
	require.NoError(t, err)
	require.Equal(t, expectedSignature, artifact.Signature)
	require.Len(t, artifact.ExtraData, 256)
}

func TestBuildArtifactBLS12381ReferenceJSON(t *testing.T) {
	t.Parallel()

	keyTag, err := symbiotic.KeyTagFromTypeAndId(symbiotic.KeyTypeBls12381, 1)
	require.NoError(t, err)

	privateKey, err := symbioticCrypto.NewPrivateKey(symbiotic.KeyTypeBls12381, []byte("bls12381-private-key-material"))
	require.NoError(t, err)

	registerer, _ := newTestRegisterer(t)
	operator := common.HexToAddress("0x00000000000000000000000000000000000000a3")

	artifact, err := registerer.BuildArtifact(context.Background(), privateKey, keyTag, operator)
	require.NoError(t, err)

	jsonBytes, err := artifact.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, `{
		"keyTag":33,
		"keyHex":"0x000000000000000000000000000000000108d3125037c0cee9d147d405ae9278778488ad59865f3cf672f1e4e8bdcd00f8d2aa5cfd21327189e85fd94120743600000000000000000000000000000000197bbbb67e44ca22f32ff7befd3c32f88e6147b6d45d9c08cd03cbf515c73c5f4d31801295651338c41b4b728e40aa29",
		"signatureHex":"0x000000000000000000000000000000000acd4c092dcca55810516616826fa51ba935fbc65b60edaafc98ef7d2f4277acb955bcbec1e34323c61ccacedc5ca9bc00000000000000000000000000000000038ee7556a6a8bf5b121f7eace864309edec7f60cd8a1e69aa36376f2ec6e53beca834d32fca1361a724a0026faa2ed1",
		"extraDataHex":"0x000000000000000000000000000000000b2f6423035008f782e5604a2475143d81b7776a8f5755ce539bd66ceb24af0c312bca3664ad92145363df90a6e7abcc0000000000000000000000000000000009d6c5e95d11c3d4d8ee80486b7cade6e4b58e907590e9cc0f7aa02cb3d29c1634a3518190964cb60d5d8525ffe73952000000000000000000000000000000000d3d8f0617e104c3871b0b6dd671db3c154c2abca7b0b945939fc2c39c45a9733194a3a0236a6b91ca13da8405ab4638000000000000000000000000000000001114fab73df7d7988e480d7749e2ffbf53ce09e3459a2ab7c63b0c50d373da0ae91324aad6abab1c9e83b3f5a64de21c"
	}`, string(jsonBytes))
}

func TestRegisterUsesArtifactBytesForTransaction(t *testing.T) {
	t.Parallel()

	keyTag, err := symbiotic.KeyTagFromTypeAndId(symbiotic.KeyTypeBlsBn254, 1)
	require.NoError(t, err)

	privateKey, err := symbioticCrypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, []byte("bls-private-key-material"))
	require.NoError(t, err)

	registerer, evmClient := newTestRegisterer(t)
	operator := common.HexToAddress("0x00000000000000000000000000000000000000a4")

	artifact, err := registerer.BuildArtifact(context.Background(), privateKey, keyTag, operator)
	require.NoError(t, err)

	_, err = registerer.Register(context.Background(), privateKey, keyTag, operator)
	require.NoError(t, err)

	require.Equal(t, testNetworkConfig.KeysProvider, evmClient.registerKeyAddress)
	require.Equal(t, keyTag, evmClient.registerKeyTag)
	require.Equal(t, artifact.Key, evmClient.registeredKey)
	require.Equal(t, artifact.Signature, evmClient.registeredSignature)
	require.Equal(t, artifact.ExtraData, evmClient.registeredExtraData)
}

var testEip712Domain = symbiotic.Eip712Domain{
	Name:              "Key Registry",
	Version:           "1",
	ChainId:           big.NewInt(31337),
	VerifyingContract: common.HexToAddress("0x0000000000000000000000000000000000000abc"),
}

var testNetworkConfig = symbiotic.NetworkConfig{
	KeysProvider: symbiotic.CrossChainAddress{
		ChainId: 31337,
		Address: common.HexToAddress("0x0000000000000000000000000000000000000def"),
	},
}

type testEVMClient struct {
	eip712RequestedFor  symbiotic.CrossChainAddress
	registerKeyAddress  symbiotic.CrossChainAddress
	registerKeyTag      symbiotic.KeyTag
	registeredKey       symbiotic.CompactPublicKey
	registeredSignature symbiotic.RawSignature
	registeredExtraData []byte
}

func newTestRegisterer(t *testing.T) (*Registerer, *testEVMClient) {
	t.Helper()

	evmClient := &testEVMClient{}
	registerer, err := NewRegisterer(Config{EVMClient: evmClient})
	require.NoError(t, err)
	return registerer, evmClient
}

func (t *testEVMClient) GetCurrentEpoch(ctx context.Context) (symbiotic.Epoch, error) {
	return 42, nil
}

func (t *testEVMClient) GetConfig(ctx context.Context, timestamp symbiotic.Timestamp, epoch symbiotic.Epoch) (symbiotic.NetworkConfig, error) {
	return testNetworkConfig, nil
}

func (t *testEVMClient) GetEip712Domain(ctx context.Context, addr symbiotic.CrossChainAddress) (symbiotic.Eip712Domain, error) {
	t.eip712RequestedFor = addr
	return testEip712Domain, nil
}

func (t *testEVMClient) GetEpochStart(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.Timestamp, error) {
	return 1000, nil
}

func (t *testEVMClient) RegisterKey(
	ctx context.Context,
	addr symbiotic.CrossChainAddress,
	keyTag symbiotic.KeyTag,
	key symbiotic.CompactPublicKey,
	signature symbiotic.RawSignature,
	extraData []byte,
) (symbiotic.TxResult, error) {
	t.registerKeyAddress = addr
	t.registerKeyTag = keyTag
	t.registeredKey = append(symbiotic.CompactPublicKey(nil), key...)
	t.registeredSignature = append(symbiotic.RawSignature(nil), signature...)
	t.registeredExtraData = append([]byte(nil), extraData...)
	return symbiotic.TxResult{}, nil
}
