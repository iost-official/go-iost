class Storage1 {
    constructor() {
        this.num = "99"
        this.str = "yeah"
    }

    put(k, v) {
        return IOSTContractStorage.put(k, v);
    }

    get(k, v) {
        return IOSTContractStorage.get(k, v)
    }

    delete(k) {
        return IOSTContractStorage.del(k);
    }

    getThisNum() {
        return this.num
    }

    getThisStr() {
        return this.str
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