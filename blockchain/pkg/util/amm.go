package util

import (
	"math/big"
)

// Automated Market Maker Functions

// ---------------------
// Q96 = 2^96 (for sqrtPriceX96 format). memo. 2의 n승 구하는 방법.
var Q96 = new(big.Int).Lsh(big.NewInt(1), 96)

// -----------------------------------------------------------------------------
// Utilities
// -----------------------------------------------------------------------------

// memo. 기본적인 수학적 수식은 일치하나, solidity에서는 Pow가 없어 multiplier을 미리 계산해두고 사용하여 아래식과 약간의 오차가 발생.
// TickToSqrtPriceX96 converts a tick to sqrt(price) in Q96 format
// func TickToSqrtPriceX96(tick int) *big.Int {
// 	// price = 1.0001^tick
// 	price := math.Pow(1.0001, float64(tick))
// 	sqrtPrice := math.Sqrt(price)
// 	r := new(big.Float).Mul(big.NewFloat(sqrtPrice), new(big.Float).SetInt(Q96))
// 	out := new(big.Int)
// 	r.Int(out)
// 	return out
// }

func TickToSqrtPriceX96(tick int) *big.Int {
	absTick := tick
	if tick < 0 {
		absTick = -tick
	}

	// Clamp to allowed tick range
	maxTick := 887272
	if absTick > maxTick {
		panic("Tick out of range")
	}

	ratio := new(big.Int)
	ratio.SetString("340282366920938463463374607431768211456", 10) // 1 << 128

	/* memo. multipliers are pre-computed constants.
	1.0001^tick requires a power/exponent function, but Solidity has no floating point and no pow().
	instead of computing: `sqrt(1.0001^tick)`, we rewrite tick into its binary representation. */
	multipliers := []string{
		"fffcb933bd6fad37aa2d162d1a594001",
		"fff97272373d413259a46990580e213a",
		"fff2e50f5f656932ef12357cf3c7fdcc",
		"ffe5caca7e10e4e61c3624eaa0941cd0",
		"ffcb9843d60f6159c9db58835c926644",
		"ff973b41fa98c081472e6896dfb254c0",
		"ff2ea16466c96a3843ec78b326b52861",
		"fe5dee046a99a2a811c461f1969c3053",
		"fcbe86c7900a88aedcffc83b479aa3a4",
		"f987a7253ac413176f2b074cf7815e54",
		"f3392b0822b70005940c7a398e4b70f3",
		"e7159475a2c29b7443b29c7fa6e889d9",
		"d097f3bdfd2022b8845ad8f792aa5825",
		"a9f746462d870fdf8a65dc1f90e061e5",
		"70d869a156d2a1b890bb3df62baf32f7",
		"31be135f97d08fd981231505542fcfa6",
		"9aa508b5b7a84e1c677de54f3e99bc9",
		"5d6af8dedb81196699c329225ee604",
		"2216e584f5fa1ea926041bedfe98",
		"48a170391f7dc42444e8fa2",
	}

	for i := 0; i < 20; i++ {
		if (absTick & (1 << i)) != 0 {
			mul := new(big.Int)
			mul.SetString(multipliers[i], 16)
			ratio.Mul(ratio, mul)
			ratio.Rsh(ratio, 128)
		}
	}

	if tick > 0 {
		// ratio = (2^256 - 1) / ratio
		max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
		ratio = max.Div(max, ratio)
	}

	// Round up then shift down to Q96
	ratio.Add(ratio, new(big.Int).SetUint64(0xFFFFFFFF))
	ratio.Rsh(ratio, 32)

	return ratio
}

// big division: (a * b) / c  with rounding down
func mulDiv(a, b, c *big.Int) *big.Int {
	num := new(big.Int).Mul(a, b)
	return new(big.Int).Div(num, c)
}

// -----------------------------------------------------------------------------
// Core Logic
// -----------------------------------------------------------------------------

// ComputeAmounts calculates the required token0 and token1 amounts and liquidity L.
// Inputs:
//   - sqrtPriceX96: current sqrt price (from safelyGetStateOfAMM)
//   - tick, tickLower, tickUpper: position ticks
//   - amount0Max, amount1Max: how much you WANT to supply (your budgets)
//
// Returns: amount0Required, amount1Required, liquidity L
// -----------------------------------------------------------------------------
func ComputeAmounts(
	sqrtPriceX96 *big.Int,
	tick int,
	tickLower int,
	tickUpper int,
	amount0Max *big.Int,
	amount1Max *big.Int,
) (amount0 *big.Int, amount1 *big.Int, L *big.Int) {

	// compute sqrtPriceLower / sqrtPriceUpper
	sqrtLower := TickToSqrtPriceX96(tickLower)
	sqrtUpper := TickToSqrtPriceX96(tickUpper)

	// convert to big.Float for intermediate calcs
	sP := new(big.Float).SetInt(sqrtPriceX96)
	sL := new(big.Float).SetInt(sqrtLower)
	sU := new(big.Float).SetInt(sqrtUpper)

	// -------------------------------------------------------------------------
	// CASE 1: Price <= tickLower → only token0 required
	// amount0 = L * (sqrtPriceUpper - sqrtPriceLower) * Q96 / (sqrtPriceLower * sqrtPriceUpper)
	// amount1 = 0
	// -------------------------------------------------------------------------
	if tick <= tickLower {
		// L0 = amount0Max * (sqrtL * sqrtU) / (sqrtU - sqrtL) / Q96
		numer := new(big.Float).Mul(
			new(big.Float).Mul(new(big.Float).SetInt(amount0Max), sL),
			sU,
		)
		numer.Quo(numer, new(big.Float).SetInt(Q96))
		denom := new(big.Float).Sub(sU, sL)

		Lf := new(big.Float).Quo(numer, denom)
		L = new(big.Int)
		Lf.Int(L)

		amount0 = amount0Max
		amount1 = big.NewInt(0)
		return
	}

	// -------------------------------------------------------------------------
	// CASE 3: Price >= tickUpper → only token1 required
	// amount0 = 0
	// amount1 = L * (sqrtPriceUpper - sqrtPriceLower) / Q96
	// -------------------------------------------------------------------------
	if tick >= tickUpper {
		// L1 = amount1Max * Q96 / (sqrtU - sqrtL)
		numer := new(big.Float).Mul(new(big.Float).SetInt(amount1Max), new(big.Float).SetInt(Q96))
		denom := new(big.Float).Sub(sU, sL)
		Lf := new(big.Float).Quo(numer, denom)

		L = new(big.Int)
		Lf.Int(L)

		amount1 = amount1Max
		amount0 = big.NewInt(0)
		return
	}

	// -------------------------------------------------------------------------
	// CASE 2: Price inside range → token0 + token1
	// For a position with range [tickLower, tickUpper] and current price sqrtP:
	// L0 = amount0 * (sqrtP * sqrtU) / (sqrtU - sqrtP) / Q96
	// L1 = amount1 * Q96 / (sqrtP - sqrtL)
	// -------------------------------------------------------------------------

	// Liquidity from amount0:
	// L0 = amount0Max * (sqrtP * sqrtU) / (sqrtU - sqrtP) / Q96
	denom0 := new(big.Float).Sub(sU, sP)
	numer0 := new(big.Float).Mul(
		new(big.Float).Mul(new(big.Float).SetInt(amount0Max), sP),
		sU,
	)
	numer0.Quo(numer0, new(big.Float).SetInt(Q96))
	L0f := new(big.Float).Quo(numer0, denom0)
	L0 := new(big.Int)
	L0f.Int(L0)

	// Liquidity from amount1:
	// L1 = amount1Max * Q96 / (sqrtP - sqrtL)
	denom1 := new(big.Float).Sub(sP, sL)
	numer1 := new(big.Float).Mul(new(big.Float).SetInt(amount1Max), new(big.Float).SetInt(Q96))
	L1f := new(big.Float).Quo(numer1, denom1)
	L1 := new(big.Int)
	L1f.Int(L1)

	// Choose the MIN liquidity (must satisfy both token budgets)
	if L0.Cmp(L1) < 0 {
		L = L0
	} else {
		L = L1
	}

	// Now compute actual required amounts using L.
	Lf := new(big.Float).SetInt(L)

	// amount0 = L * (sqrtU - sqrtP) * Q96 / (sqrtP * sqrtU)
	{
		numer := new(big.Float).Mul(Lf, new(big.Float).Sub(sU, sP))
		numer.Mul(numer, new(big.Float).SetInt(Q96))
		denom := new(big.Float).Mul(sP, sU)
		a0f := new(big.Float).Quo(numer, denom)
		amount0 = new(big.Int)
		a0f.Int(amount0)
	}

	// amount1 = L * (sqrtP - sqrtL) / Q96
	{
		a1f := new(big.Float).Mul(Lf, new(big.Float).Sub(sP, sL))
		a1f.Quo(a1f, new(big.Float).SetInt(Q96))
		amount1 = new(big.Int)
		a1f.Int(amount1)
	}

	return
}

/*
Liquidity is an abstract numeric value used inside the AMM math to relate prices to amounts.
It is not token0 or token1 amounts — it is the scaling constant of the curve.

Price P = (amount1 / amount0)
Liquidity L = constant relating amount0 & amount1 to P


*/
