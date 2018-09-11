let IOSTContractStorage = (function () {

    let storage = new IOSTStorage;

    let simpleStorage = function () {
        this.put = function (k, v) {
            if (typeof v !== 'string') {
                throw new Error("storage put must be string");
            }
            return storage.put(k, v);
        };
        this.get = function (k) {
            return storage.get(k);
        };
        this.del = function (k) {
            return storage.del(k);
        }
    };
    let simpleStorageObj = new simpleStorage;

    let mapStorage = function () {
        this.mapPut = function (k, f, v) {
            if (typeof v !== 'string') {
                throw new Error("storage mapPut must be string");
            }
            return storage.mapPut(k, f, v);
        };
        this.mapHas = function (k, f) {
            return storage.mapHas(k, f);
        };
        this.mapGet = function (k, f) {
            return storage.mapGet(k, f);
        };
        this.mapKeys = function (k) {
            return [];
        };
        this.mapDel = function (k, f) {
            return storage.mapDel(k, f);
        }
    };
    let mapStorageObj = new mapStorage;

    let globalStorage = function () {
        this.get = function (key) {
            return storage.globalGet(c, k);
        }
    };
    let globalStorageObj = new globalStorage;

    return {
        put: simpleStorageObj.put,
        get: simpleStorageObj.get,
        del: simpleStorageObj.del,
        mapPut: mapStorageObj.mapPut,
        mapHas: mapStorageObj.mapHas,
        mapGet: mapStorageObj.mapGet,
        mapKeys: mapStorageObj.mapKeys,
        mapDel: mapStorageObj.mapDel,
        globalGet: globalStorageObj.get,
    }
})();

module.exports = IOSTContractStorage;

