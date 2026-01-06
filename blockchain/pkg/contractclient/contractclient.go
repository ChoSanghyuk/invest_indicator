package contractclient

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	contracttypes "investindicator/blockchain/pkg/types"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ContractClient struct {
	contractAddress common.Address
	abi             *abi.ABI
	client          *ethclient.Client
	chainId         *big.Int
	defaultGasLimit *big.Int
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

func NewContractClient(client *ethclient.Client, contractAddress common.Address, abi *abi.ABI, opts ...Option) *ContractClient {
	chainID := big.NewInt(0)
	if client != nil {
		cid, err := client.ChainID(context.Background())
		if err != nil {
			// todo. logging
		}
		chainID = cid
	}

	cc := &ContractClient{
		contractAddress: contractAddress,
		abi:             abi,
		client:          client,
		chainId:         chainID,
	}

	for _, opt := range opts {
		opt(cc)
	}

	return cc
}

// Option is a functional option for configuring ContractClient
type Option func(*ContractClient)

func WithDefaultGasLimit(gasLimit *big.Int) Option {
	return func(cc *ContractClient) {
		cc.defaultGasLimit = gasLimit
	}
}

func (cm *ContractClient) Call(from *common.Address, method string, args ...interface{}) ([]interface{}, error) {

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

func (cm *ContractClient) Send(priority contracttypes.Priority, from *common.Address, privateKey *ecdsa.PrivateKey, method string, args ...interface{}) (common.Hash, error) {
	return cm.send(priority, nil, from, privateKey, method, args...)
}

func (cm *ContractClient) SendWithValue(priority contracttypes.Priority, value *big.Int, from *common.Address, privateKey *ecdsa.PrivateKey, method string, args ...interface{}) (common.Hash, error) {
	return cm.send(priority, value, from, privateKey, method, args...)
}

func (cm *ContractClient) send(priority contracttypes.Priority, value *big.Int, from *common.Address, privateKey *ecdsa.PrivateKey, method string, args ...interface{}) (common.Hash, error) {
	if from == nil {
		from = &common.Address{}
	}
	packed, err := cm.abi.Pack(method, args...)
	if err != nil {
		return common.Hash{}, errors.Join(fmt.Errorf("%s Send 시, abi Pack Error", method), err)
	}

	fmt.Println("packed :", common.Bytes2Hex(packed))

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
	gasLimit, err = cm.client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:  *from,
		To:    &cm.contractAddress,
		Data:  packed,
		Value: nil, //big.NewInt(),
	})
	if err != nil {
		if cm.defaultGasLimit != nil {
			gasLimit = cm.defaultGasLimit.Uint64()
		} else {
			return common.Hash{}, errors.Join(fmt.Errorf("%s Send 시, EstimateGas Error", method), err)
		}
	}
	if priority == contracttypes.High {
		gasLimit = gasLimit * 2
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

func (cm *ContractClient) unparseTxData(txData string, method string) error {

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

func (cm *ContractClient) TestSend(priority contracttypes.Priority, from *common.Address, privateKeyHex string, method string) (common.Hash, error) {
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

	packed := common.Hex2Bytes("3593564c000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000686c74a80000000000000000000000000000000000000000000000000000000000000003000604000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000018000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000f4240000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002bb97ef9ef8734c71904d8002f8b6bc66dd9c48a6e0001f4b31f66aa3c1e785363f0875a1b74e27b85fd66c70000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000060000000000000000000000000b31f66aa3c1e785363f0875a1b74e27b85fd66c70000000000000000000000001682f533c2359834167e5e4e108c1bfb69920e7800000000000000000000000000000000000000000000000000000000000000190000000000000000000000000000000000000000000000000000000000000060000000000000000000000000b31f66aa3c1e785363f0875a1b74e27b85fd66c7000000000000000000000000b4dd4fb3d4bced984cce972991fb100488b5922300000000000000000000000000000000000000000000000000c4e7233be3d9df0c")

	gasLimit := uint64(398130)

	if priority == contracttypes.High {
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

func (cm *ContractClient) GetReceipt(txHash common.Hash) (*contracttypes.TxReceipt, error) {

	var r *contracttypes.TxReceipt

	err := cm.client.Client().CallContext(context.Background(), &r, "eth_getTransactionReceipt", txHash)
	if err == nil && r == nil {
		return nil, ethereum.NotFound
	}

	return r, nil
}

func (cm *ContractClient) ParseReceipt(receipt *contracttypes.TxReceipt) (string, error) {

	events := make([]*contracttypes.EventInfo, len(receipt.Logs))
	for i, log := range receipt.Logs {

		eventInfo := contracttypes.EventInfo{}
		events[i] = &eventInfo

		if log.Address != cm.contractAddress {
			continue // 내 컨트랙트에서 발생한 것 아니면 패쓰하기
		}
		eventInfo.Address = log.Address
		eventInfo.Index = log.Index

		var abiEvent *abi.Event
		for _, event := range cm.abi.Events {
			// fmt.Printf("event.ID.Hex(): %s | log.Topics[0].Hex(): %s\n", event.ID.Hex(), log.Topics[0].Hex())
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
			// memo. 자기 자신의 receipt일 경우에는 idx < len(indexed) 필요 없음. log.Topics이 시그니처 + indexed params로 구성되기 때문.
			// 다만, 컨트랙트 내부에서 다른 컨트랙트 호출되어서 찍히는 로그는 제대로 파싱을 못하기에 여기서 오류가 생김
			if input.Indexed && idx < len(indexed) {
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

func (cm *ContractClient) ContractAddress() *common.Address {
	return &cm.contractAddress
}

func (cm *ContractClient) ChainId() *big.Int {
	return cm.chainId
}

func (cm *ContractClient) Abi() *abi.ABI {
	return cm.abi
}

// PrintFunctionSelectors prints a mapping of function selectors (method IDs) to function names
// from the contract's ABI. This is useful for debugging and understanding contract interfaces.
func (cm *ContractClient) PrintFunctionSelectors() {
	fmt.Println("=== Function Selector Mapping ===")
	fmt.Printf("Contract Address: %s\n\n", cm.contractAddress.Hex())
	fmt.Printf("%-12s %-30s %s\n", "Selector", "Function Name", "Signature")
	fmt.Println(strings.Repeat("-", 80))

	// Iterate through all methods in the ABI
	for name, method := range cm.abi.Methods {
		// Get the 4-byte selector (method ID)
		selector := hex.EncodeToString(method.ID)

		// Build the full signature
		signature := buildMethodSignature(&method)

		// Print the mapping
		fmt.Printf("0x%-10s %-30s %s\n", selector, name, signature)
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Total functions: %d\n", len(cm.abi.Methods))
}

// GetFunctionSelectors returns a map of function selectors to function information
// Key: selector hex string (e.g., "a9059cbb")
// Value: map with "name" and "signature" keys
func (cm *ContractClient) GetFunctionSelectors() map[string]map[string]string {
	selectors := make(map[string]map[string]string)

	for name, method := range cm.abi.Methods {
		selector := hex.EncodeToString(method.ID)
		signature := buildMethodSignature(&method)

		selectors[selector] = map[string]string{
			"name":      name,
			"signature": signature,
		}
	}

	return selectors
}

func (cm *ContractClient) TransactionData(hash common.Hash) ([]byte, error) {
	tx, _, err := cm.client.TransactionByHash(context.Background(), hash)
	if err != nil {
		return nil, err
	}

	return tx.Data(), nil
}

// DecodeTransaction decodes raw transaction input data using the contract's ABI
func (cm *ContractClient) DecodeTransaction(data []byte) (*contracttypes.DecodedTransaction, error) {
	if len(data) < 4 {
		return nil, errors.New("transaction data too short: must be at least 4 bytes for method selector")
	}

	// Extract method selector (first 4 bytes)
	methodSelector := data[:4]

	// Find the method by selector
	method, err := cm.abi.MethodById(methodSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to find method by selector %s: %w", hex.EncodeToString(methodSelector), err)
	}

	// Unpack the arguments
	args, err := method.Inputs.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack arguments for method %s: %w", method.Name, err)
	}

	// Build decoded parameters
	params := make([]contracttypes.DecodedParam, len(method.Inputs))
	for i, input := range method.Inputs {
		value := args[i]

		// Convert special types for better JSON representation
		value = convertValueForJSON(value, input.Type)

		params[i] = contracttypes.DecodedParam{
			Name:  input.Name,
			Type:  input.Type.String(),
			Value: value,
		}
	}

	// Build method signature
	signature := buildMethodSignature(method)

	return &contracttypes.DecodedTransaction{
		ContractAddress: cm.contractAddress,
		MethodName:      method.Name,
		MethodSignature: signature,
		Parameters:      params,
		RawData:         data,
	}, nil
}

// DecodeTransactionHex decodes hex-encoded transaction data
func (cm *ContractClient) DecodeTransactionHex(hexData string) (*contracttypes.DecodedTransaction, error) {
	// Remove 0x prefix if present
	if len(hexData) >= 2 && hexData[:2] == "0x" {
		hexData = hexData[2:]
	}

	data, err := hex.DecodeString(hexData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex data: %w", err)
	}

	return cm.DecodeTransaction(data)
}

// DecodeByHash fetches a transaction by hash and decodes its input data
func (cm *ContractClient) DecodeByHash(txHash common.Hash) (*contracttypes.DecodedTransaction, error) {
	tx, _, err := cm.client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction %s: %w", txHash.Hex(), err)
	}

	return cm.DecodeTransaction(tx.Data())
}

/*********************************** internal utils *********************************************/

// buildMethodSignature constructs the full method signature string
func buildMethodSignature(method *abi.Method) string {
	var inputs []string
	for _, input := range method.Inputs {
		inputs = append(inputs, input.Type.String())
	}
	return fmt.Sprintf("%s(%s)", method.Name, joinStrings(inputs, ","))
}

// joinStrings joins strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// convertValueForJSON converts ABI values to JSON-friendly representations
func convertValueForJSON(value interface{}, abiType abi.Type) interface{} {
	switch abiType.T {
	case abi.AddressTy:
		if addr, ok := value.(common.Address); ok {
			return addr.Hex()
		}
	case abi.BytesTy, abi.FixedBytesTy:
		switch v := value.(type) {
		case []byte:
			return "0x" + hex.EncodeToString(v)
		case [1]byte:
			return "0x" + hex.EncodeToString(v[:])
		case [2]byte:
			return "0x" + hex.EncodeToString(v[:])
		case [3]byte:
			return "0x" + hex.EncodeToString(v[:])
		case [4]byte:
			return "0x" + hex.EncodeToString(v[:])
		case [8]byte:
			return "0x" + hex.EncodeToString(v[:])
		case [16]byte:
			return "0x" + hex.EncodeToString(v[:])
		case [20]byte:
			return "0x" + hex.EncodeToString(v[:])
		case [32]byte:
			return "0x" + hex.EncodeToString(v[:])
		}
	case abi.IntTy, abi.UintTy:
		if bigInt, ok := value.(*big.Int); ok {
			return bigInt.String()
		}
	case abi.SliceTy, abi.ArrayTy:
		return convertSliceForJSON(value, abiType.Elem)
	case abi.TupleTy:
		return convertTupleForJSON(value, abiType)
	}
	return value
}

// convertSliceForJSON converts slice/array values for JSON representation
func convertSliceForJSON(value interface{}, elemType *abi.Type) interface{} {
	if elemType == nil {
		return value
	}

	switch slice := value.(type) {
	case []common.Address:
		result := make([]string, len(slice))
		for i, addr := range slice {
			result[i] = addr.Hex()
		}
		return result
	case []*big.Int:
		result := make([]string, len(slice))
		for i, v := range slice {
			result[i] = v.String()
		}
		return result
	case [][]byte:
		result := make([]string, len(slice))
		for i, v := range slice {
			result[i] = "0x" + hex.EncodeToString(v)
		}
		return result
	}

	return value
}

// convertTupleForJSON converts tuple values for JSON representation
func convertTupleForJSON(value interface{}, abiType abi.Type) interface{} {
	if abiType.TupleElems == nil {
		return value
	}

	// Handle struct types - convert to map for better JSON representation
	result := make(map[string]interface{})

	// Use reflection to extract struct fields
	switch v := value.(type) {
	case struct {
		From   common.Address
		To     common.Address
		Stable bool
	}:
		result["from"] = v.From.Hex()
		result["to"] = v.To.Hex()
		result["stable"] = v.Stable
		return result
	}

	return value
}
