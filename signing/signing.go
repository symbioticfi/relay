package signing

import (
	"middleware-offchain/bls"
)

// Signing coordinates the P2P and ETH services
type Signing struct {
	keyPair *bls.KeyPair
}

// NewSigner creates a new signer service
func NewSigning(keyPair *bls.KeyPair) (*Signing, error) {
	return &Signing{
		keyPair: keyPair,
	}, nil
}

// SignMessage signs a message
//func (n Signing) SignMessage(msg []byte) (pubKey []byte, signatureBytes []byte, msgHash []byte, err error) {
//	msgHash = crypto.Keccak256(msg)
//
//	signature, err := n.keyPair.Sign(msgHash)
//	if err != nil {
//		return nil, nil, nil, fmt.Errorf("failed to sign message: %w", err)
//	}
//
//	return n.keyPair.PublicKeyG1.Marshal(), signature.Marshal(), msgHash, nil
//}
