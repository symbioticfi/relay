package main

// BroadcastSignature broadcasts a signature request to all peers
//func (s *SignerClient) BroadcastSignature(msgHash string, signature []byte, pubKey []byte) error {
//	// Create signature request
//	req := SignatureMessage{
//		MessageHash: msgHash,
//		Signature:   signature,
//		PublicKey:   pubKey,
//	}
//
//	data, err := json.Marshal(req)
//	if err != nil {
//		return fmt.Errorf("failed to marshal signature request: %w", err)
//	}
//
//	msg := entity.P2PMessage{
//		Type:      entity.TypeSignatureRequest,
//		Sender:    s.p2p.HostID().String(),
//		Timestamp: time.Now().Unix(),
//		Data:      data,
//	}
//
//	if err := s.p2p.Broadcast(msg); err != nil {
//		return fmt.Errorf("failed to broadcast signature request: %w", err)
//	}
//
//	log.Println("Broadcasted signature request to all peers")
//
//	return nil
//}
