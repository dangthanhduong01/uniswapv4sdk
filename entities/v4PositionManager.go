package entities

import (
	"math/big"

	"github.com/dangthanhduong01/uniswapv4-sdk/utils"
	"github.com/ethereum/go-ethereum/common"

	core "github.com/daoleno/uniswap-sdk-core/entities"
)

type V4PositionPlanner struct {
	V4Planner
}

func (p *V4PositionPlanner) AddMint(pool Pool, tickLower, tickUpper int, liquidity *big.Int, amount0Max, amount1Max *big.Int, owner common.Address, hookData byte) error {
	poolKey, err := GetPoolKey(pool.Currency0, pool.Currency1, pool.Fee, pool.TickSpacing, pool.Hooks)
	if err != nil {
		return err
	}
	inputs := []interface{}{
		poolKey,
		tickLower,
		tickUpper,
		liquidity.String(),
		amount0Max.String(),
		amount1Max.String(),
		owner,
		hookData,
	}
	_, err = p.AddActions(MINT_POSITION, inputs)
	return err
}

func (p *V4PositionPlanner) AddDecrease(tokenId, liquidity, amount0Min, amount1Min *big.Int, hookData byte) error {
	inputs := []interface{}{
		tokenId.String(),
		amount0Min.String(),
		amount1Min.String(),
		hookData,
	}
	_, err := p.AddActions(TAKE_PAIR, inputs)
	return err
}

func (p *V4PositionPlanner) AddBurn(tokenId, amount0Min, amount1Min *big.Int, hookData byte) error {
	inputs := []interface{}{
		tokenId.String(),
		amount0Min.String(),
		amount1Min.String(),
		hookData,
	}
	_, err := p.AddActions(BURN_POSITION, inputs)
	return err
}

func (p *V4PositionPlanner) AddSettlePair(currency0, currency1 *core.Currency) error {
	var err error
	curr0, err := utils.ToAddress(*currency0)
	if err != nil {
		return err
	}
	curr1, err := utils.ToAddress(*currency1)
	if err != nil {
		return err
	}
	inputs := []interface{}{curr0, curr1}
	_, err = p.AddActions(SETTLE_PAIR, inputs)
	return err
}

func (p *V4PositionPlanner) AddTakePair(currency0, currency1 *core.Currency, recipient common.Address) error {
	var err error
	curr0, err := utils.ToAddress(*currency0)
	if err != nil {
		return err
	}
	curr1, err := utils.ToAddress(*currency1)
	if err != nil {
		return err
	}

	inputs := []interface{}{curr0, curr1, recipient.Hex()}
	_, err = p.AddActions(TAKE_PAIR, inputs)
	return err
}

func (p *V4PositionPlanner) AddSweep(currency *core.Currency, to common.Address) error {
	curr, err := utils.ToAddress(*currency)
	if err != nil {
		return err
	}
	input := []interface{}{curr, to}
	_, err = p.AddActions(SWEEP, input)
	return err
}
