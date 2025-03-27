package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/multiformats/go-multiaddr"
)

func main() {
	listenAddr := flag.String("listen", "/ip4/0.0.0.0/tcp/0", "Address to listen on")
	ethEndpoint := flag.String("eth", "http://localhost:8545", "Ethereum RPC endpoint")
	contractAddr := flag.String("contract", "", "Contract address")
	privateKey := flag.String("key", "", "Ethereum private key (without 0x prefix)")
	flag.Parse()
	//
	//if *contractAddr == "" || *privateKey == "" {
	//	log.Fatal("Contract address and private key are required")
	//}

	// Parse the listen address
	addr, err := multiaddr.NewMultiaddr(*listenAddr)
	if err != nil {
		log.Fatalf("Invalid listen address: %s", err)
	}

	// Create the service context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create storage
	storage := NewStorage()

	// Create and start ETH service
	ethService, err := NewETHService(*ethEndpoint, *contractAddr, storage)
	if err != nil {
		log.Fatalf("Failed to create ETH service: %s", err)
	}

	// Start ETH service in background
	go ethService.Start(ctx, storage, 30*time.Second)

	// Create the P2P service
	service, err := NewP2PService(ctx, []multiaddr.Multiaddr{addr}, storage, *privateKey)
	if err != nil {
		log.Fatalf("Failed to create P2P service: %s", err)
	}

	// Start the P2P service
	if err := service.Start(); err != nil {
		log.Fatalf("Failed to start service: %s", err)
	}
	defer service.Stop()

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-sigCh
	fmt.Println("Shutting down...")
}
