package blockchain

import (
	"investindicator/blockchain/uniswap"
	"math/big"
	"os"
	"testing"
)

func TestBlockchain(t *testing.T) {

	pk := os.Getenv("PK")

	us, err := uniswap.NewUniswapClient(uniswap.NewUniswapClientConfig(
		"https://api.avax.network/ext/bc/C/rpc",
		pk,
		"0x94b75331AE8d42C1b61065089B7d48FE14aA73b7",
		"0x000000000022D473030F116dDEE9F6B43aC78BA3",
		big.NewInt(300000),
	))
	if err != nil {
		panic(err)
	}
	bt := NewBlockChainTrader(us, nil, nil)

	err = bt.SwapUsdtUsdc(true)
	if err != nil {
		t.Fatal(err)
	}

}
