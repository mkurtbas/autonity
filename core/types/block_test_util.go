package types

import (
	"github.com/clearmatics/autonity/common"
	"math/big"
)

// old geth header
type GethHeader struct {
	ParentHash  common.Hash
	UncleHash   common.Hash
	Coinbase    common.Address
	Root        common.Hash
	TxHash      common.Hash
	ReceiptHash common.Hash
	Bloom       Bloom
	Difficulty  *big.Int
	Number      *big.Int
	GasLimit    uint64
	GasUsed     uint64
	Time        uint64
	Extra       []byte
	MixDigest   common.Hash
	Nonce       BlockNonce
}

func CopyGethHeader(h *GethHeader) *GethHeader {
	cpy := *h
	if cpy.Difficulty = new(big.Int); h.Difficulty != nil {
		cpy.Difficulty.Set(h.Difficulty)
	}
	if cpy.Number = new(big.Int); h.Number != nil {
		cpy.Number.Set(h.Number)
	}
	if len(h.Extra) > 0 {
		cpy.Extra = make([]byte, len(h.Extra))
		copy(cpy.Extra, h.Extra)
	}
	return &cpy
}

type GethBlock struct {
	Header       *GethHeader
	uncles       []*GethHeader
	transactions Transaction
}

func NewGethBlock(h *GethHeader) *GethBlock {
	b := &GethBlock{Header: CopyGethHeader(h)}

	b.Header.TxHash = EmptyRootHash
	b.Header.ReceiptHash = EmptyRootHash
	b.Header.UncleHash = EmptyUncleHash

	return b
}

func (b *GethBlock) Hash() common.Hash {
	return rlpHash(b.Header)
}
