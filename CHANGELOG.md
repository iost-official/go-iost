## v3.5.0
Mon Mar  8 17:06:41 CST 2021

Compatible upgrade.
- Migrate to go.mod. go-iost does not need GOPATH any more
- Upgrade go version to 1.16
- Upgrade base docker image to ubuntu18.04
- iWallet can be built when CGO\_ENABLED=0
- Upgrade some third party dependencies including protobuf grpc


## v3.4.0
Wed Nov 25 18:06:58 CST 2020

Breaking Change! Every node should upgrade before 2020.12.10.
- reserve 'husd' token for Huobi (breaking change)
- add original contract source code in `getContract` rpc. (only for new contracts)

## v3.3.6
Mon Aug 31 20:31:02 CST 2020
- Fixed a memory issue
- Remove deprecated keystore format
- Enable more linters
- Improve system abi 

## v3.3.5
Wed Jun 10 13:28:27 CST 2020
- Hotfix for a bug where some corner case tx cannot be confirmed

## v3.3.4
Thu May 28 14:34:58 CST 2020
- Upgrade Go to 1.14
- Mining nodes have better logs for troubleshooting
- Security fixes

## v3.3.3 
Fri Apr 10 08:36:50 UTC 2020

- Add ListContractStorage rpc
- Mitigate TimeError of rpc
- iwallet can execute a new message call immediately without creating a transaction on the block chain.


## v3.3.2
Mon Mar  2 12:23:18 CST 2020

- Switch dependency management to go module.
- Upgrade dependencies.
- Add a tool to prune blockchain storage to save disk usage.

## v3.3.1
Fri Dec 13 11:51:32 CST 2019

- Add ripemd160 hash function to chain api, make atomic swap with BTC and ETH possible.
- Fix some CGO memleak of VM.


## v3.3.0
Fri Sep 20 15:21:41 CST 2019

- Optimize code of common.
- Enrich go-sdk APIs.
- Adjust p2p default config.
- In addition to the `active` permission, other permissions must be declared in the `signer` field.
- The `returns` field is cleared when the transaction fails to execute; the `returns` and `receipts` fields are cleared when the ram fails.

## v3.2.0
Mon Jul 29 15:59:39 CST 2019

- Optimize the snapshot function, new nodes can quickly complete data synchronization.
- Reorganized code of `common` package.
- RPC increases the number of query contract votes.
- Reorganized code of RPC.
- Fix bug with high CPU usage.

## v3.1.1

Wed Jun 12 11:18:28 CST 2019

- Reorganized code of blockcache.
- Add detail log for blockcache.
- Fix the bug of synchronous fork in testnet.
- Fix bugs in RPC concurrent requests.
- Fix bug with publish contract crash.

## v3.1.0

Mon May  6 11:46:27 CST 2019

- Reorganized code module.
- Add hard fork framework and a hard-fork upgrade:
  - AmountLimit in a transaction cannot contain the same token
  - Optimize the logic of RequireAuth function
  - Fix mapDel bug
- Improve synchronization performance.
- Fix blockcache memory leak and data race.
- Add more detailed logs.
- Fix bug for Fixed library.
- Optimized transaction verification processing.

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
