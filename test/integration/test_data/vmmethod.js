class Contract {
    init() {
    }

    contractName() {
        return BlockChain.contractName();
    }

    receiptf(data) {
        BlockChain.receipt(data);
    }

    event(data) {
        BlockChain.event(data);
    }
}

module.exports = Contract;