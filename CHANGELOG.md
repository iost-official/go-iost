## v3.0.10

Mon Apr 22 14:47:14 CST 2019

- Improve virtual machine stability.

## v3.0.9

Fri Apr 12 14:18:46 CST 2019

- Add goroutine pool for request handler of synchro.
- Clean up code for consensus/pob.
- Clean up code for core/txpool.
- Improve stability for vm.

## v3.0.8

Tue Apr  9 11:16:15 CST 2019

- Update metrics of system info.
- Add the onlyIssuerCanTransfer field to the tokeninfo interface.
- Add a more detailed system recovery log.
- Fix sync module bug.
- Fix the panic of Fixed lib.
- Fix bugs that generate block time errors.
- Fix the install command of the Makefile.
- Fix iwallet multi-signature function.

## v3.0.7

Fri Mar 29 16:47:15 CST 2019

- Add an RPC interface prompt message.
- Refactor code structure.

## v3.0.6

Thu Mar 28 12:50:15 CST 2019

- Add server time in getNodeInfo.
- Add tokeninfo RPC.
- Refactor code structure.

## v3.0.5

Fri Mar 22 13:55:49 CST 2019

- Add new RPC api.
- Add refine amount.
- Optimize synchronization logic.

## v3.0.4

Sat Mar 16 20:58:46 CST 2019

- Optimize RPC stability.

## v3.0.3

Fri Mar 15 16:36:17 CST 2019

- Fix high cpu load problem.
- Tune synchro module parameters.
- Improve stability for vm.

## v3.0.2

Wed Mar 13 16:46:33 CST 2019

- Rewrite synchronization module.
- Improve iwallet availability.
- Reduce the data size of p2p routing syncing.
- RPC: Add query interface of vote reward and block reward.
- RPC: Transaction query returns information to increase the Block number.

## v3.0.1

Thu Feb 28 10:49:56 CST 2019

- Fix memory leaks during synchronization.
- Optimize the routing table policy of the p2p module.

## v3.0.0

Sat Feb 23 13:16:25 CST 2019

- Add base metrics.
- Add iwallet permission replacement.
- Optimize verification of transaction time and block time.

## v3.0.0 rc5

Thu Feb 21 14:52:28 CST 2019

- Fix read auth bug.
- Optimize P2P routing table query logic.
- Optimize JS contract safety.
- Optimize synchronizer block hash query logic.

## v3.0.0 rc4

Mon Feb 18 20:46:06 CST 2019

- Update genesis contract, including 'producer_vote' and 'bonus'.
- Fix a witness bug.
- Fix a panic issue in vm crypto library.
- Fix verification error in 'publish'.
- Fix safety issue in 'batchVerify'.

## v3.0.0 rc3

Thu Feb 14 16:37:23 CST 2019

- Improve stability for consensus and synchronizer module.

## v3.0.0 rc2

Wed Feb 13 16:08:40 CST 2019

- Improve stability for consensus module.
- Optimize usability of iWallet.

## v3.0.0 rc1

Mon Feb 11 18:14:20 CST 2019

- Improve stability.
- Add 'verify' method to 'IOSTCrypto' object in order to verify signature.

## Everest v2.5.0

Thu Jan 31 14:29:15 CST 2019

- Fix known issue.
- Improve stability.
- Add system contract command to iWallet.

## Everest v2.4.0

Mon Jan 28 16:54:43 CST 2019

- Shrink docker image size.
- Fix bug: iwallet 'compile' command fails to generate contract abi on Linux.
- Increase block packing time from 300ms to 500ms.
- Modify 'maxTxLimitTime' from 100ms to 200ms.
- Increase gas charged for 'setCode'.
- Disable account referrer reward.
- Add a reserved field in transaction.

## Everest v2.3.1

Sat Jan 19 18:32:18 CST 2019

- Add can_update to exchange.iost contract.

## Everest v2.3.0

Sat Jan 19 17:25:28 CST 2019

- Increase minimum GasLimit of transaction from 5000 to 6000.
- RPC: add "voteInfos" to getAccount api.
- Complete vote, dividend test.
- iWallet remove npm package dependencies.
- Add 'exchange.iost' system contract, used for creating accounts and transferring.

## Everest v2.2.1

Sat Jan 12 20:36:05 CST 2019

- Fix `iWallet` bug.

## Everest v2.2.0

Sat Jan 12 17:37:34 CST 2019

- Change naming style of all functions in system contracts from `ThisNameStyle` to `thisNameStyle`.
- Remove `keypair pubkey ID`. Change pubkey ID ( "IOST" + `base58`(pubkey_bytes + `crc32`(pubkey_bytes)) to `base58`(pubkey_bytes) for simplicity;
  `signup` and `vote` will be affected.
- Add transaction replay protection.
- Add `ChainID` field in each transaction.
- Rewrite serialization to make it simpler and more efficient.
- Update RPC; redefine `groups` to `group_names` in struct `Permission`.
- Adjust genesis configuration.
- Add `contractOwner()` system method to JS smart contract.

## Everest v2.1.0

Thu Jan  3 22:11:50 CST 2019

- Pledge 1Token, get 100000Gas immediately, 1token can generate 100000Gas per day, and the user holds the Gas limit of 300,000 times of the pledge token.
- Increase Gas charge; now a JS contract transfer costs about 37902Gas, and a system contract (iost.token) costs about 7668Gas.
- System Contract: buying ram fee changes from 1% to 2%, then the handling fee will be destroyed.
- RAM can be rent, but can't be traded.
- Add the fields interface of querying state db; i.e. , you can query all the tokens of an account.
- Add `publish` for iwallet; it publishes js and abi to blockchain. The old `compile` command now can only generate abi file; it cannot publish contract any longer.
- Iwallet needs amount_limit for every transaction, i.e. `"iost:300|ram:200"` or `"*:unlimited"`.
- Command flag changes for iwallet: gasLimit -> gas_limit, gasRatio -> gas_ratio (edited).
- More friendly error message in RPC and iWallet.
- API `getAccount` returns detailed ram info: used, available, total.
- System Contract: `SignUp` needs the creator pledging 10IOST for gas when creating new account.
