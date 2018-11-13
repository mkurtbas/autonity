// Copyright 2017 The go-ethereum Authors
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

package soma

import (
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
)

// API is a user facing RPC API to allow controlling the signer and voting
// mechanisms of the proof-of-authority scheme.
type API struct {
	chain consensus.ChainReader
	soma  *Soma
}

// GetGovernanceAddress returns the address to which the governance is deployed
func (api *API) GetGovernanceAddress() common.Address {
	// Return the address of the governance contract
	return api.soma.somaContract
}

// GetValidatorsAtBlock returns validators at block N
func (api *API) GetValidatorsAtBlock(number uint64) ([]common.Address, error) {
	// Make sure the number is not ahead of the latest block
	if number > api.chain.CurrentHeader().Number.Uint64() {
		return nil, errors.New("Error: Block not mined yet")
	}

	// Get header
	header := api.chain.GetHeaderByNumber(number)

	// Instantiate new state database
	sdb := state.NewDatabase(api.soma.db)
	statedb, _ := state.New(header.Root, sdb)

	// Signature of function being called defined by Soma interface
	functionSig := "getValidators()"

	sender := vm.AccountRef(api.soma.signer)
	gas := uint64(0xFFFFFFFF)
	value := new(big.Int).SetUint64(0x00)

	evm := getEVM(api.chain, header, api.soma.signer, api.soma.signer, statedb)

	// Pad address for ABI encoding
	input := crypto.Keccak256Hash([]byte(functionSig)).Bytes()

	// Call ActiveValidators()
	ret, gas, vmerr := evm.Call(sender, api.soma.somaContract, input, gas, value)
	if vmerr != nil {
		return nil, vmerr
	}

	const def = `[{ "name" : "method", "outputs": [{ "type": "address[]" }] }]`
	funcAbi, err := abi.JSON(strings.NewReader(def))
	if err != nil {
		return nil, vmerr
	}

	var output []common.Address
	err = funcAbi.Unpack(&output, "method", ret)
	if err != nil {
		return nil, err
	}

	return output, vmerr

}

// GetRecentsAtBlock returns validators at block N
func (api *API) GetRecentsAtBlock(number uint64) ([]common.Address, error) {
	// Make sure the number is not ahead of the latest block
	if number > api.chain.CurrentHeader().Number.Uint64() {
		return nil, errors.New("Error: Block not mined yet")
	}

	// Get header
	header := api.chain.GetHeaderByNumber(number)

	// Instantiate new state database
	sdb := state.NewDatabase(api.soma.db)
	statedb, _ := state.New(header.Root, sdb)

	// Signature of function being called defined by Soma interface
	functionSig := "getRecents()"

	sender := vm.AccountRef(api.soma.signer)
	gas := uint64(0xFFFFFFFF)
	value := new(big.Int).SetUint64(0x00)

	evm := getEVM(api.chain, header, api.soma.signer, api.soma.signer, statedb)

	// Pad address for ABI encoding
	input := crypto.Keccak256Hash([]byte(functionSig)).Bytes()

	// Call ActiveValidators()
	ret, gas, vmerr := evm.Call(sender, api.soma.somaContract, input, gas, value)
	if vmerr != nil {
		return nil, vmerr
	}

	const def = `[{ "name" : "method", "outputs": [{ "type": "address[]" }] }]`
	funcAbi, err := abi.JSON(strings.NewReader(def))
	if err != nil {
		return nil, vmerr
	}

	var output []common.Address
	err = funcAbi.Unpack(&output, "method", ret)
	if err != nil {
		return nil, err
	}

	return output, vmerr

}

// GetActivesAtBlock returns validators at block N
func (api *API) GetActivesAtBlock(number uint64) ([]common.Address, error) {
	// Make sure the number is not ahead of the latest block
	if number > api.chain.CurrentHeader().Number.Uint64() {
		return nil, errors.New("Error: Block not mined yet")
	}

	// Get header
	header := api.chain.GetHeaderByNumber(number)

	// Instantiate new state database
	sdb := state.NewDatabase(api.soma.db)
	statedb, _ := state.New(header.Root, sdb)

	// Signature of function being called defined by Soma interface
	functionSig := "getActives()"

	sender := vm.AccountRef(api.soma.signer)
	gas := uint64(0xFFFFFFFF)
	value := new(big.Int).SetUint64(0x00)

	evm := getEVM(api.chain, header, api.soma.signer, api.soma.signer, statedb)

	// Pad address for ABI encoding
	input := crypto.Keccak256Hash([]byte(functionSig)).Bytes()

	// Call ActiveValidators()
	ret, gas, vmerr := evm.Call(sender, api.soma.somaContract, input, gas, value)
	if vmerr != nil {
		return nil, vmerr
	}

	const def = `[{ "name" : "method", "outputs": [{ "type": "address[]" }] }]`
	funcAbi, err := abi.JSON(strings.NewReader(def))
	if err != nil {
		return nil, vmerr
	}

	var output []common.Address
	err = funcAbi.Unpack(&output, "method", ret)
	if err != nil {
		return nil, err
	}

	return output, vmerr

}

// GetThresholdAtBlock returns validators at block N
func (api *API) GetThresholdAtBlock(number uint64) (*big.Int, error) {
	// Make sure the number is not ahead of the latest block
	if number > api.chain.CurrentHeader().Number.Uint64() {
		return big.NewInt(0), errors.New("Error: Block not mined yet")
	}

	// Get header
	header := api.chain.GetHeaderByNumber(number)

	// Instantiate new state database
	sdb := state.NewDatabase(api.soma.db)
	statedb, _ := state.New(header.Root, sdb)

	// Signature of function being called defined by Soma interface
	functionSig := "threshold()"

	sender := vm.AccountRef(api.soma.signer)
	gas := uint64(0xFFFFFFFF)
	value := new(big.Int).SetUint64(0x00)

	evm := getEVM(api.chain, header, api.soma.signer, api.soma.signer, statedb)

	// Pad address for ABI encoding
	input := crypto.Keccak256Hash([]byte(functionSig)).Bytes()

	// Call ActiveValidators()
	ret, gas, vmerr := evm.Call(sender, api.soma.somaContract, input, gas, value)
	if vmerr != nil {
		return big.NewInt(1), vmerr
	}

	const def = `[{"name" : "int", "constant" : false, "outputs": [ { "type": "uint256" } ]}]`
	funcAbi, err := abi.JSON(strings.NewReader(def))
	if err != nil {
		return big.NewInt(1), err
	}

	// marshal int
	var Int *big.Int
	err = funcAbi.Unpack(&Int, "int", ret)
	if err != nil {
		return big.NewInt(1), consensus.ErrPrunedAncestor
	}

	return Int, nil

}
