'use strict';
class Storage1 {
    constructor() {
        storage.put("num", "99");
        storage.put("str", "year");
    }

    put(k, v) {
        return storage.put(k, v);
    }

    get(k, v) {
        return storage.get(k, v)
    }

    delete(k) {
        return storage.del(k);
    }

    getThisNum() {
        return storage.get("num")
    }

    getThisStr() {
        return storage.get("str")
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
}

module.exports = Storage1;