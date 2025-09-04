package walletSDK

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

// Client represents a quantum blockchain client
type Client struct {
	endpoint   string
	httpClient *http.Client
}

// NewClient creates a new blockchain client
func NewClient(endpoint string) *Client {
	return &Client{
		endpoint:   endpoint,
		httpClient: &http.Client{},
	}
}

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      int         `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
	ID      int             `json:"id"`
}

// RPCError represents an RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// call makes a JSON-RPC call
func (c *Client) call(method string, params interface{}) (json.RawMessage, error) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.endpoint, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var rpcResp JSONRPCResponse
	err = json.Unmarshal(body, &rpcResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

// GetChainID returns the chain ID
func (c *Client) GetChainID() (*big.Int, error) {
	result, err := c.call("eth_chainId", nil)
	if err != nil {
		return nil, err
	}

	var chainIDStr string
	err = json.Unmarshal(result, &chainIDStr)
	if err != nil {
		return nil, err
	}

	chainID := new(big.Int)
	chainID.SetString(chainIDStr[2:], 16) // Remove 0x prefix
	return chainID, nil
}

// GetBlockNumber returns the current block number
func (c *Client) GetBlockNumber() (*big.Int, error) {
	result, err := c.call("eth_blockNumber", nil)
	if err != nil {
		return nil, err
	}

	var blockNumStr string
	err = json.Unmarshal(result, &blockNumStr)
	if err != nil {
		return nil, err
	}

	blockNum := new(big.Int)
	blockNum.SetString(blockNumStr[2:], 16) // Remove 0x prefix
	return blockNum, nil
}

// GetBalance returns the balance of an address
func (c *Client) GetBalance(address types.Address) (*big.Int, error) {
	params := []interface{}{address.Hex(), "latest"}
	result, err := c.call("eth_getBalance", params)
	if err != nil {
		return nil, err
	}

	var balanceStr string
	err = json.Unmarshal(result, &balanceStr)
	if err != nil {
		return nil, err
	}

	balance := new(big.Int)
	balance.SetString(balanceStr[2:], 16) // Remove 0x prefix
	return balance, nil
}

// GetNonce returns the transaction count (nonce) for an address
func (c *Client) GetNonce(address types.Address) (uint64, error) {
	params := []interface{}{address.Hex(), "latest"}
	result, err := c.call("eth_getTransactionCount", params)
	if err != nil {
		return 0, err
	}

	var nonceStr string
	err = json.Unmarshal(result, &nonceStr)
	if err != nil {
		return 0, err
	}

	nonce := new(big.Int)
	nonce.SetString(nonceStr[2:], 16) // Remove 0x prefix
	return nonce.Uint64(), nil
}

// GetBlock returns a block by number
func (c *Client) GetBlock(blockNumber *big.Int) (*types.Block, error) {
	params := []interface{}{fmt.Sprintf("0x%x", blockNumber), true}
	result, err := c.call("eth_getBlockByNumber", params)
	if err != nil {
		return nil, err
	}

	var block types.Block
	err = json.Unmarshal(result, &block)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

// GetBlockByHash returns a block by hash
func (c *Client) GetBlockByHash(hash types.Hash) (*types.Block, error) {
	params := []interface{}{hash.Hex(), true}
	result, err := c.call("eth_getBlockByHash", params)
	if err != nil {
		return nil, err
	}

	var block types.Block
	err = json.Unmarshal(result, &block)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

// GetGasPrice returns the current gas price
func (c *Client) GetGasPrice() (*big.Int, error) {
	result, err := c.call("eth_gasPrice", nil)
	if err != nil {
		return nil, err
	}

	var gasPriceStr string
	err = json.Unmarshal(result, &gasPriceStr)
	if err != nil {
		return nil, err
	}

	gasPrice := new(big.Int)
	gasPrice.SetString(gasPriceStr[2:], 16) // Remove 0x prefix
	return gasPrice, nil
}

// EstimateGas estimates gas for a transaction
func (c *Client) EstimateGas(from, to types.Address, value *big.Int, data []byte) (uint64, error) {
	params := []interface{}{
		map[string]interface{}{
			"from":  from.Hex(),
			"to":    to.Hex(),
			"value": fmt.Sprintf("0x%x", value),
			"data":  fmt.Sprintf("0x%x", data),
		},
	}

	result, err := c.call("eth_estimateGas", params)
	if err != nil {
		return 0, err
	}

	var gasStr string
	err = json.Unmarshal(result, &gasStr)
	if err != nil {
		return 0, err
	}

	gas := new(big.Int)
	gas.SetString(gasStr[2:], 16) // Remove 0x prefix
	return gas.Uint64(), nil
}

// GetSupportedAlgorithms returns supported quantum algorithms
func (c *Client) GetSupportedAlgorithms() (map[string]interface{}, error) {
	result, err := c.call("quantum_getSupportedAlgorithms", nil)
	if err != nil {
		return nil, err
	}

	var algorithms map[string]interface{}
	err = json.Unmarshal(result, &algorithms)
	if err != nil {
		return nil, err
	}

	return algorithms, nil
}

// Wallet represents a quantum-resistant wallet
type Wallet struct {
	address    types.Address
	privateKey []byte
	algorithm  crypto.SignatureAlgorithm
	client     *Client
}

// NewWallet creates a new quantum wallet
func NewWallet(algorithm crypto.SignatureAlgorithm, client *Client) (*Wallet, error) {
	var privateKey []byte
	var publicKey []byte

	switch algorithm {
	case crypto.SigAlgDilithium:
		priv, pub, genErr := crypto.GenerateDilithiumKeyPair()
		if genErr != nil {
			return nil, genErr
		}
		privateKey = priv.Bytes()
		publicKey = pub.Bytes()

	case crypto.SigAlgFalcon:
		priv, pub, genErr := crypto.GenerateFalconKeyPair()
		if genErr != nil {
			return nil, genErr
		}
		privateKey = priv.Bytes()
		publicKey = pub.Bytes()

	default:
		return nil, fmt.Errorf("unsupported algorithm: %v", algorithm)
	}

	address := types.PublicKeyToAddress(publicKey)

	return &Wallet{
		address:    address,
		privateKey: privateKey,
		algorithm:  algorithm,
		client:     client,
	}, nil
}

// LoadWallet loads a wallet from private key bytes
func LoadWallet(privateKey []byte, algorithm crypto.SignatureAlgorithm, client *Client) (*Wallet, error) {
	var publicKey []byte

	switch algorithm {
	case crypto.SigAlgDilithium:
		priv, err := crypto.DilithiumPrivateKeyFromBytes(privateKey)
		if err != nil {
			return nil, err
		}
		// Extract public key from private key
		pubKeyData := priv.Bytes()[:crypto.DilithiumPublicKeySize]
		publicKey = pubKeyData

	case crypto.SigAlgFalcon:
		priv, err := crypto.FalconPrivateKeyFromBytes(privateKey)
		if err != nil {
			return nil, err
		}
		// Extract public key from private key
		pubKeyData := priv.Bytes()[:crypto.FalconPublicKeySize]
		publicKey = pubKeyData

	default:
		return nil, fmt.Errorf("unsupported algorithm: %v", algorithm)
	}

	address := types.PublicKeyToAddress(publicKey)

	return &Wallet{
		address:    address,
		privateKey: privateKey,
		algorithm:  algorithm,
		client:     client,
	}, nil
}

// GetAddress returns the wallet address
func (w *Wallet) GetAddress() types.Address {
	return w.address
}

// GetBalance returns the wallet balance
func (w *Wallet) GetBalance() (*big.Int, error) {
	return w.client.GetBalance(w.address)
}

// GetNonce returns the wallet nonce
func (w *Wallet) GetNonce() (uint64, error) {
	return w.client.GetNonce(w.address)
}

// CreateTransaction creates a new quantum transaction
func (w *Wallet) CreateTransaction(to *types.Address, value *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) (*types.QuantumTransaction, error) {
	chainID, err := w.client.GetChainID()
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	nonce, err := w.GetNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	tx := types.NewQuantumTransaction(chainID, nonce, to, value, gasLimit, gasPrice, data)

	// Sign the transaction
	err = tx.SignTransaction(w.privateKey, w.algorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}

// SendTransaction creates and sends a transaction
func (w *Wallet) SendTransaction(to types.Address, value *big.Int, data []byte) (types.Hash, error) {
	gasPrice, err := w.client.GetGasPrice()
	if err != nil {
		return types.ZeroHash, fmt.Errorf("failed to get gas price: %w", err)
	}

	gasLimit, err := w.client.EstimateGas(w.address, to, value, data)
	if err != nil {
		gasLimit = 21000 // Default gas limit
	}

	tx, err := w.CreateTransaction(&to, value, gasLimit, gasPrice, data)
	if err != nil {
		return types.ZeroHash, err
	}

	// TODO: Implement raw transaction sending
	// For now, return the transaction hash
	return tx.Hash(), nil
}

// Transfer sends a simple value transfer
func (w *Wallet) Transfer(to types.Address, amount *big.Int) (types.Hash, error) {
	return w.SendTransaction(to, amount, []byte{})
}

// GetPrivateKey returns the private key (use with caution)
func (w *Wallet) GetPrivateKey() []byte {
	return w.privateKey
}

// GetAlgorithm returns the signature algorithm
func (w *Wallet) GetAlgorithm() crypto.SignatureAlgorithm {
	return w.algorithm
}

// ExportPrivateKey exports the private key as hex string
func (w *Wallet) ExportPrivateKey() string {
	return fmt.Sprintf("0x%x", w.privateKey)
}

// ImportPrivateKey imports a private key from hex string
func ImportPrivateKey(privateKeyHex string, algorithm crypto.SignatureAlgorithm, client *Client) (*Wallet, error) {
	if len(privateKeyHex) > 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKey, err := types.HexToBytes(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key hex: %w", err)
	}

	return LoadWallet(privateKey, algorithm, client)
}

// Helper function to convert hex to bytes
func hexToBytes(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, fmt.Errorf("hex string has odd length")
	}

	result := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		b, err := parseHexByte(s[i : i+2])
		if err != nil {
			return nil, err
		}
		result[i/2] = b
	}

	return result, nil
}

func parseHexByte(s string) (byte, error) {
	var result byte
	for i, c := range s {
		var val byte
		if c >= '0' && c <= '9' {
			val = byte(c - '0')
		} else if c >= 'a' && c <= 'f' {
			val = byte(c - 'a' + 10)
		} else if c >= 'A' && c <= 'F' {
			val = byte(c - 'A' + 10)
		} else {
			return 0, fmt.Errorf("invalid hex character: %c", c)
		}

		if i == 0 {
			result = val << 4
		} else {
			result |= val
		}
	}

	return result, nil
}

// Add HexToBytes function to types package
func init() {
	types.HexToBytes = hexToBytes
}
