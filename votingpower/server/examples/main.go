package main

import (
	"context"
	"log"
	"net"
	"time"

	serverv1 "github.com/symbioticfi/relay/votingpower/server/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:9090")
	if err != nil {
		log.Fatalf("listen failed: %v", err)
	}

	grpcServer := grpc.NewServer()
	serverv1.RegisterVotingPowerProviderServiceServer(grpcServer, serverv1.NewServer())
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	reflection.Register(grpcServer)

	log.Printf("voting power example server listening on :9090")
	log.Printf("serving static response at %s", time.Now().UTC().Format(time.RFC3339))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve failed: %v", err)
	}
}
