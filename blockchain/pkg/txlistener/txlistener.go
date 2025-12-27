package txlistener

import (
	"context"
	"errors"
	"fmt"
	"time"

	contracttypes "investindicator/blockchain/pkg/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	// ErrTimeout is returned when the transaction is not mined within the timeout period
	ErrTimeout = errors.New("transaction receipt timeout")

	// ErrTransactionFailed is returned when the transaction status is 0 (failed)
	ErrTransactionFailed = errors.New("transaction failed")
)

// TxListener waits for transactions to be mined on the blockchain
type TxListener struct {
	client       *ethclient.Client
	PollInterval time.Duration
	Timeout      time.Duration
}

// Option is a functional option for configuring TxListener
type Option func(*TxListener)

// WithPollInterval sets the polling interval for checking transaction receipts
func WithPollInterval(interval time.Duration) Option {
	return func(tl *TxListener) {
		tl.PollInterval = interval
	}
}

// WithTimeout sets the maximum time to wait for transaction confirmation
func WithTimeout(timeout time.Duration) Option {
	return func(tl *TxListener) {
		tl.Timeout = timeout
	}
}

// NewTxListener creates a new transaction listener with the given client and options
// Default configuration: 2s poll interval, 5min timeout
func NewTxListener(client *ethclient.Client, opts ...Option) *TxListener {
	tl := &TxListener{
		client:       client,
		PollInterval: 2 * time.Second, // Default 2s poll interval
		Timeout:      5 * time.Minute, // Default 5min poll interval
	}

	for _, opt := range opts {
		opt(tl)
	}
	return tl
}

// WaitForTransaction waits for a transaction to be mined and returns its receipt
// Uses the configured poll interval and timeout from the TxListener instance
func (tl *TxListener) WaitForTransaction(txHash common.Hash) (*contracttypes.TxReceipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), tl.Timeout)
	defer cancel()

	ticker := time.NewTicker(tl.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: transaction %s not mined within %v", ErrTimeout, txHash.Hex(), tl.Timeout)

		case <-ticker.C:
			receipt, err := tl.getReceipt(txHash)
			if err != nil {
				// If receipt not found, continue polling
				if errors.Is(err, ethereum.NotFound) {
					continue
				}
				// Other errors should be returned
				return nil, fmt.Errorf("failed to get receipt for transaction %s: %w", txHash.Hex(), err)
			}

			// Receipt found - check if transaction was successful
			if receipt.Status == "0x0" {
				return receipt, fmt.Errorf("%w: transaction %s status is 0x0", ErrTransactionFailed, txHash.Hex())
			}

			return receipt, nil
		}
	}
}

// getReceipt retrieves the transaction receipt from the blockchain
func (tl *TxListener) getReceipt(txHash common.Hash) (*contracttypes.TxReceipt, error) {
	var receipt *contracttypes.TxReceipt

	err := tl.client.Client().CallContext(context.Background(), &receipt, "eth_getTransactionReceipt", txHash)
	if err == nil && receipt == nil {
		return nil, ethereum.NotFound
	}

	return receipt, err
}
