class Storage1 {
    constructor() {
    }

    put(k, v) {
        return IOSTContractStorage.put(k, v);
    }

    get(k) {
        return IOSTContractStorage.get(k)
    }

    delete(k) {
        return IOSTContractStorage.del(k);
    }

    /*
    mset(k, f, v) {
        return IOSTContract
    }

    mget(k) {
        return IOSTContractStorage.get(k)
    }

    mdelete(k) {
        return IOSTContractStorage.del(k);
    }
    */
};

module.exports = Storage1;