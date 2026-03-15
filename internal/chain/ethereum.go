package chain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"golang.org/x/crypto/sha3"
)

// EthereumClient is a placeholder for Ethereum JSON-RPC interaction.
type EthereumClient struct {
	nodeURL string
}

func NewEthereumClient(nodeURL string) *EthereumClient {
	return &EthereumClient{nodeURL: nodeURL}
}

// GenerateAddress creates a new Ethereum address (simplified, no HD wallet).
func (c *EthereumClient) GenerateAddress() (address string, privateKey string, err error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate key: %w", err)
	}

	pubBytes := elliptic.MarshalCompressed(key.PublicKey.Curve, key.PublicKey.X, key.PublicKey.Y)
	hash := sha3.NewLegacyKeccak256()
	hash.Write(pubBytes[1:])
	addr := hash.Sum(nil)[12:]

	return "0x" + hex.EncodeToString(addr), hex.EncodeToString(key.D.Bytes()), nil
}

// ScanDeposits polls for new deposits. Placeholder.
func (c *EthereumClient) ScanDeposits() {
	log.Printf("ETH deposit scanner started (node: %s) — placeholder", c.nodeURL)
	// TODO: Poll blocks, check for transfers to deposit addresses
}

// SendTransaction submits a withdrawal transaction. Placeholder.
func (c *EthereumClient) SendTransaction(to string, amount string) (txHash string, err error) {
	log.Printf("ETH send %s to %s — placeholder", amount, to)
	return "0x" + hex.EncodeToString(make([]byte, 32)), nil
}
