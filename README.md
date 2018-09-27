<p>
<img src="https://upload-images.jianshu.io/upload_images/2056014-6069383bda7bd291.jpeg?imageMogr2/auto-orient/" >
</p>


# IOSBlockchain - A Secure & Scalable Blockchain for Smart Services

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

## Docker Installation

### Modify Args

In [account/account.go](https://github.com/iost-official/go-iost/blob/master/account/account.go), change boot-up node as below

```
var (
	MainAccount    Account
	//GenesisAccount = map[string]float64{
	//	"2BibFrAhc57FAd3sDJFbPqjwskBJb5zPDtecPWVRJ1jxT": 3400000000,
	//	"tUFikMypfNGxuJcNbfreh8LM893kAQVNTktVQRsFYuEU":  3200000000,
	//	"s1oUQNTcRKL7uqJ1aRqUMzkAkgqJdsBB7uW9xrTd85qB":  3100000000,
	//	"22zr9ows3qndmAjnkiPFex26taATEaEfjGkatVCr5akSU": 3000000000,
	//	"wSKjLjqWbhH2LcJFwTW9Nfq9XPdhb4pw9KCM7QGtemZG":  2900000000,
	//	"oh7VBi17aQvG647cTfhhoRGby3tH55o3Qv7YHWD5q8XU":  2800000000,
	//	"28mKnLHaVvc1YRKc9CWpZxCpo2gLVCY3RL5nC9WbARRym": 2600000000,
	//}
	// //local net
	GenesisAccount = map[string]float64{
		"iWgLQj3VTPN4dZnomuJMMCggv22LFw4nAkA6bmrVsmCo":  13400000000,
		"281pWKbjMYGWKf2QHXUKDy4rVULbF61WGCZoi4PiKhbEk": 13200000000,
		"bj38rN9xdqBa4eiMi1vPjcUwdMyZmQhvYbVA6cnHyQCH":  13100000000,
	}
```
### Build docker image

Generate a docker image based on the current code. Note that the binary file in the build directory will be the centos system binary file.
```
make image
```
### Set environment variable
Set some required environment variables.
#### Linux
```
export DOCKER_IMAGE="iost-node:1.0.3-$(git rev-parse --short HEAD)"
export PROJECT=`pwd`
export LOCAL_IP="hostname -i"
```
#### Mac OS X
```
export DOCKER_IMAGE="iost-node:1.0.3-$(git rev-parse --short HEAD)"
export PROJECT=`pwd`
export LOCAL_IP="$(ipconfig getifaddr en0)"
```

### Run register server
Run the register server, note that mode needs to be private.
```
mkdir -p test/data/register
docker run -d -p 30304:30304 --name iost_register \
       -v $PROJECT/test/data/register:/workdir/data \
       $DOCKER_IMAGE ./register --mode private
```

### Run three iservers
First create three iserver working directories, then generate three iserver configuration files, and finally run the servers.
```
mkdir -p test/data/iserver0
mkdir -p test/data/iserver1
mkdir -p test/data/iserver2
cp test/template/iserver0.yml test/data/iserver0/iserver.yml
cp test/template/iserver1.yml test/data/iserver1/iserver.yml
cp test/template/iserver2.yml test/data/iserver2/iserver.yml
sed -i '.bak' "s/{{LOCAL_IP}}/${LOCAL_IP}/g" test/data/iserver0/iserver.yml
sed -i '.bak' "s/{{LOCAL_IP}}/${LOCAL_IP}/g" test/data/iserver1/iserver.yml
sed -i '.bak' "s/{{LOCAL_IP}}/${LOCAL_IP}/g" test/data/iserver2/iserver.yml

docker run -d -p 30302:30302 -p 30303:30303 -p 8080:8080 --name iost_iserver0 \
       -v $PROJECT/test/data/iserver0:/var/lib/iserver \
       $DOCKER_IMAGE ./start.sh
docker run -d -p 30312:30312 -p 30313:30313 -p 8081:8080 --name iost_iserver1 \
       -v $PROJECT/test/data/iserver1:/var/lib/iserver \
       $DOCKER_IMAGE ./start.sh
docker run -d -p 30322:30322 -p 30323:30323 -p 8082:8080 --name iost_iserver2 \
       -v $PROJECT/test/data/iserver2:/var/lib/iserver \
       $DOCKER_IMAGE ./start.sh
```

### Check the logs
You can view the iserver0 logs in the `test/data/iserver0/logs/current.log` file，for example：
```
tail -f test/data/iserver0/logs/current.log
tail -f test/data/iserver1/logs/current.log
tail -f test/data/iserver2/logs/current.log
```

### Exit all server
Exit all server normally.
```
docker stop iost_iserver0
docker stop iost_iserver1
docker stop iost_iserver2
docker stop iost_register
```

### Remove all server
Clean up all server.
```
docker rm -f iost_iserver0
docker rm -f iost_iserver1
docker rm -f iost_iserver2
docker rm -f iost_register
```

## Smart Contract Handbook

### Overview

For the request of agile development and concepts verifying, IOST Testnet uses Lua as our extensible smart
contract programming language. And a pre-compiler is provided to add or remove features of original Lua script.

Supported features list below:

1. On blockchain storage of key-value pair, vector types like Lua table supported
2. IOST transfer
3. Deposit to or Withdraw from contract account
3. Inter-contract API calls
4. Multi signature

Some of Lua's feature such as require are not supported yet, and will be supported during next updates.

### Ruler of Gas

An EVM-like Gas system will be used to solve halting problem，and pay to nodes.

The PUBLISHER will pay the contract fee, which equals actually gas cost multiplied by annonced gas price.

At least 0.01 IOST will be charged for each transaction to avoid hostile attack.

### Way to Publish Transactions

Use `iwallet -h` for more infomations.

1. Compile source code to .sc file.
`iwallet compile <lua_file>`
2. (optional) Sign .sc file as participant.
`iwallet sign <sc_file>`
3. Publish smart contract.
`iwallet publish <sc_file> [sig_file]`

### API

IOST smart contract is an API oriental smart contract, API declaration looks like below.
```
--- main
-- your own comments
-- @gas_limit 11
-- @gas_price 0.0001
-- @param_cnt 0
-- @return_cnt 1
function main()
    Put("hello", "world")
    return "success"
end--f
--- sayHi
-- @param_cnt 1
-- @return_cnt 1
function sayHi(name)
    return "hi " .. name
end--f
```

Pre-compiler will read comments after ---, and args in comments will be loaded.
Attention：function definition should end with "end--f"

Definition of args:

```
--- <function_name>   name of API
-- @param_cnt         counts of params
-- @return_cnt        counts of return value
-- @gas_limit         Gas limit, only the first declaration of a file will be used.
-- @gas_price         Gas price, act as above
-- @privilege         Privilege of API, default _public_
```

IOST currently support _private_ and _public_ . _public_ means it could be called by everyone, and _private_ means only PUBLISHER of this contract could call it (in inter-contract call).

Further privilege control will be add in future

A main function must be provided in contract. It will be called ONCE when transaction accepted on blockchain. the param count in main function is 0, and return value will be aborted.

Do not use anything outside function, for those codes will be clear in pre-compile phase.

### IOST API
```
Put(key, value) -> bool               -- Write key-value pair to blockchain，only float，string are support
Get(key) -> bool, value               -- Read value of a key, which written by contract itself
Transfer(from, to, account) -> bool                     -- IOST transfer，sender's signature should included
Call(ContractID, apiName, args) -> bool, value...       -- inter-contract call, return API's returns
Deposit(from, value) -> bool          -- deposit IOST to contract account
Withdraw(to, value)  -> bool          -- get IOST from contract account
Random(probability)  -> number        -- give probability and return a blockchain-random true/false result
Now() -> value                        -- return timestamp in seconds
Witness() -> string                   -- return current block's witness, in base 58 encode
Height() -> number                    -- return height of block
ParentHash() -> number                -- return last byte of parent hash
ToJson(table) -> bool, jsonStr        -- convert lua table to a json string
ParseJson(jsonStr) -> bool, table     -- parse lua table from a json string
```
### Playground

Playground is a debug tool of IOST smart contract. Input init values and source codes, and result will be printed.


Usage

```
playground [-v init_values.yml] [source_code_1][source_code_2]...
```

flag -v specified .yaml files which imply init variant, source codes will be compiled and run by order.
you can see more by using `playground -h`

## iwallet

### iwallet instructions

|Command|Contents|Description
|:-:|:-:|:-:|
help|Help about any command|using iwallet -h to get further infomation
account|Account manage|./iwallet account -c id
balance|check balance of specified account|./iwallet balance ~/.ssh/id_secp.pub
block|print block info, default find by block number reversed|./iwallet block 0
check|check .sc file|./iwallet check ./test/a.test.sc
compile|Compile contract files to smart contract|./iwallet compile ./test/a.test.lua
publish|sign to a .sc file with .sig files, and publish it|./iwallet publish ./test/a.test.sc ./test/a.test.sig -k ~/.ssh/id_secp
sign|Sign to .sc file|./iwallet sign ./test/a.test.sc -k ~/.ssh/id_secp
transaction|find transactions|./iwallet transaction -n 2 -p tB4Bc8G7bMEJ3SqFPJtsuXXixbEUDXrYfE5xH4uFmHaV
value|check value of a specified key|./iwallet value "iost"

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
