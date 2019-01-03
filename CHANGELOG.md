## v2.1.0

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
