package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// DecodedParam represents a single decoded parameter from transaction data
type DecodedParam struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// DecodedTransaction represents a fully decoded transaction
type DecodedTransaction struct {
	ContractAddress common.Address `json:"contract"`
	MethodName      string         `json:"method"`
	MethodSignature string         `json:"signature"`
	Parameters      []DecodedParam `json:"parameters"`
	RawData         []byte         `json:"rawData,omitempty"`
}

// TxReceipt represents a transaction receipt with additional fields
type TxReceipt struct {
	BlockHash         common.Hash  `json:"blockHash"`
	BlockNumber       string       `json:"blockNumber"`
	ContractAddress   string       `json:"contractAddress"`
	CumulativeGasUsed string       `json:"cumulativeGasUsed"`
	EffectiveGasPrice string       `json:"effectiveGasPrice"`
	From              string       `json:"from"`
	GasUsed           string       `json:"gasUsed"`
	Logs              []*types.Log `json:"logs"`
	Bloom             types.Bloom  `json:"logsBloom"`
	RevertReason      string       `json:"revertReason"`
	Status            string       `json:"status"`
	To                string       `json:"to"`
	TxHash            common.Hash  `json:"transactionHash" gencodec:"required"`
	TransactionIndex  string       `json:"transactionIndex"`
	Type              string       `json:"type"`
}

// EventInfo represents a parsed event from transaction receipt
type EventInfo struct {
	Address   common.Address         `json:"address"`
	EventName string                 `json:"event"`
	Index     uint                   `json:"index"`
	Parameter map[string]interface{} `json:"parameter"`
}
