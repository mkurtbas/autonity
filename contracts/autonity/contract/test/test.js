'use strict';
const assert = require('assert');
const utils = require('./test-utils');
//todo: move gas analysis to separate js file

contract('Autonity', function (accounts) {
    const validatorsList = [
        accounts[1],
        accounts[2],
        accounts[3],
        accounts[4],
        accounts[5],
    ];

    const whiteList = [
        "enode://d73b857969c86415c0c000371bcebd9ed3cca6c376032b3f65e58e9e2b79276fbc6f59eb1e22fcd6356ab95f42a666f70afd4985933bd8f3e05beb1a2bf8fdde@172.25.0.11:30303",
        "enode://1f207dfb3bcbbd338fbc991ec13e40d204b58fe7275cea48cfeb53c2c24e1071e1b4ef2959325fe48a5893de8ff37c73a24a412f367e505e5dec832813da546a@172.25.0.12:30303",
        "enode://e766ac390e2d99b559aef773c3656fa8d50df2310496ac26ca6c3fc84e21dabb8a0162cc8e34f938d45e0a8ed04955f8ddf1c380182f8ef17a3f08885064505f@172.25.0.13:30303",
        "enode://438a5c2cd8fdc2ecbc508bf7362e41c0f0c3754ba1d3267127a3756324caf45e6546b02140e2144b205aeb372c96c5df9641485f721dc7c5b27eb9e35f5d887b@172.25.0.14:30303",
        "enode://3ce6c053cb563bfd94f4e0e248510a07ccee1bc836c9784da1816dba4b10564e7be1ba42e0bd8d73c8f6274f8e9878dc13814adb381c823264265c06048b4b59@172.25.0.15:30303"
    ];
    const userTypes = [2, 2, 2, 2, 2];
    const stakes = [100, 90, 80, 110, 120];
    const commisionRate = [0, 0, 0, 0, 0];
    const minGasPrice = 0;
    const bondPeriod = 100;
    const committeeSize = 1000;
    const operator = accounts[0];
    const deployer = accounts[8];
    const version = "v0.0.0";
    let token;

    describe('Metrics', function() { // test failing

        beforeEach(async function(){
            token = await utils.deployContract(validatorsList, whiteList,
                userTypes, stakes, commisionRate, operator, minGasPrice, bondPeriod, committeeSize, version, {from: accounts[8]} );
        });

        it('test dump network Economic metric data.', async function () {
            let data = await token.dumpEconomicsMetricData({from: operator});
            let minGasPrice = await token.getMinimumGasPrice({from: operator});
            let sum = 0;
            for (let i = 0; i < data.accounts.length; i++) {
                let stake = await token.getAccountStake(data.accounts[i], {from: operator});
                assert.deepEqual(Number(data.stakes[i]), Number(stake));
                sum += Number(data.stakes[i])
            }

            assert.deepEqual(data.accounts, validatorsList);
            assert.deepEqual(Number(data.mingasprice), Number(minGasPrice))
            assert.deepEqual(Number(data.stakesupply), sum)
        });
    });

    describe('Initial state', function() {
        beforeEach(async function(){
            token = await utils.deployContract(validatorsList, whiteList,
                userTypes, stakes, commisionRate, operator, minGasPrice, bondPeriod, committeeSize, version, { from:accounts[8]} );
        });

        it('test validator can get initial validator list', async function () {
            let getValidatorsResult = await token.getValidators({from: operator});
            assert.deepEqual(getValidatorsResult, validatorsList);
        });

        it('test non validator account can get initial validator list', async function () {
            let getValidatorsResult = await token.getValidators({from: accounts[7]});
            assert.deepEqual(getValidatorsResult, validatorsList)
        });

        it('test default minimum gas price equals to 0', async function () {
            let minGasPrice = await token.getMinimumGasPrice({from: operator});
            assert(0 == minGasPrice, "default min gas price different of zero");
        });

        it('test validator can get the initial enodes whitelist', async function () {
            let enodesWhitelist = await token.getWhitelist({from: operator});
            assert.deepEqual(enodesWhitelist, whiteList);
        });
    });

    describe('Fee Distribution', function() {

        beforeEach(async function(){
            token = await utils.deployContract(validatorsList, whiteList,
                userTypes, stakes, commisionRate, operator, minGasPrice, bondPeriod, committeeSize, version, { from:accounts[0]} );
        });

        it('test redistribution fails with empty balance', async function () {
            try {
                await token.finalize(10000, {from: deployer});
                assert.fail('Expected throw not received', r);
            } catch (e) {

            }
        });

        it('test redistribution fails with not deployer', async function () {
            let st = await token.getStakeholders();

            for (let i = 0; i < st.length; i++) {
                await web3.eth.sendTransaction({from: st[i], to: token.address, value: 10000});
            }
            let balance = await web3.eth.getBalance(token.address);
            if (balance < 10000) {
                assert.fail("incorrect balance")
            }
            try {
                await token.finalize(10000, {from: accounts[0]});
                assert.fail('Expected throw not received', r);
            } catch (e) {

            }
        });

        it('test redistribution', async function () {
            let st = await token.getStakeholders({from: operator});

            for (let i = 0; i < st.length; i++) {
                await web3.eth.sendTransaction({from: st[i], to: token.address, value: 10000});
            }

            let balances = [];
            for (let i = 0; i < st.length; i++) {
                balances[i] = await web3.eth.getBalance(st[i]);
            }

            let performAmount = 10000;
            let totalStake= stakes.reduce((a,b) => a + b);
            let stakeholdersPart = stakes.map(element => element * performAmount / totalStake);

            await token.finalize(performAmount, {from: operator});

            let balancesAfter = [];
            for (let i = 0; i < st.length; i++) {
                balancesAfter[i] = await web3.eth.getBalance(st[i]);
            }

            for (let i = 0; i < st.length; i++) {
                let check = web3.utils.toBN(balancesAfter[i])
                    .sub(web3.utils.toBN(balances[i]))
                    .eq(web3.utils.toBN(stakeholdersPart[i]));

                assert(check, "not equal")
            }
        });
    });

    describe('Governance - System Operator', function() {

        beforeEach(async function(){
            token = await utils.deployContract(validatorsList, whiteList,
                userTypes, stakes, commisionRate, operator, minGasPrice, bondPeriod, committeeSize, version, { from:accounts[8]} );
        });

        it('test Governance operator can add/remove to whitelist', async function () {
            let enode = "enode://testenode";
            let tx = await token.addValidator(accounts[8], 20, enode, {from: operator});
            // console.log("\tGas used to add validator to whitelist = " + tx.receipt.gasUsed.toString() + " gas");
            let getValidatorsResult = await token.getWhitelist({from: operator});
            let expected = whiteList.slice();
            expected.push(enode);
            assert.deepEqual(getValidatorsResult, expected);

            tx = await token.removeUser(accounts[8], {from: operator});
            // console.log("\tGas used to remove val from whitelist = " + tx.receipt.gasUsed.toString() + " gas");
            getValidatorsResult = await token.getWhitelist({from: accounts[1]});
            assert.deepEqual(getValidatorsResult, whiteList);
        });

        it('test add validator and check that it is in get validator list', async function () {
            let expected = validatorsList.slice();
            expected.push(accounts[7]);
            let tx = await token.addValidator(accounts[7], 100, "not nil enode", {from: operator});
            // console.log("\tGas used to add new validator = " + tx.receipt.gasUsed.toString() + " gas");
            let getValidatorsResult = await token.getValidators({from: operator});
            assert.deepEqual(expected, getValidatorsResult);

            tx = await token.removeUser(accounts[7], {from: operator});
            // console.log("\tGas used to remove validator = " + tx.receipt.gasUsed.toString() + " gas");
            getValidatorsResult = await token.getValidators({from: operator});
            assert.deepEqual(validatorsList, getValidatorsResult)
        });

        it('test create participant account check it and remove it', async function () {
            let tx = await token.addParticipant(accounts[9], "some enode", {from: operator});
            //console.log("\tGas used to add participant = " + tx.receipt.gasUsed.toString() + " gas");
            let addMemberResult = await token.checkMember(accounts[9]);

            assert(true === addMemberResult);

            tx = await token.removeUser(accounts[9], {from: operator});
            // console.log("\tGas used to remove participant = " + tx.receipt.gasUsed.toString() + " gas");
            let removeMemberResult = await token.checkMember(accounts[9]);

            assert(false === removeMemberResult);
        });

        it('test non validator cannot add validator', async function () {

            try {
                let r = await token.addValidator(accounts[7], {from: accounts[6]})

            } catch (e) {
                let getValidatorsResult = await token.getValidators({from: operator});
                assert.deepEqual(getValidatorsResult, validatorsList);
                return
            }

            assert.fail('Expected throw not received');
        });

        it('test non Governance operator cannot add validator', async function () {
            let enode = "enode://testenode";
            try {
                let r = await token.addValidator(accounts[6], 20, enode, {from: accounts[6]});
                assert.fail('Expected throw not received');

            } catch (e) {
                //skip error
                let getWhitelistResult = await token.getWhitelist({from: operator});
                assert.deepEqual(getWhitelistResult, whiteList);
            }
        });

        it('test non Governance operator cannot remove user', async function () {
            let enode = whiteList[0];
            try {
                let r = await token.removeUser(accounts[2], {from: accounts[6]});
                assert.fail('Expected throw not received', r);

            } catch (e) {
                let getWhitelistResult = await token.getWhitelist({from: operator});
                assert.deepEqual(getWhitelistResult, whiteList);
            }
        });

        it('test create/remove participants by non governance operator', async function () {
            let errorOnAddNewMember = false;
            let errorOnRemoveMember = false;

            try {
                await token.addParticipant(accounts[8], "some enode", {from: accounts[7]});
            } catch (e) {
                errorOnAddNewMember = true
            }
            let addMemberResult = await token.checkMember(accounts[8]);
            assert(false === addMemberResult);

            await token.addParticipant(accounts[8], "some enode", {from: operator});

            addMemberResult = await token.checkMember(accounts[8]);
            assert(true === addMemberResult);


            try {
                await token.removeUser(accounts[8], {from: accounts[7]});
            } catch (e) {
                errorOnRemoveMember = true
            }

            let removeMemberResult = await token.checkMember(accounts[8]);
            assert(true === removeMemberResult);
            await token.removeUser(accounts[8], {from: operator});

            removeMemberResult = await token.checkMember(accounts[8]);
            assert(false === removeMemberResult);

            assert(true === errorOnAddNewMember);
            assert(true === errorOnRemoveMember);

        });
    });

    describe('Committee Selection', function () {

        beforeEach(async function(){
            token = await utils.deployContract(validatorsList, whiteList,
                userTypes, stakes, commisionRate, operator, minGasPrice, bondPeriod, committeeSize, version,  { from:accounts[8]} );
        });

        it('test set max committee size by operator account', async function () {

            await token.setCommitteeSize(4, {from: operator});
            let maxCommitteeSize = await token.getMaxCommitteeSize();
            assert(4 == maxCommitteeSize, "maxCommittee size was not set correctly");

        });

        it('test regular validator cannot set the max committee size', async function() {

            let initMaxCommitteeSize = await token.getMaxCommitteeSize();

            try {
                let r = await token.setCommitteeSize(50, {from: accounts[6]});
                assert.fail('Expected throw not received', r);

            } catch (e) {
                let maxCommitteeSize = await token.getMaxCommitteeSize();
                assert.deepEqual(initMaxCommitteeSize,  maxCommitteeSize);
            }
        });

        it('test set committee when committee size is equal than the number of validators', async function() {

            let validators = await token.getValidators();
            let committeeSize = validators.length;

            await token.setCommitteeSize(committeeSize, {from: operator});
            let maxCommitteeSize = await token.getMaxCommitteeSize();
            assert(committeeSize == maxCommitteeSize, "maxCommittee size was not set correctly");

            try {
                let r = await token.setCommittee({from: deployer});
                assert.fail('Expected throw not received', r);
            } catch (e) {

            }
            let committeeResult = await token.getCommittee();
            let committeeValidators = [];

            for (let i = 0; i < committeeResult.length; i++) {
                committeeValidators.push(committeeResult[i][0])
            }

            assert.deepEqual(committeeValidators.sort(), validators.sort(), "Committee should be equal than validator set");

        });

        it('test set committee when committee size is greater than the number of validators', async function() {
            let validators = await token.getValidators();
            let committeeSize = validators.length + 5;

            await token.setCommitteeSize(committeeSize, {from: operator});
            let maxCommitteeSize = await token.getMaxCommitteeSize();
            assert(committeeSize == maxCommitteeSize, "maxCommittee size was not set correctly");

            try {
                let r = await token.setCommittee({from: deployer});
                assert.fail('Expected throw not received', r);
            } catch (e) {

            }
            let committeeResult = await token.getCommittee();
            let committeeValidators = [];

            for (let i = 0; i < committeeResult.length; i++) {
                committeeValidators.push(committeeResult[i][0])
            }

            assert.deepEqual(committeeValidators.sort(), validators.sort(), "Committee should be equal than validator set");

        });

        it('test set committee when committee size is smaller than the number of validators', async function() {

            try {
                let validators = await token.getValidators();
                let committeeSize = validators.length - 2;

                await token.setCommitteeSize(committeeSize, {from: operator});
                let maxCommitteeSize = await token.getMaxCommitteeSize();
                assert(committeeSize == maxCommitteeSize, "maxCommittee size was not set correctly");

                let r = await token.setCommittee({from: deployer});
                assert.fail('Expected throw not received', r);

                let committeeResult = await token.getCommittee();
                let committeeValidators = [];

                for (let i = 0; i < r.length; i++) {
                    committeeValidators.push(r[i][0])
                }

                // Mock committee selection
                let indexesToBeRemoved = [1,2]
                while(indexesToBeRemoved.length) {
                    validators.splice(indexesToBeRemoved.pop(), 1);
                }
                assert.deepEqual(committeeValidators.sort(), validators.sort(), "Error while creating new committee");
            }catch (e) {

            }

        });
    });

    describe('Stake Token', function() {

        beforeEach(async function(){
            token = await utils.deployContract(validatorsList, whiteList,
                userTypes, stakes, commisionRate, operator, minGasPrice, bondPeriod, committeeSize, version,  { from:accounts[8]} );
        });

        it('test create account, add stake, check that it is added, remove stake', async function () {
            await token.addStakeholder(accounts[7], "some enode", 0, {from: operator});
            let getStakeResult = await token.getStake({from: accounts[7]});
            assert(0 == getStakeResult, "unexpected tokens");

            let tx = await token.mintStake(accounts[7], 100, {from: operator});
            // console.log("\tGas used to mint stake = " + tx.receipt.gasUsed.toString() + " gas");
            getStakeResult = await token.getStake({from: accounts[7]});
            assert(100 == getStakeResult, "tokens are not minted");
            tx = await token.redeemStake(accounts[7], 100, {from: operator});

            // console.log("\tGas used to redeem stake = " + tx.receipt.gasUsed.toString() + " gas");

            getStakeResult = await token.getStake({from: accounts[7]});
            assert(0 == getStakeResult, "unexpected tokens");

            await token.removeUser(accounts[7], {from: operator});
        });

        it('test create account, get error when redeem empty stake', async function () {

            await token.addStakeholder(accounts[7], "some enode", 0, {from: operator});
            let getStakeResult = await token.getStake({from: accounts[7]});
            assert(0 == getStakeResult, "unexpected tokens not minted");

            try {
                await token.redeemStake(accounts[7], 100, {from: operator});
                assert.fail('Expected throw not received');
            } catch (e) {
                getStakeResult = await token.getStake({from: accounts[7]});
                assert(0 == getStakeResult, "unexpected tokens");
                await token.removeUser(accounts[7], {from: operator});
            }
        });

        it('test transfer stake', async function () {
            let getStakeResult = await token.getStake({from: validatorsList[2]});
            assert(stakes[2] == getStakeResult, "unexpected tokens");

            await token.redeemStake(accounts[2], 50, {from: operator});
            let balance_after = await token.getStake({from: validatorsList[2]});
            assert(stakes[2] - 50, balance_after);

            await token.removeUser(accounts[4], {from: operator});
            await token.removeUser(accounts[5], {from: operator});
        });

        it('test validator can get users list', async function () {
            var getValidatorsResult = await token.retrieveState({from: operator});
            let addresses = getValidatorsResult[0];
            let types = getValidatorsResult[1];
            let stake = getValidatorsResult[2];
            let enodes = getValidatorsResult[3];
            let commisionRate = getValidatorsResult[4];
            assert.deepEqual(getValidatorsResult[0], validatorsList);

            await token.addStakeholder(accounts[7], "some enode", 0, {from: operator});
            let getStakeResult = await token.getStake({from: accounts[7]});
            assert(0 == getStakeResult, "unexpected tokens");

            await token.addStakeholder(accounts[5], "some enode", 0, {from: operator});
            getStakeResult = await token.getStake({from: accounts[5]});
            assert(0 == getStakeResult, "unexpected tokens");

            await token.mintStake(accounts[7], 100, {from: operator});

            getStakeResult = await token.getStake({from: accounts[7]});
            assert(100 == getStakeResult, "tokens are not minted");

            let tx = await token.send(accounts[5], 50, {from: accounts[7]});
            // console.log("\tGas used to send state token = " + tx.receipt.gasUsed.toString() + " gas");

            getStakeResult = await token.getStake({from: accounts[7]});
            assert(50 == getStakeResult, "unexpected tokens");

            getStakeResult = await token.getStake({from: accounts[5]});
            assert(50 == getStakeResult, "unexpected tokens");

            await token.redeemStake(accounts[7], 50, {from: operator});
            await token.redeemStake(accounts[5], 50, {from: operator});
            await token.removeUser(accounts[7], {from: operator});
            await token.removeUser(accounts[5], {from: operator});
        });
    });
});
