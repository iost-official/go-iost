class Contract {
    init() {
        storage.mapPut("nest1", "f", "1")
    }

    test(from, to) {
        storage.mapPut("nest1", "f", "11111")
        blockchain.transfer(from, to, "100", "memo")
    }

    sh1(conName) {
        blockchain.call(conName, "sh2", JSON.stringify([blockchain.contractName()]));
    }

    sh3(conName) {
        blockchain.call(conName, "sh4", JSON.stringify([blockchain.contractName()]));
    }

    sh5(conName) {
        blockchain.call(conName, "sh6", JSON.stringify([blockchain.contractName()]));
    }

    sh7(conName) {
        blockchain.call(conName, "sh8", JSON.stringify([blockchain.contractName()]));
    }
}

module.exports = Contract;
