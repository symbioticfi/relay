package storage

import (
	"log"
	"math/big"
	"sync"
)

// actually RAM storage
type Storage struct {
	mutex sync.RWMutex
	// epoch -> messageHash -> signatures
	signatures map[*big.Int]map[string][]Signature
	hasSigned  map[*big.Int]map[string]bool
}

type Signature struct {
	Signature []byte
	PublicKey []byte
}

func NewStorage() *Storage {
	return &Storage{
		signatures: make(map[*big.Int]map[string][]Signature),
		hasSigned:  make(map[*big.Int]map[string]bool),
	}
}

func (s *Storage) AddSignature(epoch *big.Int, messageHash string, signature Signature) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// Initialize maps for this epoch if they don't exist
	if _, ok := s.signatures[epoch]; !ok {
		s.signatures[epoch] = make(map[string][]Signature)
	}
	if _, ok := s.hasSigned[epoch]; !ok {
		s.hasSigned[epoch] = make(map[string]bool)
	}

	// Check if this public key has already signed this message in this epoch
	pubKeyStr := string(signature.PublicKey)
	if s.hasSigned[epoch][pubKeyStr] {
		// Log or handle the double signing attempt
		// This could be expanded to include more detailed logging or alerting
		log.Println("Warning: Double signing attempt detected for epoch", epoch.String(), "by public key", pubKeyStr)
	}

	s.signatures[epoch][messageHash] = append(s.signatures[epoch][messageHash], signature)
	s.hasSigned[epoch][string(signature.PublicKey)] = true
}

func (s *Storage) GetSignatures(epoch *big.Int) map[string][]Signature {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Check if the epoch exists in the map
	if epochSignatures, ok := s.signatures[epoch]; ok {
		return epochSignatures
	}

	// Return empty map if epoch doesn't exist
	return make(map[string][]Signature)
}
