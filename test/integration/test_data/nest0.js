class Contract {
    init() {
        storage.put("nest0", "0")
    }

    call(conName, abiName, args) {
        storage.put("nest0", "000")
        blockchain.call(conName, abiName, args)
    }

    sh0(conName) {
        blockchain.call(conName, "sh1", JSON.stringify([blockchain.contractName()]));
    }

    sh2(conName) {
        blockchain.call(conName, "sh3", JSON.stringify([blockchain.contractName()]));
    }

    sh4(conName) {
        blockchain.call(conName, "sh5", JSON.stringify([blockchain.contractName()]));
    }

    sh6(conName) {
        blockchain.call(conName, "sh7", JSON.stringify([blockchain.contractName()]));
    }
}

module.exports = Contract;
