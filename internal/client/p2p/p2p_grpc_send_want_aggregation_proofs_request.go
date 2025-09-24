package p2p

import (
	"context"
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

	return protoToEntityAggregationProofResponse(response), nil
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
		SignatureTargetIds: lo.Map(req.SignatureTargetIDs, func(hash common.Hash, _ int) string {
			return hash.Hex()
		}),
	}
}

// protoToEntityAggregationProofResponse converts protobuf WantAggregationProofsResponse to entity
func protoToEntityAggregationProofResponse(resp *prototypes.WantAggregationProofsResponse) entity.WantAggregationProofsResponse {
	proofs := make(map[common.Hash]entity.AggregationProof)

	for hashStr, protoProof := range resp.GetProofs() {
		// Convert aggregation proof
		proof := entity.AggregationProof{
			MessageHash: protoProof.GetMessageHash(),
			Proof:       protoProof.GetProof(),
		}

		proofs[common.HexToHash(hashStr)] = proof
	}

	return entity.WantAggregationProofsResponse{
		Proofs: proofs,
	}
}

// protoToEntityAggregationProofRequest converts protobuf WantAggregationProofsRequest to entity
func protoToEntityAggregationProofRequest(req *prototypes.WantAggregationProofsRequest) entity.WantAggregationProofsRequest {
	signatureTargetIDs := make([]common.Hash, len(req.GetSignatureTargetIds()))

	for i, hashStr := range req.GetSignatureTargetIds() {
		signatureTargetIDs[i] = common.HexToHash(hashStr)
	}

	return entity.WantAggregationProofsRequest{
		SignatureTargetIDs: signatureTargetIDs,
	}
}

// entityToProtoAggregationProofResponse converts entity WantAggregationProofsResponse to protobuf
func entityToProtoAggregationProofResponse(resp entity.WantAggregationProofsResponse) *prototypes.WantAggregationProofsResponse {
	proofs := make(map[string]*prototypes.AggregationProof)

	for hash, proof := range resp.Proofs {
		// Convert aggregation proof
		protoProof := &prototypes.AggregationProof{
			MessageHash: proof.MessageHash,
			Proof:       proof.Proof,
		}

		proofs[hash.Hex()] = protoProof
	}

	return &prototypes.WantAggregationProofsResponse{
		Proofs: proofs,
	}
}
