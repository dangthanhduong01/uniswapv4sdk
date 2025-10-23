package utils

import (
	"bytes"
	"errors"

	core "github.com/daoleno/uniswap-sdk-core/entities"
)

var (
	ErrDifferentChain = errors.New("ChainIds of the two currencies are different")
)

func SortsBefore(currencyA, currencyB *core.Token) (bool, error) {
	if currencyA.IsNative() {
		return true, nil
	}
	if currencyB.IsNative() {
		return false, nil
	}
	if currencyA.ChainId() != currencyB.ChainId() {
		return false, ErrDifferentChain
	}
	if currencyA.Address == currencyB.Address {
		return false, nil
	}
	return bytes.Compare(currencyA.Address[:], currencyB.Address[:]) < 0, nil
}
