package util

import (
	"fmt"
	"investindicator/blockchain/pkg/types"
	"math"
	"math/big"
	"strings"
)

// Validation and helper functions for liquidity staking operations

// ValidateStakingRequest validates input parameters for staking operation
// Returns error if validation fails, nil otherwise
func ValidateStakingRequest(maxWAVAX, maxUSDC *big.Int, rangeWidth, slippagePct int) error {
	// Range width validation (1-20 tick ranges)
	if rangeWidth <= 0 || rangeWidth > 20 {
		return fmt.Errorf("range width must be between 1 and 20, got %d (valid examples: 2, 6, 10)", rangeWidth)
	}

	// Slippage validation (1-50 percent)
	if slippagePct <= 0 || slippagePct > 50 {
		return fmt.Errorf("slippage tolerance must be between 1 and 50 percent, got %d", slippagePct)
	}

	// Amount validation
	if maxWAVAX == nil || maxWAVAX.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("maxWAVAX must be > 0")
	}
	if maxUSDC == nil || maxUSDC.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("maxUSDC must be > 0")
	}

	return nil
}

// CalculateTickBounds calculates tick bounds from current tick and range width
// rangeWidth N means ±(N/2) tick ranges from current tick
// Returns tickLower, tickUpper, or error if bounds invalid
// Edge case: For extreme ticks near ±887272, bounds are clamped to valid range
func CalculateTickBounds(currentTick int32, rangeWidth int, tickSpacing int) (int32, int32, error) {
	const maxTick = 887272

	halfWidth := rangeWidth / 2
	// tickIndex := int(currentTick) / tickSpacing
	tickIndex := int(math.Round(float64(currentTick) / float64(tickSpacing)))

	// Calculate raw bounds
	rawTickLower := (tickIndex - halfWidth) * tickSpacing
	rawTickUpper := (tickIndex + halfWidth) * tickSpacing

	// Clamp to valid tick range for edge cases near ±maxTick
	// This handles extreme ticks where calculated bounds would exceed limits
	tickLower := int32(rawTickLower)
	tickUpper := int32(rawTickUpper)

	if tickLower < -maxTick {
		tickLower = -maxTick
	}
	if tickLower > maxTick {
		tickLower = maxTick
	}
	if tickUpper < -maxTick {
		tickUpper = -maxTick
	}
	if tickUpper > maxTick {
		tickUpper = maxTick
	}

	// Validate tickLower < tickUpper (should always be true after clamping)
	if tickLower >= tickUpper {
		return 0, 0, fmt.Errorf("tickLower (%d) must be < tickUpper (%d) - current tick %d with range width %d creates invalid bounds", tickLower, tickUpper, currentTick, rangeWidth)
	}

	return tickLower, tickUpper, nil
}

// CalculateMinAmount calculates minimum amount with slippage protection
// amountMin = amountDesired * (100 - slippagePct) / 100
func CalculateMinAmount(amountDesired *big.Int, slippagePct int) *big.Int {
	if amountDesired == nil {
		return big.NewInt(0)
	}

	// amountMin = amountDesired * (100 - slippagePct) / 100
	multiplier := big.NewInt(int64(100 - slippagePct))
	divisor := big.NewInt(100)

	result := new(big.Int).Mul(amountDesired, multiplier)
	result.Div(result, divisor)

	return result
}

// ExtractGasCost extracts gas cost from transaction receipt
// Returns gas cost in wei (GasUsed * EffectiveGasPrice)
func ExtractGasCost(receipt *types.TxReceipt) (*big.Int, error) {
	if receipt == nil {
		return nil, fmt.Errorf("receipt is nil")
	}

	// Parse GasUsed from string
	gasUsed := new(big.Int)
	if _, ok := gasUsed.SetString(receipt.GasUsed, 0); !ok {
		return nil, fmt.Errorf("failed to parse GasUsed: %s", receipt.GasUsed)
	}

	// Parse EffectiveGasPrice from string
	gasPrice := new(big.Int)
	if _, ok := gasPrice.SetString(receipt.EffectiveGasPrice, 0); !ok {
		return nil, fmt.Errorf("failed to parse EffectiveGasPrice: %s", receipt.EffectiveGasPrice)
	}

	// Calculate gas cost
	gasCost := new(big.Int).Mul(gasUsed, gasPrice)

	return gasCost, nil
}

// IsCriticalError determines if an error is critical and requires immediate halt (T015)
// Critical errors require immediate strategy halt, non-critical errors use threshold-based logic
// Implements error classification from research.md R6
func IsCriticalError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Critical error patterns that require immediate halt
	criticalPatterns := []string{
		"insufficient balance",
		"insufficient funds",
		"nft not owned",
		"not owner",
		"transaction reverted",
		"execution reverted",
		"invalid position state",
		"position does not exist",
		"unauthorized",
		"contract paused",
	}

	for _, pattern := range criticalPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}
