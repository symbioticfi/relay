package p2p

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	prototypes "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/tracing"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

const grpcProtocolTag protocol.ID = "/relay/v1/grpc"

// SendWantSignaturesRequest sends a synchronous signature request to a peer
func (s *Service) SendWantSignaturesRequest(ctx context.Context, request entity.WantSignaturesRequest) (entity.WantSignaturesResponse, error) {
	ctx, span := tracing.StartClientSpan(ctx, "p2p.SendWantSignaturesRequest",
		tracing.AttrSignatureCount.Int(len(request.WantSignatures)),
	)
	defer span.End()

	ctx = log.WithComponent(ctx, "p2p")

	tracing.AddEvent(span, "converting_request")
	// Convert entity request to protobuf
	protoReq, err := entityToProtoRequest(request)
	if err != nil {
		tracing.RecordError(span, err)
		return entity.WantSignaturesResponse{}, errors.Errorf("failed to convert request: %w", err)
	}

	tracing.AddEvent(span, "selecting_peer")
	// Select a peer for the request
	peerID, err := s.selectPeerForSync()
	if err != nil {
		tracing.RecordError(span, err)
		return entity.WantSignaturesResponse{}, errors.Errorf("failed to select peer: %w", err)
	}

	tracing.SetAttributes(span, tracing.AttrPeerID.String(peerID.String()))

	tracing.AddEvent(span, "sending_request")
	// Send request to the selected peer
	response, err := s.sendRequestToPeer(ctx, peerID, protoReq)
	if err != nil {
		tracing.RecordError(span, err)
		return entity.WantSignaturesResponse{}, errors.Errorf("failed to get signatures from peer %s: %w", peerID, err)
	}

	tracing.AddEvent(span, "converting_response")
	// Convert protobuf response to entity
	entityResp := protoToEntityResponse(ctx, response)

	tracing.SetAttributes(span,
		tracing.AttrSignatureCount.Int(len(entityResp.Signatures)),
	)

	tracing.AddEvent(span, "request_completed")
	return entityResp, nil
}

// sendRequestToPeer sends a gRPC request to a specific peer
func (s *Service) sendRequestToPeer(ctx context.Context, peerID peer.ID, req *prototypes.WantSignaturesRequest) (*prototypes.WantSignaturesResponse, error) {
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

	response, err := client.WantSignatures(requestCtx, req)
	if err != nil {
		return nil, errors.Errorf("gRPC request failed: %w", err)
	}

	return response, nil
}

// createGRPCConnection creates a gRPC connection to a peer over libp2p
func (s *Service) createGRPCConnection(ctx context.Context, peerID peer.ID) (*grpc.ClientConn, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, peerIdStr string) (net.Conn, error) {
			targetPeer, err := peer.Decode(peerIdStr)
			if err != nil {
				return nil, err
			}

			conn, err := gostream.Dial(ctx, s.host, targetPeer, grpcProtocolTag)
			if err != nil {
				return nil, err
			}

			return conn, nil
		}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxP2PMessageSize),
			grpc.MaxCallSendMsgSize(maxP2PMessageSize),
		),
	}

	// Attach tracing interceptors whenever we have a valid span context so trace headers propagate
	if spanCtx := trace.SpanFromContext(ctx).SpanContext(); spanCtx.IsValid() {
		dialOpts = append(dialOpts,
			grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		)
	}

	conn, err := grpc.NewClient("passthrough:///"+peerID.String(), dialOpts...)
	if err != nil {
		return nil, errors.Errorf("failed to create gRPC client: %w", err)
	}

	return conn, nil
}

// entityToProtoRequest converts entity.WantSignaturesRequest to protobuf
func entityToProtoRequest(req entity.WantSignaturesRequest) (*prototypes.WantSignaturesRequest, error) {
	wantSignatures := make(map[string][]byte)

	for hash, bitmap := range req.WantSignatures {
		// Serialize roaring bitmap to bytes
		bitmapBytes, err := bitmap.ToBytes()
		if err != nil {
			return nil, errors.Errorf("failed to serialize bitmap for hash %s: %w", hash.Hex(), err)
		}

		wantSignatures[hash.Hex()] = bitmapBytes
	}

	return &prototypes.WantSignaturesRequest{
		WantSignatures: wantSignatures,
	}, nil
}

// protoToEntityResponse converts protobuf WantSignaturesResponse to entity
func protoToEntityResponse(ctx context.Context, resp *prototypes.WantSignaturesResponse) entity.WantSignaturesResponse {
	signatures := make(map[common.Hash][]entity.ValidatorSignature)

	for hashStr, sigList := range resp.GetSignatures() {
		// Convert validator signatures
		var validatorSigs []entity.ValidatorSignature
		for _, protoSig := range sigList.GetSignatures() {
			pubKey, err := crypto.NewPublicKey(symbiotic.KeyTag(protoSig.GetSignature().GetKeyTag()).Type(), protoSig.GetSignature().GetPublicKey())
			if err != nil {
				slog.WarnContext(ctx, "Failed to parse public key from peer[WantSignaturesResponse], skipping signature", "error", err)
				continue
			}
			sig := entity.ValidatorSignature{
				ValidatorIndex: protoSig.GetValidatorIndex(),
				Signature: symbiotic.Signature{
					MessageHash: protoSig.GetSignature().GetMessageHash(),
					KeyTag:      symbiotic.KeyTag(protoSig.GetSignature().GetKeyTag()),
					Epoch:       symbiotic.Epoch(protoSig.GetSignature().GetEpoch()),
					PublicKey:   pubKey,
					Signature:   protoSig.GetSignature().GetSignature(),
				},
			}
			validatorSigs = append(validatorSigs, sig)
		}

		signatures[common.HexToHash(hashStr)] = validatorSigs
	}

	return entity.WantSignaturesResponse{
		Signatures: signatures,
	}
}

// selectPeerForSync selects a single peer for synchronous signature requests
func (s *Service) selectPeerForSync() (peer.ID, error) {
	peers := s.host.Network().Peers()
	if len(peers) == 0 {
		return "", errors.Errorf("no peers available for sync: %w", entity.ErrNoPeers)
	}

	//nolint:gosec // G404: non-cryptographic random selection
	selectedPeer := peers[rand.IntN(len(peers))]
	return selectedPeer, nil
}

// protoToEntityRequest converts protobuf WantSignaturesRequest to entity
func protoToEntityRequest(req *prototypes.WantSignaturesRequest) (entity.WantSignaturesRequest, error) {
	wantSignatures := make(map[common.Hash]entity.Bitmap)

	for hashStr, bitmapBytes := range req.GetWantSignatures() {
		// Deserialize roaring bitmap from bytes
		bitmap := entity.NewBitmap()
		if _, err := bitmap.FromBuffer(bitmapBytes); err != nil {
			return entity.WantSignaturesRequest{}, errors.Errorf("failed to deserialize bitmap for hash %s: %w", hashStr, err)
		}

		wantSignatures[common.HexToHash(hashStr)] = bitmap
	}

	return entity.WantSignaturesRequest{
		WantSignatures: wantSignatures,
	}, nil
}

// entityToProtoResponse converts entity WantSignaturesResponse to protobuf
func entityToProtoResponse(resp entity.WantSignaturesResponse) *prototypes.WantSignaturesResponse {
	signatures := make(map[string]*prototypes.ValidatorSignatureList)

	for hash, sigList := range resp.Signatures {
		// Convert validator signatures
		var protoSigs []*prototypes.ValidatorSignature
		for _, validatorSig := range sigList {
			protoSig := &prototypes.ValidatorSignature{
				ValidatorIndex: validatorSig.ValidatorIndex,
				Signature: &prototypes.Signature{
					MessageHash: validatorSig.Signature.MessageHash,
					KeyTag:      uint32(validatorSig.Signature.KeyTag),
					Epoch:       uint64(validatorSig.Signature.Epoch),
					Signature:   validatorSig.Signature.Signature,
					PublicKey:   validatorSig.Signature.PublicKey.Raw(),
				},
			}
			protoSigs = append(protoSigs, protoSig)
		}

		signatures[hash.Hex()] = &prototypes.ValidatorSignatureList{
			Signatures: protoSigs,
		}
	}

	return &prototypes.WantSignaturesResponse{
		Signatures: signatures,
	}
}
