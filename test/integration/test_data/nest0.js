class Contract {
    init() {
        storage.put("nest0", "0")
    }

    call(conName, abiName, args) {
        storage.put("nest0", "000")
        BlockChain.call(conName, abiName, args)
    }

    sh0(conName) {
        BlockChain.call(conName, "sh1", JSON.stringify([BlockChain.contractName()]));
    }

    sh2(conName) {
        BlockChain.call(conName, "sh3", JSON.stringify([BlockChain.contractName()]));
    }

    sh4(conName) {
        BlockChain.call(conName, "sh5", JSON.stringify([BlockChain.contractName()]));
    }

    sh6(conName) {
        BlockChain.call(conName, "sh7", JSON.stringify([BlockChain.contractName()]));
    }
}

module.exports = Contract;