let IOSTContractStorage = (function () {

    let storage = new IOSTStorage;

    let simpleStorage = function () {
        this.put = function (k, v) {
            if (typeof v !== 'string') {
                throw new Error("storage put must be string");
            }
            return storage.put(k, v, "");
        };
        this.get = function (k) {
            return storage.get(k, "");
        };
        this.has = function (k) {
            return storage.has(k, "");
        };
        this.del = function (k) {
            return storage.del(k, "");
        }
    };
    let simpleStorageObj = new simpleStorage;

    let mapStorage = function () {
        this.mapPut = function (k, f, v) {
            if (typeof v !== 'string') {
                throw new Error("storage mapPut must be string");
            }
            return storage.mapPut(k, f, v, "");
        };
        this.mapHas = function (k, f) {
            return storage.mapHas(k, f, "");
        };
        this.mapGet = function (k, f) {
            return storage.mapGet(k, f, "");
        };
        this.mapKeys = function (k) {
            return JSON.parse(storage.mapKeys(k, ""));
        };
        this.mapLen = function (k) {
            return storage.mapLen(k, "");
        };
        this.mapDel = function (k, f) {
            return storage.mapDel(k, f, "");
        }
    };
    let mapStorageObj = new mapStorage;

    let globalStorage = function () {
        this.get = function (key) {
            return storage.globalGet(c, k, "");
        }
        this.has = function (key) {
            return storage.globalHas(c, k, "");
        }
        this.mapHas = function (k, f) {
            return storage.globalMapHas(k, f, "");
        };
        this.mapGet = function (k, f) {
            return storage.globalMapGet(k, f, "");
        };
        this.mapKeys = function (k) {
            return JSON.parse(storage.globalMapKeys(k, ""));
        };
        this.mapLen = function (k) {
            return storage.globalMapLen(k, "");
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