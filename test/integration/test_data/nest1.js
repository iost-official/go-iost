class Contract {
    init() {
        storage.mapPut("nest1", "f", "1")
    }

    test(from, to) {
        storage.mapPut("nest1", "f", "11111")
        BlockChain.transfer(from, to, "100", "memo")
    }

    sh1(conName) {
        BlockChain.call(conName, "sh2", JSON.stringify([BlockChain.contractName()]));
    }

    sh3(conName) {
        BlockChain.call(conName, "sh4", JSON.stringify([BlockChain.contractName()]));
    }

    sh5(conName) {
        BlockChain.call(conName, "sh6", JSON.stringify([BlockChain.contractName()]));
    }

    sh7(conName) {
        BlockChain.call(conName, "sh8", JSON.stringify([BlockChain.contractName()]));
    }
}

module.exports = Contract;