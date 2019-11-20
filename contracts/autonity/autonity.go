package autonity

import (
	"errors"
	"fmt"
	"github.com/clearmatics/autonity/accounts/abi"
	"github.com/clearmatics/autonity/common"
	"github.com/clearmatics/autonity/consensus"
	"github.com/clearmatics/autonity/core/state"
	"github.com/clearmatics/autonity/core/types"
	"github.com/clearmatics/autonity/core/vm"
	"github.com/clearmatics/autonity/log"
	"github.com/clearmatics/autonity/metrics"
	"github.com/clearmatics/autonity/params"
	"math/big"
	"reflect"
	"sort"
	"strings"
	"sync"
)

func NewAutonityContract(
	bc Blockchainer,
	canTransfer func(db vm.StateDB, addr common.Address, amount *big.Int) bool,
	transfer func(db vm.StateDB, sender, recipient common.Address, amount *big.Int),
	GetHashFn func(ref *types.Header, chain ChainContext) func(n uint64) common.Hash,
) *Contract {
	return &Contract{
		bc:          bc,
		canTransfer: canTransfer,
		transfer:    transfer,
		GetHashFn:   GetHashFn,
		//SavedValidatorsRetriever: SavedValidatorsRetriever,
		users:		 nil,
	}
}

const (
	Participant uint8 = iota
	Stakeholder
	Validator
)

const (
	/* example metrics ID:
	contract/user/0xefqefea...214dafaff/validator/stake
	contract/user/0xefqefea...214dafaff/stakeholder/stake
	contract/user/0xefqefea...214dafaff/participant/stake
	contract/user/0xefqefea...214dafaff/validator/balance
	contract/user/0xefqefea...214dafaff/stakeholder/balance
	contract/user/0xefqefea...214dafaff/participant/balance
	contract/user/0xefqefea...214dafaff/validator/commissionrate
	contract/user/0xefqefea...214dafaff/stakeholder/commissionrate
	contract/user/0xefqefea...214dafaff/participant/commissionrate
	*/
	// template: contract/user/common.address/[validator|stakeholder|participant]/[stake|balance|commissionrate]
	UserMetricIDTemplate      = "contract/user/%s/%s/%s"

	// counting for each block will introduce metric ID increasing in metric registry, it will exhaust the memory.
	// template: contract/user/common.address/[validator|stakeholder|participant]/reward
	BlockRewardDistributionMetricIDTemplate = "contract/user/%s/%s/reward"
	BlockRewardMetricID = "contract/transactionfee"

	GlobalMetricIDGasPrice    = "contract/global/minimum/gasprice"
	GloablMetricIDStakeSupply = "contract/global/stakesupply"
)

type ChainContext interface {
	// Engine retrieves the chain's consensus engine.
	Engine() consensus.Engine

	// GetHeader returns the hash corresponding to their hash.
	GetHeader(common.Hash, uint64) *types.Header
}
type Blockchainer interface {
	ChainContext
	GetVMConfig() *vm.Config
	Config() *params.ChainConfig

	UpdateEnodeWhitelist(newWhitelist *types.Nodes)
	ReadEnodeWhitelist(openNetwork bool) *types.Nodes
}

type Contract struct {
	address                  common.Address
	contractABI              *abi.ABI
	bc                       Blockchainer
	SavedValidatorsRetriever func(i uint64) ([]common.Address, error)

	usersMutex				 sync.RWMutex
	users					 []common.Address

	canTransfer func(db vm.StateDB, addr common.Address, amount *big.Int) bool
	transfer    func(db vm.StateDB, sender, recipient common.Address, amount *big.Int)
	GetHashFn   func(ref *types.Header, chain ChainContext) func(n uint64) common.Hash
	sync.RWMutex
}

func (ac *Contract) generateRewardDistributionMetricsID(address common.Address, role uint8) string {
	userType := ac.resolveUserTypeName(role)
	blockMetricsID := fmt.Sprintf(BlockRewardDistributionMetricIDTemplate, address.String(), userType)
	return blockMetricsID
}

func (ac *Contract) resolveUserTypeName(role uint8) string {
	ret := "unknown"
	switch role {
	case Validator:
		ret = "validator"
	case Stakeholder:
		ret = "stakeholder"
	case Participant:
		ret = "participant"
	}
	return ret
}

func (ac *Contract) generateUserMetricsID(address common.Address, role uint8) (stakeID string,
	balanceID string, commissionRateID string, err error) {
	if role > Validator {
		return "", "", "", errors.New("invalid parameter")
	}
	userType := ac.resolveUserTypeName(role)
	stakeID = fmt.Sprintf(UserMetricIDTemplate, address.String(), userType, "stake")
	balanceID = fmt.Sprintf(UserMetricIDTemplate, address.String(), userType, "balance")
	commissionRateID = fmt.Sprintf(UserMetricIDTemplate, address.String(), userType, "commissionrate")
	return stakeID, balanceID, commissionRateID, nil
}

func (ac *Contract) removeMetricsFromRegistry(user common.Address) {

	// clean up metrics which counts user's stake, balance and commission rate.
	for role := Participant; role <= Validator; role ++ {
		if stakeID, balanceID, commissionRateID, err := ac.generateUserMetricsID(user, role); err == nil {
			metrics.DefaultRegistry.Unregister(stakeID)
			metrics.DefaultRegistry.Unregister(balanceID)
			metrics.DefaultRegistry.Unregister(commissionRateID)
		}
	}
	// clean up metrics which counts user's reward.
	rewardDistributionMetricID := ac.generateRewardDistributionMetricsID(user, Stakeholder)
	metrics.DefaultRegistry.Unregister(rewardDistributionMetricID)
}

/*
*  CleanUselessMetrics clean up metric memory from ETH-Metric Registry framework used by removed users.
*  Note: when node restart, those metrics registered in the metric registry are auto released.
*/
func (ac *Contract) CleanUselessMetrics(addresses []common.Address) {
	if len(addresses) == 0 {
		return
	}
	ac.usersMutex.Lock()
	defer ac.usersMutex.Unlock()
	if ac.users == nil {
		ac.users = addresses
		return
	}

	for _, user := range ac.users {
		found := false
		for _, address := range addresses {
			if user == address {
				found = true
				break
			}
		}

		if found == false {
			// to clean up metrics of users who was removed.
			ac.removeMetricsFromRegistry(user)
		}
	}
	// load the latest set.
	ac.users = addresses
}

// measure metrics of user's meta data by regarding of network economic.
func (ac *Contract) MeasureMetricsOfNetworkEconomic(header *types.Header, stateDB *state.StateDB) {
	if header == nil || stateDB == nil {
		log.Warn("invalid parameter")
		return
	}
	// skip the measurement of genesis block.
	if header.Number.Int64() < 1 {
		return
	}

	deployer := ac.bc.Config().AutonityContractConfig.Deployer
	sender := vm.AccountRef(deployer)
	gas := uint64(0xFFFFFFFF)
	evm := ac.getEVM(header, deployer, stateDB)

	ABI, err := ac.abi()
	if err != nil {
		return
	}

	input, err := ABI.Pack("dumpNetworkEconomicsData")
	if err != nil {
		log.Warn("cannot pack the method: ", err.Error())
		return
	}

	value := new(big.Int).SetUint64(0x00)
	ret, _, vmerr := evm.Call(sender, ac.Address(), input, gas, value)
	log.Debug("bytes return from contract: ", ret)
	if vmerr != nil {
		log.Warn("Error Autonity Contract dumpNetworkEconomics")
		return
	}

	v := struct {
		Accounts []common.Address `abi:"accounts"`
		Usertypes []uint8	`abi:"usertypes"`
		Stakes []*big.Int	`abi:"stakes"`
		Commissionrates []*big.Int `abi:"commissionrates"`
		Mingasprice *big.Int `abi:"mingasprice"`
		Stakesupply *big.Int `abi:"stakesupply"`
	}{make([]common.Address, 32), make([]uint8, 32), make([]*big.Int, 32),
		make([]*big.Int, 32), new(big.Int), new(big.Int)}

	if err := ABI.Unpack(&v, "dumpNetworkEconomicsData", ret); err != nil { // can't work with aliased types
		log.Warn("Could not unpack dumpNetworkEconomicsData returned value", "err", err, "header.num",
			header.Number.Uint64())
		return
	}

	// measure global metrics
	gasPriceGauge := metrics.GetOrRegisterGauge(GlobalMetricIDGasPrice, nil)
	stakeTotalSupplyGauge := metrics.GetOrRegisterGauge(GloablMetricIDStakeSupply, nil)
	gasPriceGauge.Update(v.Mingasprice.Int64())
	stakeTotalSupplyGauge.Update(v.Stakesupply.Int64())

	// measure user metrics
	if len(v.Accounts) != len(v.Usertypes) || len(v.Accounts) != len(v.Stakes) ||
		len(v.Accounts) != len(v.Commissionrates) {
		log.Warn("mismatched data set dumped from autonity contract")
		return
	}

	for i := 0; i < len(v.Accounts); i++ {
		user := v.Accounts[i]
		userType := v.Usertypes[i]
		stake := v.Stakes[i]
		rate := v.Commissionrates[i]
		balance := stateDB.GetBalance(user)

		log.Debug("user: ", user, "userType: ", userType, "stake: ", stake, "rate: ", rate, "balance: ", balance)

		// generate metric ID.
		stakeID, balanceID, commmissionRateID, err := ac.generateUserMetricsID(user, userType)
		if err != nil {
			log.Warn("generateUserMetricsID failed.")
			return
		}

		// get or create metrics from default registry.
		stakeGauge := metrics.GetOrRegisterGauge(stakeID, nil)
		balanceGauge := metrics.GetOrRegisterGauge(balanceID, nil)
		commissionRateGauge := metrics.GetOrRegisterGauge(commmissionRateID, nil)

		// submit data to registry.
		stakeGauge.Update(stake.Int64())
		balanceGauge.Update(balance.Int64())
		commissionRateGauge.Update(rate.Int64())
	}

	// clean up useless metrics if there were.
	ac.CleanUselessMetrics(v.Accounts)
	return
}

//// Instantiates a new EVM object which is required when creating or calling a deployed contract
func (ac *Contract) getEVM(header *types.Header, origin common.Address, statedb *state.StateDB) *vm.EVM {
	coinbase, _ := types.Ecrecover(header)
	evmContext := vm.Context{
		CanTransfer: ac.canTransfer,
		Transfer:    ac.transfer,
		GetHash:     ac.GetHashFn(header, ac.bc),
		Origin:      origin,
		Coinbase:    coinbase,
		BlockNumber: header.Number,
		Time:        new(big.Int).SetUint64(header.Time),
		GasLimit:    header.GasLimit,
		Difficulty:  header.Difficulty,
		GasPrice:    new(big.Int).SetUint64(0x0),
	}
	vmConfig := *ac.bc.GetVMConfig()
	evm := vm.NewEVM(evmContext, statedb, ac.bc.Config(), vmConfig)
	return evm
}

// deployContract deploys the contract contained within the genesis field bytecode
func (ac *Contract) DeployAutonityContract(chain consensus.ChainReader, header *types.Header, statedb *state.StateDB) (common.Address, error) {
	// Convert the contract bytecode from hex into bytes
	contractBytecode := common.Hex2Bytes(chain.Config().AutonityContractConfig.Bytecode)
	evm := ac.getEVM(header, chain.Config().AutonityContractConfig.Deployer, statedb)
	sender := vm.AccountRef(chain.Config().AutonityContractConfig.Deployer)

	//todo do we need it?
	//validators, err = ac.SavedValidatorsRetriever(1)
	//sort.Sort(validators)

	//We need to append to data the constructor's parameters
	//That should always be genesis validators

	contractABI, err := ac.abi()

	if err != nil {
		log.Error("abi.JSON returns err", "err", err)
		return common.Address{}, err
	}

	ln := len(chain.Config().AutonityContractConfig.GetValidatorUsers())
	validators := make(common.Addresses, 0, ln)
	enodes := make([]string, 0, ln)
	accTypes := make([]*big.Int, 0, ln)
	participantStake := make([]*big.Int, 0, ln)
	for _, v := range chain.Config().AutonityContractConfig.Users {
		validators = append(validators, v.Address)
		enodes = append(enodes, v.Enode)
		accTypes = append(accTypes, big.NewInt(int64(v.Type.GetID())))
		participantStake = append(participantStake, big.NewInt(int64(v.Stake)))
	}

	constructorParams, err := contractABI.Pack("",
		validators,
		enodes,
		accTypes,
		participantStake,
		chain.Config().AutonityContractConfig.Operator,
		new(big.Int).SetUint64(chain.Config().AutonityContractConfig.MinGasPrice))
	if err != nil {
		log.Error("contractABI.Pack returns err", "err", err)
		return common.Address{}, err
	}

	data := append(contractBytecode, constructorParams...)
	gas := uint64(0xFFFFFFFF)
	value := new(big.Int).SetUint64(0x00)

	// Deploy the Autonity contract
	_, contractAddress, _, vmerr := evm.Create(sender, data, gas, value)
	if vmerr != nil {
		log.Error("evm.Create returns err", "err", vmerr)
		return contractAddress, vmerr
	}
	ac.Lock()
	ac.address = contractAddress
	ac.Unlock()
	log.Info("Deployed Autonity Contract", "Address", contractAddress.String())

	return contractAddress, nil
}

func (ac *Contract) ContractGetValidators(chain consensus.ChainReader, header *types.Header, statedb *state.StateDB) ([]common.Address, error) {
	if header.Number.Cmp(big.NewInt(1)) == 0 && ac.SavedValidatorsRetriever != nil {
		return ac.SavedValidatorsRetriever(1)
	}
	sender := vm.AccountRef(chain.Config().AutonityContractConfig.Deployer)
	gas := uint64(0xFFFFFFFF)
	evm := ac.getEVM(header, chain.Config().AutonityContractConfig.Deployer, statedb)
	contractABI, err := ac.abi()
	if err != nil {
		return nil, err
	}

	input, err := contractABI.Pack("getValidators")
	if err != nil {
		return nil, err
	}

	value := new(big.Int).SetUint64(0x00)
	//A standard call is issued - we leave the possibility to modify the state
	ret, _, vmerr := evm.Call(sender, ac.Address(), input, gas, value)
	if vmerr != nil {
		return nil, vmerr
	}

	var addresses []common.Address
	if err := contractABI.Unpack(&addresses, "getValidators", ret); err != nil { // can't work with aliased types
		log.Error("Could not unpack getValidators returned value", "err", err)
		return nil, err
	}

	sortableAddresses := common.Addresses(addresses)
	sort.Sort(sortableAddresses)
	return sortableAddresses, nil
}

var ErrAutonityContract = errors.New("could not call Autonity contract")

func (ac *Contract) UpdateEnodesWhitelist(state *state.StateDB, block *types.Block) error {
	newWhitelist, err := ac.GetWhitelist(block, state)
	if err != nil {
		log.Error("could not call contract", "err", err)
		return ErrAutonityContract
	}

	ac.bc.UpdateEnodeWhitelist(newWhitelist)
	return nil
}

func (ac *Contract) GetWhitelist(block *types.Block, db *state.StateDB) (*types.Nodes, error) {
	var (
		newWhitelist *types.Nodes
		err          error
	)

	if block.Number().Uint64() == 1 {
		// use genesis block whitelist
		newWhitelist = ac.bc.ReadEnodeWhitelist(false)
	} else {
		// call retrieveWhitelist contract function
		newWhitelist, err = ac.callGetWhitelist(db, block.Header())
	}

	return newWhitelist, err
}

//blockchain

func (ac *Contract) callGetWhitelist(state *state.StateDB, header *types.Header) (*types.Nodes, error) {
	// Needs to be refactored somehow
	deployer := ac.bc.Config().AutonityContractConfig.Deployer
	sender := vm.AccountRef(deployer)
	gas := uint64(0xFFFFFFFF)
	evm := ac.getEVM(header, deployer, state)

	ABI, err := ac.abi()
	if err != nil {
		return nil, err
	}

	input, err := ABI.Pack("getWhitelist")
	if err != nil {
		return nil, err
	}

	ret, _, vmerr := evm.StaticCall(sender, ac.Address(), input, gas)
	if vmerr != nil {
		log.Error("Error Autonity Contract getWhitelist()")
		return nil, vmerr
	}

	var returnedEnodes []string
	if err := ABI.Unpack(&returnedEnodes, "getWhitelist", ret); err != nil { // can't work with aliased types
		log.Error("Could not unpack getWhitelist returned value")
		return nil, err
	}

	return types.NewNodes(returnedEnodes, false), nil
}

func (ac *Contract) GetMinimumGasPrice(block *types.Block, db *state.StateDB) (uint64, error) {
	if block.Number().Uint64() <= 1 {
		return ac.bc.Config().AutonityContractConfig.MinGasPrice, nil
	}

	return ac.callGetMinimumGasPrice(db, block.Header())
}

func (ac *Contract) SetMinimumGasPrice(block *types.Block, db *state.StateDB, price *big.Int) error {
	if block.Number().Uint64() <= 1 {
		return nil
	}

	return ac.callSetMinimumGasPrice(db, block.Header(), price)
}

func (ac *Contract) callGetMinimumGasPrice(state *state.StateDB, header *types.Header) (uint64, error) {
	// Needs to be refactored somehow
	deployer := ac.bc.Config().AutonityContractConfig.Deployer
	sender := vm.AccountRef(deployer)
	gas := uint64(0xFFFFFFFF)
	evm := ac.getEVM(header, deployer, state)

	ABI, err := ac.abi()
	if err != nil {
		return 0, err
	}

	input, err := ABI.Pack("getMinimumGasPrice")
	if err != nil {
		return 0, err
	}

	value := new(big.Int).SetUint64(0x00)
	ret, _, vmerr := evm.Call(sender, ac.Address(), input, gas, value)
	if vmerr != nil {
		log.Error("Error Autonity Contract getMinimumGasPrice()")
		return 0, vmerr
	}

	minGasPrice := new(big.Int)
	if err := ABI.Unpack(&minGasPrice, "getMinimumGasPrice", ret); err != nil { // can't work with aliased types
		log.Error("Could not unpack minGasPrice returned value", "err", err, "header.num", header.Number.Uint64())
		return 0, err
	}

	return minGasPrice.Uint64(), nil
}

func (ac *Contract) callSetMinimumGasPrice(state *state.StateDB, header *types.Header, price *big.Int) error {
	// Needs to be refactored somehow
	deployer := ac.bc.Config().AutonityContractConfig.Deployer
	sender := vm.AccountRef(deployer)
	gas := uint64(0xFFFFFFFF)
	evm := ac.getEVM(header, deployer, state)

	ABI, err := ac.abi()
	if err != nil {
		return err
	}

	input, err := ABI.Pack("setMinimumGasPrice")
	if err != nil {
		return err
	}

	_, _, vmerr := evm.Call(sender, ac.Address(), input, gas, price)
	if vmerr != nil {
		log.Error("Error Autonity Contract getMinimumGasPrice()")
		return vmerr
	}
	return nil
}

func (ac *Contract) PerformRedistribution(header *types.Header, db *state.StateDB, gasUsed *big.Int) error {
	if header.Number.Uint64() <= 1 {
		return nil
	}
	return ac.callPerformRedistribution(db, header, gasUsed)
}

func (ac *Contract) callPerformRedistribution(state *state.StateDB, header *types.Header, blockGas *big.Int) error {
	// Needs to be refactored somehow
	deployer := ac.bc.Config().AutonityContractConfig.Deployer

	sender := vm.AccountRef(deployer)
	gas := uint64(0xFFFFFFFF)
	evm := ac.getEVM(header, deployer, state)

	ABI, err := ac.abi()
	if err != nil {
		return err
	}

	input, err := ABI.Pack("performRedistribution", blockGas)
	if err != nil {
		log.Error("Error Autonity Contract callPerformRedistribution()", "err", err)
		return err
	}

	value := new(big.Int).SetUint64(0x00)

	ret, _, vmerr := evm.Call(sender, ac.Address(), input, gas, value)
	if vmerr != nil {
		log.Error("Error Autonity Contract callPerformRedistribution()", "err", err)
		return vmerr
	}

	// after reward distribution, update metrics with the return values.
	v := struct {
		Holders []common.Address `abi:"stakeholders"`
		Rewardfractions []*big.Int	`abi:"rewardfractions"`
		Amount *big.Int `abi:"amount"`
	}{make([]common.Address, 32), make([]*big.Int, 32),	new(big.Int)}

	if err := ABI.Unpack(&v, "performRedistribution", ret); err != nil { // can't work with aliased types
		log.Error("Could not unpack performRedistribution returned value", "err", err, "header.num", header.Number.Uint64())
		return nil
	}

	if len(v.Holders) != len(v.Rewardfractions) {
		log.Warn("reward fractions does not distribute to all stake holder.")
		return nil
	}

	// submit metrics to registry.
	for i := 0; i < len(v.Holders); i++ {
		rewardDistributionMetricID := ac.generateRewardDistributionMetricsID(v.Holders[i], Stakeholder)
		rwdDistributionMetric := metrics.GetOrRegisterCounter(rewardDistributionMetricID, nil)
		rwdDistributionMetric.Inc(v.Rewardfractions[i].Int64())
	}

	txFeeCounter := metrics.GetOrRegisterCounter(BlockRewardMetricID, nil)
	txFeeCounter.Inc(v.Amount.Int64())

	return nil
}

func (ac *Contract) ApplyPerformRedistribution(transactions types.Transactions, receipts types.Receipts, header *types.Header, statedb *state.StateDB) error {
	log.Info("ApplyPerformRedistribution", "header", header.Number.Uint64())
	if header.Number.Cmp(big.NewInt(1)) < 1 {
		return nil
	}
	blockGas := new(big.Int)
	for i, tx := range transactions {
		blockGas.Add(blockGas, new(big.Int).Mul(tx.GasPrice(), new(big.Int).SetUint64(receipts[i].GasUsed)))
	}
	log.Info("execution start ApplyPerformRedistribution", "balance", statedb.GetBalance(ac.Address()), "block", header.Number.Uint64(), "gas", blockGas.Uint64())
	if blockGas.Cmp(new(big.Int)) == 0 {
		log.Info("execution start ApplyPerformRedistribution with 0 gas", "balance", statedb.GetBalance(ac.Address()), "block", header.Number.Uint64())
		return nil
	}
	return ac.PerformRedistribution(header, statedb, blockGas)
}

func (ac *Contract) Address() common.Address {
	if reflect.DeepEqual(ac.address, common.Address{}) {
		addr, err := ac.bc.Config().AutonityContractConfig.GetContractAddress()
		if err != nil {
			log.Error("Cant get contract address", "err", err)
		}
		return addr
	}
	return ac.address
}

func (ac *Contract) abi() (*abi.ABI, error) {
	ac.Lock()
	defer ac.Unlock()
	if ac.contractABI != nil {
		return ac.contractABI, nil
	}
	ABI, err := abi.JSON(strings.NewReader(ac.bc.Config().AutonityContractConfig.ABI))
	if err != nil {
		return nil, err
	}
	ac.contractABI = &ABI
	return ac.contractABI, nil

}
