## v2.1.0

Thu Jan  3 22:11:50 CST 2019

- Pledge 1Token, get 100000Gas immediately, 1token can generate 100000Gas per day, and the user holds the Gas limit of 300,000 times of the pledge token.
- To increase the Gas charge, the JS contract transfer is about 37902Gas, and the transfer of the system contract token.iost is about 7668Gas.
- The user will charge a 2% handling fee for the purchase of RAM with the system, and the handling fee will be destroyed.
- RAM can be rent, but can't be traded.
- Add the fields interface of querying state db. For example, you can query all the tokens of an account.
- Smart contracts and transactions involving token transactions must set "amountLimit" in abi and tx with format "iost": "100", and wildcard format like "*": "unlimited" is supported now

