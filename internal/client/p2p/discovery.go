package p2p

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

const mdnsServiceTag = "symbnet-p2p-messaging"

type peerAddable interface {
	AddPeer(peerID peer.AddrInfo) error
}

type DiscoveryNorifee struct {
	ctx         context.Context
	host        host.Host
	discovery   mdns.Service
	peerAddable peerAddable
}

func NewDiscoveryService(ctx context.Context, pa peerAddable, h host.Host) (*DiscoveryNorifee, error) {
	dn := &DiscoveryNorifee{
		ctx:         ctx,
		host:        h,
		peerAddable: pa,
	}
	dn.discovery = mdns.NewMdnsService(h, mdnsServiceTag, dn)
	return dn, nil
}

func (dn *DiscoveryNorifee) Start() error {
	if err := dn.discovery.Start(); err != nil {
		return errors.Errorf("failed to start discovery: %w", err)
	}
	slog.InfoContext(dn.ctx, "discovery started", "localPeerID", dn.host.ID(), "localAddrs", dn.host.Addrs())
	return nil
}

func (dn *DiscoveryNorifee) Close() error {
	if err := dn.discovery.Close(); err != nil {
		return errors.Errorf("failed to close discovery: %w", err)
	}
	return nil
}

func (dn *DiscoveryNorifee) HandlePeerFound(pi peer.AddrInfo) {
	err := dn.peerAddable.AddPeer(pi)
	slog.InfoContext(dn.ctx, "peer found", "peer", pi.ID)

	if err != nil {
		slog.ErrorContext(dn.ctx, "failed to add peer", "peer", pi.ID, "err", err)
		return
	}

	slog.InfoContext(dn.ctx, "peer added", "peer", pi.ID)
}
