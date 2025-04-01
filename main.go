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

	"offchain-middleware/bls"
	"offchain-middleware/eth"
	"offchain-middleware/network"
	"offchain-middleware/p2p"
	"offchain-middleware/storage"

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
	storage := storage.NewStorage()

	ethClient, err := eth.NewEthClient(*ethEndpoint, *contractAddr)
	if err != nil {
		log.Fatalf("Failed to create ETH service: %s", err)
	}

	// Create the P2P service
	p2pService, err := p2p.NewP2PService(ctx, []multiaddr.Multiaddr{addr}, storage)
	if err != nil {
		log.Fatalf("Failed to create P2P service: %s", err)
	}

	// Start the P2P service
	if err := p2pService.Start(); err != nil {
		log.Fatalf("Failed to start service: %s", err)
	}
	defer p2pService.Stop()

	keyPair, err := bls.GenerateKeyOrLoad(*privateKey)
	if err != nil {
		log.Fatalf("Failed to create key pair: %s", err)
	}

	networkService, err := network.NewNetworkService(p2pService, ethClient, storage, keyPair)
	if err != nil {
		log.Fatalf("Failed to create network service: %s", err)
	}

	if err := networkService.Start(time.Minute); err != nil {
		log.Fatalf("Failed to start network service: %s", err)
	}

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-sigCh
	fmt.Println("Shutting down...")
}
