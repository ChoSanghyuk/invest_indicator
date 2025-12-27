package util

import (
	"fmt"
	"math/big"
)

// Strategy calculation functions

// SqrtPriceToPrice converts sqrtPriceX96 to human-readable price (T014)
// Formula: price = (sqrtPrice / 2^96)^2
// For WAVAX/USDC pool (assuming token0=WAVAX, token1=USDC):
//
//	priceUSDCperWAVAX = price * (10^(18-6)) // Adjust for decimals
//
// Used for stability detection and ratio calculations from research.md R1
func SqrtPriceToPrice(sqrtPriceX96 *big.Int) *big.Float {
	if sqrtPriceX96 == nil || sqrtPriceX96.Sign() == 0 {
		return big.NewFloat(0)
	}

	// Convert to big.Float
	sqrtPrice := new(big.Float).SetInt(sqrtPriceX96)

	// Divide by Q96 (2^96)
	q96 := new(big.Float).SetInt(Q96)
	sqrtPriceNormalized := new(big.Float).Quo(sqrtPrice, q96)

	// Square to get price
	price := new(big.Float).Mul(sqrtPriceNormalized, sqrtPriceNormalized)
	// decimalAdjustment := new(big.Float).SetInt64(1_000_000_000_000) // 10^12

	return price
}

// CalculateRebalanceAmounts calculates swap amounts needed to achieve 50:50 value ratio (T017)
// Uses value-based proportional rebalancing with current pool price from research.md R3
// Returns: tokenToSwap (0=WAVAX, 1=USDC), swapAmount, error
func CalculateRebalanceAmounts(
	wavaxBalance *big.Int,
	usdcBalance *big.Int,
	sqrtPriceX96 *big.Int,
) (tokenToSwap int, swapAmount *big.Int, err error) {
	if wavaxBalance == nil || usdcBalance == nil || sqrtPriceX96 == nil {
		return 0, nil, fmt.Errorf("nil input parameters")
	}

	// Get current pool price (USDC per WAVAX)
	price := SqrtPriceToPrice(sqrtPriceX96)

	// Adjust for decimals: WAVAX has 18 decimals, USDC has 6 decimals
	// Price needs to be adjusted by 10^(18-6) = 10^12
	// decimalAdjustment := new(big.Float).SetInt64(1_000_000_000_000) // 10^12 // !IMPORTANT. 필요없음. adjustment를 안 해야 정상 작동
	// priceUSDCperWAVAX := new(big.Float).Mul(price, decimalAdjustment)
	// fmt.Printf("priceUSDCperWAVAX: %v\n", priceUSDCperWAVAX)

	// Calculate current values in USDC terms
	wavaxBalanceFloat := new(big.Float).SetInt(wavaxBalance)
	usdcBalanceFloat := new(big.Float).SetInt(usdcBalance)

	wavaxValueInUSDC := new(big.Float).Mul(wavaxBalanceFloat, price)
	totalValue := new(big.Float).Add(wavaxValueInUSDC, usdcBalanceFloat)
	fmt.Printf("wavaxValueInUSDC: %v\n", wavaxValueInUSDC)
	fmt.Printf("totalValue: %v\n", totalValue)

	// Target 50% of total value in each token
	targetUSDC := new(big.Float).Quo(totalValue, big.NewFloat(2))
	targetWAVAXValue := new(big.Float).Quo(totalValue, big.NewFloat(2))
	fmt.Printf("targetUSDC: %v\n", targetUSDC)
	fmt.Printf("targetWAVAXValue: %v\n", targetWAVAXValue)

	// Determine which token to swap and how much
	usdcDiff := new(big.Float).Sub(usdcBalanceFloat, targetUSDC)

	// If USDC > target, swap USDC to WAVAX
	if usdcDiff.Sign() > 0 {
		// Swap excess USDC to WAVAX
		swapAmountFloat := usdcDiff
		swapAmount = new(big.Int)
		swapAmountFloat.Int(swapAmount)

		// Ensure positive and non-zero
		if swapAmount.Sign() <= 0 {
			return 0, big.NewInt(0), nil // No swap needed
		}

		return 1, swapAmount, nil // tokenToSwap=1 (USDC)
	}

	// If WAVAX > target, swap WAVAX to USDC
	wavaxDiff := new(big.Float).Sub(wavaxValueInUSDC, targetWAVAXValue)
	if wavaxDiff.Sign() > 0 {
		// Convert excess WAVAX value to WAVAX amount
		excessWAVAXAmount := new(big.Float).Quo(wavaxDiff, price)
		swapAmount = new(big.Int)
		excessWAVAXAmount.Int(swapAmount)

		// Ensure positive and non-zero
		if swapAmount.Sign() <= 0 {
			return 0, big.NewInt(0), nil // No swap needed
		}

		return 0, swapAmount, nil // tokenToSwap=0 (WAVAX)
	}

	// Already balanced
	return 0, big.NewInt(0), nil
}
