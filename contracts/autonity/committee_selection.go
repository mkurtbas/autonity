package autonity

import (
	"github.com/clearmatics/autonity/common"
	"github.com/clearmatics/autonity/core/state"
	"github.com/clearmatics/autonity/core/types"
	"github.com/clearmatics/autonity/core/vm"
	"github.com/clearmatics/autonity/log"
	"math/big"
)

// refer to autonity contract abt spec, keep in same meta.
type Committee struct {
	Accounts        []common.Address `abi:"accounts"`
	Stakes          []*big.Int       `abi:"stakes"`
}

func (ac *Contract) callGetCommittee(header *types.Header, stateDB *state.StateDB) (*Committee, error) {
	if header.Number.Cmp(big.NewInt(1)) == 0 && ac.SavedValidatorsRetriever != nil {
		return ac.SavedCommitteeRetriever(1)
	}
	// prepare abi and evm context
	deployer := ac.bc.Config().AutonityContractConfig.Deployer
	sender := vm.AccountRef(deployer)
	gas := uint64(0xFFFFFFFF)
	evm := ac.getEVM(header, deployer, stateDB)

	ABI, err := ac.abi()
	if err != nil {
		return nil, err
	}

	// pack the function which dump the data from contract.
	input, err := ABI.Pack("getCommittee")
	if err != nil {
		log.Warn("cannot pack the method: ", err.Error())
		return nil, err
	}

	// call evm.
	ret, _, vmerr := evm.StaticCall(sender, ac.Address(), input, gas)
	log.Debug("bytes return from contract: ", ret)
	if vmerr != nil {
		log.Error("Error Autonity Contract getCommittee")
		return nil, vmerr
	}

	// marshal the data from bytes arrays into specified structure.
	c := Committee{make([]common.Address, 32), make([]*big.Int, 32)}

	if err := ABI.Unpack(&c, "getCommittee", ret); err != nil { // can't work with aliased types
		log.Error("Could not unpack getCommittee returned value", "err", err, "header.num",
			header.Number.Uint64())
		return nil, err
	}
	return &c, nil
}

func (ac *Contract) SetCommittee(header *types.Header, stateDB *state.StateDB) (*Committee, error) {
	if header.Number.Uint64() <= 1 {
		return nil, nil
	}
	return ac.callSetCommittee(header, stateDB)
}

func (ac *Contract) callSetCommittee(header *types.Header, stateDB *state.StateDB) (*Committee, error) {

	// prepare abi and evm context
	deployer := ac.bc.Config().AutonityContractConfig.Deployer
	sender := vm.AccountRef(deployer)
	gas := uint64(0xFFFFFFFF)
	evm := ac.getEVM(header, deployer, stateDB)

	ABI, err := ac.abi()
	if err != nil {
		return nil, err
	}

	// pack the function which dump the data from contract.
	input, err := ABI.Pack("setCommittee")
	if err != nil {
		log.Warn("cannot pack the method: ", err.Error())
		return nil, err
	}

	// call evm.
	value := new(big.Int).SetUint64(0x00)
	ret, _, vmerr := evm.Call(sender, ac.Address(), input, gas, value)
	log.Debug("bytes return from contract: ", ret)
	if vmerr != nil {
		log.Error("Error Autonity Contract setCommittee")
		return nil, vmerr
	}

	// marshal the data from bytes arrays into specified structure.
	c := Committee{make([]common.Address, 32), make([]*big.Int, 32)}

	if err := ABI.Unpack(&c, "setCommittee", ret); err != nil { // can't work with aliased types
		log.Error("Could not unpack getCommittee returned value", "err", err, "header.num",
			header.Number.Uint64())
		return nil, err
	}
	return &c, nil

	//TODO: Need to save the new committee to in the client state

}
