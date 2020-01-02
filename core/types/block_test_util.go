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

type GethBlock struct {
	Header       *GethHeader
	uncles       []*GethHeader
	transactions Transaction
}

func (b *GethBlock) Hash() common.Hash {
	return rlpHash(b.Header)
}
