class Storage2 {
    put(k, v) {
        storage.put(k, v);
    }

    mset(k, f, v) {
        storage.mapPut(k, f, v);
    }

    ghas(c, k) {
        return storage.globalHas(c, k);
    }

    gget(c, k) {
        return storage.globalGet(c, k);
    }

    gmhas(c, k, f) {
        return storage.globalMapHas(c, k ,f);
    }

    gmget(c, k, f) {
        return storage.globalMapGet(c, k ,f);
    }

    gmkeys(c, k) {
        return storage.globalMapKeys(c, k);
    }

    gmlen(c, k) {
        return storage.globalMapLen(c, k);
    }
}

module.exports = Storage2;