<p>
<img src="https://upload-images.jianshu.io/upload_images/2056014-6069383bda7bd291.jpeg?imageMogr2/auto-orient/" >
</p>


# IOSBlockchain - A Secure & Scalable Blockchain for Smart Services

[![Go Report Card](https://goreportcard.com/badge/github.com/iost-official/go-iost)](https://goreportcard.com/report/github.com/iost-official/go-iost)
[![Build Status](https://travis-ci.org/iost-official/go-iost.svg?branch=develop)](https://travis-ci.org/iost-official/go-iost)

The Internet of Services (IOS) offers a secure & scalable infrastructure for online service providers. Its high TPS, scalable and secure blockchain, and privacy protection scales social and economic cooperation to a new level. For more information about IOS Blockchain, please refer to our [whitepaper](https://github.com/iost-official/Documents)

The source codes released in this repo represent part of our early alpha-quality codes, with the purpose to demonstrate our development progress. Currently, there are still parts of our project have been intensively working on in private repos. We will gradually release our code.

## Features of Everest v0.5

* IOST consensus mechanism: Proof of Believability (PoB) up and running
* A Faucet that provides testnet tokens to users
* A Wallet to store and send tokens
* Ability to run IOST testnet nodes
* A Blockchain Explorer to display transactions and blocks
* Build private IOST test networks
* Smart contracts system featuring key-value pair storage, inter-contract API calls and multiple signatures
* API oriented smart contracts to quickly write and deploy code
* A smart contract ‘Playground’ to test smart contracts locally
* A ‘Lucky Bet’ game running on testnet
* Real-time monitoring metrics and sending of alerts
* Deployed on real network environment, running on multiple nodes deployed worldwide in Tokyo, London, Singapore, Mumbai, Frankfurt, Seoul and Montreal

## TPS: Throughput Testing Outcome

Our private suite of experiments used virtual machine clusters of around 1700 and 900 slave nodes in Amazon’s Elastic Compute Cloud (EC2) with each single machine equipped with 1.73Ghz Xeon LC-3518, 32 GB memory, 256 KiB L2 Cache. The EC2 instances emulate semi-synthetic and real-world environments with deployment across 9 data centers in Asia, Europe and America.

We deployed our testnet node instances with the PoB consensus engine on up to 6 shards and achieved an average throughput between 7000-8000 transactions per second.

## Proof of Believability (PoB) up and running

Our proprietary consensus algorithm, Proof-of-Believability, is up and running in a real environment. The source code is available to view in the [consensus folder](https://github.com/iost-official/go-iost/tree/master/consensus) of the [IOST GitHub project](https://github.com/iost-official/go-iost).

The PoB consensus uses a node’s contribution and stake balance to gain block creation chances by becoming a validator. The algorithm factors in both transactions verified and token balance to determine the contribution to the network.

A challenge faced by traditional PoS consensus mechanisms is the tendency towards centralization. In order to mitigate this risk, after validating a block, the PoB system will automatically clear any remaining Servi token balance.

Servi tokens are implemented this way with the following desired properties: non-tradable, self-destructive and self-issuing. Each transaction verification counts as 1 Servi and each staked IOST counts as 1, the combination of the 2 is used to efficiently select validators.

## Upcoming releases and our plan for next stage

* We will have 2 or more major updated releases of the IOST testnet before the launch of our mainnet
* More extensive testing and general infrastructure developments
* Increased stability of node lifecycle
* Integration of Layer 1 and Layer 2 scaling solutions with our research progress
* Additional functionality and security checks for smart contracts
* PoB running under increasingly diverse environments
* IOST Virtual Machine optimization
* More documentation guidelines for developers



## Prerequisites

* Go 1.9 or newer (Go 1.11 recommended)
* `git-lfs` (v2.5.2 recommended)
* Rocksdb v5.14.3 or newer
* Docker v18.06.0-ce or newer (older versions are not tested)

## Environment

Currently, you can develop and deploy on below environments:

* [Mac OS X](https://developers.iost.io/docs/en/4-running-iost-node/Environment-Configuration/#mac-os-x)
* [Ubuntu/Linux](https://developers.iost.io/docs/en/4-running-iost-node/Environment-Configuration/#ubuntu-linux)
* [Docker](https://developers.iost.io/docs/en/4-running-iost-node/Environment-Configuration/#docker)

## Get Repository

Run the command to get the code repository:

```
git clone git@github.com:iost-official/go-iost.git
```

## Build

Run the command to compile and generate file in the `target` directory:

```
make build
```

## Run

Run the command to run a local node. Check iServer setup here: [iServer](iServer).

```
./target/iserver -f config/iserver.yml
```

## Docker

### Run

Run the command to run a local node with the docker:

```
docker run -it --rm iostio/iost-node:1.0.0
```

## Access the Testnet

### Update config

Change genesis settings as below:

```
genesis:
  creategenesis: true
  witnessinfo:
  - IOST2g5LzaXkjAwpxCnCm29HK69wdbyRKbfG4BQQT7Yuqk57bgTFkY
  - "100000000000000000"
  - IOST22TgXbjvgrDd3DuMkVufcWbYDMy7vpmQoCgZXmgi8eqM7doxWD
  - "100000000000000000"
  - IOSTAXksR6rKvmkjJyzhJJkDsG3yip47BJJWmbSTYqwqoNErBoN2k
  - "100000000000000000"
  - IOSTFPe9aXhZMmyvy6BsmgeucKEgzXy3zHMhsBFFeqNtKsqy98sbX
  - "100000000000000000"
  - IOST23xQCcviwn7AGxDnJbkL2Sjh8ijsKL6sPJWAkVEP8jACHLGknX
  - "100000000000000000"
  - IOST2CxDxZJwo2Useu2kMvZRTmMpHiwrK4UzQRLEQccLTfAmY9Z4Up
  - "100000000000000000"
  - IOSTKbYwTYpGZUTQqnmnbQAeJKhCBAMfW3pNvtJn6nEtVj6aozGMQ
  - "100000000000000000"
  - IOSTxUBnFHNBb22TSU8ruiEPfVUx6utxxbUcat3ZaDmtZea4roPES
  - "100000000000000000"
  - IOSTpWBkze9vPL3rxmnobgVN6s6WwHUFJGMo7wFcAHwkbhij3eDZY
  - "100000000000000000"
  - IOST27LJHEEBZ8oNqQR9EhutmybLuNdeitNfWdkoFk8MwQ2pSbifig
  - "100000000000000000"
  - IOST2AcBEJawoVzg4MW6UcvQsP6p6mSwACF7bbroNU2jBtE3MDSt6G
  - "100000000000000000"
  votecontractpath: config/
```

Change the settings of `p2p.seednodes` as below:

```
p2p:
  seednodes:
  - /ip4/18.218.255.180/tcp/30000/ipfs/12D3KooWLwNFzkAf3fRmjVRc9MGcn89J8HpityXbtLtdCtPSHDg1
```

Among the settings, the network IDs of seed nodes can be changed. Seed nodes of the testnet is shown below:

| Name   | Region | Network ID                                                                              |
| ------ | ------ | --------------------------------------------------------------------------------------- |
| node16 | US East | /ip4/18.218.255.180/tcp/30000/ipfs/12D3KooWLwNFzkAf3fRmjVRc9MGcn89J8HpityXbtLtdCtPSHDg1 |
| node17 | US West | /ip4/52.9.253.198/tcp/30000/ipfs/12D3KooWABS9bLYUnvmLYeuZvkgL2WY3TLHJDbmG2tUWB4GfJJiq   |
| node19 | Mumbai   | /ip4/13.127.153.57/tcp/30000/ipfs/12D3KooWAx1pZHvUq73UGMSXqjUBsKBKgXFoFBoXZZAhfvM9HnVr  |
| node20 | Seoul   | /ip4/52.79.231.23/tcp/30000/ipfs/12D3KooWCsq3Lfxe8E17anTred2o7X4cSZ77faai8hkHH611RjMp   |
| node21 | Singapore | /ip4/13.229.176.106/tcp/30000/ipfs/12D3KooWKGK1ah5JgMEic2dH8oYE3LMEZLBJUzCNP165tPaQnaW9 |
| node22 | Sydney   | /ip4/13.238.140.219/tcp/30000/ipfs/12D3KooWGHmaxL8LmRpvXoFPNYj3FavYgqqEBks4YPVUL6KRcQFs |
| node23 | Canada | /ip4/52.60.78.2/tcp/30000/ipfs/12D3KooWAivafPT52QEf2eStdXS4DjiRyLCGhLanvVgJ7hhbqans     |
| node24 | Germany   | /ip4/52.58.16.220/tcp/30000/ipfs/12D3KooWPKjYYL4tvbUQF2VzA1mg6XsByA8GVN4anDfrRxp9qdxm   |
| node25 | Ireland | /ip4/18.202.100.127/tcp/30000/ipfs/12D3KooWDL2BdvSR65kS2z8LX8142ksX35mNFWhtVpK6a24WXBoV |
| node26 | UK   | /ip4/35.176.96.113/tcp/30000/ipfs/12D3KooWHfCWdXnKkTqFYNh8AhrjJ21v7RrTTuwSBLztHgGLWYyX  |
| node27 | France   | /ip4/52.47.133.32/tcp/30000/ipfs/12D3KooWScNNuMLh1AEnWoNppXKY6qwVVGrvzYF4dKQxBMmnwW3b   |
| node28 | Brazil   | /ip4/52.67.231.15/tcp/30000/ipfs/12D3KooWRJxjPsVxRR7spvfRPRWzvGKZrWggRj5kEiqyS4tzPq78   |
| node40 | Tokyo   | /ip4/52.192.86.141/tcp/30000/ipfs/12D3KooWS4kyTpyjEA8ixqFGT7uLd4mAh4fYbYNYhaPYNEWE69BA  |

### Run iServer

Connect to Testnet by runing iServer with updated config:

```
./target/iserver -f config/iserver.yml
```

## Documents
Detailed documentation please visit: [Developer Documents](https://developers.iost.io/)




## Community & Resources

Make sure to check out these resources as well for more information and to keep up to date with all the latest news about IOST project and team.

[/r/IOSToken on Reddit](https://www.reddit.com/r/IOStoken)

[Telegram](https://t.me/officialios)

[Twitter](https://twitter.com/IOStoken)

[Official website](https://iost.io)

## How to report bugs or raise an issue with the Testnet
  As the IOST blockchain is still in its early stages, our team would love to see developers test our network. We will release specifics for a bug bounty in the near future. In order to provide feedback and report any bugs for the Everest testnet, please visit [explorer.iost.io/#/feedback](https://explorer.iost.io/#/feedback) or join our community channels on  [Twitter](https://twitter.com/IOStoken), [Reddit](https://www.reddit.com/r/IOStoken/), [Telegram](https://t.me/officialios) and [Discord](https://discordapp.com/invite/arbQt6w).

  We encourage you to get involved and play with our testnet. As always, let us know your thoughts and we look forward to continuing to improve the IOST blockchain.

  Happy hacking!
