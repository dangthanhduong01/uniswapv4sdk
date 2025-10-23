package entities

import (
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Param struct {
	name  string
	value interface{}
}

type V4RouterAction struct {
	actionName string
	actionType Actions
	params     []Param
}

type V4RouterCall struct {
	actions []V4RouterAction
}

type SwapExactInSingle struct {
	poolKey          PoolKey
	zeroForOne       bool
	amountIn         *big.Int
	amountOutMinimum *big.Int
	hookData         []byte
}

type SwapExactIn struct {
	currencyIn       *big.Int
	path             []PathKey
	amountIn         *big.Int
	amountOutMinimum *big.Int
}

type SwapExactOutSingle struct {
	poolKey         PoolKey
	zeroForOne      bool
	amountOut       *big.Int
	amountInMaximum *big.Int
	hookData        []byte
}

type SwapExactOut struct {
	currencyOut     *big.Int
	path            []PathKey
	amountIn        *big.Int
	amountInMaximum *big.Int
}

type V4BaseActionsParser struct {
}

// func ParseCallData(calldata []byte) V4RouterCall {
// func ParseCallData(calldata []byte) *V4RouterCall {
// 	actions, inputs, err := abiEnCoder(calldata)
// 	if err != nil {
// 		return nil
// 	}
// 	actionTypes := getActions(actions)

// 	for i, actionType := range actionTypes {

// 	}

// }

func getActions(actions []byte) []Actions {
	var actionTypes []Actions
	for i := 2; i < len(actions); i += 1 {
		b := actions[i]
		actionTypes = append(actionTypes, Actions(b))
	}
	return actionTypes
}

func abiEnCoder(calldata []byte) ([]byte, [][]byte, error) {
	bytesType, _ := abi.NewType("bytes", "", nil)
	bytesArrayType, _ := abi.NewType("bytes[]", "", nil)

	arguments := abi.Arguments{
		{Type: bytesType},
		{Type: bytesArrayType},
	}

	decoded, err := arguments.Unpack(calldata)
	if err != nil {
		return nil, nil, err
	}

	actions, _ := decoded[0].([]byte)
	inputsIface, _ := decoded[1].([][]byte)

	return actions, inputsIface, nil

}

func parsePoolKey(data []interface{}) PoolKey {
	currency0 := data[0].(common.Address)
	currency1 := data[1].(common.Address)
	fee := toInt64(data[2])
	tickSpacing := toInt64(data[3])
	hooks := data[4].(common.Address)

	return PoolKey{
		Currency0:   currency0,
		Currency1:   currency1,
		Fee:         fee,
		TickSpacing: tickSpacing,
		Hooks:       hooks,
	}
}

func parsePathKey(data []interface{}) PathKey {
	intermediateCurrency := data[0].(common.Address)
	fee := toInt64(data[1])
	tickSpacing := toInt64(data[2])
	hooks := data[3].(common.Address)
	hookData := data[4].([]byte)

	return PathKey{
		IntermediateCurrency: intermediateCurrency,
		Fee:                  fee,
		TickSpacing:          tickSpacing,
		Hooks:                hooks,
		HookData:             hookData,
	}
}

func parseV4ExactIn(data []interface{}) SwapExactIn {
	currencyIn, _ := data[0].(*big.Int)
	rawPath := data[1].([]interface{})
	amountIn, _ := data[2].(*big.Int)
	amountOutMinimum, _ := data[3].(*big.Int)

	var paths []PathKey
	for _, pathKey := range rawPath {
		// Giả sử parsePathKey nhận vào string, nếu không thì sửa lại cho phù hợp
		pk := parsePathKey(pathKey.([]interface{}))
		paths = append(paths, pk)
	}

	return SwapExactIn{
		currencyIn:       currencyIn,
		path:             paths,
		amountIn:         amountIn,
		amountOutMinimum: amountOutMinimum,
	}
}

// Helper function to convert interface{} to int
func toInt64(val interface{}) int64 {
	switch v := val.(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	// case int64:
	// 	return int(v)
	case float64:
		return int64(v)
	case string:
		i, _ := strconv.Atoi(v)
		return int64(i)
	default:
		return 0
	}
}
