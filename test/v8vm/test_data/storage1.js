'use strict';
class Storage1 {
    constructor() {
        storage.put("num", JSON.stringify(99));
        storage.put("str", JSON.stringify("yeah"));
    }

    put(k, v) {
        return storage.put(k, v);
    }

    get(k) {
        return storage.get(k);
    }

    delete(k) {
        return storage.del(k);
    }

    getThisNum() {
        return JSON.parse(storage.get("num"));
    }

    getThisStr() {
        return JSON.parse(storage.get("str"));
    }

    mset(k, f, v) {
        return storage.mapPut(k, f, JSON.stringify(v));
    }

    mget(k, f) {
        return JSON.parse(storage.mapGet(k, f));
    }

    mhas(k, f) {
        return storage.mapHas(k, f);
    }

    mdelete(k, f) {
        return storage.mapDel(k, f);
    }
}

module.exports = Storage1;