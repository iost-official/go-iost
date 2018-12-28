'use strict';

class StorageOp {
    constructor() {
    }

    doPut(num, data) {
        if (data === undefined || data === null) {
            data = "value";
        }
        for (let i = 0; i < num; i++) {
            storage.put("key" + i, data)
        }
    }

    doGet(num, data) {
        if (data === undefined || data === null) {
            data = "value";
        }
        storage.put("key", data)
        for (let i = 0; i < num; i++) {
            storage.get("key")
        }
    }

    doHas(num, data) {
        if (data === undefined || data === null) {
            data = "value";
        }
        storage.put("key", data)
        for (let i = 0; i < num; i++) {
            storage.has("key")
        }
    }

    doDel(num) {
        for (let i = 0; i < num; i++) {
            storage.del("key")
        }
    }

    doMapPut(num, data) {
        if (data === undefined || data === null) {
            data = "value";
        }
        for (let i = 0; i < num; i++) {
            storage.mapPut("key", "field", data)
        }
    }

    doMapGet(num, data) {
        if (data === undefined || data === null) {
            data = "value";
        }
        storage.mapPut("key", "field", data)
        for (let i = 0; i < num; i++) {
            storage.mapGet("key", "field")
        }
    }

    doMapHas(num, data) {
        if (data === undefined || data === null) {
            data = "value";
        }
        storage.mapPut("key", "field", data)
        for (let i = 0; i < num; i++) {
            storage.mapHas("key", "field")
        }
    }

    doMapDel(num) {
        for (let i = 0; i < num; i++) {
            storage.mapDel("key", "field")
        }
    }

    doMapKeys(num) {
        storage.mapPut("key", "field", "value")
        for (let i = 0; i < num; i++) {
            storage.mapKeys("key")
        }
    }

    doMapLen(num) {
        storage.mapPut("key", "field", "value")
        for (let i = 0; i < num; i++) {
            storage.mapLen("key")
        }
    }
};

module.exports = StorageOp;