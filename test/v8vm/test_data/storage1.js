'use strict';
class Storage1 {
    constructor() {
        storage.put("num", JSON.stringify(99));
        storage.put("str", JSON.stringify("yeah"));
    }

    put(k, v) {
        storage.put(k, v);
    }

    has(k) {
        return storage.has(k);
    }

    get(k) {
        let ret = storage.get(k);
        return ret;
    }

    delete(k) {
        storage.del(k);
    }

    getThisNum() {
        return JSON.parse(storage.get("num"));
    }

    getThisStr() {
        return JSON.parse(storage.get("str"));
    }

    mset(k, f, v) {
        storage.mapPut(k, f, JSON.stringify(v));
    }

    mget(k, f) {
        return JSON.parse(storage.mapGet(k, f));
    }

    mhas(k, f) {
        return storage.mapHas(k, f);
    }

    mkeys(k) {
        return storage.mapKeys(k);
    }

    mlen(k) {
        return storage.mapLen(k);
    }

    mdelete(k, f) {
        storage.mapDel(k, f);
    }
}

module.exports = Storage1;