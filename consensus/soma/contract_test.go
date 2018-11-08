package soma

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/compiler"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
)

func TestEVMRuntimeCall(t *testing.T) {
	/*
		pragma solidity ^0.4.25;

		contract Test {
			function test() public pure returns(string) {
				return "Hello Test!!!";
			}
		}
	*/
	contractBytecode := "608060405260043610610041576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063f8a8fd6d14610046575b600080fd5b34801561005257600080fd5b5061005b6100d6565b6040518080602001828103825283818151815260200191508051906020019080838360005b8381101561009b578082015181840152602081019050610080565b50505050905090810190601f1680156100c85780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b60606040805190810160405280600d81526020017f48656c6c6f2054657374212121000000000000000000000000000000000000008152509050905600a165627a7a723058207d86d1462ac765f7f77965f34f8ad38a8fa270361ddfe7def03b516d6d6e4d120029"
	// (new Buffer(utils.sha3('test()'), 16)).toString().slice(0,8+2)
	input, err := hex.DecodeString("f8a8fd6d")
	if err != nil {
		t.Log(err)
	}

	ret, _, err := runtime.Execute(common.Hex2Bytes(contractBytecode), input, nil)
	if err != nil {
		t.Log(err)
	}
	// firstPart := ret[:32] // what is this?
	// secondPart := ret[32:(32*2)] // size of the string (which is 13)
	retStr := ret[(32 * 2) : (32*2)+13] // third part the data itself
	if "Hello Test!!!" != string(retStr) {
		t.Error("Call() result different from expected: ", ret)
	}
}

func MakePreState(db ethdb.Database, accounts core.GenesisAlloc) *state.StateDB {
	sdb := state.NewDatabase(db)
	statedb, _ := state.New(common.Hash{}, sdb)
	for addr, a := range accounts {
		statedb.SetCode(addr, a.Code)
		statedb.SetNonce(addr, a.Nonce)
		statedb.SetBalance(addr, a.Balance)
		for k, v := range a.Storage {
			statedb.SetState(addr, k, v)
		}
	}
	// Commit and re-open to start with a clean state.
	root, _ := statedb.Commit(false)
	statedb, _ = state.New(root, sdb)
	return statedb
}

func TestStateDBChanges(t *testing.T) {
	genesisHash := common.Hash{}
	// START STATE DB
	memorydb := ethdb.NewMemDatabase() // generates memory db (this could the LevelDB)
	sdb := state.NewDatabase(memorydb) // thread safe DB wrapper
	statedb, _ := state.New(common.Hash{}, sdb)

	userKey, _ := crypto.GenerateKey()
	userAddr := crypto.PubkeyToAddress(userKey.PublicKey)
	statedb.SetBalance(userAddr, big.NewInt(1000000000))
	statedb.SetNonce(userAddr, uint64(0))
	/*
		statedb.SetCode(addr, a.Code)
		for k, v := range a.Storage {
			statedb.SetState(addr, k, v)
		}
	*/

	// COMPILE CONTRACT
	basePath := os.Getenv("GOPATH") + "/src/github.com/clearmatics/autonity/"
	testContractPath := basePath + "consensus/soma/test.sol"
	contracts, err := compiler.CompileSolidity("", testContractPath)
	if err != nil {
		t.Error("ERROR failed to compile test.sol:", err)
	}
	testContract := contracts[testContractPath+":Test"]
	t.Logf("Bytecode: %s\n", testContract.Code)

	// START EVM
	// evmContext := vm.Context{} //core.NewEVMContext()
	vmTestBlockHash := func(n uint64) common.Hash {
		if n == 0 {
			return genesisHash
		}
		return common.BytesToHash(crypto.Keccak256([]byte(big.NewInt(int64(n)).String())))
	}
	evmContext := vm.Context{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     vmTestBlockHash,
		Origin:      userAddr,
		Coinbase:    userAddr,
		BlockNumber: new(big.Int).SetUint64(0x00),
		Time:        new(big.Int).SetUint64(0x01),
		GasLimit:    uint64(0x0f4240),
		Difficulty:  new(big.Int).SetUint64(0x0100),
		GasPrice:    new(big.Int).SetUint64(0x3b9aca00),
	}
	chainConfig := params.AllSomaProtocolChanges
	vmconfig := vm.Config{}
	/*
		type Config struct {
			// Debug enabled debugging Interpreter options
			Debug bool
			// Tracer is the op code logger
			Tracer Tracer
			// NoRecursion disabled Interpreter call, callcode,
			// delegate call and create.
			NoRecursion bool
			// Enable recording of SHA3/keccak preimages
			EnablePreimageRecording bool
			// JumpTable contains the EVM instruction table. This
			// may be left uninitialised and will be set to the default
			// table.
			JumpTable [256]operation
		}
	*/
	evm := vm.NewEVM(evmContext, statedb, chainConfig, vmconfig) //vmconfig)

	// DEPLOY CONTRACT
	sender := vm.AccountRef(userAddr)
	data := common.Hex2Bytes(testContract.Code[2:])
	gas := uint64(1000000)
	value := new(big.Int).SetUint64(0x00)
	ret, contractAddress, gas, vmerr := evm.Create(sender, data, gas, value)
	t.Log("====== CREATE =======")
	t.Logf("Contract:\n%s\n", hex.Dump(ret))
	t.Log("Address: ", contractAddress.String())
	t.Log("Gas: ", gas)
	t.Log("Error: ", vmerr)

	//contractAddress = common.HexToAddress("0x00")
	//statedb.SetNonce(contractAddress, uint64(0))
	//statedb.SetCode(contractAddress, data)

	// CALL
	functionSig := "test()"
	t.Log("====== CALL =======", functionSig)
	input := crypto.Keccak256Hash([]byte(functionSig)).Bytes()
	ret, gas, vmerr = evm.Call(sender, contractAddress, input, gas, value)
	t.Logf("Result:\n%s\n", hex.Dump(ret))
	t.Log("Gas: ", gas)
	t.Log("Error: ", vmerr)

	// commit makes current state saved into DB
	root, _ := statedb.Commit(true)
	t.Logf("Trie root: 0x%x\n", root)

	//t.Logf("Contract address code:\n%s\n", hex.Dump(statedb.GetCode(contractAddress)))
	t.Log(statedb.GetBalance(userAddr))
	t.Logf("memorydb Keys: %#v\n", memorydb.Keys())
	t.Logf("UserAddr: 0x%x\tHash: 0x%x\n", userAddr.Bytes(), crypto.Keccak256Hash(userAddr.Bytes()).Bytes())
	t.Logf("ContractAddress: 0x%x\tHash: 0x%x\n", contractAddress.Bytes(), crypto.Keccak256Hash(contractAddress.Bytes()).Bytes())

	// printDB(sdb)
}

func TestEVMContractDeployment(t *testing.T) {
	initialBalance := big.NewInt(1000000000)
	userKey, _ := crypto.GenerateKey()
	userAddr := crypto.PubkeyToAddress(userKey.PublicKey)
	coinbaseKey, _ := crypto.GenerateKey()
	coinbaseAddr := crypto.PubkeyToAddress(coinbaseKey.PublicKey)
	originKey, _ := crypto.GenerateKey()
	originAddr := crypto.PubkeyToAddress(originKey.PublicKey)

	alloc := make(core.GenesisAlloc)
	alloc[userAddr] = core.GenesisAccount{
		Balance: initialBalance,
	}
	statedb := MakePreState(ethdb.NewMemDatabase(), alloc)

	vmTestBlockHash := func(n uint64) common.Hash {
		return common.BytesToHash(crypto.Keccak256([]byte(big.NewInt(int64(n)).String())))
	}
	context := vm.Context{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     vmTestBlockHash,
		Origin:      originAddr,
		Coinbase:    coinbaseAddr,
		BlockNumber: new(big.Int).SetUint64(0x00),
		Time:        new(big.Int).SetUint64(0x01),
		GasLimit:    uint64(0x0f4240),
		Difficulty:  new(big.Int).SetUint64(0x0100),
		GasPrice:    new(big.Int).SetUint64(0x3b9aca00),
	}
	vmconfig := vm.Config{}
	//vmconfig.NoRecursion = true
	evm := vm.NewEVM(context, statedb, params.MainnetChainConfig, vmconfig) //vmconfig)

	// CREATE
	sender := vm.AccountRef(userAddr)
	/*
		pragma solidity ^0.4.25;

		contract Test {
			function test() public pure returns(string) {
				return "Hello Test!!!";
			}

			int private count = 0;
			function incrementCounter() public {
				count += 1;
			}
			function decrementCounter() public {
				count -= 1;
			}
			function getCount() public view returns (int) {
				return count;
			}
		}
	*/
	contractBytecode := "60806040526000805534801561001457600080fd5b506101e6806100246000396000f300608060405260043610610062576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680635b34b96614610067578063a87d942c1461007e578063f5c5ad83146100a9578063f8a8fd6d146100c0575b600080fd5b34801561007357600080fd5b5061007c610150565b005b34801561008a57600080fd5b50610093610162565b6040518082815260200191505060405180910390f35b3480156100b557600080fd5b506100be61016b565b005b3480156100cc57600080fd5b506100d561017d565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156101155780820151818401526020810190506100fa565b50505050905090810190601f1680156101425780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b60016000808282540192505081905550565b60008054905090565b60016000808282540392505081905550565b60606040805190810160405280600d81526020017f48656c6c6f2054657374212121000000000000000000000000000000000000008152509050905600a165627a7a723058201b0858a814ecee293d6f73f3c8ed4b76a898989e7e0c3796fdb8db6a6c16884b0029"
	data := common.Hex2Bytes(contractBytecode)
	gas := uint64(1000000)
	value := new(big.Int).SetUint64(0x00)
	ret, contractAddress, gas, vmerr := evm.Create(sender, data, gas, value)
	t.Log("====== CREATE =======")
	t.Logf("Contract:\n%s\n", hex.Dump(ret))
	t.Log("Address: ", contractAddress.String())
	t.Log("Gas: ", gas)
	t.Log("Error: ", vmerr)
	// CALL
	functionSig := "test()"
	t.Log("====== CALL =======", functionSig)
	input := crypto.Keccak256Hash([]byte(functionSig)).Bytes()
	ret, gas, vmerr = evm.Call(sender, contractAddress, input, gas, value)
	t.Logf("Result:\n%s\n", hex.Dump(ret))
	t.Log("Gas: ", gas)
	t.Log("Error: ", vmerr)
	// CALL
	functionSig = "getCount()"
	t.Log("====== CALL =======", functionSig)
	input = crypto.Keccak256Hash([]byte(functionSig)).Bytes()
	ret, gas, vmerr = evm.Call(sender, contractAddress, input, gas, value)
	t.Logf("Result:\n%s\n", hex.Dump(ret))
	t.Log("Gas: ", gas)
	t.Log("Error: ", vmerr)
	totalIncrements := new(big.Int).SetUint64(5)
	for i := uint64(0); i < totalIncrements.Uint64(); i++ {
		// CALL
		functionSig = "incrementCounter()"
		t.Log("====== CALL =======", functionSig)
		input = crypto.Keccak256Hash([]byte(functionSig)).Bytes()
		ret, gas, vmerr = evm.Call(sender, contractAddress, input, gas, value)
		t.Logf("Result:\n%s\n", hex.Dump(ret))
		t.Log("Gas: ", gas)
		t.Log("Error: ", vmerr)
	}
	// CALL
	functionSig = "getCount()"
	t.Log("====== CALL =======", functionSig)
	input = crypto.Keccak256Hash([]byte(functionSig)).Bytes()
	ret, gas, vmerr = evm.Call(sender, contractAddress, input, gas, value)
	t.Logf("Result:\n%s\n", hex.Dump(ret))
	t.Log("Gas: ", gas)
	t.Log("Error: ", vmerr)

	funcType := `[{"type": "uint256"}]`
	funcDef := fmt.Sprintf(`[{ "name" : "method", "outputs": %s}]`, funcType)
	testAbi, err := abi.JSON(strings.NewReader(funcDef))
	if err != nil {
		t.Log("Error", err)
	}

	output, err := testAbi.Methods["method"].Outputs.UnpackValues(ret)
	if err == nil {
		t.Log("Error", err)
	}
	t.Log(output)

	resultTotalIncrements := new(big.Int).SetBytes(ret)
	if resultTotalIncrements.Uint64() != totalIncrements.Uint64() {
		t.Error("Increments n smart contract and expected differ\n", "result: ", resultTotalIncrements, " expected: ", totalIncrements)
	}

	// CALL (TRANSFER)
	initialBalanceUser := statedb.GetBalance(userAddr)
	t.Log("initialBalanceUser:\t", initialBalanceUser)
	initialBalanceCoinbase := statedb.GetBalance(coinbaseAddr)
	t.Log("initialBalanceCoinbase:\t", initialBalanceCoinbase)
	initialBalanceorigin := statedb.GetBalance(originAddr)
	t.Log("initialBalanceorigin:\t", initialBalanceorigin)

	t.Log("====== CALL ======= TRANSFER")
	input = []byte{}
	value = new(big.Int).SetUint64(0x100)
	ret, gas, vmerr = evm.Call(sender, originAddr, nil, gas, value)
	t.Logf("Result:\n%s\n", hex.Dump(ret))
	t.Log("Gas: ", gas)
	t.Log("Error: ", vmerr)

	statedb.Finalise(true) // clean dirty objects

	finalBalanceUser := statedb.GetBalance(userAddr)
	t.Log("finalBalanceUser:\t\t", finalBalanceUser)
	finalBalanceCoinbase := statedb.GetBalance(coinbaseAddr)
	t.Log("finalBalanceCoinbase:\t", finalBalanceCoinbase)
	finalBalanceorigin := statedb.GetBalance(originAddr)
	t.Log("finalBalanceorigin:\t", finalBalanceorigin)

	var transferredValue big.Int
	transferredValue.Sub(finalBalanceorigin, initialBalanceorigin)
	if transferredValue.Cmp(value) != 0 {
		t.Error("Unexpected balance in origin account!")
	}
}

// func TestEVMSomaContractDeployment(t *testing.T) {
// 	contractBytecode := "0x60806040523480156200001157600080fd5b50604051620015ad380380620015ad8339810180604052810190808051820192919050505060008090505b81518110156200014157600160008084848151811015156200005a57fe5b9060200190602002015173ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff02191690831515021790555060038282815181101515620000c657fe5b9060200190602002015190806001815401808255809150509060018203906000526020600020016000909192909190916101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505080806001019150506200003c565b620001606003805490506200025e640100000000026401000000009004565b60058190555062000185600380549050620002a6640100000000026401000000009004565b6006819055506001600554101515620002565760043390806001815401808255809150509060018203906000526020600020016000909192909190916101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505060018060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff0219169083151502179055505b5050620002c5565b6000806001831415620002755760009150620002a0565b6002831415620002895760019150620002a0565b60016002848115156200029857fe5b040190508091505b50919050565b6000806001600284811515620002b857fe5b0401905080915050919050565b6112d880620002d56000396000f3006080604052600436106100af576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063044d77a4146100b45780630f2e98e91461010f5780632be1c4661461016a57806335aa2e44146101c557806342cde4e8146102325780635b25cd621461025d578063632a9a52146102b857806365e94b4e146102e357806371ff924f146103505780637fad141e1461039357806383686b28146103aa575b600080fd5b3480156100c057600080fd5b506100f5600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610401565b604051808215151515815260200191505060405180910390f35b34801561011b57600080fd5b50610150600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610456565b604051808215151515815260200191505060405180910390f35b34801561017657600080fd5b506101ab600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610476565b604051808215151515815260200191505060405180910390f35b3480156101d157600080fd5b506101f06004803603810190808035906020019092919050505061058e565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34801561023e57600080fd5b506102476105cc565b6040518082815260200191505060405180910390f35b34801561026957600080fd5b5061029e600480360381019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506105d2565b604051808215151515815260200191505060405180910390f35b3480156102c457600080fd5b506102cd6105f2565b6040518082815260200191505060405180910390f35b3480156102ef57600080fd5b5061030e600480360381019080803590602001909291905050506105f8565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34801561035c57600080fd5b50610391600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610636565b005b34801561039f57600080fd5b506103a861079e565b005b3480156103b657600080fd5b506103eb600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610c16565b6040518082815260200191505060405180910390f35b60008060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff169050919050565b60016020528060005260406000206000915054906101000a900460ff1681565b6000816000808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff161515610539576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601d8152602001807f566f746572206973206e6f74206163746976652076616c696461746f7200000081525060200191505060405180910390fd5b600160008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff16915050919050565b60038181548110151561059d57fe5b906000526020600020016000915054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b60055481565b60006020528060005260406000206000915054906101000a900460ff1681565b60065481565b60048181548110151561060757fe5b906000526020600020016000915054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b336000808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff1615156106f7576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601d8152602001807f566f746572206973206e6f74206163746976652076616c696461746f7200000081525060200191505060405180910390fd5b600260008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008154809291906001019190505550600654600260008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410151561079a5761079982610c2e565b5b5050565b6000336000808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff161515610861576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601d8152602001807f566f746572206973206e6f74206163746976652076616c696461746f7200000081525060200191505060405180910390fd5b33600160008273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff16151515610924576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260198152602001807f566f74657220697320726563656e742076616c696461746f720000000000000081525060200191505060405180910390fd5b60006005541115610c11576005546004805490501015610a005760043390806001815401808255809150509060018203906000526020600020016000909192909190916101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505060018060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff021916908315150217905550610c10565b6005546004805490501415610c0f5760006001600060046000815481101515610a2557fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff021916908315150217905550600092505b600160048054905003831015610b5757600460018401815481101515610ac857fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16600484815481101515610b0257fe5b9060005260206000200160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508280600101935050610aa6565b336004600160048054905003815481101515610b6f57fe5b9060005260206000200160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060018060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff0219169083151502179055505b5b5b505050565b60026020528060005260406000206000915090505481565b6000806000808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff168015610c8e57506001600380549050115b156110ca5760008060008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff021916908315150217905550600091505b600380549050821015610e6c57600382815481101515610d0b57fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff161415610e5f57600382815481101515610d7857fe5b9060005260206000200160006101000a81549073ffffffffffffffffffffffffffffffffffffffff02191690558190505b600160038054905003811015610e5a57600360018201815481101515610dcb57fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16600382815481101515610e0557fe5b9060005260206000200160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508080600101915050610da9565b610e6c565b8180600101925050610cef565b6003805480919060019003610e81919061125b565b50600160008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff16156110c5576000600160008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff021916908315150217905550600091505b6004805490508210156110ae57600482815481101515610f4d57fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1614156110a157600482815481101515610fba57fe5b9060005260206000200160006101000a81549073ffffffffffffffffffffffffffffffffffffffff02191690558190505b60016004805490500381101561109c5760046001820181548110151561100d57fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1660048281548110151561104757fe5b9060005260206000200160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508080600101915050610feb565b6110ae565b8180600101925050610f31565b60048054809190600190036110c3919061125b565b505b611188565b60016000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff02191690831515021790555060038390806001815401808255809150509060018203906000526020600020016000909192909190916101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550505b6000600260008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055506111db6003805490506111fa565b6005819055506111ef60038054905061123d565b600681905550505050565b600080600183141561120f5760009150611237565b60028314156112215760019150611237565b600160028481151561122f57fe5b040190508091505b50919050565b600080600160028481151561124e57fe5b0401905080915050919050565b815481835581811115611282578183600052602060002091820191016112819190611287565b5b505050565b6112a991905b808211156112a557600081600090555060010161128d565b5090565b905600a165627a7a723058205908f4e9c5d67192a62e4c749ecf71a13145ed72f03e25199b6c4e6d767ade310029"

// 	initialBalance := big.NewInt(1000000000)
// 	userKey, _ := crypto.GenerateKey()
// 	userAddr := crypto.PubkeyToAddress(userKey.PublicKey)
// 	coinbaseKey, _ := crypto.GenerateKey()
// 	coinbaseAddr := crypto.PubkeyToAddress(coinbaseKey.PublicKey)
// 	originKey, _ := crypto.GenerateKey()
// 	originAddr := crypto.PubkeyToAddress(originKey.PublicKey)
// 	gas := uint64(0xFFFFFFFFFF)
// 	value := new(big.Int).SetUint64(0x00)

// 	alloc := make(core.GenesisAlloc)
// 	alloc[userAddr] = core.GenesisAccount{
// 		Balance: initialBalance,
// 	}

// 	db := ethdb.NewMemDatabase()
// 	statedb := MakePreState(db, alloc)

// 	vmTestBlockHash := func(n uint64) common.Hash {
// 		return common.BytesToHash(crypto.Keccak256([]byte(big.NewInt(int64(n)).String())))
// 	}
// 	context := vm.Context{
// 		CanTransfer: core.CanTransfer,
// 		Transfer:    core.Transfer,
// 		GetHash:     vmTestBlockHash,
// 		Origin:      originAddr,
// 		Coinbase:    coinbaseAddr,
// 		BlockNumber: new(big.Int).SetUint64(0x00),
// 		Time:        new(big.Int).SetUint64(0x01),
// 		GasLimit:    uint64(0xFFFFFFFFFF),
// 		Difficulty:  new(big.Int).SetUint64(0x0100),
// 		GasPrice:    new(big.Int).SetUint64(0x3b9aca00),
// 	}
// 	vmconfig := vm.Config{}
// 	evm := vm.NewEVM(context, statedb, params.MainnetChainConfig, vmconfig) //vmconfig)

// 	// CREATE
// 	sender := vm.AccountRef(userAddr)

// 	// Compile the Soma contract and then deploy it natively
// 	binaryBytes := common.Hex2Bytes(contractBytecode)
// 	encodedAddress := [32]byte{}
// 	copy(encodedAddress[12:], userAddr[:])

// 	abiEncoding := common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000001")
// 	constructorArgs := append(abiEncoding[:], encodedAddress[:]...)
// 	somaCode := append(binaryBytes[:], constructorArgs[:]...)
// 	ret, contractAddress, gas, vmerr := evm.Create(sender, somaCode, gas, value)

// 	deployContract(&testerChainReader{db: db}, somaCode)

// 	// t.Log("====== CREATE =======")
// 	// t.Logf("Contract:\n%s\n", hex.Dump(ret))
// 	// t.Log("Address: ", contractAddress.String())
// 	// t.Log("Gas: ", gas)
// 	// t.Log("Error: ", vmerr)

// 	// functionSig := "ActiveValidator(address)"
// 	// t.Log("====== CALL =======", functionSig)
// 	// input := crypto.Keccak256Hash([]byte(functionSig)).Bytes()[:4]
// 	// inputData := append(input[:], encodedAddress[:]...)
// 	// ret, gas, vmerr = evm.Call(sender, contractAddress, inputData, gas, value)
// 	// t.Logf("Result:\n%s\n", hex.Dump(ret))
// 	// t.Log("Gas: ", gas)
// 	// t.Log("Error: ", vmerr)

// }

func CompileSoma() string {
	basePath := os.Getenv("GOPATH") + "/src/github.com/ethereum/go-ethereum/consensus/soma/"
	contractPath := basePath + "Soma.sol"

	contracts, err := compiler.CompileSolidity("", contractPath)
	if err != nil {
		log.Fatal("ERROR failed to compile Soma.sol:", err)
	}

	governContract := contracts[basePath+"Soma.sol:Soma"]
	governBinStr, _ := getContractBytecodeAndABI(governContract)

	return governBinStr
}

func getContractBytecodeAndABI(c *compiler.Contract) (string, string) {
	cABIBytes, err := json.Marshal(c.Info.AbiDefinition)
	if err != nil {
		log.Fatal("ERROR marshalling contract ABI:", err)
	}

	contractBinStr := c.Code[2:]
	contractABIStr := string(cABIBytes)
	return contractBinStr, contractABIStr
}
