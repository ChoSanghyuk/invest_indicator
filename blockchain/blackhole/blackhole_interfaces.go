package blackholedex

import (
	"crypto/ecdsa"
	"investindicator/blockchain/pkg/types"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// ContractClientInterface combines all contract interaction capabilities
type ContractClient interface {
	TxSender
	TxReader
	TxDecoder
	Abi() *abi.ABI
}

// TxSender defines methods for sending transactions to the blockchain
type TxSender interface {
	// Send executes a contract method with transaction
	Send(priority types.Priority, from *common.Address, privateKey *ecdsa.PrivateKey, method string, args ...interface{}) (common.Hash, error)

	// SendWithValue executes a contract method with transaction and native token value
	SendWithValue(priority types.Priority, value *big.Int, from *common.Address, privateKey *ecdsa.PrivateKey, method string, args ...interface{}) (common.Hash, error)
}

// TxReader defines methods for reading blockchain and contract state
type TxReader interface {
	// Call executes a read-only contract method (does not create transaction)
	Call(from *common.Address, method string, args ...interface{}) ([]interface{}, error)

	// GetReceipt retrieves transaction receipt by hash
	GetReceipt(txHash common.Hash) (*types.TxReceipt, error)

	// ParseReceipt parses events from transaction receipt
	ParseReceipt(receipt *types.TxReceipt) (string, error)

	// TransactionData retrieves raw transaction input data by hash
	TransactionData(hash common.Hash) ([]byte, error)

	// ContractAddress returns the contract address this client is bound to
	ContractAddress() *common.Address

	// ChainId returns the chain ID
	ChainId() *big.Int
}

// TxDecoder defines methods for decoding transaction data
type TxDecoder interface {
	// DecodeTransaction decodes raw transaction input data using the contract's ABI
	DecodeTransaction(data []byte) (*types.DecodedTransaction, error)

	// DecodeTransactionHex decodes hex-encoded transaction data
	DecodeTransactionHex(hexData string) (*types.DecodedTransaction, error)

	// DecodeByHash fetches a transaction by hash and decodes its input data
	DecodeByHash(txHash common.Hash) (*types.DecodedTransaction, error)
}

type TxListener interface {
	WaitForTransaction(txHash common.Hash) (*types.TxReceipt, error)
}
