package chain

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/sha3"
)

type EthereumClient struct {
	nodeURL    string
	httpClient *http.Client
}

func NewEthereumClient(nodeURL string) *EthereumClient {
	return &EthereumClient{
		nodeURL:    nodeURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type jsonRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *EthereumClient) call(ctx context.Context, method string, params ...interface{}) (json.RawMessage, error) {
	if params == nil {
		params = []interface{}{}
	}
	reqBody, _ := json.Marshal(jsonRPCRequest{JSONRPC: "2.0", Method: method, Params: params, ID: 1})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.nodeURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rpcResp jsonRPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("invalid JSON-RPC response: %w", err)
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}
	return rpcResp.Result, nil
}

// GetBlockNumber returns the latest block number from the node.
func (c *EthereumClient) GetBlockNumber(ctx context.Context) (uint64, error) {
	result, err := c.call(ctx, "eth_blockNumber")
	if err != nil {
		return 0, err
	}
	var hexStr string
	if err := json.Unmarshal(result, &hexStr); err != nil {
		return 0, err
	}
	n := new(big.Int)
	n.SetString(strings.TrimPrefix(hexStr, "0x"), 16)
	return n.Uint64(), nil
}

// Block represents an Ethereum block with its transactions.
type Block struct {
	Number       string        `json:"number"`
	Transactions []Transaction `json:"transactions"`
}

// Transaction represents an Ethereum transaction within a block.
type Transaction struct {
	Hash  string `json:"hash"`
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value"`
}

// GetBlockByNumber retrieves a block with full transaction objects.
func (c *EthereumClient) GetBlockByNumber(ctx context.Context, blockNum uint64) (*Block, error) {
	hexStr := fmt.Sprintf("0x%x", blockNum)
	result, err := c.call(ctx, "eth_getBlockByNumber", hexStr, true)
	if err != nil {
		return nil, err
	}
	if string(result) == "null" {
		return nil, fmt.Errorf("block %d not found", blockNum)
	}
	var block Block
	if err := json.Unmarshal(result, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

// ParseWeiToEther converts a hex wei value to an ether string with 18 decimals.
func ParseWeiToEther(hexValue string) string {
	wei := new(big.Int)
	wei.SetString(strings.TrimPrefix(hexValue, "0x"), 16)
	if wei.Sign() == 0 {
		return "0"
	}
	ether := new(big.Float).SetInt(wei)
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	ether.Quo(ether, divisor)
	return ether.Text('f', 18)
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

// SendTransaction submits a withdrawal transaction. Placeholder for now.
func (c *EthereumClient) SendTransaction(to string, amount string) (txHash string, err error) {
	log.Printf("ETH send %s to %s — placeholder (requires signing key)", amount, to)
	return "0x0000000000000000000000000000000000000000000000000000000000000000", nil
}
