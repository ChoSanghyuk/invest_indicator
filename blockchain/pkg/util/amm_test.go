package util

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

// memo. SqrtPrice는 두 tick 사이의 값이기에, safelyGetStateOfAMM 결과로 나오는 sprtPrice 랑 tick 완벽하게 매칭되지 않음
func TestTickToSqrtPriceX96(t *testing.T) {

	sqrtPrice := TickToSqrtPriceX96(-249428)

	expected, _ := big.NewInt(0).SetString("304011615425126403287043", 10)
	assert.Equal(t, expected, sqrtPrice)
}

func TestComputeAmounts(t *testing.T) {

	sqrtPriceX96, _ := big.NewInt(0).SetString("287961100762493435308188", 10)
	tick := -250513
	tickLower := -250800
	tickUpper := -248800
	amount0Max := big.NewInt(1000000000000000000)
	amount1Max := big.NewInt(13000000)
	amount0, amount1, l := ComputeAmounts(sqrtPriceX96, tick, tickLower, tickUpper, amount0Max, amount1Max)

	t.Log("amount0:", amount0)
	t.Log("amount1:", amount1)
	t.Log("liquidity:", l)

	// Verify we got non-zero results
	assert.Greater(t, l.Cmp(big.NewInt(0)), 0, "liquidity should be > 0")
	assert.Greater(t, amount0.Cmp(big.NewInt(0)), -1, "amount0 should be >= 0")
	assert.Greater(t, amount1.Cmp(big.NewInt(0)), -1, "amount1 should be >= 0")

	// Verify amounts don't exceed the max budget
	assert.LessOrEqual(t, amount0.Cmp(amount0Max), 0, "amount0 should not exceed amount0Max")
	assert.LessOrEqual(t, amount1.Cmp(amount1Max), 0, "amount1 should not exceed amount1Max")
}
