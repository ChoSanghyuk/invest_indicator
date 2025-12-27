package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"investindicator/blockchain/pkg/contractclient"
	txtypes "investindicator/blockchain/pkg/types"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

var (
	senderAddr common.Address = common.HexToAddress("0x0000000000000000000000000000000000000001")
	thisAddr   common.Address = common.HexToAddress("0x0000000000000000000000000000000000000002")
)

type UniswapClientConfig struct {
	url      string // "https://api.avax.network/ext/bc/C/rpc"
	pk       string
	urAddr   string   // universalRouter address // "0x94b75331AE8d42C1b61065089B7d48FE14aA73b7"
	pmAddr   string   // permit address // "0x000000000022D473030F116dDEE9F6B43aC78BA3"
	gasLimit *big.Int // big.NewInt(300000)
}

func NewUniswapClientConfig(url string, pk string, urAddr string, pmAddr string, gasLimit *big.Int) *UniswapClientConfig {
	return &UniswapClientConfig{
		url:      url,
		pk:       pk,
		urAddr:   urAddr,
		pmAddr:   pmAddr,
		gasLimit: gasLimit,
	}
}

type UniswapClient struct {
	pk     *ecdsa.PrivateKey
	gl     *big.Int
	myAddr common.Address
	uo     *contractclient.ContractClient // universalrouter orchestrator
	po     *contractclient.ContractClient // permit2 orchestrator
}

// pk: 앞에 0x 없는 hex형 데이터
func NewUniswapClient(conf *UniswapClientConfig) (*UniswapClient, error) {
	client, err := ethclient.Dial(conf.url)
	if err != nil {
		return nil, err
	}

	// Private Key & Address
	privateKey, err := crypto.HexToECDSA(conf.pk)
	if err != nil {
		return nil, err
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// universalRouter
	urAbi, err := abi.JSON(strings.NewReader(universalRouterAbiJSON))
	if err != nil {
		return nil, err
	}
	universalRouter := contractclient.NewContractClient(client, common.HexToAddress(conf.urAddr), &urAbi)

	// permitRouter
	permitAbi, err := abi.JSON(bytes.NewReader(permit2Json))
	if err != nil {
		return nil, err
	}
	permit2Orchestrator := contractclient.NewContractClient(client, common.HexToAddress(conf.pmAddr), &permitAbi)

	return &UniswapClient{
		pk:     privateKey,
		myAddr: address,
		uo:     universalRouter,
		po:     permit2Orchestrator,
		gl:     conf.gasLimit,
	}, nil
}

const (
	PERMIT2_PERMIT   byte = 0x0a
	WRAP_ETH         byte = 0x0b
	UNWRAP_ETH       byte = 0x0c
	V3_SWAP_EXACT_IN byte = 0x00
	PAY_PORTION      byte = 0x06
	SWEEP            byte = 0x04
)

type PermitSingle struct {
	Details     PermitDetails  `json:"details"`
	Spender     common.Address `json:"spender"`
	SigDeadline *big.Int       `json:"sigDeadline"`
}

type PermitDetails struct {
	Token      common.Address `json:"token"`
	Amount     *big.Int       `json:"amount"`
	Expiration *big.Int       `json:"expiration"`
	Nonce      *big.Int       `json:"nonce"`
}

type PermitNonce struct {
	Amount     *big.Int `json:"amount"`
	Expiration *big.Int `json:"expiration"`
	Nonce      *big.Int `json:"nonce"`
}

type WrapETH struct {
	Recipient common.Address
	Amount    *big.Int
}

type V3Path struct {
	TokenIn  common.Address
	Fee      *big.Int // 500
	TokenOut common.Address
}

type V3SwapExactIn struct {
	Recipient    common.Address // 0000000000000000000000000000000000000000000000000000000000000002
	AmountIn     *big.Int
	AmountOutMin *big.Int
	Path         []byte
	PayerIsUser  bool
}

type PayPortion struct {
	Token     common.Address
	Recipient common.Address
	Bips      *big.Int // 25
}

type Sweep struct {
	Token        common.Address
	Recipient    common.Address
	AmountOutMin *big.Int
}

func (u *UniswapClient) Approve(token common.Address, spender common.Address, amount *big.Int, expiration *big.Int) (*common.Hash, error) {

	tx, err := u.po.Send(txtypes.Standard, u.gl, &u.myAddr, u.pk, "approve", token, spender, amount, expiration) // todo. gasfee config?
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func (u *UniswapClient) SwapNativeForToken(tokenOut common.Address, amountIn *big.Int, amountOutMin *big.Int) (*common.Hash, error) {
	commands := []byte{}
	inputs := [][]byte{}

	wrapCmd, wrapInput, err := u.wrapETH(amountIn)
	if err != nil {
		return nil, err
	}
	commands = append(commands, wrapCmd)
	inputs = append(inputs, wrapInput)

	tokenIn := common.HexToAddress("0xb31f66aa3c1e785363f0875a1b74e27b85fd66c7") // WAVAX
	//V3_SWAP_EXACT_IN
	swapCmd, swapInput, err := u.swapExactIn(tokenIn, tokenOut, amountIn, amountOutMin)
	if err != nil {
		return nil, err
	}
	commands = append(commands, swapCmd)
	inputs = append(inputs, swapInput)

	// SWEEP
	sweepCmd, sweepInput, err := u.sweep(tokenOut, amountOutMin)
	if err != nil {
		return nil, err
	}
	commands = append(commands, sweepCmd)
	inputs = append(inputs, sweepInput)

	tx, err := u.uo.SendWithValue(txtypes.Standard, u.gl, amountIn, &u.myAddr, u.pk, "execute", commands, inputs, big.NewInt(time.Now().Add(time.Hour*1).Unix()))
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func (u *UniswapClient) Swap(tokenIn common.Address, tokenOut common.Address, amountIn *big.Int, amountOutMin *big.Int) (*common.Hash, error) {

	commands := []byte{} //V4_SWAP
	inputs := [][]byte{}

	// permit2 수행 여부 확인
	nonceResult, err := u.permit2Nonce(tokenIn, *u.uo.ContractAddress())
	if err != nil {
		return nil, err
	}

	// permit된 amount보다 금액이 더 큰 경우 또는 현재 시각이 permit된 만료시간보다 클 경우, permit 재수행
	if amountIn.Cmp(nonceResult.Amount) > 0 || time.Now().Unix() > nonceResult.Expiration.Int64() {
		// PERMIT2_PERMIT
		permitCmd, permitInput, err := u.permit2(PermitSingle{
			Details: PermitDetails{
				Token:      tokenIn,
				Amount:     big.NewInt(amountIn.Int64() * 100),
				Expiration: big.NewInt(time.Now().AddDate(0, 1, 0).Unix()), // 권한은 한달 씩
				Nonce:      nonceResult.Nonce,
			},
			Spender:     *u.uo.ContractAddress(),
			SigDeadline: big.NewInt(time.Now().Add(time.Hour * 1).Unix()), //big.NewInt(1759366801), // 서명한 1시간만 유효
		})
		if err != nil {
			return nil, err
		}
		commands = append(commands, permitCmd)
		inputs = append(inputs, permitInput)
	}

	//V3_SWAP_EXACT_IN
	swapCmd, swapInput, err := u.swapExactIn(tokenIn, tokenOut, amountIn, amountOutMin)
	if err != nil {
		return nil, err
	}
	commands = append(commands, swapCmd)
	inputs = append(inputs, swapInput)

	// PAY_PORTION(한번만 테스트용으로 수행)
	// payPortionCmd, payPortionInput, err := u.payPortion(common.HexToAddress("0xb97ef9ef8734c71904d8002f8b6bc66dd9c48a6e"), common.HexToAddress("7ffc3dbf3b2b50ff3a1d5523bc24bb5043837b14"))
	// if err != nil {
	// 	return nil, err
	// }
	// commands = append(commands, payPortionCmd)
	// inputs = append(inputs, payPortionInput)

	// SWEEP
	sweepCmd, sweepInput, err := u.sweep(tokenOut, amountOutMin)
	if err != nil {
		return nil, err
	}
	commands = append(commands, sweepCmd)
	inputs = append(inputs, sweepInput)

	tx, err := u.uo.Send(txtypes.Standard, u.gl, &u.myAddr, u.pk, "execute", commands, inputs, big.NewInt(time.Now().Add(time.Hour*1).Unix()))
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

/****************************************  inner method ****************************************/

func (u *UniswapClient) wrapETH(amountIn *big.Int) (command byte, input []byte, err error) {
	wrapETHABI, err := abi.JSON(strings.NewReader(wrapEthAbiJSON))
	if err != nil {
		return 0, nil, err
	}
	wrapETHInput, err := wrapETHABI.Pack("wrapETH", thisAddr, amountIn)
	if err != nil {
		return 0, nil, err
	}

	return WRAP_ETH, wrapETHInput[4:], nil // function sig omit
}

func (u *UniswapClient) permit2(permitSingle PermitSingle) (command byte, input []byte, err error) {

	/********************  permit sig data 생성하기 ********************/

	finalHash, err := u.permit2Hash(permitSingle)
	if err != nil {
		return 0, nil, err
	}

	sig, err := crypto.Sign(finalHash, u.pk)
	if err != nil {
		return 0, nil, err
	}
	if sig[64] < 27 { // IMPORTANT
		sig[64] += 27
	}

	/*. permit 파라미터 최종 pack   */
	permit2ABI, err := abi.JSON(strings.NewReader(permit2AbiJSON))
	if err != nil {
		return 0, nil, err
	}

	permitInput, err := permit2ABI.Pack("permit", permitSingle.Details.Token, permitSingle.Details.Amount, permitSingle.Details.Expiration, permitSingle.Details.Nonce, permitSingle.Spender, permitSingle.SigDeadline, sig)
	if err != nil {
		return 0, nil, err
	}

	return PERMIT2_PERMIT, permitInput[4:], nil
}

func (u *UniswapClient) swapExactIn(tokenIn common.Address, tokenOut common.Address, amountIn *big.Int, amountOutMin *big.Int) (command byte, input []byte, err error) {

	v3SwapExactInABI, err := abi.JSON(strings.NewReader(v3SwapExactInAbiJSON))
	if err != nil {
		return 0, nil, err
	}

	path := []byte{}
	// fee := 500
	// feeBytes := []byte{
	// 	byte(fee >> 16), // most significant byte
	// 	byte(fee >> 8),
	// 	byte(fee),
	// }
	feeBytes := []byte{0x00, 0x01, 0xf4} // 500

	path = append(path, tokenIn.Bytes()...)
	path = append(path, feeBytes...)
	path = append(path, tokenOut.Bytes()...)

	swap := V3SwapExactIn{
		Recipient:    thisAddr,
		AmountIn:     amountIn,
		AmountOutMin: amountOutMin,
		Path:         path,
		PayerIsUser:  true,
	}

	swapInput, err := v3SwapExactInABI.Pack("V3SwapExactIn", swap.Recipient, swap.AmountIn, swap.AmountOutMin, swap.Path, swap.PayerIsUser)
	if err != nil {
		return 0, nil, err
	}

	return V3_SWAP_EXACT_IN, swapInput[4:], nil // function sig omit
}

// Token:     common.HexToAddress("0xb97ef9ef8734c71904d8002f8b6bc66dd9c48a6e"),
// Recipient: common.HexToAddress("7ffc3dbf3b2b50ff3a1d5523bc24bb5043837b14"),
func (u *UniswapClient) payPortion(token common.Address, recipient common.Address) (command byte, input []byte, err error) {

	// commands = append(commands, PAY_PORTION)
	payPortion := PayPortion{
		Token:     token,
		Recipient: recipient,
		Bips:      big.NewInt(25),
	}

	payPortionABI, err := abi.JSON(strings.NewReader(payPortionAbiJSON))
	if err != nil {
		return 0, nil, err
	}
	input, err = payPortionABI.Pack("PayPortion", payPortion.Token, payPortion.Recipient, payPortion.Bips)
	if err != nil {
		return 0, nil, err
	}

	return PAY_PORTION, input[4:], nil // function sig omit
}

func (u *UniswapClient) sweep(tokenOut common.Address, amountOutMin *big.Int) (command byte, input []byte, err error) {
	sweep := Sweep{
		Token:        tokenOut,
		Recipient:    u.myAddr,
		AmountOutMin: amountOutMin,
	}

	sweepABI, err := abi.JSON(strings.NewReader(sweepAbiJSON))
	if err != nil {
		return 0, nil, err
	}
	sweepInput, err := sweepABI.Pack("Sweep", sweep.Token, sweep.Recipient, sweep.AmountOutMin)
	if err != nil {
		return 0, nil, err
	}

	return SWEEP, sweepInput[4:], nil // function sig omit
}

func (u *UniswapClient) permit2Nonce(
	token common.Address,
	spender common.Address,
) (*PermitNonce, error) {

	result, err := u.po.Call(&u.myAddr, "allowance", u.myAddr, token, spender) // Amount, Expiration, Nonce
	if err != nil {
		return nil, err
	}

	return &PermitNonce{
		Amount:     result[0].(*big.Int),
		Expiration: result[1].(*big.Int),
		Nonce:      result[2].(*big.Int),
	}, nil
}

func (u *UniswapClient) permit2Hash(permitSingle PermitSingle) ([]byte, error) {

	/*
		apitypes.TypedData를 통해 solidity의 struct 타입을 go로 구현한다.
		func TypedDataAndHash(typedData TypedData) ([]byte, string, error)를 통해 EIP-712의 _hashTypedData(permitBatch.hash()) 부분을 구현.
		이후 그 데이터를 서명하여 이후에 붙여야 함.
	*/
	chainId := u.po.ChainId() //big.NewInt(43114)

	permit2Addr := u.po.ContractAddress()

	domain := apitypes.TypedDataDomain{
		Name:              "Permit2",
		ChainId:           (*math.HexOrDecimal256)(chainId),
		VerifyingContract: permit2Addr.Hex(),
	}

	types := apitypes.Types{
		"EIP712Domain": {
			{Name: "name", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		},
		"PermitDetails": {
			{Name: "token", Type: "address"},
			{Name: "amount", Type: "uint160"},
			{Name: "expiration", Type: "uint48"},
			{Name: "nonce", Type: "uint48"},
		},
		"PermitSingle": {
			{Name: "details", Type: "PermitDetails"},
			{Name: "spender", Type: "address"},
			{Name: "sigDeadline", Type: "uint256"},
		},
	}

	message := apitypes.TypedDataMessage{
		"details": map[string]interface{}{
			"token":      permitSingle.Details.Token.Hex(),
			"amount":     permitSingle.Details.Amount.String(),
			"expiration": permitSingle.Details.Expiration,
			"nonce":      permitSingle.Details.Nonce,
		},
		"spender":     permitSingle.Spender.Hex(),
		"sigDeadline": permitSingle.SigDeadline.String(),
	}

	typedData := apitypes.TypedData{
		Types:       types,
		PrimaryType: "PermitSingle",
		Message:     message,
		Domain:      domain,
	}

	/*
		AllowanceTransfer.sol의 37 line. permitSingle.hash() 하는 부분
		struct 데이터의 hash는 TYPEHASH와 같이 function sig를 가지고, 필드들을 encode하는 작업.
	*/
	finalHash, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, fmt.Errorf("failed to hash struct: %v", err)
	}

	return finalHash, nil
}
