package p2p

import (
	"fmt"
	"log"
	"testing"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsub_pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	prototypes "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
)

func TestHandleSignatureReadyMessage_WithOversizedPublicKey_ReturnsError(t *testing.T) {
	service := createTestService(t, false, nil)

	oversizedPubKey := make([]byte, maxPubKeySize+1)
	signature := &prototypes.Signature{
		KeyTag:      1,
		Epoch:       100,
		PublicKey:   oversizedPubKey,
		Signature:   []byte("test-signature"),
		MessageHash: []byte("test-hash"),
	}

	signatureData, err := proto.Marshal(signature)
	require.NoError(t, err)

	p2pMsg := &prototypes.P2PMessage{
		Data: signatureData,
	}

	p2pMsgData, err := proto.Marshal(p2pMsg)
	require.NoError(t, err)

	pubSubMsg := &pubsub.Message{
		Message: &pubsub_pb.Message{
			Data: p2pMsgData,
			From: []byte(service.host.ID()),
		},
		ReceivedFrom: service.host.ID(),
	}

	err = service.handleSignatureReadyMessage(pubSubMsg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("public key %x size exceeds maximum", oversizedPubKey))
}

func TestHandleSignatureReadyMessage_WithOversizedSignature_ReturnsError(t *testing.T) {
	service := createTestService(t, false, nil)

	oversizedSig := make([]byte, maxSignatureSize+1)
	signature := &prototypes.Signature{
		KeyTag:      1,
		Epoch:       100,
		PublicKey:   []byte("valid-pub-key"),
		Signature:   oversizedSig,
		MessageHash: []byte("test-hash"),
	}

	signatureData, err := proto.Marshal(signature)
	require.NoError(t, err)

	p2pMsg := &prototypes.P2PMessage{
		Data: signatureData,
	}

	p2pMsgData, err := proto.Marshal(p2pMsg)
	require.NoError(t, err)

	pubSubMsg := &pubsub.Message{
		Message: &pubsub_pb.Message{
			Data: p2pMsgData,
			From: []byte(service.host.ID()),
		},
		ReceivedFrom: service.host.ID(),
	}

	err = service.handleSignatureReadyMessage(pubSubMsg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("signature %x size exceeds maximum", oversizedSig))
}

func TestHandleSignatureReadyMessage_WithOversizedMessageHash_ReturnsError(t *testing.T) {
	service := createTestService(t, false, nil)

	oversizedHash := make([]byte, maxMsgHashSize+1)
	signature := &prototypes.Signature{
		KeyTag:      1,
		Epoch:       100,
		PublicKey:   []byte("valid-pub-key"),
		Signature:   []byte("valid-signature"),
		MessageHash: oversizedHash,
	}

	signatureData, err := proto.Marshal(signature)
	require.NoError(t, err)

	p2pMsg := &prototypes.P2PMessage{
		Data: signatureData,
	}

	p2pMsgData, err := proto.Marshal(p2pMsg)
	require.NoError(t, err)

	pubSubMsg := &pubsub.Message{
		Message: &pubsub_pb.Message{
			Data: p2pMsgData,
			From: []byte(service.host.ID()),
		},
		ReceivedFrom: service.host.ID(),
	}

	err = service.handleSignatureReadyMessage(pubSubMsg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("message hash %x size exceeds maximum", oversizedHash))
}

func TestHandleSignatureReadyMessage_WithInvalidPublicKey_ReturnsError(t *testing.T) {
	service := createTestService(t, false, nil)

	signature := &prototypes.Signature{
		KeyTag:      1,
		Epoch:       100,
		PublicKey:   []byte("invalid-pub-key-format"),
		Signature:   []byte("valid-signature"),
		MessageHash: []byte("test-hash"),
	}

	signatureData, err := proto.Marshal(signature)
	require.NoError(t, err)

	p2pMsg := &prototypes.P2PMessage{
		Data: signatureData,
	}

	p2pMsgData, err := proto.Marshal(p2pMsg)
	require.NoError(t, err)

	pubSubMsg := &pubsub.Message{
		Message: &pubsub_pb.Message{
			Data: p2pMsgData,
			From: []byte(service.host.ID()),
		},
		ReceivedFrom: service.host.ID(),
	}

	err = service.handleSignatureReadyMessage(pubSubMsg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse public key")
}

func TestHandleAggregatedProofReadyMessage_WithOversizedMessageHash_ReturnsError(t *testing.T) {
	service := createTestService(t, false, nil)

	oversizedHash := make([]byte, maxMsgHashSize+1)
	aggProof := &prototypes.AggregationProof{
		KeyTag:      1,
		Epoch:       100,
		MessageHash: oversizedHash,
		Proof:       []byte("test-proof"),
	}

	aggProofData, err := proto.Marshal(aggProof)
	require.NoError(t, err)

	p2pMsg := &prototypes.P2PMessage{
		Data: aggProofData,
	}

	p2pMsgData, err := proto.Marshal(p2pMsg)
	require.NoError(t, err)

	pubSubMsg := &pubsub.Message{
		Message: &pubsub_pb.Message{
			Data: p2pMsgData,
			From: []byte(service.host.ID()),
		},
		ReceivedFrom: service.host.ID(),
	}

	err = service.handleAggregatedProofReadyMessage(pubSubMsg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("aggregation proof message hash %x size exceeds maximum", oversizedHash))
}

func TestHandleAggregatedProofReadyMessage_WithOversizedProof_ReturnsError(t *testing.T) {
	service := createTestService(t, false, nil)

	oversizedProof := make([]byte, maxProofSize+1)
	aggProof := &prototypes.AggregationProof{
		KeyTag:      1,
		Epoch:       100,
		MessageHash: []byte("test-hash"),
		Proof:       oversizedProof,
	}

	aggProofData, err := proto.Marshal(aggProof)
	require.NoError(t, err)

	p2pMsg := &prototypes.P2PMessage{
		Data: aggProofData,
	}

	p2pMsgData, err := proto.Marshal(p2pMsg)
	require.NoError(t, err)

	pubSubMsg := &pubsub.Message{
		Message: &pubsub_pb.Message{
			Data: p2pMsgData,
			From: []byte(service.host.ID()),
		},
		ReceivedFrom: service.host.ID(),
	}

	err = service.handleAggregatedProofReadyMessage(pubSubMsg)

	require.Error(t, err)
	log.Println(err.Error())
	assert.Contains(t, err.Error(), fmt.Sprintf("aggregation proof %x size exceeds maximum", oversizedProof))
}

func TestUnmarshalMessage_WithInvalidP2PMessage_ReturnsError(t *testing.T) {
	invalidData := []byte("invalid protobuf data")

	pubSubMsg := &pubsub.Message{
		Message: &pubsub_pb.Message{
			Data: invalidData,
		},
	}

	var signature prototypes.Signature
	_, err := unmarshalMessage(pubSubMsg, &signature)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal message")
}

func TestUnmarshalMessage_WithInvalidInnerMessage_ReturnsError(t *testing.T) {
	p2pMsg := &prototypes.P2PMessage{
		Data: []byte("invalid inner protobuf data"),
	}

	p2pMsgData, err := proto.Marshal(p2pMsg)
	require.NoError(t, err)

	pubSubMsg := &pubsub.Message{
		Message: &pubsub_pb.Message{
			Data: p2pMsgData,
		},
	}

	var signature prototypes.Signature
	_, err = unmarshalMessage(pubSubMsg, &signature)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal message")
}

func TestExtractSenderInfo_WithInvalidPeerID_ReturnsError(t *testing.T) {
	invalidPeerID := peer.ID("invalid-peer-id-format")

	pubSubMsg := &pubsub.Message{
		Message: &pubsub_pb.Message{
			From: []byte(invalidPeerID),
		},
		ReceivedFrom: invalidPeerID,
	}

	_, err := extractSenderInfo(pubSubMsg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to extract public key")
}
