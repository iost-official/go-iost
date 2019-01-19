## Everest v2.3.0

Sat Jan 19 17:25:28 CST 2019

- Increase minimum GasLimit of transaction from 5000 to 6000.
- RPC: add "voteInfos" to getAccount api
- Complete vote, dividend test
- iwallet remove npm package dependencies
- Add exchange.iost system contract, used for creating accounts and transferring

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
- User will be charge a 2% handling fee for the purchase of RAM with the system, then the handling fee will be destroyed.
- RAM can be rent, but can't be traded.
- Add the fields interface of querying state db; i.e. , you can query all the tokens of an account.

- Add `publish` for iwallet; it publishes js and abi to blockchain. The old `compile` command now can only generate abi file; it cannot publish contract any longer.
- Iwallet needs amount_limit for every transaction, i.e. `"iost:300|ram:200"` or `"*:unlimited"`.
- Command flag changes for iwallet: gasLimit -> gas_limit, gasRatio -> gas_ratio (edited).
- More friendly error message in RPC and iWallet.
- API `getAccount` returns detailed ram info: used, available, total
- System Contract: `SignUp` needs the creator pledging 10IOST for gas when creating new account
- System Contract: buying ram fee changes from 1% to 2%
