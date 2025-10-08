package evmtxbroker

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EvmTxBroker struct {
	contractAddress common.Address
	abi             *abi.ABI
	client          *ethclient.Client
	chainId         *big.Int
}

/*

func (cm *EvmContractCodec) ChainId() (*big.Int, error) {
	chainID, err := cm.client.ChainID(context.Background())
	if err != nil {
		return nil, errors.Join(errors.New("ChainID Get Error"), err)
	}
	return chainID, nil
}
*/

func NewEvmTxBroker(client *ethclient.Client, contractAddress common.Address, abi *abi.ABI) *EvmTxBroker {
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil // todo.
	}

	return &EvmTxBroker{
		contractAddress: contractAddress,
		abi:             abi,
		client:          client,
		chainId:         chainID,
	}
}

func (cm *EvmTxBroker) Call(from *common.Address, method string, args ...interface{}) ([]interface{}, error) {

	if from == nil {
		from = &common.Address{}
	}
	packed, err := cm.abi.Pack(method, args...)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("%s Call 시, abi Pack Error", method), err)
	}

	raw, err := cm.client.CallContract(context.Background(), ethereum.CallMsg{
		From: *from,
		To:   &cm.contractAddress,
		Data: packed,
	}, nil)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("%s Call 시, CallContract Error", method), err)
	}

	rtn, err := cm.abi.Unpack(method, raw)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("%s Call 시, abi Unpack Error", method), err)
	}

	return rtn, nil
}

func (cm *EvmTxBroker) Send(priority Priority, fixedGasLimit *big.Int, from *common.Address, privateKey *ecdsa.PrivateKey, method string, args ...interface{}) (common.Hash, error) {
	return cm.send(priority, fixedGasLimit, nil, from, privateKey, method, args...)
}

func (cm *EvmTxBroker) SendWithValue(priority Priority, fixedGasLimit *big.Int, value *big.Int, from *common.Address, privateKey *ecdsa.PrivateKey, method string, args ...interface{}) (common.Hash, error) {
	return cm.send(priority, fixedGasLimit, value, from, privateKey, method, args...)
}

func (cm *EvmTxBroker) send(priority Priority, fixedGasLimit *big.Int, value *big.Int, from *common.Address, privateKey *ecdsa.PrivateKey, method string, args ...interface{}) (common.Hash, error) {
	if from == nil {
		from = &common.Address{}
	}
	packed, err := cm.abi.Pack(method, args...)
	if err != nil {
		return common.Hash{}, errors.Join(fmt.Errorf("%s Send 시, abi Pack Error", method), err)
	}

	// fmt.Println("packed :", common.Bytes2Hex(packed))

	nonce, err := cm.client.PendingNonceAt(context.Background(), *from)
	if err != nil {
		return common.Hash{}, errors.Join(fmt.Errorf("%s Send 시, PendingNonceAt Error", method), err)
	}

	// Get gas price and estimate gas limit
	gasPrice, err := cm.client.SuggestGasPrice(context.Background())
	if err != nil {
		return common.Hash{}, errors.Join(fmt.Errorf("%s Send 시, SuggestGasPrice Error", method), err)
	}

	gasLimit := uint64(0)
	// Estimate gas limit
	if fixedGasLimit == nil {
		gasLimit, err = cm.client.EstimateGas(context.Background(), ethereum.CallMsg{
			From:  *from,
			To:    &cm.contractAddress,
			Data:  packed,
			Value: nil, //big.NewInt(),
		})
		if err != nil {
			return common.Hash{}, errors.Join(fmt.Errorf("%s Send 시, EstimateGas Error", method), err)
		}
		if priority == High {
			gasLimit = gasLimit * 2
		}
	} else {
		gasLimit = fixedGasLimit.Uint64()
	}

	// Calculate gas tip cap (priority fee) - typically 1-2 Gwei
	gasTipCap := big.NewInt(1500000000) // 1.5 Gwei

	// Calculate gas fee cap (max fee per gas) - base fee + priority fee
	// For most networks, base fee + 2 Gwei is reasonable
	gasFeeCap := new(big.Int).Add(gasPrice, big.NewInt(2000000000)) // base fee + 2 Gwei
	// EIP-1559에서는 baseFee가 자동으로 소각(burn) => validator에게 별도로 주는 팁이 priorityFee(보통 2Gwei)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:    cm.chainId,
		Nonce:      nonce,
		GasTipCap:  gasTipCap, // a.k.a. maxPriorityFeePerGas
		GasFeeCap:  gasFeeCap, // a.k.a. maxFeePerGas
		Gas:        gasLimit,
		To:         &cm.contractAddress,
		Value:      value,
		Data:       packed,
		AccessList: nil, // Access list는 특정 컨트랙트를 호출할 때, 호출자가 접근할 컨트랙트의 주소 및 slot 키값들의 목록을 미리 저장
	})

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(cm.chainId), privateKey)
	if err != nil {
		return common.Hash{}, errors.Join(fmt.Errorf("%s Send 시, SignTx Error", method), err)
	}

	// Send transaction
	err = cm.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return common.Hash{}, errors.Join(fmt.Errorf("%s Send 시, SendTransaction Error", method), err)
	}

	return signedTx.Hash(), nil
}

func (cm *EvmTxBroker) unparseTxData(txData string, method string) error {

	// hex to bytes
	txDataBytes, err := hex.DecodeString(txData)
	if err != nil {
		return errors.Join(fmt.Errorf("txData 파싱 시, hex.DecodeString Error"), err)
	}

	unpack, err := cm.abi.Unpack(method, txDataBytes[4:])
	if err != nil {
		return errors.Join(fmt.Errorf("txData 파싱 시, abi Unpack Error"), err)
	}

	fmt.Println(unpack)

	return nil
}

func (cm *EvmTxBroker) TestSend(priority Priority, from *common.Address, privateKeyHex string, method string) (common.Hash, error) {
	if from == nil {
		from = &common.Address{}
	}

	nonce, err := cm.client.PendingNonceAt(context.Background(), *from)
	if err != nil {
		return common.Hash{}, errors.Join(fmt.Errorf("%s Send 시, PendingNonceAt Error", ""), err)
	}

	// Get gas price and estimate gas limit
	gasPrice, err := cm.client.SuggestGasPrice(context.Background())
	if err != nil {
		return common.Hash{}, errors.Join(fmt.Errorf("%s Send 시, SuggestGasPrice Error", ""), err)
	}

	// Estimate gas limit
	// gasLimit, err := cm.client.EstimateGas(context.Background(), ethereum.CallMsg{
	// 	From:  *from,
	// 	To:    &cm.contractAddress,
	// 	Data:  packed,
	// 	Value: nil, //big.NewInt(),
	// })
	// if err != nil {
	// 	return common.Hash{}, errors.Join(fmt.Errorf("%s Send 시, EstimateGas Error", method), err)
	// }

	packed := common.Hex2Bytes("3593564c000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000686c74a80000000000000000000000000000000000000000000000000000000000000003000604000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000018000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000f4240000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002bb97ef9ef8734c71904d8002f8b6bc66dd9c48a6e0001f4b31f66aa3c1e785363f0875a1b74e27b85fd66c70000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000060000000000000000000000000b31f66aa3c1e785363f0875a1b74e27b85fd66c70000000000000000000000001682f533c2359834167e5e4e108c1bfb69920e7800000000000000000000000000000000000000000000000000000000000000190000000000000000000000000000000000000000000000000000000000000060000000000000000000000000b31f66aa3c1e785363f0875a1b74e27b85fd66c7000000000000000000000000b4dd4fb3d4bced984cce972991fb100488b5922300000000000000000000000000000000000000000000000000c4e7233be3d9df0c")

	gasLimit := uint64(398130)

	if priority == High {
		gasLimit = gasLimit * 2
	}

	// Calculate gas tip cap (priority fee) - typically 1-2 Gwei
	gasTipCap := big.NewInt(1500000000) // 1.5 Gwei

	// Calculate gas fee cap (max fee per gas) - base fee + priority fee
	// For most networks, base fee + 2 Gwei is reasonable
	gasFeeCap := new(big.Int).Add(gasPrice, big.NewInt(2000000000)) // base fee + 2 Gwei

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:    cm.chainId,
		Nonce:      nonce,
		GasTipCap:  gasTipCap, // a.k.a. maxPriorityFeePerGas
		GasFeeCap:  gasFeeCap, // a.k.a. maxFeePerGas
		Gas:        gasLimit,
		To:         &cm.contractAddress,
		Value:      nil,
		Data:       packed,
		AccessList: nil, // Access list는 특정 컨트랙트를 호출할 때, 호출자가 접근할 컨트랙트의 주소 및 slot 키값들의 목록을 미리 저장
	})

	// Sign transaction
	privateKey, err := crypto.HexToECDSA(privateKeyHex[2:])
	if err != nil {
		return common.Hash{}, errors.Join(fmt.Errorf("%s Send 시, HexToECDSA Error", method), err)
	}

	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(cm.chainId), privateKey)
	if err != nil {
		return common.Hash{}, errors.Join(fmt.Errorf("%s Send 시, SignTx Error", method), err)
	}

	// Send transaction
	err = cm.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return common.Hash{}, errors.Join(fmt.Errorf("%s Send 시, SendTransaction Error", method), err)
	}

	return signedTx.Hash(), nil
}

type TxReceipt struct {
	// types.Receipt
	// RevertReason string `json:"revertReason"`
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

func (cm *EvmTxBroker) GetReceipt(txHash common.Hash) (*TxReceipt, error) {

	var r *TxReceipt

	err := cm.client.Client().CallContext(context.Background(), &r, "eth_getTransactionReceipt", txHash)
	if err == nil && r == nil {
		return nil, ethereum.NotFound
	}

	return r, nil
}

type EventInfo struct {
	Address   common.Address         `json:"address"`
	EventName string                 `json:"event"`
	Index     uint                   `json:"index"`
	Parameter map[string]interface{} `json:"parameter"`
}

func (cm *EvmTxBroker) ParseReceipt(receipt *TxReceipt) (string, error) {

	events := make([]*EventInfo, len(receipt.Logs))
	for i, log := range receipt.Logs {

		eventInfo := EventInfo{}
		events[i] = &eventInfo

		eventInfo.Address = log.Address
		eventInfo.Index = log.Index

		var abiEvent *abi.Event
		for _, event := range cm.abi.Events {
			fmt.Printf("event.ID.Hex(): %s | log.Topics[0].Hex(): %s\n", event.ID.Hex(), log.Topics[0].Hex())
			if event.ID.Hex() == log.Topics[0].Hex() {
				abiEvent = &event
				break
			}
		}
		if abiEvent == nil {
			continue
		}

		eventInfo.EventName = abiEvent.Name

		paramMap := make(map[string]interface{})
		eventInfo.Parameter = paramMap

		err := abiEvent.Inputs.UnpackIntoMap(paramMap, log.Data)
		if err != nil {
			return "", err
		}

		indexed := make([]abi.Argument, len(log.Topics)-1)
		idx := 0
		for _, input := range abiEvent.Inputs {
			if input.Indexed {
				indexed[idx] = input
				idx++
			}
		}

		err = abi.ParseTopicsIntoMap(paramMap, indexed, log.Topics[1:])
		if err != nil {
			return "", err
		}

		// []byte 일 때, string 변환 추가
		for i, input := range indexed {
			if input.Type.T == abi.FixedBytesTy || input.Type.T == abi.BytesTy {
				topic := log.Topics[i+1]
				paramMap[input.Name] = topic.Hex()
			}
		}

	}

	jsonData, err := json.Marshal(events)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func (cm *EvmTxBroker) ContractAddress() *common.Address {
	return &cm.contractAddress
}

func (cm *EvmTxBroker) ChainId() *big.Int {
	return cm.chainId
}
