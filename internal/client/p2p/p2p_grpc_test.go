package p2p

import (
	"bufio"
	"context"
	"io"
	"net"
	"testing"
	"time"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-gostream"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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
		require.NoError(t, err)
		defer listener.Close()

		v2.RegisterSymbioticP2PServiceServer(grpcServer, &p2pHandler{})
		close(ready)

		require.NoError(t, grpcServer.Serve(listener))
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

func TestServerClient(t *testing.T) {
	srvHost, err := libp2p.New()
	require.NoError(t, err)
	clientHost, err := libp2p.New()
	require.NoError(t, err)
	defer srvHost.Close()
	defer clientHost.Close()

	srvHost.Peerstore().AddAddrs(clientHost.ID(), clientHost.Addrs(), peerstore.PermanentAddrTTL)
	clientHost.Peerstore().AddAddrs(srvHost.ID(), srvHost.Addrs(), peerstore.PermanentAddrTTL)

	var tag protocol.ID = "/testitytest"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		listener, err := gostream.Listen(srvHost, tag)
		if err != nil {
			t.Error(err)
			return
		}
		defer listener.Close()

		if listener.Addr().String() != srvHost.ID().String() {
			t.Error("bad listener address")
			return
		}

		servConn, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer servConn.Close()

		reader := bufio.NewReader(servConn)
		for {
			msg, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Error(err)
				return
			}
			if msg != "is libp2p awesome?\n" {
				t.Errorf("Bad incoming message: %s", msg)
				return
			}

			_, err = servConn.Write([]byte("yes it is\n"))
			if err != nil {
				t.Error(err)
				return
			}
		}
	}()

	clientConn, err := gostream.Dial(ctx, clientHost, srvHost.ID(), tag)
	if err != nil {
		t.Fatal(err)
	}

	if clientConn.LocalAddr().String() != clientHost.ID().String() {
		t.Fatal("Bad LocalAddr")
	}

	if clientConn.RemoteAddr().String() != srvHost.ID().String() {
		t.Fatal("Bad RemoteAddr")
	}

	if clientConn.LocalAddr().Network() != gostream.Network {
		t.Fatal("Bad Network()")
	}

	err = clientConn.SetDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	err = clientConn.SetReadDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	err = clientConn.SetWriteDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	_, err = clientConn.Write([]byte("is libp2p awesome?\n"))
	if err != nil {
		t.Fatal(err)
	}

	reader := bufio.NewReader(clientConn)
	resp, err := reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}

	if string(resp) != "yes it is\n" {
		t.Errorf("Bad response: %s", resp)
	}

	err = clientConn.Close()
	if err != nil {
		t.Fatal(err)
	}
	<-done
}
