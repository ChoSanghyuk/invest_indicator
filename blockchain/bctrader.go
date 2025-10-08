package blockchain

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type BlockChainTrader struct {
	us *UniswapClient
}

func NewBlockChainTrader(us *UniswapClient) *BlockChainTrader {
	return &BlockChainTrader{
		us: us,
	}
}

func (b *BlockChainTrader) SwapUsdtUsdc(isUsdcIn bool) error {

	usdc := common.HexToAddress("0xb97ef9ef8734c71904d8002f8b6bc66dd9c48a6e")
	usdt := common.HexToAddress("0x9702230A8Ea53601f5cD2dc00fDBc13d4dF4A8c7")
	amountIn := big.NewInt(1e6)
	amountOutMin := big.NewInt(9e5)

	var tokenIn, tokenOut common.Address
	if isUsdcIn {
		tokenIn, tokenOut = usdc, usdt
	} else {
		tokenIn, tokenOut = usdt, usdc
	}

	tx, err := b.us.Swap(tokenIn, tokenOut, amountIn, amountOutMin)
	if err != nil {
		return fmt.Errorf("[SwapUsdtUsdc swap 오류 발생] %s", err)
	}

	var i int
	for i = 0; i < 10; i++ {
		receipt, err := b.us.uo.GetReceipt(*tx)
		if err != nil {
			time.Sleep(1 * time.Second)
		} else {
			if receipt.Status != "0x1" {
				return fmt.Errorf("SwapUsdtUsdc tx 오류 발생. tx: %s", tx.Hex())
			}
			break
		}
	}

	if i == 10 {
		return fmt.Errorf("SwapUsdtUsdc tx 조회 실패. 시도 횟수 10회 초과. tx: %s", tx.Hex())
	}

	return nil
}
