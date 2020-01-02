package core

import (
	"github.com/clearmatics/autonity/common"
	"github.com/clearmatics/autonity/core/rawdb"
	"github.com/clearmatics/autonity/core/state"
	"github.com/clearmatics/autonity/core/types"
	"github.com/clearmatics/autonity/ethdb"
	"github.com/clearmatics/autonity/params"
	"math/big"
	"sync"
)

type gethGenesis struct {
	Config     *params.ChainConfig
	Nonce      uint64
	Timestamp  uint64
	ExtraData  []byte
	GasLimit   uint64
	Difficulty *big.Int
	Mixhash    common.Hash
	Coinbase   common.Address
	Alloc      GenesisAlloc

	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	Number     uint64
	GasUsed    uint64
	ParentHash common.Hash

	mu sync.RWMutex
}

func (g *gethGenesis) getExtraData() []byte {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return append([]byte{}, g.ExtraData...)
}

func (g *gethGenesis) ToBlock(db ethdb.Database) *types.GethBlock {
	if db == nil {
		db = rawdb.NewMemoryDatabase()
	}
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(db))
	for addr, account := range g.Alloc {
		statedb.AddBalance(addr, account.Balance)
		statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
		for key, value := range account.Storage {
			statedb.SetState(addr, key, value)
		}
	}
	root := statedb.IntermediateRoot(false)

	g.mu.RLock()
	diff := big.NewInt(0)
	if g.Difficulty == nil {
		diff.Set(params.GenesisDifficulty)
	} else {
		diff.Set(g.Difficulty)
	}
	g.mu.RUnlock()

	// old block header
	head := &types.GethHeader{
		Number:     new(big.Int).SetUint64(g.Number),
		Nonce:      types.EncodeNonce(g.Nonce),
		Time:       g.Timestamp,
		ParentHash: g.ParentHash,
		Extra:      g.getExtraData(),
		GasLimit:   g.GasLimit,
		GasUsed:    g.GasUsed,
		Difficulty: diff,
		MixDigest:  g.Mixhash,
		Coinbase:   g.Coinbase,
		Root:       root,
	}
	if g.GasLimit == 0 {
		head.GasLimit = params.GenesisGasLimit
	}
	statedb.Commit(false)
	statedb.Database().TrieDB().Commit(root, true)

	// old block
	return types.NewGethBlock(head)
}
