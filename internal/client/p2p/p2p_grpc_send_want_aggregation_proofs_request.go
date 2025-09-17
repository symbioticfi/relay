package p2p

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/samber/lo"

	"github.com/symbioticfi/relay/core/entity"
	prototypes "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
	"github.com/symbioticfi/relay/pkg/log"
)

// SendWantAggregationProofsRequest sends a synchronous aggregation proof request to a peer
func (s *Service) SendWantAggregationProofsRequest(ctx context.Context, request entity.WantAggregationProofsRequest) (entity.WantAggregationProofsResponse, error) {
	ctx = log.WithComponent(ctx, "p2p")

	// Convert entity request to protobuf
	protoReq := entityToProtoAggregationProofRequest(request)

	// Select a peer for the request
	peerID, err := s.selectPeerForSync()
	if err != nil {
		return entity.WantAggregationProofsResponse{}, errors.Errorf("failed to select peer: %w", err)
	}

	// Send request to the selected peer
	response, err := s.sendAggregationProofRequestToPeer(ctx, peerID, protoReq)
	if err != nil {
		return entity.WantAggregationProofsResponse{}, errors.Errorf("failed to get aggregation proofs from peer %s: %w", peerID, err)
	}

	// Convert protobuf response to entity
	entityResp, err := protoToEntityAggregationProofResponse(response)
	if err != nil {
		return entity.WantAggregationProofsResponse{}, errors.Errorf("failed to convert aggregation proof response: %w", err)
	}

	return entityResp, nil
}

// sendAggregationProofRequestToPeer sends a gRPC aggregation proof request to a specific peer
func (s *Service) sendAggregationProofRequestToPeer(ctx context.Context, peerID peer.ID, req *prototypes.WantAggregationProofsRequest) (*prototypes.WantAggregationProofsResponse, error) {
	// Create gRPC connection over libp2p stream
	conn, err := s.createGRPCConnection(ctx, peerID)
	if err != nil {
		return nil, errors.Errorf("failed to create gRPC connection to peer %s: %w", peerID, err)
	}
	defer conn.Close()

	// Create gRPC client and send request
	client := prototypes.NewSymbioticP2PServiceClient(conn)

	requestCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	response, err := client.WantAggregationProofs(requestCtx, req)
	if err != nil {
		return nil, errors.Errorf("gRPC aggregation proof request failed: %w", err)
	}

	return response, nil
}

// entityToProtoAggregationProofRequest converts entity.WantAggregationProofsRequest to protobuf
func entityToProtoAggregationProofRequest(req entity.WantAggregationProofsRequest) *prototypes.WantAggregationProofsRequest {
	return &prototypes.WantAggregationProofsRequest{
		RequestHashes: lo.Map(req.RequestHashes, func(hash common.Hash, _ int) string {
			return hash.Hex()
		}),
	}
}

// protoToEntityAggregationProofResponse converts protobuf WantAggregationProofsResponse to entity
func protoToEntityAggregationProofResponse(resp *prototypes.WantAggregationProofsResponse) (entity.WantAggregationProofsResponse, error) {
	proofs := make(map[common.Hash]entity.AggregationProof)

	for hashStr, protoProof := range resp.GetProofs() {
		// Parse hash from hex string
		hashBytes, err := hex.DecodeString(hashStr)
		if err != nil {
			return entity.WantAggregationProofsResponse{}, errors.Errorf("failed to decode hash %s: %w", hashStr, err)
		}

		hash := common.BytesToHash(hashBytes)

		// Convert aggregation proof
		proof := entity.AggregationProof{
			VerificationType: entity.VerificationType(protoProof.GetVerificationType()),
			MessageHash:      protoProof.GetMessageHash(),
			Proof:            protoProof.GetProof(),
		}

		proofs[hash] = proof
	}

	return entity.WantAggregationProofsResponse{
		Proofs: proofs,
	}, nil
}

// protoToEntityAggregationProofRequest converts protobuf WantAggregationProofsRequest to entity
func protoToEntityAggregationProofRequest(req *prototypes.WantAggregationProofsRequest) (entity.WantAggregationProofsRequest, error) {
	requestHashes := make([]common.Hash, len(req.GetRequestHashes()))

	for i, hashStr := range req.GetRequestHashes() {
		// Parse hash from hex string
		hashBytes, err := hex.DecodeString(hashStr)
		if err != nil {
			return entity.WantAggregationProofsRequest{}, errors.Errorf("failed to decode hash %s: %w", hashStr, err)
		}

		requestHashes[i] = common.BytesToHash(hashBytes)
	}

	return entity.WantAggregationProofsRequest{
		RequestHashes: requestHashes,
	}, nil
}

// entityToProtoAggregationProofResponse converts entity WantAggregationProofsResponse to protobuf
func entityToProtoAggregationProofResponse(resp entity.WantAggregationProofsResponse) *prototypes.WantAggregationProofsResponse {
	proofs := make(map[string]*prototypes.AggregationProof)

	for hash, proof := range resp.Proofs {
		hashKey := hex.EncodeToString(hash.Bytes())

		// Convert aggregation proof
		protoProof := &prototypes.AggregationProof{
			VerificationType: uint32(proof.VerificationType),
			MessageHash:      proof.MessageHash,
			Proof:            proof.Proof,
		}

		proofs[hashKey] = protoProof
	}

	return &prototypes.WantAggregationProofsResponse{
		Proofs: proofs,
	}
}
