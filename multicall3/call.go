package multicall3

import "github.com/ethereum/go-ethereum/common"

type Call[Q any, R any] struct {
	Target common.Address
	Method string
	Args   []Q
	Result R
}
