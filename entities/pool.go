package entities

import (
	"errors"
	"math/big"

	v3constants "github.com/KyberNetwork/pancake-v3-sdk/constants"
	v3sdk "github.com/KyberNetwork/pancake-v3-sdk/entities"
	v3utils "github.com/KyberNetwork/pancake-v3-sdk/utils"
	"github.com/dangthanhduong01/uniswapv4-sdk/constants"
	v4utils "github.com/dangthanhduong01/uniswapv4-sdk/utils"
	core "github.com/daoleno/uniswap-sdk-core/entities"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	ErrFeeTooHigh         = errors.New("fee too high")
	ErrInvalidSqrtRatio96 = errors.New("invalid sqrtRatioX96")
	ErrHookNotEnabled     = errors.New("hook not enabled")
)

type StepComputations struct {
	SqrtPriceStartX96 *big.Int `json:"sqrtPriceStartX96"`
	TickNext          int      `json:"tickNext"`
	Initialized       bool     `json:"initialized"`
	SqrtPriceNextX96  *big.Int `json:"sqrtPriceNextX96"`
	AmountIn          *big.Int `json:"amountIn"`
	AmountOut         *big.Int `json:"amountOut"`
	FeeAmount         *big.Int `json:"feeAmount"`
}

type PoolKey struct {
	Currency0   common.Address
	Currency1   common.Address
	Fee         int64
	TickSpacing int64
	Hooks       common.Address
}

type Pool struct {
	Currency0        *core.Token
	Currency1        *core.Token
	Fee              int64
	TickSpacing      int64
	SqrtRatioX96     *big.Int
	Hooks            common.Address
	Liquidity        *big.Int
	TickCurrent      int
	TickDataProvider v3sdk.TickDataProvider
	PoolKey          PoolKey
	PoolId           []byte

	token0Price *core.Price
	token1Price *core.Price
}

type SwapResult struct {
	AmountCalculated      *big.Int
	SqrtRatioX96          *big.Int
	Liquidity             *big.Int
	RemainingTargetAmount *big.Int
	CurrentTick           int
	CrossInitTickLoops    int
}

type GetOutputAmountResult struct {
	ReturnedAmount     *core.CurrencyAmount
	RemainingAmountIn  *core.CurrencyAmount
	NewPoolState       *Pool
	CrossInitTickLoops int
}

type GetInputAmountResult struct {
	ReturnedAmount     *core.CurrencyAmount
	RemainingAmountOut *core.CurrencyAmount
	NewPoolState       *Pool
	CrossInitTickLoops int
}

func GetPoolKey(currencyA, currencyB *core.Token,
	fee int64, tickSpacing int64, hooks common.Address) (*PoolKey, error) {
	token0 := currencyA
	token1 := currencyB
	isSorted, err := currencyA.SortsBefore(currencyB)
	if err != nil {
		return nil, err
	}
	if !isSorted {
		token0 = currencyB
		token1 = currencyA
	}

	currency0Addr := constants.AddressZero
	currency1Addr := constants.AddressZero

	if !token0.IsNative() {
		currency0Addr = token0.Wrapped().Address
	}
	if !token1.IsNative() {
		currency1Addr = token1.Wrapped().Address
	}

	return &PoolKey{
		currency0Addr,
		currency1Addr,
		fee,
		tickSpacing,
		hooks,
	}, nil
}

func GetPoolId(currencyA *core.Token, currencyB *core.Token,
	fee int64, tickSpacing int64, hooks common.Address) ([]byte, error) {
	isSorted, err := currencyA.SortsBefore(currencyB)
	if err != nil {
		return nil, err
	}
	token0 := currencyA
	token1 := currencyB
	if !isSorted {
		token0 = currencyB
		token1 = currencyA
	}

	currency0Addr := constants.AddressZero
	currency1Addr := constants.AddressZero

	if !token0.IsNative() {
		currency0Addr = token0.Wrapped().Address
	}
	if !token1.IsNative() {
		currency1Addr = token1.Wrapped().Address
	}

	salt := crypto.Keccak256(abiEncode(currency0Addr, currency1Addr, fee, tickSpacing, hooks))

	return salt, nil
}

func abiEncode(addressA, addressB common.Address, fee int64, tickSpacing int64, hooks common.Address) []byte {
	addressTy, _ := abi.NewType("address", "address", nil)
	uint256Ty, _ := abi.NewType("uint256", "uint256", nil)

	arguments := abi.Arguments{
		{Type: addressTy},
		{Type: addressTy},
		{Type: uint256Ty},
		{Type: uint256Ty},
		{Type: addressTy},
	}

	bytes, _ := arguments.Pack(
		addressA,
		addressB,
		big.NewInt(fee),
		tickSpacing,
		hooks,
	)
	return bytes
}

func NewPool(tokenA, tokenB *core.Token,
	fee int64, tickSpacing int64,
	hooks common.Address, sqrtRatioX96 *big.Int, liquidity *big.Int,
	tickCurrent int, ticks v3sdk.TickDataProvider) (*Pool, error) {
	if fee >= 1000000 {
		return nil, ErrFeeTooHigh
	}
	tickCurrentSqrtRatioX96, err := v3utils.GetSqrtRatioAtTick(tickCurrent)
	if err != nil {
		return nil, err
	}
	nextTickSqrtRatioX96, err := v3utils.GetSqrtRatioAtTick(tickCurrent + 1)
	if err != nil {
		return nil, err
	}

	if sqrtRatioX96.Cmp(tickCurrentSqrtRatioX96) < 0 || sqrtRatioX96.Cmp(nextTickSqrtRatioX96) > 0 {
		return nil, ErrInvalidSqrtRatio96
	}

	token0 := tokenA
	token1 := tokenB
	isSorted, err := tokenA.SortsBefore(tokenB)
	if err != nil {
		return nil, err
	}
	if !isSorted {
		token0 = tokenB
		token1 = tokenA
	}
	poolKey, err := GetPoolKey(token0, token1, fee, tickSpacing, hooks)
	if err != nil {
		return nil, err
	}
	poolId, err := GetPoolId(token0, token1, fee, tickSpacing, hooks)
	if err != nil {
		return nil, err
	}

	return &Pool{
		Currency0:        token0,
		Currency1:        token1,
		Fee:              fee,
		SqrtRatioX96:     sqrtRatioX96,
		Liquidity:        liquidity,
		TickCurrent:      tickCurrent,
		TickDataProvider: ticks,
		Hooks:            hooks,
		PoolKey:          *poolKey,
		PoolId:           poolId,
	}, nil
}

func (p *Pool) Token0() *core.Token {
	return p.Currency0
}

func (p *Pool) Token1() *core.Token {
	return p.Currency1
}

/**
 * Returns true if the currency is either currency0 or currency1
 * @param currency The currency to check
 * @returns True if currency is either currency0 or currency1
 */
func (p *Pool) InvolvesToken(currency *core.Token) bool {
	return p.Currency0.Equal(currency) || p.Currency1.Equal(currency)
}

/**
 * v4-only involvesToken convenience method, used for mixed route ETH <-> WETH connection only
 * @param currency
 */
func (p *Pool) V4InvolvesToken(currency *core.Token) bool {
	return p.InvolvesToken(currency) ||
		currency.Wrapped().Equal(p.Currency0) ||
		currency.Wrapped().Equal(p.Currency1) ||
		currency.Wrapped().Equal(p.Currency0.Wrapped()) ||
		currency.Wrapped().Equal(p.Currency1.Wrapped())
}

// func (p *Pool) Currency0Price() *core.Price {
// 	if (p.token0Price != nil) {
// }

func (p *Pool) Token0Price() *core.Price {
	if p.token0Price != nil {
		return p.token0Price
	}
	p.token0Price = core.NewPrice(p.Currency0, p.Currency1, v3constants.Q192, new(big.Int).Mul(p.SqrtRatioX96, p.SqrtRatioX96))
	return p.token0Price
}

func (p *Pool) Token1Price() *core.Price {
	if p.token1Price != nil {
		return p.token1Price
	}
	p.token1Price = core.NewPrice(p.Currency1, p.Currency0, new(big.Int).Mul(p.SqrtRatioX96, p.SqrtRatioX96), v3constants.Q192)
	return p.token1Price
}

func (p *Pool) PriceOf(token *core.Token) (*core.Price, error) {
	if !p.InvolvesToken(token) {
		return nil, v3sdk.ErrTokenNotInvolved
	}
	if p.Currency0.Equal(token) {
		return p.Token0Price(), nil
	}
	return p.Token1Price(), nil
}

func (p *Pool) ChainID() uint {
	return p.Currency0.ChainId()
}

func (p *Pool) GetOutputAmount(inputAmount *core.CurrencyAmount, sqrtPriceLimitX96 *big.Int) (*GetOutputAmountResult, error) {
	if !(inputAmount.Currency.IsToken() && p.InvolvesToken(inputAmount.Currency.Wrapped())) {
		return nil, v3sdk.ErrTokenNotInvolved
	}
	zeroForOne := inputAmount.Currency.Equal(p.Currency0)

	swapResult, err := p.swap(zeroForOne, inputAmount.Quotient(), sqrtPriceLimitX96)
	if err != nil {
		return nil, err
	}
	var outputToken *core.Token
	if zeroForOne {
		outputToken = p.Currency1
	} else {
		outputToken = p.Currency0
	}

	pool, err := NewPool(
		p.Currency0,
		p.Currency1,
		p.Fee,
		p.TickSpacing,
		p.Hooks,
		swapResult.SqrtRatioX96,
		swapResult.Liquidity,
		swapResult.CurrentTick,
		p.TickDataProvider,
	)
	if err != nil {
		return nil, err
	}

	return &GetOutputAmountResult{
		ReturnedAmount:     core.FromRawAmount(outputToken, new(big.Int).Mul(swapResult.AmountCalculated, v3constants.NegativeOne)),
		RemainingAmountIn:  core.FromRawAmount(inputAmount.Currency, swapResult.RemainingTargetAmount),
		NewPoolState:       pool,
		CrossInitTickLoops: swapResult.CrossInitTickLoops,
	}, nil
}

func (p *Pool) GetInputAmount(outputAmount *core.CurrencyAmount, sqrtPriceLimitX96 *big.Int) (*GetInputAmountResult, error) {
	if !(outputAmount.Currency.IsToken() && p.InvolvesToken(outputAmount.Currency.Wrapped())) {
		return nil, v3sdk.ErrTokenNotInvolved
	}
	zeroForOne := outputAmount.Currency.Equal(p.Currency1)
	swapResult, err := p.swap(zeroForOne, new(big.Int).Mul(outputAmount.Quotient(), v3constants.NegativeOne), sqrtPriceLimitX96)
	if err != nil {
		return nil, err
	}

	var inputToken *core.Token
	if zeroForOne {
		inputToken = p.Currency0
	} else {
		inputToken = p.Currency1
	}

	pool, err := NewPool(
		p.Currency0,
		p.Currency1,
		p.Fee,
		p.TickSpacing,
		p.Hooks,
		swapResult.SqrtRatioX96,
		swapResult.Liquidity,
		swapResult.CurrentTick,
		p.TickDataProvider,
	)
	if err != nil {
		return nil, err
	}

	// return core.FromRawAmount(inputToken, swapResult.AmountCalculated), pool, nil
	return &GetInputAmountResult{
		ReturnedAmount:     core.FromRawAmount(inputToken, swapResult.AmountCalculated),
		RemainingAmountOut: core.FromRawAmount(outputAmount.Currency, swapResult.RemainingTargetAmount),
		NewPoolState:       pool,
		CrossInitTickLoops: swapResult.CrossInitTickLoops,
	}, nil
}

func (p *Pool) swap(zeroForOne bool, amountSpecified, sqrtPriceLimitX96 *big.Int) (*SwapResult, error) {
	if !p.hookImpactsSwap() {
		return p.v3Swap(zeroForOne, amountSpecified, sqrtPriceLimitX96)

	} else {
		return nil, errors.New("unsupported hook")
	}
}

func (p *Pool) v3Swap(zeroForOne bool, amountSpecified, sqrtPriceLimitX96 *big.Int) (*SwapResult, error) {
	var err error
	if sqrtPriceLimitX96 == nil {
		if zeroForOne {
			sqrtPriceLimitX96 = new(big.Int).Add(v3utils.MinSqrtRatio, v3constants.One)
		} else {
			sqrtPriceLimitX96 = new(big.Int).Sub(v3utils.MaxSqrtRatio, v3constants.One)
		}
	}

	if zeroForOne {
		if sqrtPriceLimitX96.Cmp(v3utils.MinSqrtRatio) < 0 {
			return nil, v3sdk.ErrSqrtPriceLimitX96TooLow
		}
		if sqrtPriceLimitX96.Cmp(p.SqrtRatioX96) >= 0 {
			return nil, v3sdk.ErrSqrtPriceLimitX96TooHigh
		}
	} else {
		if sqrtPriceLimitX96.Cmp(v3utils.MaxSqrtRatio) > 0 {
			return nil, v3sdk.ErrSqrtPriceLimitX96TooHigh
		}
		if sqrtPriceLimitX96.Cmp(p.SqrtRatioX96) <= 0 {
			return nil, v3sdk.ErrSqrtPriceLimitX96TooLow
		}
	}

	exactInput := amountSpecified.Cmp(v3constants.Zero) >= 0

	state := struct {
		amountSpecifiedRemaining *big.Int
		amountCalculated         *big.Int
		sqrtPriceX96             *big.Int
		tick                     int
		liquidity                *big.Int
	}{
		amountSpecifiedRemaining: amountSpecified,
		amountCalculated:         v3constants.Zero,
		sqrtPriceX96:             p.SqrtRatioX96,
		tick:                     p.TickCurrent,
		liquidity:                p.Liquidity,
	}

	// crossInitTickLoops is the number of loops that cross an initialized tick.
	// We only count when tick passes an initialized tick, since gas only significant in this case.
	crossInitTickLoops := 0

	// start swap while loop
	for state.amountSpecifiedRemaining.Cmp(v3constants.Zero) != 0 && state.sqrtPriceX96.Cmp(sqrtPriceLimitX96) != 0 {
		var step StepComputations
		step.SqrtPriceStartX96 = state.sqrtPriceX96

		// because each iteration of the while loop rounds, we can't optimize this code (relative to the smart contract)
		// by simply traversing to the next available tick, we instead need to exactly replicate
		// tickBitmap.nextInitializedTickWithinOneWord
		// step.TickNext, step.Initialized = p.TickDataProvider.NextInitializedTickWithinOneWord(state.tick, zeroForOne, p.tickSpacing())
		step.TickNext, step.Initialized, err = p.TickDataProvider.NextInitializedTickIndex(state.tick, zeroForOne)
		if err != nil {
			return nil, err
		}

		if step.TickNext < v3utils.MinTick {
			step.TickNext = v3utils.MinTick
		} else if step.TickNext > v3utils.MaxTick {
			step.TickNext = v3utils.MaxTick
		}

		step.SqrtPriceNextX96, err = v3utils.GetSqrtRatioAtTick(step.TickNext)
		if err != nil {
			return nil, err
		}
		var targetValue *big.Int
		if zeroForOne {
			if step.SqrtPriceNextX96.Cmp(sqrtPriceLimitX96) < 0 {
				targetValue = sqrtPriceLimitX96
			} else {
				targetValue = step.SqrtPriceNextX96
			}
		} else {
			if step.SqrtPriceNextX96.Cmp(sqrtPriceLimitX96) > 0 {
				targetValue = sqrtPriceLimitX96
			} else {
				targetValue = step.SqrtPriceNextX96
			}
		}

		state.sqrtPriceX96, step.AmountIn, step.AmountOut, step.FeeAmount, err = v3utils.ComputeSwapStep(state.sqrtPriceX96, targetValue, state.liquidity, state.amountSpecifiedRemaining, v3constants.FeeAmount(p.Fee))
		if err != nil {
			return nil, err
		}

		if exactInput {
			state.amountSpecifiedRemaining = new(big.Int).Sub(state.amountSpecifiedRemaining, new(big.Int).Add(step.AmountIn, step.FeeAmount))
			state.amountCalculated = new(big.Int).Sub(state.amountCalculated, step.AmountOut)
		} else {
			state.amountSpecifiedRemaining = new(big.Int).Add(state.amountSpecifiedRemaining, step.AmountOut)
			state.amountCalculated = new(big.Int).Add(state.amountCalculated, new(big.Int).Add(step.AmountIn, step.FeeAmount))
		}

		// TODO
		if state.sqrtPriceX96.Cmp(step.SqrtPriceNextX96) == 0 {
			// if the tick is initialized, run the tick transition
			if step.Initialized {
				tick, err := p.TickDataProvider.GetTick(step.TickNext)
				if err != nil {
					return nil, err
				}
				liquidityNet := tick.LiquidityNet
				// if we're moving leftward, we interpret liquidityNet as the opposite sign
				// safe because liquidityNet cannot be type(int128).min
				if zeroForOne {
					liquidityNet = new(big.Int).Mul(liquidityNet, v3constants.NegativeOne)
				}
				state.liquidity = v3utils.AddDelta(state.liquidity, liquidityNet)

				crossInitTickLoops++
			}
			if zeroForOne {
				state.tick = step.TickNext - 1
			} else {
				state.tick = step.TickNext
			}
		} else if state.sqrtPriceX96.Cmp(step.SqrtPriceStartX96) != 0 {
			// recompute unless we're on a lower tick boundary (i.e. already transitioned ticks), and haven't moved
			state.tick, err = v3utils.GetTickAtSqrtRatio(state.sqrtPriceX96)
			if err != nil {
				return nil, err
			}
		}
	}

	return &SwapResult{
		AmountCalculated:      state.amountCalculated,
		SqrtRatioX96:          state.sqrtPriceX96,
		Liquidity:             state.liquidity,
		CurrentTick:           state.tick,
		RemainingTargetAmount: state.amountSpecifiedRemaining,
		CrossInitTickLoops:    crossInitTickLoops,
	}, nil
}

func (p *Pool) hookImpactsSwap() bool {
	var hook v4utils.Hook
	permission, err := (&hook).HasSwapPermissions(p.Hooks.Hex())
	if err != nil {
		return false
	}
	return permission
}
