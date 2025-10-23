package entities

import (
	"errors"
	"math/big"

	v3constants "github.com/KyberNetwork/pancake-v3-sdk/constants"
	"github.com/dangthanhduong01/uniswapv4-sdk/constants"
	core "github.com/daoleno/uniswap-sdk-core/entities"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Actions byte

// Các hằng số mô phỏng enum Actionss với giá trị hex đã xác định.
const (
	// pool actions
	// liquidity actions
	INCREASE_LIQUIDITY Actions = 0x00
	DECREASE_LIQUIDITY Actions = 0x01
	MINT_POSITION      Actions = 0x02
	BURN_POSITION      Actions = 0x03

	// for fee on transfer tokens
	// INCREASE_LIQUIDITY_FROM_DELTAS = 0x04,
	// MINT_POSITION_FROM_DELTAS = 0x05,

	// swapping
	SWAP_EXACT_IN_SINGLE  Actions = 0x06
	SWAP_EXACT_IN         Actions = 0x07
	SWAP_EXACT_OUT_SINGLE Actions = 0x08
	SWAP_EXACT_OUT        Actions = 0x09

	// settling
	SETTLE      Actions = 0x0b
	SETTLE_ALL  Actions = 0x0c
	SETTLE_PAIR Actions = 0x0d

	// taking
	TAKE         Actions = 0x0e
	TAKE_ALL     Actions = 0x0f
	TAKE_PORTION Actions = 0x10
	TAKE_PAIR    Actions = 0x11

	CLOSE_CURRENCY Actions = 0x12
	// CLEAR_OR_TAKE = 0x13,
	SWEEP Actions = 0x14

	// for wrapping/unwrapping native
	// WRAP = 0x15,
	UNWRAP Actions = 0x16
)

type Subparser int

const (
	V4SwapExactInSingle  Subparser = iota // 0
	V4SwapExactIn                         // 1
	V4SwapExactOutSingle                  // 2
	V4SwapExactOut                        // 3
	Poolkey                               // 4
)

type ParamType struct {
	Name      string
	Type      string
	Subparser Subparser
}

const POOL_KEY_STRUCT = "(address currency0,address currency1,uint24 fee,int24 tickSpacing,address hooks)"
const PATH_KEY_STRUCT = "(address intermediateCurrency,uint256 fee,int24 tickSpacing,address hooks,bytes hookData)"

const SWAP_EXACT_IN_SINGLE_STRUCT = "(" + POOL_KEY_STRUCT + " poolKey,bool zeroForOne,uint128 amountIn,uint128 amountOutMinimum,bytes hookData)"

const SWAP_EXACT_IN_STRUCT = "(address currencyIn," + PATH_KEY_STRUCT + "[] path,uint128 amountIn,uint128 amountOutMinimum)"

const SWAP_EXACT_OUT_SINGLE_STRUCT = "(" + POOL_KEY_STRUCT + " poolKey,bool zeroForOne,uint128 amountOut,uint128 amountInMaximum,bytes hookData)"

const SWAP_EXACT_OUT_STRUCT = "(address currencyOut," + PATH_KEY_STRUCT + "[] path,uint128 amountOut,uint128 amountInMaximum)"

var V4_BASE_ACTIONS_ABI_DEFINITION = map[Actions][]ParamType{
	INCREASE_LIQUIDITY: {
		{Name: "tokenId", Type: "uint256"},
		{Name: "liquidity", Type: "uint256"},
		{Name: "amount0Max", Type: "uint128"},
		{Name: "amount1Max", Type: "uint128"},
		{Name: "hookData", Type: "bytes"},
	},
	DECREASE_LIQUIDITY: {
		{Name: "tokenId", Type: "uint256"},
		{Name: "liquidity", Type: "uint256"},
		{Name: "amount0Min", Type: "uint128"},
		{Name: "amount1Min", Type: "uint128"},
		{Name: "hookData", Type: "bytes"},
	},
	MINT_POSITION: {
		{Name: "poolKey", Type: POOL_KEY_STRUCT, Subparser: Poolkey},
		{Name: "tickLower", Type: "int24"},
		{Name: "tickUpper", Type: "int24"},
		{Name: "liquidity", Type: "uint256"},
		{Name: "amount0Max", Type: "uint128"},
		{Name: "amount1Max", Type: "uint128"},
		{Name: "owner", Type: "address"},
		{Name: "hookData", Type: "bytes"},
	},
	BURN_POSITION: {
		{Name: "tokenId", Type: "uint256"},
		{Name: "amount0Min", Type: "uint128"},
		{Name: "amount1Min", Type: "uint128"},
		{Name: "hookData", Type: "bytes"},
	},

	// swapping commands
	SWAP_EXACT_IN_SINGLE: {
		{Name: "swap", Type: SWAP_EXACT_IN_SINGLE_STRUCT, Subparser: V4SwapExactInSingle},
	},
	SWAP_EXACT_IN: {
		{Name: "swap", Type: SWAP_EXACT_IN_STRUCT, Subparser: V4SwapExactIn},
	},
	SWAP_EXACT_OUT_SINGLE: {
		{Name: "swap", Type: SWAP_EXACT_OUT_SINGLE_STRUCT, Subparser: V4SwapExactOutSingle},
	},
	SWAP_EXACT_OUT: {
		{Name: "swap", Type: SWAP_EXACT_OUT_STRUCT, Subparser: V4SwapExactOut},
	},
	SETTLE: {
		{Name: "currency", Type: "address"},
		{Name: "amount", Type: "uint256"},
		{Name: "payerIsUser", Type: "bool"},
	},
	SETTLE_ALL: {
		{Name: "currency", Type: "address"},
		{Name: "maxAmount", Type: "uint256"},
	},
	SETTLE_PAIR: {
		{Name: "currency0", Type: "address"},
		{Name: "currency1", Type: "address"},
	},
	TAKE: {
		{Name: "currency", Type: "address"},
		{Name: "recipient", Type: "address"},
		{Name: "amount", Type: "uint256"},
	},
	TAKE_ALL: {
		{Name: "currency", Type: "address"},
		{Name: "minAmount", Type: "uint256"},
	},
	TAKE_PORTION: {
		{Name: "currency", Type: "address"},
		{Name: "recipient", Type: "address"},
		{Name: "bips", Type: "uint256"},
	},
	TAKE_PAIR: {
		{Name: "currency0", Type: "address"},
		{Name: "currency1", Type: "address"},
		{Name: "recipient", Type: "address"},
	},
	CLOSE_CURRENCY: {
		{Name: "currency", Type: "address"},
	},
	SWEEP: {
		{Name: "currency", Type: "address"},
		{Name: "recipient", Type: "address"},
	},
	UNWRAP: {
		{Name: "amount", Type: "uint256"},
	},
}

const FULL_DELTA_AMOUNT = 0

type V4Planner struct {
	Actions []byte
	Params  [][]byte
}

func NewV4Planner() *V4Planner {
	return &V4Planner{
		Actions: []byte(constants.EmptyBytes),
		Params:  [][]byte{},
	}
}

func (p *V4Planner) AddActions(typ Actions, parameters []interface{}) (*V4Planner, error) {
	command, err := createAction(typ, parameters)
	if err != nil {
		return nil, err
	}

	p.Params = append(p.Params, command.encodedInput)
	p.Actions = append(p.Actions, byte(command.action))
	return p, nil
}

func (p *V4Planner) AddTrade(trade Trade, slippageTolerance *core.Percent) (*V4Planner, error) {
	exactOutput := true
	if trade.TradeType == core.ExactOutput {
		exactOutput = false
	}
	if exactOutput {
		if slippageTolerance == nil || slippageTolerance.LessThan(v3constants.PercentZero) {
			return nil, ErrInvalidSlippageTolerance
		}
	}
	if len(trade.Swaps) != 1 {
		return nil, errors.New("only accepts Trades with 1 swap (must break swaps into individual trades)")
	}
	actionType := SWAP_EXACT_IN
	if exactOutput {
		actionType = SWAP_EXACT_OUT
	}

	r, err := trade.Route()
	if err != nil {
		return nil, err
	}

	currencyIn := currencyAddress(r.PathInput)
	currencyOut := currencyAddress(r.PathOutput)

	if !exactOutput {
		encoded, err := EncodeRouteToPath(r, exactOutput)
		if err != nil {
			return nil, err
		}
		amountInMax, err := trade.MaximumAmountIn(slippageTolerance, nil)
		if err != nil {
			return nil, err
		}
		act, err := p.AddActions(actionType, []interface{}{
			currencyOut,
			encoded,
			amountInMax,
			trade.outputAmount.Quotient().String(),
		})
		if err != nil {
			return nil, err
		}
		return act, nil
	} else {
		encoded, err := EncodeRouteToPath(r, exactOutput)
		if err != nil {
			return nil, err
		}
		amountOutMin, err := trade.MininumAmountOut(slippageTolerance, nil)
		if err != nil {
			return nil, err
		}
		act, err := p.AddActions(actionType, []interface{}{
			currencyIn,
			encoded,
			trade.inputAmount.Quotient().String(),
			amountOutMin,
		})
		if err != nil {
			return nil, err
		}
		return act, nil
	}

}

func (p *V4Planner) AddSettle(currency *core.Currency, payerIsUser bool, amount *big.Int) (*V4Planner, error) {
	if amount == nil {
		amount = big.NewInt(FULL_DELTA_AMOUNT)
	}
	act, err := p.AddActions(SETTLE, []interface{}{currencyAddress(*currency), amount, payerIsUser})
	return act, err
}

func (p *V4Planner) AddTake(currency *core.Currency, recipient common.Address, amount *big.Int) (*V4Planner, error) {
	takeAmount := amount
	if amount == nil {
		takeAmount = big.NewInt(FULL_DELTA_AMOUNT)
	}
	return p.AddActions(TAKE, []interface{}{currencyAddress(*currency), recipient, takeAmount})
}

func (p *V4Planner) AddUnwrap(amount *big.Int) (*V4Planner, error) {
	return p.AddActions(UNWRAP, []interface{}{amount})
}

func currencyAddress(curr core.Currency) common.Address {
	if curr.IsNative() {
		return constants.AddressZero
	}
	return curr.Wrapped().Address
}

type RouterAction struct {
	action       Actions
	encodedInput []byte
}

func createAction(action Actions, parameters []interface{}) (*RouterAction, error) {
	paramTypes := V4_BASE_ACTIONS_ABI_DEFINITION[action]
	var typeStrings []string
	for _, v := range paramTypes {
		typeStrings = append(typeStrings, v.Type)
	}

	// Create string for abi.NewType
	types := make([]abi.ArgumentMarshaling, len(typeStrings))
	for i, t := range typeStrings {
		types[i] = abi.ArgumentMarshaling{Type: t}
	}
	arguments := abi.Arguments{}
	for _, t := range types {
		typ, err := abi.NewType(t.Type, "", nil)
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, abi.Argument{Type: typ})
	}
	encodeInput, err := arguments.Pack(parameters...)
	if err != nil {
		return nil, err
	}

	return &RouterAction{
		action:       action,
		encodedInput: encodeInput,
	}, nil
}
