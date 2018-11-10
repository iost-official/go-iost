let IOSTContractStorage = (function () {

    let storage = new IOSTStorage;

    let simpleStorage = function () {
        this.put = function (k, v, o) {
            if (typeof v !== 'string') {
                throw new Error("storage put must be string");
            }
            if (o === undefined) {
                o = "";
            }
            return storage.put(k, v, o);
        };
        this.get = function (k, o) {
            if (o === undefined) {
                o = "";
            }
            return storage.get(k, o);
        };
        this.has = function (k, o) {
            if (o === undefined) {
                o = "";
            }
            return storage.has(k, o);
        };
        this.del = function (k, o) {
            if (o === undefined) {
                o = "";
            }
            return storage.del(k, o);
        }
    };
    let simpleStorageObj = new simpleStorage;

    let mapStorage = function () {
        this.mapPut = function (k, f, v, o) {
            if (typeof v !== 'string') {
                throw new Error("storage mapPut must be string");
            }
            if (o === undefined) {
                o = "";
            }
            return storage.mapPut(k, f, v, o);
        };
        this.mapHas = function (k, f, o) {
            if (o === undefined) {
                o = "";
            }
            return storage.mapHas(k, f, o);
        };
        this.mapGet = function (k, f, o) {
            if (o === undefined) {
                o = "";
            }
            return storage.mapGet(k, f, o);
        };
        this.mapKeys = function (k, o) {
            if (o === undefined) {
                o = "";
            }
            return JSON.parse(storage.mapKeys(k, o));
        };
        this.mapLen = function (k, o) {
            if (o === undefined) {
                o = "";
            }
            return storage.mapLen(k, o);
        };
        this.mapDel = function (k, f, o) {
            if (o === undefined) {
                o = "";
            }
            return storage.mapDel(k, f, o);
        }
    };
    let mapStorageObj = new mapStorage;

    let globalStorage = function () {
        this.get = function (c, k, o) {
            if (o === undefined) {
                o = "";
            }
            return storage.globalGet(c, k, o);
        }
        this.has = function (c, k, o) {
            if (o === undefined) {
                o = "";
            }
            return storage.globalHas(c, k, o);
        }
        this.mapHas = function (c, k, f, o) {
            if (o === undefined) {
                o = "";
            }
            return storage.globalMapHas(c, k, f, o);
        };
        this.mapGet = function (c, k, f, o) {
            if (o === undefined) {
                o = "";
            }
            return storage.globalMapGet(c, k, f, o);
        };
        this.mapKeys = function (c, k, o) {
            if (o === undefined) {
                o = "";
            }
            return JSON.parse(storage.globalMapKeys(c, k, o));
        };
        this.mapLen = function (c, k, o) {
            if (o === undefined) {
                o = "";
            }
            return storage.globalMapLen(c, k, o);
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