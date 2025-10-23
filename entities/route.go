package entities

import (
	"errors"

	core "github.com/daoleno/uniswap-sdk-core/entities"
)

var (
	ErrRouteNoPools      = errors.New("route must have at least one pool")
	ErrAllOnSameChain    = errors.New("all pools must be on the same chain")
	ErrInputNotInvolved  = errors.New("input token not involved in route")
	ErrOutputNotInvolved = errors.New("output token not involved in route")
	ErrPathNotContinuous = errors.New("path not continuous")
)

type Route struct {
	Pools     []*Pool
	TokenPath []*core.Token
	Input     core.Currency
	Output    core.Currency

	PathInput  core.Currency
	PathOutput core.Currency

	midPrice *core.Price
}

func NewRoute(pools []*Pool, input, output core.Currency) (*Route, error) {
	if len(pools) == 0 {
		return nil, ErrRouteNoPools
	}
	chainId := pools[0].ChainID()
	for _, pool := range pools {
		if pool.ChainID() != chainId {
			return nil, ErrAllOnSameChain
		}
	}

	pathInput := getPathCurrency(input.Wrapped(), pools[0])
	pathOutput := getPathCurrency(output.Wrapped(), pools[len(pools)-1])

	currencyPath := []*core.Token{pathInput}

	for i, pool := range pools {
		currencyInputCurrency := currencyPath[i]
		if currencyInputCurrency.Equal(pool.Currency0) || currencyInputCurrency.Equal(pool.Currency1) {
			return nil, ErrPathNotContinuous
		}
		nextCurrency := pool.Currency0
		if currencyInputCurrency.Equal(pool.Currency0) {
			nextCurrency = pool.Currency1
		}

		currencyPath = append(currencyPath, nextCurrency)
	}

	if output == nil {
		output = currencyPath[len(currencyPath)-1]
	} else {
		if !pools[len(pools)-1].InvolvesToken(output.Wrapped()) {
			return nil, ErrOutputNotInvolved
		}
	}

	return &Route{
		Pools:      pools,
		TokenPath:  currencyPath,
		Input:      input,
		Output:     output,
		PathInput:  pathInput,
		PathOutput: pathOutput,
	}, nil
}

func (r *Route) ChainID() uint {
	return r.Pools[0].ChainID()
}

func (r *Route) MidPrice() (*core.Price, error) {
	if r.midPrice != nil {
		return r.midPrice, nil
	}

	var (
		nextInput *core.Token
		price     *core.Price
	)
	if r.Pools[0].Currency0.Equal(r.Input) {
		nextInput = r.Pools[0].Currency1
		price = r.Pools[0].Token0Price()
	} else {
		nextInput = r.Pools[0].Currency0
		price = r.Pools[0].Token1Price()
	}
	price, err := reducePrice(nextInput, price, r.Pools[1:])
	if err != nil {
		return nil, err
	}

	r.midPrice = core.NewPrice(r.Input, r.Output, price.Denominator, price.Numerator)
	return r.midPrice, nil
}

func reducePrice(nextInput *core.Token, price *core.Price, pools []*Pool) (*core.Price, error) {
	var err error
	for _, p := range pools {
		if nextInput.Equal(p.Currency0) {
			nextInput = p.Currency1
			price, err = price.Multiply(p.Token0Price())
			if err != nil {
				return nil, err
			}
		} else {
			nextInput = p.Currency0
			price, err = price.Multiply(p.Token1Price())
			if err != nil {
				return nil, err
			}
		}

	}
	return price, nil
}
