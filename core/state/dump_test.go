// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package state

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
)

// Tests that an empty state is not scheduled for syncing.
func TestCalcDirties(t *testing.T) {
	// test for 1 balance change
	addr := common.HexToAddress("0x823140710bf13990e4500136726d8b55")
	entries := []journalEntry{
		balanceChange{
			account: &addr,
		},
	}
	dirties := calcDirties(entries)
	assert.NotNil(t, dirties[addr])
	assert.Equal(t, 1, dirties[addr].balanceChange)
	assert.Equal(t, 0, len(dirties[addr].storageChange))

	// test for 1 storage change
	entries = []journalEntry{
		storageChange{
			account: &addr,
			key:     common.BytesToHash([]byte("key")),
		},
	}
	dirties = calcDirties(entries)
	assert.NotNil(t, dirties[addr])
	assert.Equal(t, 0, dirties[addr].balanceChange)
	assert.Equal(t, 1, len(dirties[addr].storageChange))
	_, exist := dirties[addr].storageChange[common.BytesToHash([]byte("key"))]
	assert.True(t, exist)

	// test for 1 balance change and 1 storage change
	entries = []journalEntry{
		balanceChange{
			account: &addr,
		},
		storageChange{
			account: &addr,
			key:     common.BytesToHash([]byte("key2")),
		},
	}
	dirties = calcDirties(entries)
	assert.NotNil(t, dirties[addr])
	assert.Equal(t, 1, dirties[addr].balanceChange)
	assert.Equal(t, 1, len(dirties[addr].storageChange))
	_, exist = dirties[addr].storageChange[common.BytesToHash([]byte("key2"))]
	assert.True(t, exist)
}

type testDirtyAction struct {
	name    string
	fn      func(testDirtyAction, *StateDB)
	thash   string
	addr    common.Address
	balance int64
	storage map[common.Hash]common.Hash
}

func TestDumpDirtySnapshot(t *testing.T) {
	bhash := common.HexToHash("0xbd921b799e372549034755b7523e80634ad73d8eddeced84ef35114825192095")
	addr1 := common.HexToAddress("0x823140710bf13990e4500136726d8b55")
	addr2 := common.HexToAddress("0xfb45ca3e7e9d9d8a1cbc94f9d7c4144c46c030d6")

	state, _ := New(common.Hash{}, NewDatabase(ethdb.NewMemDatabase()))

	actions := []testDirtyAction{
		{
			name: "SetBalance",
			fn: func(a testDirtyAction, s *StateDB) {
				s.SetBalance(a.addr, big.NewInt(10))
			},
			thash:   "2d61b42398ed84a52515ac480f11e2e01519153a13064857c881a5333713e066",
			addr:    addr1,
			balance: int64(10),
		},
		{
			name: "SetState",
			fn: func(a testDirtyAction, s *StateDB) {
				for key, val := range a.storage {
					s.SetState(a.addr, key, val)
				}
			},
			thash: "9d33cf1edc701217b648704bb46d57513db29d9c863f8a31c5e42ff2275c4f7f",
			addr:  addr1,
			storage: map[common.Hash]common.Hash{
				common.BytesToHash([]byte("key")): common.BytesToHash([]byte("value")),
			},
		},
		{
			name: "SetBalance and SetState",
			fn: func(a testDirtyAction, s *StateDB) {
				s.SetBalance(a.addr, big.NewInt(a.balance))
				for key, val := range a.storage {
					s.SetState(a.addr, key, val)
				}
			},
			thash:   "4117c87e070c3203418779d7d2dfbacb0a569854b419f3f6772eee4fe10d1bb0",
			addr:    addr1,
			balance: int64(20),
			storage: map[common.Hash]common.Hash{
				common.BytesToHash([]byte("key2")): common.BytesToHash([]byte("value2")),
			},
		},
		{
			name: "SetBalance for another account",
			fn: func(a testDirtyAction, s *StateDB) {
				s.SetBalance(a.addr, big.NewInt(a.balance))
			},
			thash:   "1c86f46c42c868a8a274061470c258a935fa298323b578c901d9fa2a6904ccca",
			addr:    addr2,
			balance: int64(30),
		},
	}

	// apply transactins
	for ti, testCase := range actions {
		state.Prepare(common.BytesToHash(common.Hex2Bytes(testCase.thash)), bhash, ti)
		testCase.fn(testCase, state)
		state.DumpDirtySnapshot()
	}

	dirtyDump := state.DumpDirty()
	assert.NotNil(t, dirtyDump)
	assert.Equal(t, len(actions), len(dirtyDump.Transactions))

	for i, dumpTransaction := range dirtyDump.Transactions {
		assert.Equal(t, actions[i].thash, dumpTransaction.TxHash,
			"TxHash is not equal, want:%v, got:%v, in case %v", actions[i].thash, dumpTransaction.TxHash, actions[i].name)
		assert.Equal(t, 1, len(dumpTransaction.Accounts))
		dirtyAccount, exist := dumpTransaction.Accounts[common.Bytes2Hex(actions[i].addr.Bytes())]
		assert.True(t, exist)
		if actions[i].balance != 0 {
			assert.NotNil(t, dirtyAccount.Balance)
			assert.Equal(t, fmt.Sprintf("%d", actions[i].balance), *dirtyAccount.Balance)
		}
		if len(actions[i].storage) > 0 {
			assert.Equal(t, len(actions[i].storage), len(dirtyAccount.Storage))
			for key, val := range actions[i].storage {
				assert.Equal(t, common.Bytes2Hex(val.Bytes()), dirtyAccount.Storage[common.Bytes2Hex(key.Bytes())])
			}
		}
	}
}
