class Contract {
    init() {
        storage.put("nest0", "0")
    }

    call(conName, abiName, args) {
        storage.put("nest0", "000")
        BlockChain.call(conName, abiName, args)
    }
}

module.exports = Contract;