package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

const (
	protocolID     = "/p2p/example/1.0.0"
	mdnsServiceTag = "example-mdns"
)

type Message struct {
	Text string `json:"text"`
}

func main() {
	ctx := context.Background()

	// 1. Создаем новый p2p-хост
	host, err := libp2p.New()
	if err != nil {
		log.Fatalf("Failed to create libp2p host: %v", err)
	}

	// 2. Устанавливаем обработчик сообщений
	host.SetStreamHandler(protocol.ID(protocolID), handleStream)

	// 3. Запускаем mDNS для обнаружения пиров
	service := mdns.NewMdnsService(host, mdnsServiceTag, &discoveryNotifee{Host: host})
	defer service.Close()
	service.Start()

	log.Printf("Host ID: %s", host.ID().String())
	for _, addr := range host.Addrs() {
		log.Printf("Listening on: %s/p2p/%s", addr, host.ID().String())
	}

	// 4. Отправляем сообщение вручную через консоль
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()

		for peerID := range connectedPeers {
			err := sendMessage(ctx, host, peerID, Message{Text: text})
			if err != nil {
				log.Printf("Failed to send message to %s: %v", peerID.String(), err)
			}
		}
	}
}

// Глобальный список подключенных пиров
var connectedPeers = make(map[peer.ID]struct{})

type discoveryNotifee struct {
	Host host.Host
}

// Вызывается, когда найден пир через mDNS
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if pi.ID == n.Host.ID() {
		return // не подключаться к себе
	}

	log.Printf("Discovered peer: %s", pi.ID.String())
	err := n.Host.Connect(ctx, pi)
	if err != nil {
		log.Printf("Failed to connect to peer: %v", err)
		return
	}

	log.Printf("Connected to peer: %s", pi.ID.String())
	connectedPeers[pi.ID] = struct{}{}
}

// Обработчик входящего потока
func handleStream(stream network.Stream) {
	defer stream.Close()

	reader := bufio.NewReader(stream)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		log.Printf("Failed to read from stream: %v", err)
		return
	}

	var msg Message
	err = json.Unmarshal(data, &msg)
	if err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	log.Printf("Received message: %s", msg.Text)
}

// Отправить сообщение на конкретного пира
func sendMessage(ctx context.Context, h host.Host, peerID peer.ID, msg Message) error {
	stream, err := h.NewStream(ctx, peerID, protocol.ID(protocolID))
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	data = append(data, '\n')

	_, err = stream.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to stream: %w", err)
	}

	return nil
}
