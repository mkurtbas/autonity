package soma

import (
	golog "log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

// Instantiates a new EVM object which is required when creating or calling a deployed contract
func getEVM(chain consensus.ChainReader, header *types.Header, coinbase, origin common.Address, statedb *state.StateDB) *vm.EVM {
	GetHashFn := func(ref *types.Header) func(n uint64) common.Hash {
		var cache map[uint64]common.Hash

		return func(n uint64) common.Hash {
			// If there's no hash cache yet, make one
			if cache == nil {
				cache = map[uint64]common.Hash{
					ref.Number.Uint64() - 1: ref.ParentHash,
				}
			}
			// Try to fulfill the request from the cache
			if hash, ok := cache[n]; ok {
				return hash
			}
			// Not cached, iterate the blocks and cache the hashes
			for header := chain.GetHeader(ref.ParentHash, ref.Number.Uint64()-1); header != nil; header = chain.GetHeader(header.ParentHash, header.Number.Uint64()-1) {
				cache[header.Number.Uint64()-1] = header.ParentHash
				if n == header.Number.Uint64()-1 {
					return header.ParentHash
				}
			}
			return common.Hash{}
		}
	}

	gasPrice := new(big.Int).SetUint64(0x0)
	evmContext := vm.Context{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     GetHashFn(header),
		Origin:      origin,
		Coinbase:    coinbase,
		BlockNumber: header.Number,
		Time:        header.Time,
		GasLimit:    header.GasLimit,
		Difficulty:  header.Difficulty,
		GasPrice:    gasPrice,
	}
	chainConfig := params.AllSomaProtocolChanges
	vmconfig := vm.Config{}
	evm := vm.NewEVM(evmContext, statedb, chainConfig, vmconfig)
	return evm
}

// deployContract deploys the contract contained within the genesis field bytecode
func deployContract(chain consensus.ChainReader, bytecodeStr string, userAddr common.Address, header *types.Header, statedb *state.StateDB) (common.Address, error) {
	// Convert the contract bytecode from hex into bytes
	contractBytecode := common.Hex2Bytes(bytecodeStr[2:])

	evm := getEVM(chain, header, userAddr, userAddr, statedb)

	sender := vm.AccountRef(userAddr)
	data := contractBytecode
	gas := uint64(0xFFFFFFFF)
	value := new(big.Int).SetUint64(0x00)

	// Deploy the Soma validator governance contract
	_, contractAddress, gas, vmerr := evm.Create(sender, data, gas, value)

	if vmerr != nil {
		return contractAddress, vmerr
	}
	log.Info("Deployed Soma Governance Contract", "Address", contractAddress.String())

	return contractAddress, nil
}

// callActiveValidators queries the active validator set contained in the deployed Soma contract.
// Returns true/false if the the address is an active validator and false if not.
func callActiveValidators(chain consensus.ChainReader, userAddr common.Address, contractAddress common.Address, header *types.Header, db ethdb.Database) (bool, error) {
	// Signature of function being called defined by Soma interface
	functionSig := "Validator(address)"

	// Instantiate new state database
	sdb := state.NewDatabase(db)
	statedb, _ := state.New(header.Root, sdb)

	sender := vm.AccountRef(userAddr)
	gas := uint64(0xFFFFFFFF)
	evm := getEVM(chain, header, userAddr, userAddr, statedb)

	// Pad address for ABI encoding
	encodedAddress := [32]byte{}
	copy(encodedAddress[12:], userAddr[:])
	input := crypto.Keccak256Hash([]byte(functionSig)).Bytes()[:4]
	inputData := append(input[:], encodedAddress[:]...)

	// Call ActiveValidators()
	ret, gas, vmerr := evm.StaticCall(sender, contractAddress, inputData, gas)
	if vmerr != nil {
		return false, vmerr
	}

	const def = `[{ "name" : "method", "outputs": [{ "type": "bool" }] }]`
	funcAbi, err := abi.JSON(strings.NewReader(def))
	if err != nil {
		return false, vmerr
	}

	var output bool
	err = funcAbi.Unpack(&output, "method", ret)
	if err != nil {
		return false, err
	}

	return output, nil
}

// callRecentValidators queries the recent validator set contained in the deployed Soma contract.
// Returns true if address is not a recent validator and false if they are.
func callRecentValidators(chain consensus.ChainReader, userAddr common.Address, contractAddress common.Address, header *types.Header, db ethdb.Database) (bool, error) {
	// Signature of function being called defined by Soma interface
	functionSig := "RecentValidator(address)"

	// Instantiate new state database
	sdb := state.NewDatabase(db)
	statedb, _ := state.New(header.Root, sdb)

	sender := vm.AccountRef(userAddr)
	gas := uint64(0xFFFFFFFF)
	evm := getEVM(chain, header, userAddr, userAddr, statedb)

	// Pad address for ABI encoding
	encodedAddress := [32]byte{}
	copy(encodedAddress[12:], userAddr[:])
	input := crypto.Keccak256Hash([]byte(functionSig)).Bytes()[:4]
	inputData := append(input[:], encodedAddress[:]...)

	// Call ActiveValidators()
	ret, gas, vmerr := evm.StaticCall(sender, contractAddress, inputData, gas)
	if vmerr != nil {
		return false, vmerr
	}

	const def = `[{ "name" : "method", "outputs": [{ "type": "bool" }] }]`
	funcAbi, err := abi.JSON(strings.NewReader(def))
	if err != nil {
		return false, vmerr
	}

	var output bool
	err = funcAbi.Unpack(&output, "method", ret)
	if err != nil {
		return false, err
	}

	return output, nil

}

func calculateDifficulty(chain consensus.ChainReader, userAddr common.Address, contractAddress common.Address, header *types.Header, db ethdb.Database) (*big.Int, error) {
	// Signature of function being called defined by Soma interface
	functionSig := "calculateDifficulty(address)"

	// Instantiate new state database
	sdb := state.NewDatabase(db)
	statedb, _ := state.New(header.Root, sdb)

	sender := vm.AccountRef(userAddr)
	gas := uint64(0xFFFFFFFF)
	evm := getEVM(chain, header, userAddr, userAddr, statedb)

	// Pad address for ABI encoding
	encodedAddress := [32]byte{}
	copy(encodedAddress[12:], userAddr[:])
	input := crypto.Keccak256Hash([]byte(functionSig)).Bytes()[:4]
	inputData := append(input[:], encodedAddress[:]...)

	// Call ActiveValidators()
	ret, gas, vmerr := evm.StaticCall(sender, contractAddress, inputData, gas)
	if vmerr != nil {
		return big.NewInt(1), vmerr
	}

	const def = `[{"name" : "int", "constant" : false, "outputs": [ { "type": "uint256" } ]}]`
	funcAbi, err := abi.JSON(strings.NewReader(def))
	if err != nil {
		return big.NewInt(1), vmerr
	}

	// marshal int
	var Int *big.Int
	err = funcAbi.Unpack(&Int, "int", ret)
	log.Info("calculateDifficulty", "Difficulty", Int)
	if err != nil {
		golog.Println(err)
		return big.NewInt(1), vmerr
	}

	return Int, nil
}

// updateGovernance when a validator attempts to submit a block the
func updateGovernance(chain consensus.ChainReader, userAddr common.Address, contractAddress common.Address, header *types.Header, statedb *state.StateDB) error {
	// Signature of function being called defined by Soma interface
	functionSig := "UpdateGovernance()"

	sender := vm.AccountRef(userAddr)
	gas := uint64(0xFFFFFFFF)
	value := new(big.Int).SetUint64(0x00)

	evm := getEVM(chain, header, userAddr, userAddr, statedb)

	// Pad address for ABI encoding
	input := crypto.Keccak256Hash([]byte(functionSig)).Bytes()

	// Call ActiveValidators()
	_, gas, vmerr := evm.Call(sender, contractAddress, input, gas, value)
	if vmerr != nil {
		return vmerr
	}

	return nil

}

// getThreshold returns the threshold of validators for use with calculating the correct out of turn wiggle
func getThreshold(chain consensus.ChainReader, userAddr common.Address, contractAddress common.Address, header *types.Header, db ethdb.Database) ([]byte, error) {
	// Instantiate new state database
	sdb := state.NewDatabase(db)
	statedb, _ := state.New(header.Root, sdb)

	// Signature of function being called defined by Soma interface
	functionSig := "threshold()"

	sender := vm.AccountRef(userAddr)
	gas := uint64(0xFFFFFFFF)
	value := new(big.Int).SetUint64(0x00)

	evm := getEVM(chain, header, userAddr, userAddr, statedb)

	// Pad address for ABI encoding
	input := crypto.Keccak256Hash([]byte(functionSig)).Bytes()

	// Call ActiveValidators()
	ret, gas, vmerr := evm.Call(sender, contractAddress, input, gas, value)
	if vmerr != nil {
		return nil, vmerr
	}

	return ret, nil

}

// getThreshold returns the threshold of validators for use with calculating the correct out of turn wiggle
func getValsNumber(chain consensus.ChainReader, userAddr common.Address, contractAddress common.Address, header *types.Header, db ethdb.Database) ([]byte, error) {
	// Instantiate new state database
	sdb := state.NewDatabase(db)
	statedb, _ := state.New(header.Root, sdb)

	// Signature of function being called defined by Soma interface
	functionSig := "getValsNumber()"

	sender := vm.AccountRef(userAddr)
	gas := uint64(0xFFFFFFFF)
	value := new(big.Int).SetUint64(0x00)

	evm := getEVM(chain, header, userAddr, userAddr, statedb)

	// Pad address for ABI encoding
	input := crypto.Keccak256Hash([]byte(functionSig)).Bytes()

	// Call ActiveValidators()
	ret, gas, vmerr := evm.Call(sender, contractAddress, input, gas, value)
	if vmerr != nil {
		return nil, vmerr
	}

	return ret, nil

}

// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
// that a new block should have based on the previous blocks in the chain and the
// current signer.
func calcDifficulty(chain consensus.ChainReader, parent *types.Header, soma *Soma) *big.Int {
	log.Info("CalcDifficulty", "ParentHash", chain.CurrentHeader().ParentHash)
	if parent.Number.Uint64() == 1 {
		if soma.config.Deployer == soma.signer {
			return new(big.Int).Set(diffInTurn)
		}
		return new(big.Int).Set(diffNoTurn)
	} else {
		result, _ := calculateDifficulty(chain, soma.signer, soma.somaContract, parent, soma.db)
		if result.Uint64() == 2 {
			return new(big.Int).Set(diffInTurn)
		}
		return new(big.Int).Set(diffNoTurn)
	}
}
