<p>
<img src="https://i.loli.net/2019/02/24/5c72b13b98f8e.jpeg" >
</p>

# IOST - A Scalable & Developer Friendly Blockchain 

[![Go Report Card](https://goreportcard.com/badge/github.com/iost-official/go-iost)](https://goreportcard.com/report/github.com/iost-official/go-iost)
[![Build Status](https://api.travis-ci.com/iost-official/go-iost.svg?branch=master)](https://api.travis-ci.com/iost-official/go-iost.svg?branch=master)
[![codecov](https://codecov.io/gh/iost-official/go-iost/branch/master/graph/badge.svg)](https://codecov.io/gh/iost-official/go-iost)

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


