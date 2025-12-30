package uniswap

import (
	"os"
	"path/filepath"
	"runtime"
)

// 컨트랙트를 호출함에 있어, 전체 abi가 아닌 일부만 필요한 경우 해당 부분만 별도로 정의해서 사용
const (
	universalRouterAbiJSON = `[{
		"name": "execute",
		"type": "function",
		"inputs": [
			{"name": "commands", "type": "bytes"},
			{"name": "inputs", "type": "bytes[]"},
			{"name": "deadline", "type": "uint256"}
		],
		"outputs": [],
		"stateMutability": "payable"
	}]`
	apprvoeAbiJSON = `[{
			"inputs": [
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "address", 
				"name": "spender",
				"type": "address"
			},
			{
				"internalType": "uint160",
				"name": "amount",
				"type": "uint160"
			},
			{
				"internalType": "uint48",
				"name": "expiration",
				"type": "uint48"
			}
			],
			"name": "approve",
			"outputs": [],
			"stateMutability": "nonpayable",
			"type": "function"
		}
		]`
	permit2AbiJSON = `[{
		"name": "permit",
		"type": "function",
		"inputs": [
			{ "type": "address" },
			{ "type": "uint160" },
			{ "type": "uint48" },
			{ "type": "uint48" },
			{ "type": "address" },
			{ "type": "uint256" },
			{ "type": "bytes" }
		],
		"outputs": []
	}]`
	wrapEthAbiJSON = `[{
		"name": "wrapETH",
		"type": "function",
		"inputs": [
			{ "type": "address" },
			{ "type": "uint256" }
		],
		"outputs": []
	}]`
	v3SwapExactInAbiJSON = `[{
		"name": "V3SwapExactIn",
		"type": "function",
		"inputs": [
			{ "type": "address" },
			{ "type": "uint256" },
			{ "type": "uint256" },
			{ "type": "bytes"	},
			{ "type": "bool" }
		],
		"outputs": []
	}]`
	payPortionAbiJSON = `[{
		"name": "PayPortion",
		"type": "function",
		"inputs": [
			{ "type": "address" },
			{ "type": "address" },
			{ "type": "uint256" }
		],
		"outputs": []
	}]`
	sweepAbiJSON = `[{
		"name": "Sweep",
		"type": "function",
		"inputs": [
			{ "type": "address" },
			{ "type": "address" },
			{ "type": "uint256" }
		],
		"outputs": []
	}]`
)

var (
	permit2Json []byte
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	// Change working directory to that location
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
	permit2Json, err = os.ReadFile("../abi/Permit2.json")
	if err != nil {
		panic(err)
	}

}
