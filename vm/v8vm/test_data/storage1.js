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

    /*
    mset(k, f, v) {
        return IOSTContract
    }

    mget(k) {
        return storage.get(k)
    }

    mdelete(k) {
        return storage.del(k);
    }
    */
};

module.exports = Storage1;