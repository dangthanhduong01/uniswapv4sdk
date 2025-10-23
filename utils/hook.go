package utils

import (
	"errors"
	"math/big"
	"regexp"
)

var (
	ErrInvalidAddress = "invalid Address"
)

type HookOption string

const (
	AfterRemoveLiquidityReturnsDelta HookOption = "afterRemoveLiquidityReturnsDelta"
	AfterAddLiquidityReturnsDelta    HookOption = "afterAddLiquidityReturnsDelta"
	AfterSwapReturnsDelta            HookOption = "afterSwapReturnsDelta"
	BeforeSwapReturnsDelta           HookOption = "beforeSwapReturnsDelta"
	AfterDonate                      HookOption = "afterDonate"
	BeforeDonate                     HookOption = "beforeDonate"
	AfterSwap                        HookOption = "afterSwap"
	BeforeSwap                       HookOption = "beforeSwap"
	AfterRemoveLiquidity             HookOption = "afterRemoveLiquidity"
	BeforeRemoveLiquidity            HookOption = "beforeRemoveLiquidity"
	AfterAddLiquidity                HookOption = "afterAddLiquidity"
	BeforeAddLiquidity               HookOption = "beforeAddLiquidity"
	AfterInitialize                  HookOption = "afterInitialize"
	BeforeInitialize                 HookOption = "beforeInitialize"
)

type HookPermissions map[HookOption]bool

var hookFlagIndex = map[HookOption]int{
	AfterRemoveLiquidityReturnsDelta: 0,
	AfterAddLiquidityReturnsDelta:    1,
	AfterSwapReturnsDelta:            2,
	BeforeSwapReturnsDelta:           3,
	AfterDonate:                      4,
	BeforeDonate:                     5,
	AfterSwap:                        6,
	BeforeSwap:                       7,
	AfterRemoveLiquidity:             8,
	BeforeRemoveLiquidity:            9,
	AfterAddLiquidity:                10,
	BeforeAddLiquidity:               11,
	AfterInitialize:                  12,
	BeforeInitialize:                 13,
}

type Hook struct {
}

func (h *Hook) Permissions(addr string) (HookPermissions, error) {
	if err := _checkAddress(addr); err != nil {
		return nil, err
	}

	getPerm := func(opt HookOption) bool {
		perm, err := _hasPermission(addr, opt)
		if err != nil {
			return false
		}
		return perm
	}

	permissions := HookPermissions{
		BeforeInitialize:                 getPerm(BeforeInitialize),
		AfterInitialize:                  getPerm(AfterInitialize),
		BeforeAddLiquidity:               getPerm(BeforeAddLiquidity),
		AfterAddLiquidity:                getPerm(AfterAddLiquidity),
		BeforeRemoveLiquidity:            getPerm(BeforeRemoveLiquidity),
		AfterRemoveLiquidity:             getPerm(AfterRemoveLiquidity),
		BeforeSwap:                       getPerm(BeforeSwap),
		AfterSwap:                        getPerm(AfterSwap),
		BeforeDonate:                     getPerm(BeforeDonate),
		AfterDonate:                      getPerm(AfterDonate),
		BeforeSwapReturnsDelta:           getPerm(BeforeSwapReturnsDelta),
		AfterSwapReturnsDelta:            getPerm(AfterSwapReturnsDelta),
		AfterAddLiquidityReturnsDelta:    getPerm(AfterAddLiquidityReturnsDelta),
		AfterRemoveLiquidityReturnsDelta: getPerm(AfterRemoveLiquidityReturnsDelta),
	}
	return permissions, nil
}

func (h *Hook) HasPermission(addr string, hookOption HookOption) (bool, error) {
	if err := _checkAddress(addr); err != nil {
		return false, err
	}
	return _hasPermission(addr, hookOption)
}

// HasInitializePermissions kiểm tra xem địa chỉ có bất kỳ quyền khởi tạo nào không.
func (h *Hook) HasInitializePermissions(address string) (bool, error) {
	if err := _checkAddress(address); err != nil {
		return false, err
	}

	perm1, err1 := _hasPermission(address, BeforeInitialize)
	if err1 != nil {
		return false, err1
	}
	perm2, err2 := _hasPermission(address, AfterInitialize)
	if err2 != nil {
		return false, err2
	}

	return perm1 || perm2, nil
}

// HasLiquidityPermissions kiểm tra xem địa chỉ có bất kỳ quyền thanh khoản cơ bản nào không.
func HasLiquidityPermissions(address string) (bool, error) {
	if err := _checkAddress(address); err != nil {
		return false, err
	}

	// Lấy tất cả các quyền có thể xảy ra trong một slice
	options := []HookOption{
		BeforeAddLiquidity,
		AfterAddLiquidity,
		BeforeRemoveLiquidity,
		AfterRemoveLiquidity,
		// Các quyền delta được bỏ qua theo logic TS, nhưng sẽ được bao hàm bởi các kiểm tra khác.
		// Logic TS: "this implicitly encapsulates liquidity delta permissions" - Hàm chỉ kiểm tra 4 quyền cơ bản
	}

	for _, opt := range options {
		perm, err := _hasPermission(address, opt)
		if err != nil {
			return false, err
		}
		if perm {
			return true, nil
		}
	}

	return false, nil
}

// HasSwapPermissions kiểm tra xem địa chỉ có bất kỳ quyền hoán đổi nào không.
func (h *Hook) HasSwapPermissions(address string) (bool, error) {
	if err := _checkAddress(address); err != nil {
		return false, err
	}

	perm1, err1 := _hasPermission(address, BeforeSwap)
	if err1 != nil {
		return false, err1
	}
	perm2, err2 := _hasPermission(address, AfterSwap)
	if err2 != nil {
		return false, err2
	}

	return perm1 || perm2, nil
}

// ---

// HasDonatePermissions kiểm tra xem địa chỉ có bất kỳ quyền donate nào không.
func (h *Hook) HasDonatePermissions(address string) (bool, error) {
	if err := _checkAddress(address); err != nil {
		return false, err
	}

	perm1, err1 := _hasPermission(address, BeforeDonate)
	if err1 != nil {
		return false, err1
	}
	perm2, err2 := _hasPermission(address, AfterDonate)
	if err2 != nil {
		return false, err2
	}

	return perm1 || perm2, nil
}

func _checkAddress(addr string) error {
	if !isAddressValid(addr) {
		return errors.New(ErrInvalidAddress)
	}
	return nil
}

func isAddressValid(addr string) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	if !re.Match([]byte(addr)) {
		return false
	}
	return len(addr) > 2 && addr[:2] == "0x"
}

// Thực hiện thao tác bitwise: !!(parseInt(address, 16) & (1 << hookFlagIndex[hookOption]))
func _hasPermission(addr string, hookOption HookOption) (bool, error) {
	index, ok := hookFlagIndex[hookOption]
	if !ok {
		return false, errors.New("unknown hook option")
	}

	// 1. Chuyển địa chỉ (chuỗi hexa) thành big.Int
	addrBig := new(big.Int)
	// Loai bo tien to "0x" neu co
	if len(addr) > 2 && addr[:2] == "0x" {
		addr = addr[2:]
	}
	_, success := addrBig.SetString(addr, 16)
	if !success {
		return false, errors.New("failed to parse address hex string")
	}

	// 2. Tính toán bit mask: (1 << index)
	// Trong Go, chúng ta sử dụng dịch bit (bit shift) trên số nguyên: 1 << index
	mask := new(big.Int).Lsh(big.NewInt(1), uint(index))

	// 3. Thực hiện phép toán bitwise AND: addr & mask
	result := new(big.Int).And(addrBig, mask)

	// 4. Kiểm tra xem bit có được đặt không: (result != 0)
	// (Tương đương với `!!(...)` trong TS)
	return result.Cmp(big.NewInt(0)) != 0, nil
}
