package entities

import (
	"reflect"

	"github.com/ethereum/go-ethereum/common"
)

type PathKey struct {
	IntermediateCurrency common.Address // address
	Fee                  int64
	TickSpacing          int64
	Hooks                common.Address // address
	HookData             []byte         // bytes
}

func EncodeRouteToPath(route *Route, exactOutput bool) ([]PathKey, error) {
	pools := route.Pools

	startingCurrency := route.PathInput
	if exactOutput {
		// reverse pools
		reverse(pools)
		startingCurrency = route.PathOutput
	}
	pathKeys := make([]PathKey, 0)

	for _, pool := range pools {
		nextCurrency := pool.Currency0
		if startingCurrency.Equal(pool.Currency0) {
			nextCurrency = pool.Currency1
		}

		pathKey := PathKey{
			IntermediateCurrency: nextCurrency.Address,
			Fee:                  pool.Fee,
			TickSpacing:          pool.TickSpacing,
			Hooks:                pool.Hooks,
			HookData:             []byte("0x"),
		}

		pathKeys = append(pathKeys, pathKey)
		startingCurrency = nextCurrency
	}

	if exactOutput {
		reverse(pathKeys)
	}

	return pathKeys, nil
}

func reverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
