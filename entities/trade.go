package entities

import (
	"errors"
	"math/big"
	"sort"

	v3constants "github.com/KyberNetwork/pancake-v3-sdk/constants"
	core "github.com/daoleno/uniswap-sdk-core/entities"
)

var (
	ErrTradeHasMultipleRoutes   = errors.New("trade has multiple routes")
	ErrInvalidAmountForRoute    = errors.New("invalid amount for route")
	ErrInputCurrencyMismatch    = errors.New("input currency mismatch")
	ErrOutputCurrencyMismatch   = errors.New("output currency mismatch")
	ErrDuplicatePools           = errors.New("duplicate pools")
	ErrInvalidSlippageTolerance = errors.New("invalid slippage tolerance")
	ErrNoPools                  = errors.New("no pools")
	ErrInvalidMaxHops           = errors.New("invalid max hops")
	ErrInvalidRecursion         = errors.New("invalid recursion")
	ErrInvalidMaxSize           = errors.New("invalid max size")
	ErrMaxSizeExceeded          = errors.New("max size exceeded")
	ErrProtocolUnknown          = errors.New("protocol unknown")
)

func tradeComparator(a, b *Trade) int {
	if !a.InputAmount().Currency.Equal(b.InputAmount().Currency) {
		panic(ErrInputCurrencyMismatch)
	}
	if !a.OutputAmount().Currency.Equal(b.OutputAmount().Currency) {
		panic(ErrOutputCurrencyMismatch)
	}
	if a.OutputAmount().EqualTo(b.OutputAmount().Fraction) {
		if a.InputAmount().EqualTo(b.InputAmount().Fraction) {
			var aHops, bHops int
			for _, swap := range a.Swaps {
				aHops += len(swap.Route.TokenPath)
			}
			for _, swap := range b.Swaps {
				bHops += len(swap.Route.TokenPath)
			}
			return aHops - bHops
		}
		// trade A requires less input than trade B, so A should come first
		if a.InputAmount().LessThan(b.InputAmount().Fraction) {
			return -1
		} else {
			return 1
		}
	} else {
		// trade A has less output than trade B, so should come second
		if a.OutputAmount().LessThan(b.OutputAmount().Fraction) {
			return 1
		} else {
			return -1
		}
	}
}

type Trade struct {
	Swaps     []*Swap
	TradeType core.TradeType

	inputAmount    *core.CurrencyAmount
	outputAmount   *core.CurrencyAmount
	executionPrice *core.Price
	priceImpact    *core.Percent
}

type Swap struct {
	Route        *Route
	InputAmount  *core.CurrencyAmount
	OutputAmount *core.CurrencyAmount
}

type BestTradeOptions struct {
	MaxNumResults int
	MaxHops       int
}

func (t *Trade) Route() (*Route, error) {
	if len(t.Swaps) != 1 {
		return nil, ErrTradeHasMultipleRoutes
	}
	return t.Swaps[0].Route, nil
}

func (t *Trade) InputAmount() *core.CurrencyAmount {
	if t.inputAmount != nil {
		return t.inputAmount
	}
	inputCurrency := t.Swaps[0].InputAmount.Currency
	totalInputFromRoutes := core.FromRawAmount(inputCurrency, big.NewInt(0))
	for _, swap := range t.Swaps {
		totalInputFromRoutes = totalInputFromRoutes.Add(swap.InputAmount)
	}
	t.inputAmount = totalInputFromRoutes
	return t.inputAmount
}

func (t *Trade) OutputAmount() *core.CurrencyAmount {
	if t.outputAmount != nil {
		return t.outputAmount
	}

	outputCurrency := t.Swaps[0].OutputAmount.Currency
	totalOutputFromRoutes := core.FromRawAmount(outputCurrency, big.NewInt(0))
	for _, swap := range t.Swaps {
		totalOutputFromRoutes = totalOutputFromRoutes.Add(swap.OutputAmount)
	}
	t.outputAmount = totalOutputFromRoutes
	return t.outputAmount
}

func (t *Trade) ExecutionPrice() *core.Price {
	if t.executionPrice != nil {
		return t.executionPrice
	}

	t.executionPrice = core.NewPrice(
		t.InputAmount().Currency,
		t.OutputAmount().Currency,
		t.InputAmount().Quotient(),
		t.OutputAmount().Quotient(),
	)
	return t.executionPrice
}

func (t *Trade) PriceImpact() (*core.Percent, error) {
	if t.priceImpact != nil {
		return t.priceImpact, nil
	}

	spotOutputAmount := core.FromRawAmount(t.OutputAmount().Currency, big.NewInt(0))
	for _, swap := range t.Swaps {
		midPrice, err := swap.Route.MidPrice()
		if err != nil {
			return nil, err
		}
		quotePrice, err := midPrice.Quote(swap.InputAmount)
		if err != nil {
			return nil, err
		}
		spotOutputAmount = spotOutputAmount.Add(quotePrice)
	}

	priceImpact := spotOutputAmount.Subtract(t.OutputAmount()).Divide(spotOutputAmount.Fraction)
	t.priceImpact = core.NewPercent(priceImpact.Numerator, priceImpact.Denominator)
	return t.priceImpact, nil
}

func ExactIn(route *Route, amountIn *core.CurrencyAmount) (*Trade, error) {
	return FromRoute(route, amountIn, core.ExactInput)
}

func ExactOut(route *Route, amountOut *core.CurrencyAmount) (*Trade, error) {
	return FromRoute(route, amountOut, core.ExactOutput)
}

func FromRoute(route *Route, amount *core.CurrencyAmount, tradeType core.TradeType) (*Trade, error) {
	// amounts := make([]*core.CurrencyAmount, len(route.TokenPath))
	var (
		inputAmount  *core.CurrencyAmount
		outputAmount *core.CurrencyAmount
		// err          error
	)
	if tradeType == core.ExactInput {
		if !amount.Currency.Equal(route.Input) {
			return nil, ErrInvalidAmountForRoute
		}

		tokenAmount, err := AmountWithPathCurrency(amount, route.Pools[0])
		if err != nil {
			return nil, err
		}
		// amounts[0] = tokenAmount //amount.Wrapped()
		routeTokenPath := len(route.TokenPath)
		for i := 0; i < routeTokenPath; i++ {
			pool := route.Pools[i]
			outputAmountResult, err := pool.GetOutputAmount(tokenAmount, nil)
			if err != nil {
				return nil, err
			}
			outputAmount = outputAmountResult.ReturnedAmount
			tokenAmount = outputAmount
		}
		inputAmount = core.FromFractionalAmount(route.Input, amount.Numerator, amount.Denominator)
		outputAmount = core.FromFractionalAmount(route.Output, tokenAmount.Numerator, tokenAmount.Denominator)
	} else {
		if !amount.Currency.Equal(route.Output) {
			return nil, ErrInvalidAmountForRoute
		}
		tokenAmount, err := AmountWithPathCurrency(amount, route.Pools[len(route.Pools)-1])
		if err != nil {
			return nil, err
		}
		// amounts[len(amounts)-1] = amount.Wrapped()
		for i := len(route.TokenPath) - 1; i > 0; i-- {
			pool := route.Pools[i-1]
			inputAmountResult, err := pool.GetInputAmount(tokenAmount, nil)
			if err != nil {
				return nil, err
			}
			inputAmount = inputAmountResult.ReturnedAmount
			tokenAmount = inputAmount
		}
		inputAmount = core.FromFractionalAmount(route.Input, tokenAmount.Numerator, tokenAmount.Denominator)
		outputAmount = core.FromFractionalAmount(route.Output, amount.Numerator, amount.Denominator)
	}
	swaps := []*Swap{{
		Route:        route,
		InputAmount:  inputAmount,
		OutputAmount: outputAmount,
	}}
	// newTrade(swaps, tradeType)
	return &Trade{
		Swaps:     swaps,
		TradeType: tradeType,
	}, nil
}

type WrappedRoute struct {
	Amount *core.CurrencyAmount
	Route  *Route
}

func FromRoutes(wrappedRoutes []*WrappedRoute, tradeType core.TradeType) (*Trade, error) {
	swaps := make([]*Swap, len(wrappedRoutes))
	for i, wrappedRoute := range wrappedRoutes {
		route := wrappedRoute.Route
		amount := wrappedRoute.Amount
		trade, err := FromRoute(route, amount, tradeType)
		if err != nil {
			return nil, err
		}
		swaps[i] = trade.Swaps[0]
	}
	return &Trade{
		Swaps:     swaps,
		TradeType: tradeType,
	}, nil
}

func CreateUncheckedTrade(route *Route, amountIn, amountOut *core.CurrencyAmount, tradeType core.TradeType) (*Trade, error) {
	return &Trade{
		Swaps:     []*Swap{{Route: route, InputAmount: amountIn, OutputAmount: amountOut}},
		TradeType: tradeType,
	}, nil
}

func CreateUncheckedTradeWithMultipleRoutes(routes []*Swap, tradeType core.TradeType) (*Trade, error) {
	return newTrade(routes, tradeType)
}

func newTrade(routes []*Swap, tradeType core.TradeType) (*Trade, error) {
	inputCurrency := routes[0].InputAmount.Currency
	outputCurrency := routes[0].OutputAmount.Currency
	for _, route := range routes {
		if !inputCurrency.Wrapped().Equal(route.Route.Input.Wrapped()) {
			return nil, ErrInputCurrencyMismatch
		}
		if !outputCurrency.Wrapped().Equal(route.Route.Output.Wrapped()) {
			return nil, ErrOutputCurrencyMismatch
		}
	}

	var numPools int
	for _, route := range routes {
		numPools += len(route.Route.Pools)
	}

	var poolAddressSet = make(map[string]bool)
	for _, route := range routes {
		for _, p := range route.Route.Pools {
			poolid, err := GetPoolId(p.Currency0, p.Currency1, p.Fee, p.TickSpacing, p.Hooks)
			if err != nil {
				return nil, err
			}
			poolAddressSet[string(poolid)] = true
		}
	}
	if len(poolAddressSet) != numPools {
		return nil, ErrDuplicatePools
	}
	return &Trade{
		Swaps:     routes,
		TradeType: tradeType,
	}, nil
}

func (t *Trade) MininumAmountOut(slippageTolerance *core.Percent, amountOut *core.CurrencyAmount) (*core.CurrencyAmount, error) {
	if amountOut == nil {
		amountOut = t.OutputAmount()
	}
	if slippageTolerance.LessThan((v3constants.PercentZero)) {
		return nil, ErrInvalidSlippageTolerance
	}
	if t.TradeType == core.ExactOutput {
		return amountOut, nil
	} else {
		slippageAdjustedAmountOut := core.NewFraction(big.NewInt(1), big.NewInt(1)).
			Add(slippageTolerance.Fraction).
			Invert().
			Multiply(amountOut.Fraction).Quotient()

		return core.FromRawAmount(amountOut.Currency, slippageAdjustedAmountOut), nil
	}

}

func (t *Trade) MaximumAmountIn(slippageTolerance *core.Percent, amountIn *core.CurrencyAmount) (*core.CurrencyAmount, error) {
	if amountIn == nil {
		amountIn = t.InputAmount()
	}
	if slippageTolerance.LessThan((v3constants.PercentZero)) {
		return nil, ErrInvalidSlippageTolerance
	}
	if t.TradeType == core.ExactInput {
		return amountIn, nil
	} else {
		slippageAdjustedAmountIn := core.NewFraction(big.NewInt(1), big.NewInt(1)).
			Add(slippageTolerance.Fraction).
			Multiply(amountIn.Fraction).Quotient()

		return core.FromRawAmount(amountIn.Currency, slippageAdjustedAmountIn), nil
	}
}

func (t *Trade) WorstExecutionPrice(slippageTolerance *core.Percent) (*core.Price, error) {
	maxAmountIn, err := t.MaximumAmountIn(slippageTolerance, nil)
	if err != nil {
		return nil, err
	}
	minAmountOut, err := t.MininumAmountOut(slippageTolerance, nil)
	if err != nil {
		return nil, err
	}
	return core.NewPrice(t.InputAmount().Currency, t.OutputAmount().Currency, maxAmountIn.Quotient(), minAmountOut.Quotient()), nil
}

func BestTradeExactIn(
	pools []*Pool, currencyAmountIn, currencyAmountOut *core.CurrencyAmount,
	currencyOut core.Currency, opts *BestTradeOptions,
	currentPools []*Pool, nextAmountIn *core.CurrencyAmount, bestTrades []*Trade,
) ([]*Trade, error) {
	if len(pools) <= 0 {
		return nil, ErrNoPools
	}
	if opts == nil {
		opts = &BestTradeOptions{MaxNumResults: 3, MaxHops: 3}
	}
	tokenOut := currencyOut.Wrapped()
	if nextAmountIn == nil {
		return nil, ErrInvalidMaxHops
	}
	if !(currencyAmountIn.EqualTo(nextAmountIn.Fraction) ||
		len(currentPools) > 0) {
		return nil, ErrInvalidRecursion
	}

	amountIn := nextAmountIn.Wrapped()
	for i := 0; i < len(pools); i++ {
		pool := pools[i]
		if !pool.Currency0.Equal(amountIn.Currency) &&
			!pool.Currency1.Equal(amountIn.Currency) {
			continue
		}
		outputAmountResult, err := pool.GetOutputAmount(amountIn, nil)
		if err != nil {
			return nil, err
		}
		amountOut := outputAmountResult.ReturnedAmount
		if amountOut.Currency.Equal(tokenOut) && amountOut.Currency.IsToken() {
			r, err := NewRoute(append(currentPools, pool), currencyAmountIn.Currency, currencyOut)
			if err != nil {
				return nil, err
			}
			trade, err := FromRoute(r, currencyAmountIn, core.ExactInput)
			if err != nil {
				return nil, err
			}
			bestTrades, err = sortedInsert(bestTrades, trade, opts.MaxNumResults, tradeComparator)
			if err != nil {
				return nil, err
			}
		} else if opts.MaxHops > 1 && len(pools) > 1 {
			var poolsExcludingThisPool []*Pool
			poolsExcludingThisPool = append(poolsExcludingThisPool, pools[:i]...)
			poolsExcludingThisPool = append(poolsExcludingThisPool, pools[i+1:]...)

			// otherwise, consider all the other paths that lead from this token as long as we have not exceeded maxHops
			bestTrades, err = BestTradeExactIn(
				poolsExcludingThisPool,
				currencyAmountIn, currencyAmountOut, currencyOut,
				&BestTradeOptions{
					MaxNumResults: opts.MaxNumResults,
					MaxHops:       opts.MaxHops - 1},
				append(currentPools, pool),
				amountOut, bestTrades)
			if err != nil {
				return nil, err
			}
		}
	}
	return bestTrades, nil
}

func BestTradeExactOut(pools []*Pool, currencyIn core.Currency,
	currencyAmountOut, currencyAmountIn *core.CurrencyAmount,
	opts *BestTradeOptions,
	currentPools []*Pool,
	nextAmountOut *core.CurrencyAmount,
	bestTrades []*Trade,
) ([]*Trade, error) {
	if len(pools) <= 0 {
		return nil, ErrNoPools
	}
	if opts == nil {
		opts = &BestTradeOptions{MaxNumResults: 3, MaxHops: 3}
	}
	// tokenIn := currencyIn.Wrapped()
	if nextAmountOut == nil {
		nextAmountOut = currencyAmountOut
	}
	if opts.MaxHops <= 0 {
		return nil, ErrInvalidMaxHops
	}
	if !(currencyAmountOut.EqualTo(nextAmountOut.Fraction) || len(currentPools) > 0) {
		return nil, ErrInvalidRecursion
	}

	amountOut := nextAmountOut.Wrapped()
	for i := 0; i < len(pools); i++ {
		pool := pools[i]
		// pool irrelevant
		if !pool.Currency0.Equal(amountOut.Currency) && !pool.Currency1.Equal(amountOut.Currency) {
			continue
		}
		inputAmountResult, err := pool.GetInputAmount(amountOut, nil)
		if err != nil {
			return nil, err
		}
		amountIn := inputAmountResult.ReturnedAmount
		// we have arrived at the input token, so this is the final trade of one of the paths
		if amountIn.Currency.Equal(currencyIn) {
			r, err := NewRoute(append([]*Pool{pool}, currentPools...), currencyIn, currencyAmountOut.Currency)
			if err != nil {
				return nil, err
			}
			trade, err := FromRoute(r, currencyAmountOut, core.ExactOutput)
			if err != nil {
				return nil, err
			}
			bestTrades, err = sortedInsert(bestTrades, trade, opts.MaxNumResults, tradeComparator)
			if err != nil {
				return nil, err
			}
		} else if opts.MaxHops > 1 && len(pools) > 1 {
			var poolsExcludingThisPool []*Pool
			poolsExcludingThisPool = append(poolsExcludingThisPool, pools[:i]...)
			poolsExcludingThisPool = append(poolsExcludingThisPool, pools[i+1:]...)

			// otherwise, consider all the other paths that arrive at this token as long as we have not exceeded maxHops
			bestTrades, err = BestTradeExactOut(poolsExcludingThisPool, currencyIn,
				currencyAmountOut, currencyAmountIn,
				&BestTradeOptions{
					MaxNumResults: opts.MaxNumResults,
					MaxHops:       opts.MaxHops - 1,
				},
				append([]*Pool{pool}, currentPools...), amountIn, bestTrades)
			if err != nil {
				return nil, err
			}
		}
	}
	return bestTrades, nil
}

func sortedInsert(items []*Trade, add *Trade, maxSize int, comparator func(a, b *Trade) int) ([]*Trade, error) {
	if maxSize <= 0 {
		return nil, ErrInvalidMaxSize
	}
	if len(items) > maxSize {
		return nil, ErrMaxSizeExceeded
	}

	isFull := len(items) == maxSize
	if isFull && comparator(items[len(items)-1], add) <= 0 {
		return items, nil
	}

	i := sort.Search(len(items), func(i int) bool {
		return comparator(add, items[i]) > 0
	})

	items = append(items, nil)
	copy(items[i+1:], items[i:])
	items[i] = add
	if isFull {
		return items[:maxSize], nil
	}
	return items, nil
}
