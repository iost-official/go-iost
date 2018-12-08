'use strict';

class StorageOp {
    constructor() {
    }

    doPut(num) {
        for (let i = 0; i < num; i++) {
            storage.put("key", "value")
        }
    }

    doGet(num) {
        storage.put("key", "value")
        for (let i = 0; i < num; i++) {
            storage.get("key")
        }
    }

    doHas(num) {
        storage.put("key", "value")
        for (let i = 0; i < num; i++) {
            storage.has("key")
        }
    }

    doDel(num) {
        for (let i = 0; i < num; i++) {
            storage.del("key")
        }
    }

    doMapPut(num) {
        for (let i = 0; i < num; i++) {
            storage.mapPut("key", "field", "value")
        }
    }

    doMapGet(num) {
        storage.mapPut("key", "field", "value")
        for (let i = 0; i < num; i++) {
            storage.mapGet("key", "field")
        }
    }

    doMapHas(num) {
        storage.mapPut("key", "field", "value")
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