package net

import (
	"crypto/rand"
)

var PeerID []byte

func GeneratePeerID() ([]byte, error) {
	peerID := make([]byte, 20)
	copy(peerID[:8], []byte("-GT0001-"))
	_, err := rand.Read(peerID[8:])
	if err != nil {
		return nil, err
	}
	return peerID, nil
}
