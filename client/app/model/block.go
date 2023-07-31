package model

import "github.com/ethereum/go-ethereum/core/types"

type Block struct {
	Number       uint64
	Hash         string
	Difficulty   uint64
	Transactions []*types.Transaction
	Uncles       []*types.Header
	Time         uint64
}
