package storage

import (
	"sync"
)

// actually RAM storage
type Storage struct {
	mutex sync.RWMutex
	// messageHash -> signatures
	signatures map[string]map[string]bool
}

func NewStorage() *Storage {
	return &Storage{
		signatures: make(map[string]map[string]bool),
	}
}

func (s *Storage) AddSignature(messageHash string, pubKey []byte, sig []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.signatures[messageHash][string(pubKey)+string(sig)] = true
}

func (s *Storage) GetSignatures(messageHash string) (pubKeys [][]byte, sigs [][]byte) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	pubKeys = make([][]byte, 0, len(s.signatures[messageHash]))
	sigs = make([][]byte, 0, len(s.signatures[messageHash]))

	for sigKey := range s.signatures[messageHash] {
		pubKeyLen := len(sigKey) / 2
		pubKey := []byte(sigKey[:pubKeyLen])
		sig := []byte(sigKey[pubKeyLen:])
		pubKeys = append(pubKeys, pubKey)
		sigs = append(sigs, sig)
	}

	return pubKeys, sigs
}
