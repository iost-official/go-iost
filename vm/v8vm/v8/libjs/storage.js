var IOSTContractStorage = (function () {

    var storage = new IOSTStorage;

    var normalValTypeList = ['number', 'string', 'boolean'];

    var MapStorage = function (key) {
        this.mapKey = key;
        this.put = function (field, value) {
            return storage.mapPut(this.mapKey, value);
        };
        this.get = function (field) {
            return storage.mapGet(this.mapKey, field);
        };
        this.del = function (field) {
            return storage.mapDel(this.mapKey, field);
        };
        this.keys = function () {
            return storage.mapKeys(this.mapKey);
        };
        this.length = function () {
            return storage.mapLen(this.mapKey);
        }
    };

    var GlobalStorage = function (contract) {
        this.contract = contract;
        this.get = function (key) {
            return storage.globalGet(contract, key);
        }
    };

    return {
        put: function (key, val) {
            let valType = typeof val;
            if (valType === 'string') {
                return storage.put(key, val);
            }
            if (val instanceof Int64) {
                return storage.put(key, val.toString())
            }
            if (val instanceof BigNumber) {
                return storage.put(key, val.toString(10))
            }
            // _native_log("storage put: " + key + " : " + JSON.stringify(val))
            return storage.put(key, JSON.stringify(val))
        },
        get: function (key, val) {
            let valType = typeof val;
            if (valType === 'string') {
                return storage.get(key);
            }
            if (normalValTypeList.indexOf(valType) !== -1) {
                let valInStorage = storage.get(key);
                return JSON.parse(valInStorage);
            }

            if (val instanceof Int64) {
                let valInStorage = storage.get(key);
                return new Int64(valInStorage);
            }

            if (val instanceof BigNumber) {
                let valInStorage = storage.get(key);
                return new BigNumber(valInStorage);
            }

            let valInStorage = storage.get(key);
            return JSON.parse(valInStorage);
            // return new MapStorage(key);
        },
        GlobalStorage: GlobalStorage
    }
})();

module.exports = IOSTContractStorage;

