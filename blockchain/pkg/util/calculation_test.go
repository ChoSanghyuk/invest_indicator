package util

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSqrtPriceToPrice(t *testing.T) {

	val, _ := big.NewInt(0).SetString("280057970020625981233062", 0)
	priceRaw := SqrtPriceToPrice(val)

	// expected, _ := big.NewInt(0).SetString("304011615425126403287043", 10)
	// assert.Equal(t, expected, sqrtPrice)
	price, _ := priceRaw.Float64()

	fmt.Printf("%v\n", val)
	fmt.Printf("%v\n", priceRaw)
	fmt.Printf("%v\n", price)
}

func TestCalculateRebalanceAmounts(t *testing.T) {

	// 1AVAX = 12.49 USDC일 때의 값
	sqrtPrice, _ := big.NewInt(0).SetString("280057970020625981233062", 0)
	// price := 12.49

	t.Run("USDC_TO_WAVAX", func(t *testing.T) {
		wavaxBalance := big.NewInt(2 * 1000000000000000000) // 2AVAX. 25USDC
		usdcBalance := big.NewInt(50000000)                 // 50 USDC

		tokenToSwap, swapAmount, err := CalculateRebalanceAmounts(
			wavaxBalance,
			usdcBalance,
			sqrtPrice,
		)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, tokenToSwap)
		fmt.Printf("tokenToSwap : %v, swapAmount: %v\n", tokenToSwap, swapAmount)
	})

	t.Run("WAVX_TO_USDC", func(t *testing.T) {
		wavaxBalance := big.NewInt(5 * 1000000000000000000) // 5AVAX. 62.5 USDC
		usdcBalance := big.NewInt(50000000)                 // 50 USDC

		tokenToSwap, swapAmount, err := CalculateRebalanceAmounts(
			wavaxBalance,
			usdcBalance,
			sqrtPrice,
		)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, tokenToSwap)
		fmt.Printf("tokenToSwap : %v, swapAmount: %v\n", tokenToSwap, swapAmount)
	})

}

/*
279069233386509245994440
     3306379361727413336
*/
