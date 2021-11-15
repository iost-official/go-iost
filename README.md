# IOST - A Scalable & Developer Friendly Blockchain 

IOST is a smart contract platform focusing on performance and developer friendliness. 

# Features

1. The V8 JavaScript engine is integrated inside the blockchain, so you can use JavaScript to write smart contracts!
2. The blockchain is highly scalable with thousands of TPS. Meanwhile it still has a more decentralized consensus than DPoS.
3. 0.5 second block, 0.5 minute finality.
4. Free transactions. You can stake coins to get gas.

# Development

### Environments

OS: Ubuntu 18.04 or later  
Go: 1.17 or later

IOST node uses CGO V8 javascript engine, so only x64 is supported now.

### Deployment

build local binary: `make build`  
start a local devnet: `make debug`     
build docker: `make image`  


For documentation, please visit: [IOST Developer](https://developers.iost.io)

Welcome to our [tech community at telegram](https://t.me/iostdev)

Happy hacking!


