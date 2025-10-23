package entities

import (
	"fmt"

	core "github.com/daoleno/uniswap-sdk-core/entities"
)

func AmountWithPathCurrency(amount *core.CurrencyAmount, pool *Pool) (*core.CurrencyAmount, error) {
	pathCurrency := getPathCurrency(amount.Currency.Wrapped(), pool)
	if pathCurrency == nil {
		return &core.CurrencyAmount{}, fmt.Errorf("expected currency %s to be either %s or %s", amount.Currency.Symbol(), pool.Currency0.Symbol(), pool.Currency1.Symbol())
	}
	fractionalAmount := *core.FromFractionalAmount(
		pathCurrency,
		amount.Numerator,
		amount.Denominator,
	)
	return &fractionalAmount, nil
}

func getPathCurrency(currency *core.Token, pool *Pool) *core.Token {
	if pool.InvolvesToken(currency) {
		return currency
	} else if pool.InvolvesToken(currency.Wrapped()) {
		return currency.Wrapped()
	} else if pool.Currency0.Wrapped().Equal(currency) {
		return pool.Currency0
	} else if pool.Currency1.Wrapped().Equal(currency) {
		return pool.Currency1
	} else {
		return nil
		// , fmt.Errorf("Expected currency %s to be either %s or %s", currency.Symbol(), pool.Currency0.Symbol(), pool.Currency1.Symbol())
	}
}
