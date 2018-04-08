# IOSBlockchain - A Secure & Scalable Blockchain for Smart Services

The Internet of Services (IOS) offers a secure & scalable infrastructure for online service providers. Its high TPS, scalable and secure blockchain, and privacy protection scales social and economic cooperation to a new level. For more information about IOS Blockchain, please refer to our [whitepaper](https://github.com/iost-official/Documents/blob/master/Technical_White_Paper/Technical_White_Paper_English.md)

The source codes released in this repo represent part of our early alpha-quality codes, with the purpose to demonstrate our development progress. Currently, there are still parts of our project have been intensively working on in private repos. We will gradually release our code.

## Overview

This is the first release version (v0.1.0) of IOST, including the prototype of protocol, blockchain, database and testing libs.

>“I could be bounded in a nutshell, and count myself a king of infinite space!”  - Hamlet II.ii

## Components and features

Considering the seriously waste of power in POW and complex, unrobust implement of POS system, we using an auto devoting POB system to ensure consensus of IOST blockchain.

The rotating delegates are promoted by the contribution of IOST, such as recording and spreading transactions, running smart contract, and verifying transactions, which we called as Believability. Users send their transactions to recorders they trusted, recorders deal with transactions and signed their name into transactions, finally the best of recorder promoted to have a term of replica, to make new block and get reward.

For users, a good recorder can give response fastly and exactly, and user is willing to upload their transactions to such a good recorder, so the way of POB using to devote a delegate is much more natural than traditional POS. Further, cause one recorder couldn’t be the best choose of all users, the POB system has it’s power to resist centralize.


    ├── cmd
    ├── drp
    ├── eds
    ├── iosbase: base structure of IOST, including structs of UTXO, transaction, blocks
    │   └── debug: a future debug lib, debug mode and debug related functions will be included
    ├── mock-libs: auxiliary mock-libs of testing blockchain
    │   ├── asset: user account lib, including user asset related operations
    │   ├── block: block lib, including signature validation & blockchain header builder
    │   ├── market: market lib, including two different market policy
    │   └── transaction: transaction lib, including different transaction related operations
    └── protocol: POB based protocol of IOST consensus
        ├── recorder: recorder of Transactions. Collect and verify Txs from common users
        ├── replica: rotating replicas promoted from best behaved recorder, making new Blocks,  reaching consensus with each other and broadcast to every members of IOST network
        ├── view: PBFT view of current replicas, calculated by information in previews Blocks, determining the character of replicas
        ├── router: the wrap of network, in
        └── database: runtime data used in protocol

## Development Progress

v0.1.0 - MVP completed and preliminary tests conducted

v0.2.0 - Under development.

Our developers work in their own trees, then submit pull requests when they think their feature or bug fix is ready.

Although we have started to go open source on Github starting April 9th, 2018 and released certain parts of our code/repositories, please note that we have only released a portion of them at this moment.  Since the project is still at its very early stage and some codes are related to our core technology, we have decided to go open source gradually until the launch of the main net.

## Testnet

The IOS Blockchain is not yet to be used in production. We have deployed a private testnet for early stage testing as planned. We will open the testnet to public after the official release of Janus (v0.5) by the end of Q2 2018. This document will be updated then with instructions for running on the public testnet.

## Contribution

Contribution of any forms is appreciated!

Currently, our core tech team is working intensively to develop the first stable version and core blockchain structure. We will accept pull request after the first stable release published.

If you have any questions, please email to team@iost.io

## Community & Resources

Make sure to check out these resources as well for more information and to keep up to date with all the latest news about IOST project and team.

[/r/IOSToken on Reddit](https://www.reddit.com/r/IOStoken)

[Telegram](https://t.me/officialios)

[Twitter](https://twitter.com/IOStoken)

[Official website](https://iost.io)

## Disclaimer

- IOS Blockchain is unfinished and some parts are highly experimental. Use the code at your own risk.

- THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.


