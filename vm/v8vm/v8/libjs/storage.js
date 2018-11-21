let IOSTContractStorage = (function () {

    let storage = new IOSTStorage;

    let simpleStorage = function () {
        this.put = function (k, v, p) {
            if (typeof v !== 'string') {
                throw new Error("storage put must be string");
            }
            if (p === undefined) {
                p = "";
            }
            return storage.put(k, v, p);
        };
        // payer not used
        this.get = function (k) {
            let p = "";
            return storage.get(k, p);
        };
        this.has = function (k) {
            let p = "";
            return storage.has(k, p);
        };
        this.del = function (k) {
            let p = "";
            return storage.del(k, p);
        }
    };
    let simpleStorageObj = new simpleStorage;

    let mapStorage = function () {
        this.mapPut = function (k, f, v, p) {
            if (typeof v !== 'string') {
                throw new Error("storage mapPut must be string");
            }
            if (p === undefined) {
                p = "";
            }
            return storage.mapPut(k, f, v, p);
        };
        // payer not used
        this.mapHas = function (k, f) {
            let p = "";
            return storage.mapHas(k, f, p);
        };
        this.mapGet = function (k, f) {
            let p = "";
            return storage.mapGet(k, f, p);
        };
        this.mapKeys = function (k) {
            let p = "";
            return JSON.parse(storage.mapKeys(k, p));
        };
        this.mapLen = function (k) {
            let p = "";
            return storage.mapLen(k, p);
        };
        this.mapDel = function (k, f) {
            let p = "";
            return storage.mapDel(k, f, p);
        }
    };
    let mapStorageObj = new mapStorage;

    let globalStorage = function () {
        // payer not used
        this.get = function (c, k) {
            let p = "";
            return storage.globalGet(c, k, p);
        }
        this.has = function (c, k) {
            let p = "";
            return storage.globalHas(c, k, p);
        }
        this.mapHas = function (c, k, f) {
            let p = "";
            return storage.globalMapHas(c, k, f, p);
        };
        this.mapGet = function (c, k, f) {
            let p = "";
            return storage.globalMapGet(c, k, f, p);
        };
        this.mapKeys = function (c, k) {
            let p = "";
            return JSON.parse(storage.globalMapKeys(c, k, p));
        };
        this.mapLen = function (c, k) {
            let p = "";
            return storage.globalMapLen(c, k, p);
        };
    };
    let globalStorageObj = new globalStorage;

    return {
        // simply put a k-v pair, value must be string!
        // put(key, value)
        put: simpleStorageObj.put,
        // simply get a value using key.
        // get(key)
        get: simpleStorageObj.get,
        has: simpleStorageObj.has,
        // simply del a k-v pair using key.
        // del(key)
        del: simpleStorageObj.del,
        // map put a (k, f, value) pair. use k + f to find value.
        // mapPut(key, field, value)
        mapPut: mapStorageObj.mapPut,
        // map check a (k, f) pair existence. use k + f to check.
        // mapHas(key, field)
        mapHas: mapStorageObj.mapHas,
        // map Get a (k, f) pair. use k + f to find value.
        // mapGet(key, field)
        mapGet: mapStorageObj.mapGet,
        // map Get fields inside a key.
        // mapKeys(key)
        mapKeys: mapStorageObj.mapKeys,
        mapLen: mapStorageObj.mapLen,
        // map Delete a (k, f) pair. use k + f to delete value.
        // mapDel(key, field)
        mapDel: mapStorageObj.mapDel,
        // currently not suportted, dont't use.
        globalGet: globalStorageObj.get,
        globalHas: globalStorageObj.has,
        globalMapHas: globalStorageObj.mapHas,
        globalMapGet: globalStorageObj.mapGet,
        globalMapKeys: globalStorageObj.mapKeys,
        globalMapLen: globalStorageObj.mapLen,
    }
})();

module.exports = IOSTContractStorage;