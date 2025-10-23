package utils

import (
	"math/big"

	"github.com/KyberNetwork/pancake-v3-sdk/constants"
	core "github.com/daoleno/uniswap-sdk-core/entities"
)

func TickToPrice(baseToken *core.Token, quoteToken *core.Token, tick int) (*core.Price, error) {
	sqrtRatioX96, err := GetSqrtRatioAtTick(tick)
	if err != nil {
		return nil, err
	}
	ratioX192 := new(big.Int).Mul(sqrtRatioX96, sqrtRatioX96)

	sorted, err := SortsBefore(baseToken, quoteToken)
	if err != nil {
		return nil, err
	}
	if sorted {
		return core.NewPrice(baseToken, quoteToken, constants.Q192, ratioX192), nil
	}
	return core.NewPrice(baseToken, quoteToken, ratioX192, constants.Q192), nil
}

func PriceToClosestTick(price *core.Price, baseToken, quoteToken *core.Token) (int, error) {
	sorted, err := SortsBefore(baseToken, quoteToken)
	if err != nil {
		return 0, nil
	}
	var sqrtRatioX96 *big.Int
	if sorted {
		sqrtRatioX96 = EncodeSqrtRatioX96(price.Numerator, price.Denominator)
	} else {
		sqrtRatioX96 = EncodeSqrtRatioX96(price.Denominator, price.Numerator)
	}
	tick, err := GetTickAtSqrtRatio(sqrtRatioX96)
	if err != nil {
		return 0, err
	}
	nextTickPrice, err := TickToPrice(baseToken, quoteToken, tick+1)
	if err != nil {
		return 0, err
	}
	if sorted {
		if !price.LessThan(nextTickPrice.Fraction) {
			tick++
		}
	} else {
		if !price.GreaterThan(nextTickPrice.Fraction) {
			tick++
		}
	}
	return tick, nil
}
