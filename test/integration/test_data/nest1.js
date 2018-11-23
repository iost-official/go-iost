class Contract {
    init() {
        storage.mapPut("nest1", "f", "1")
    }

    test(from, to) {
        storage.mapPut("nest1", "f", "11111")
        BlockChain.transfer(from, to, "100", "memo")
    }
}

module.exports = Contract;