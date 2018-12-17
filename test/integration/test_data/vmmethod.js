class Contract {
    init() {
    }

    contractName() {
        return blockchain.contractName();
    }

    receiptf(data) {
        blockchain.receipt(data);
    }

    event(data) {
        blockchain.event(data);
    }

    putwithpayer(k, v, p) {
        storage.put(k, v, p)
    }

    get(k) {
        return storage.get(k)
    }

    mapputwithpayer(k, f, v, p) {
        storage.mapPut(k, f, v, p)
    }

    mapget(k, f) {
        return storage.mapGet(k, f)
    }
}

module.exports = Contract;
