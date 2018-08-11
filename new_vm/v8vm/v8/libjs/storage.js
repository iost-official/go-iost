var IOSTContractStorage = (function () {
    var storage = new IOSTStorage;
    return {
        put: function (key, val) {
            var ret = storage.put(key, val);
            return ret;
        },
        get: function (key) {
            var val = storage.get(key);
            return val;
        },
        del: function () {
            var ret = storage.del(key);
            return ret;
        },
    }
})();

module.exports = IOSTContractStorage

