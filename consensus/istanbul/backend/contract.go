package backend

import (
	"fmt"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/clearmatics/autonity/accounts/abi"
	"github.com/clearmatics/autonity/common"
	"github.com/clearmatics/autonity/consensus"
	"github.com/clearmatics/autonity/core"
	"github.com/clearmatics/autonity/core/state"
	"github.com/clearmatics/autonity/core/types"
	"github.com/clearmatics/autonity/core/vm"
	"github.com/clearmatics/autonity/log"
)

// Instantiates a new EVM object which is required when creating or calling a deployed contract
func (sb *backend) getEVM(chain consensus.ChainReader, header *types.Header, origin common.Address, statedb *state.StateDB) *vm.EVM {

	coinbase, _ := sb.Author(header)
	evmContext := vm.Context{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     core.GetHashFn(header, chain),
		Origin:      origin,
		Coinbase:    coinbase,
		BlockNumber: header.Number,
		Time:        header.Time,
		GasLimit:    header.GasLimit,
		Difficulty:  header.Difficulty,
		GasPrice:    new(big.Int).SetUint64(0x0),
	}
	evm := vm.NewEVM(evmContext, statedb, chain.Config(), *sb.vmConfig)
	return evm
}

// deploySomaContract deploys the contract contained within the genesis field bytecode
func (sb *backend) deploySomaContract(chain consensus.ChainReader, header *types.Header, statedb *state.StateDB) (common.Address, error) {
	// Convert the contract bytecode from hex into bytes
	contractBytecode := common.Hex2Bytes(sb.config.Bytecode)
	evm := sb.getEVM(chain, header, sb.config.Deployer, statedb)
	sender := vm.AccountRef(sb.config.Deployer)

	var validators common.Addresses
	validators, _ = sb.retrieveSavedValidators(1, chain)
	sort.Sort(validators)
	//We need to append to data the constructor's parameters
	//That should always be genesis validators

	somaAbi, err := abi.JSON(strings.NewReader(sb.config.ABI))
	if err != nil {
		return common.Address{}, err
	}

	constructorParams, err := somaAbi.Pack("", validators)
	if err != nil {
		return common.Address{}, err
	}

	data := append(contractBytecode, constructorParams...)
	gas := uint64(0xFFFFFFFF)
	value := new(big.Int).SetUint64(0x00)

	// Deploy the Soma validator governance contract
	_, contractAddress, gas, vmerr := evm.Create2(sender, data, gas, value, header.Number)
	if vmerr != nil {
		log.Error("Error Soma Governance Contract deployment")
		return contractAddress, vmerr
	}

	fmt.Println("soma contract address", contractAddress.String())
	f,_:=os.Create("/Users/boris/go/src/github.com/clearmatics/autonity/debug/deploy_state"+strconv.Itoa(i)+".json")
	fmt.Fprintln(f, string(statedb.Dump()))
	f2,_:=os.Create("/Users/boris/go/src/github.com/clearmatics/autonity/debug/code"+strconv.Itoa(i)+".txt")
	fmt.Fprintln(f2, statedb.GetCode(contractAddress))
	f3,_:=os.Create("/Users/boris/go/src/github.com/clearmatics/autonity/debug/origin_storage"+strconv.Itoa(i)+".txt")
	f4,_:=os.Create("/Users/boris/go/src/github.com/clearmatics/autonity/debug/dirty_storage"+strconv.Itoa(i)+".txt")
	//f5,_:=os.Create("/Users/boris/go/src/github.com/clearmatics/autonity/debug/spew"+strconv.Itoa(i)+".txt")
	so:=statedb.GetOrNewStateObject(contractAddress)
	fmt.Fprintln(f3, so.OriginStorage())
	fmt.Fprintln(f4, so.DirtyStorage())
	//spew.Fdump(f5, so)
	defer f.Close()
	defer f2.Close()
	defer f3.Close()
	defer f4.Close()
	//defer f5.Close()



	log.Info("Deployed Soma Governance Contract", "Address", contractAddress.String())

	return contractAddress, nil
}
var i int

func (sb *backend) contractGetValidators(chain consensus.ChainReader, header *types.Header, statedb *state.StateDB) ([]common.Address, error) {
	sender := vm.AccountRef(sb.config.Deployer)
	gas := uint64(0xFFFFFFFF)
	evm := sb.getEVM(chain, header, sb.config.Deployer, statedb)

	somaAbi, err := abi.JSON(strings.NewReader(sb.config.ABI))
	if err != nil {
		fmt.Println("abi.JSON")
		return nil, err
	}

	input, err := somaAbi.Pack("getValidators")
	if err != nil {
		fmt.Println("somaAbi.Pack")
		return nil, err
	}

	value := new(big.Int).SetUint64(0x00)
	//A standard call is issued - we leave the possibility to modify the state
	ret, gas, vmerr := evm.Call(sender, sb.somaContract, input, gas, value)
	if vmerr != nil {
		fmt.Println("evm.Call(")
		log.Error("Error Soma Governance Contract GetValidators()", "err", err)
		return nil, vmerr
	}
	var addresses []common.Address
	fmt.Println("ret",ret)
	fmt.Println("ret",string(ret))
	if err := somaAbi.Unpack(&addresses, "getValidators", ret); err != nil { // can't work with aliased types
	fmt.Println("somaAbi.Unpack")
		log.Error("Could not unpack getValidators returned value", "err", err)
		return nil, err
	}

	sortableAddresses := common.Addresses(addresses)
	sort.Sort(sortableAddresses)
	return sortableAddresses, nil
}
