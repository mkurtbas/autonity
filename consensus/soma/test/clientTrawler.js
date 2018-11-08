const rlp = require('rlp')
const net = require('net');
const Web3 = require('web3');

const secp256k1 = require('secp256k1')
const Buffer = require('safe-buffer').Buffer

const CONTRACT_ADDRESS = '0x325e08c6ca2b28253ae2280403ba9095895109f3'

const ecrecover = (msgHash, v, r, s, chainId) => {
  const signature = Buffer.concat([r,s], 64)
  const senderPubKey = secp256k1.recover(msgHash, signature, v)
  return secp256k1.publicKeyConvert(senderPubKey, false).slice(1)
}

/*
const HTTP_URI='http://localhost'
const HTTP_PORT=8501

const WS_URI='ws://localhost'
const WS_PORT=8545

const web3 = new Web3(WS_URI + ':' + WS_PORT)
*/

const sigHash = (header) => {
	const encodedHeader = rlp.encode([
		header.parentHash,
		header.sha3Uncles,
		header.miner,
		header.stateRoot,
		header.transactionsRoot,
		header.receiptsRoot,
		header.logsBloom,
		web3_0.utils.toBN(header.difficulty),
		web3_0.utils.toBN(header.number),
		header.gasLimit,
		header.gasUsed,
		web3_0.utils.toBN(header.timestamp),
		header.extraData,
		header.mixHash,
		header.nonce,
	])
	return web3_0.utils.sha3(encodedHeader)
}

const splitSignature = signatureStr => {
  const r = Buffer.from(signatureStr.slice(2).slice(0,32*2).padStart(32*2,0),'hex')
  const s = Buffer.from(signatureStr.slice(2).slice(32*2,(32*2)*2).padStart(32*2,0),'hex')
  const v = Number('0x' + signatureStr.slice(2).slice((32*2)*2))

  //console.log(`SplitSignature():\n\tr: ${r.toString('hex')}`, `\n\ts: ${s.toString('hex')}`,`\n\tv: ${v}`)

  return { r, s, v }
}

const getBlockSignerAddress = async header  => {
  const extraVanity = 32
  const extraSeal = 65 // r(32 bytes) + s(32 bytes) + v(1 byte)

  const signature = '0x' + header.extraData.slice(-(extraSeal*2))
  const extraDataUnsigned = header.extraData.slice(0,header.extraData.length-(extraSeal*2))//.padEnd(header.extraData.length,0)

  const blockHeaderNoSignature = Object.assign({},header, {extraData: extraDataUnsigned})
  const blockHashNoSignature = sigHash(blockHeaderNoSignature)

  const unsignedBlockBuffer = Buffer.from(blockHashNoSignature.slice(2),'hex')

  const signerAddress = await web3_0.eth.accounts.recover(blockHashNoSignature, signature, true)
  return signerAddress
}

const memBlocks = {}

const printHeader =  (clientId) => async header => {
  const signerAddress = await getBlockSignerAddress(header)

  const accounts = await web3_0.eth.getAccounts()
  const signerIndex = accounts.findIndex(acc => acc === signerAddress)

  hashInfo = {signerAddress, signerIndex, confirmations: 1, clientId: [clientId]}
  if(!memBlocks[header.number]){
    memBlocks[header.number] = { hash: { [header.hash]: hashInfo} }
  } else {
    if(!(memBlocks[header.number].hash[header.hash])){
      memBlocks[header.number].hash[header.hash] = hashInfo
    } else {
      memBlocks[header.number].hash[header.hash].confirmations = memBlocks[header.number].hash[header.hash].confirmations + 1
      memBlocks[header.number].hash[header.hash].clientId.push(clientId)
    }

    if(Object.keys(memBlocks[header.number].hash).length >  1)
      console.log('-------------\nBAD HASH: ', header.number,' ', JSON.stringify(memBlocks[header.number],' ',' '),'\n-------------')
  }

  console.log(`[${clientId}] Block number: ${header.number.toString().padStart(5,0)} \
Signer Index: ${signerIndex} \
Confirmations: ${memBlocks[header.number].hash[header.hash].confirmations} \
Block Hash: ${header.hash} Parent Hash: ${header.parentHash}`)
}

const getThreshold = async (contractAddr) => {
  const thresholdFunc = web3_0.utils.sha3('threshold()')
  const callRet = await web3_0.eth.call({ to: contractAddr, data: thresholdFunc})
  console.log(callRet.slice(2))
}

const getVotes = async (contractAddr) => {
  const thresholdFunc = web3_0.utils.sha3('vote()')
  const callRet = await web3_0.eth.call({ to: contractAddr, data: thresholdFunc})
  console.log(callRet.slice(2))
}

const getValidators = async (contractAddr) => {
  const validatorsFunc = web3_0.utils.sha3('validators(uint256)').slice(0,(4*2)+2)
  const validators = []
  for(let i =0; true; i++) {
    const argument = i.toString().padStart(32*2,0)
    const callRet = await web3_0.eth.call({ to: contractAddr, data: validatorsFunc+argument })
    if(callRet.length === 2) break
    else validators.push('0x'+callRet.slice(-20*2))
  }
  return validators
}

const listValidators = async (contractAddr) => {
  const validators = await getValidators(contractAddr)
  validators.forEach((acc,idx) => console.log(`(${idx.toString().padStart(2,0)}) ${acc}`))
}

const listAccounts = async () => {
  const accounts = await web3_0.eth.getAccounts()
  accounts.forEach((acc,idx) => console.log(`(${idx.toString().padStart(2,0)}) ${acc}`))
}

const castVote = async (contractAddr, from, candidate) => {
  const castVoteFunc = web3_0.utils.sha3('CastVote(address)').slice(0,(4*2)+2)
  const argument = candidate.slice(2).padStart(32*2,0)
  const data = castVoteFunc+argument
  const tx = { from, to: contractAddr, data, gas: "0xffffffff" }
  await web3_0.eth.personal.unlockAccount(from,'password',600)
  return web3_0.eth.sendTransaction(tx)
}

const removeValidator = async (contractAddr) => {
  const validators = await getValidators(contractAddr)
  if(validators.length <= 1) {
    console.log('1 or less validators in the validator list!')
    return
  }
  const leavingValidator = validators.slice(-1)[0]
  const votingValidators = validators.slice(0,-1)
  console.log(`Voting to remove validator: ${leavingValidator}\n\tVoters: ${votingValidators.join('\n\t\t')}`)
  const vontingPromi = votingValidators.map(v => castVote(contractAddr,v,leavingValidator))
  const tx = await Promise.all(vontingPromi)
  console.log(`Finished voting to remove validator: ${leavingValidator}\n${JSON.stringify(tx,' ',' ')}`)
}

const addValidator = async (contractAddr) => {
  const accounts = await web3_0.eth.getAccounts()
  const validators = await getValidators(contractAddr)
  if(validators.length >= accounts.length) {
    console.log('The validator list is longer or the same as the unlocked accounts in the node!')
    return
  }
  const candidates = accounts.map(acc => acc.toLowerCase()).filter(acc => !validators.includes(acc))
  const candidateValidator = candidates.slice(0,1)[0]
  const votingValidators = validators
  console.log(`Voting to add validator: ${candidateValidator}\n\tVoters: ${votingValidators.join('\n\t\t')}`)
  const vontingPromi = votingValidators.map(v => castVote(contractAddr,v,candidateValidator))
  const tx = await Promise.all(vontingPromi)
  console.log(`Finished voting to add validator: ${candidateValidator}\n${JSON.stringify(tx,' ',' ')}`)
}

// const onKeyPress = async key => {
//   switch (key.toString()) {
//     case '+':
//       addValidator(CONTRACT_ADDRESS)
//       break
//     case '-':
//       removeValidator(CONTRACT_ADDRESS)
//       break
//     case 'v':
//       console.log('\n================== Validators ==================')
//       await listValidators(CONTRACT_ADDRESS)
//       console.log('================================================\n')
//       break
//     case 'l':
//       console.log('\n=================== Accounts ===================')
//       await listAccounts()
//       console.log('================================================\n')
//       break
//     case 'c':
//       console.log('\n================== Vote========================')
//       await getVotes(CONTRACT_ADDRESS)
//       console.log('================================================\n')
//       break
//     case 't':
//       console.log('\n================== Threshold ===================')
//       await getThreshold(CONTRACT_ADDRESS)
//       console.log('================================================\n')
//       break
//     case 'q':
//       process.exit()
//       break
//     case 'h':
//       console.log('\n================== Help ==================\n'
//         + '(+): cast a vote to add validator (the first on accounts that is not a validator)\n'
//         + '(-): cast a vote to remove validator (the last on the list of validators)\n'
//         + '(t): print validator turn threshold\n'
//         + '(c): print validator vote threshold\n'
//         + '(v): print list of validators\n'
//         + '(l): print list of accounts\n'
//         + '(q): quit\n'
//         + '===========================================\n'
//       )
//       break
//     default:
//       console.log(`UNKNOWN KEY! (${key})`)
//       break
//   }
// }
// onKeyPress('h')

// process.stdin.setRawMode(true)
// process.stdin.on('data', onKeyPress)
const printChanged = (clientId) => (data) => {
  console.log(`CHANGED: [${clientId}] ${JSON.stringify(data)}`)
}

// Using the IPC provider in node.js
const web3_0 = new Web3('/home/user97/repos/network-setups/network-geth/data_node_0/geth.ipc', net);
const web3_1 = new Web3('/home/user97/repos/network-setups/network-geth/data_node_1/geth.ipc', net);
const web3_2 = new Web3('/home/user97/repos/network-setups/network-geth/data_node_2/geth.ipc', net);
const web3_3 = new Web3('/home/user97/repos/network-setups/network-geth/data_node_3/geth.ipc', net);
const web3_4 = new Web3('/home/user97/repos/network-setups/network-geth/data_node_4/geth.ipc', net);
// const web3_5 = new Web3('/home/user97/repos/network-setups/network-geth/data_node_5/geth.ipc', net);
// subscribe
web3_0.eth.subscribe('newBlockHeaders').on("data", printHeader(0)).on("error", console.error).on("changed", printChanged(0))
web3_1.eth.subscribe('newBlockHeaders').on("data", printHeader(1)).on("error", console.error).on("changed", printChanged(1))
web3_2.eth.subscribe('newBlockHeaders').on("data", printHeader(2)).on("error", console.error).on("changed", printChanged(2))
web3_3.eth.subscribe('newBlockHeaders').on("data", printHeader(3)).on("error", console.error).on("changed", printChanged(3))
web3_4.eth.subscribe('newBlockHeaders').on("data", printHeader(4)).on("error", console.error).on("changed", printChanged(4))
// web3_5.eth.subscribe('newBlockHeaders').on("data", printHeader(5)).on("error", console.error).on("changed", printChanged(5))
