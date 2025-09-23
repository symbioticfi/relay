package p2p

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/libp2p/go-libp2p"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/symbioticfi/relay/core/entity"
	v2 "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
)

func TestP2P_GRPC(t *testing.T) {
	srvHost, err := libp2p.New()
	require.NoError(t, err)
	clientHost, err := libp2p.New()
	require.NoError(t, err)
	defer srvHost.Close()
	defer clientHost.Close()

	srvHost.Peerstore().AddAddrs(clientHost.ID(), clientHost.Addrs(), peerstore.PermanentAddrTTL)
	clientHost.Peerstore().AddAddrs(srvHost.ID(), srvHost.Addrs(), peerstore.PermanentAddrTTL)

	grpcServer := grpc.NewServer()
	var tag protocol.ID = "/testp2pgrpc"

	done := make(chan struct{})
	ready := make(chan struct{})
	go func() {
		defer close(done)
		listener, err := gostream.Listen(srvHost, tag)
		assert.NoError(t, err)
		defer listener.Close()

		v2.RegisterSymbioticP2PServiceServer(grpcServer, &GRPCHandler{
			syncHandler: myHandler{},
		})
		close(ready)

		assert.NoError(t, grpcServer.Serve(listener))
	}()

	<-ready

	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithMax(3),
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(time.Second)),
	}
	unaryInterceptors := []grpc.UnaryClientInterceptor{grpc_retry.UnaryClientInterceptor(retryOpts...)}
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpcmiddleware.ChainUnaryClient(unaryInterceptors...)),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100*1024*1024), grpc.MaxCallSendMsgSize(100*1024*1024)),
	}

	dialOpts = append(dialOpts, grpc.WithContextDialer(func(ctx context.Context, peerIdStr string) (net.Conn, error) {
		peerID, err := peer.Decode(peerIdStr)
		if err != nil {
			return nil, err
		}

		conn, err := gostream.Dial(ctx, clientHost, peerID, tag)
		if err != nil {
			return nil, err
		}

		return conn, nil
	}))

	conn, err := grpc.NewClient("passthrough:///"+srvHost.ID().String(), dialOpts...)
	require.NoError(t, err)
	defer conn.Close()

	client := v2.NewSymbioticP2PServiceClient(conn)
	_, err = client.WantSignatures(context.Background(), &v2.WantSignaturesRequest{})
	require.NoError(t, err)

	grpcServer.GracefulStop()
	<-done
}

type myHandler struct {
	v2.UnimplementedSymbioticP2PServiceServer
}

func (m myHandler) WantSignatures(ctx context.Context, request *v2.WantSignaturesRequest) (*v2.WantSignaturesResponse, error) {
	return &v2.WantSignaturesResponse{
		Signatures: make(map[string]*v2.ValidatorSignatureList),
	}, nil
}

func (m myHandler) HandleWantSignaturesRequest(ctx context.Context, request entity.WantSignaturesRequest) (entity.WantSignaturesResponse, error) {
	return entity.WantSignaturesResponse{
		Signatures: make(map[common.Hash][]entity.ValidatorSignature),
	}, nil
}

func (m myHandler) HandleWantAggregationProofsRequest(ctx context.Context, request entity.WantAggregationProofsRequest) (entity.WantAggregationProofsResponse, error) {
	return entity.WantAggregationProofsResponse{
		Proofs: make(map[common.Hash]entity.AggregationProof),
	}, nil
}
