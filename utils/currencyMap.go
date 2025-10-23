package utils

import (
	"github.com/KyberNetwork/pancake-v3-sdk/constants"
	core "github.com/daoleno/uniswap-sdk-core/entities"
	"github.com/ethereum/go-ethereum/common"
)

func ToAddress(currency core.Currency) (common.Address, error) {
	if currency.IsNative() {
		return constants.AddressZero, nil
	}
	return currency.Wrapped().Address, nil
}
