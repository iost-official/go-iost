class Storage2 {
    get(k) {
        return storage.get(k);
    }

    mget(k, f) {
        return storage.mapGet(k ,f);
    }

    gget(c, k) {
        return storage.globalGet(c, k);
    }

    gmget(c, k, f) {
        return storage.globalMapGet(c, k ,f);
    }
}

module.exports = Storage2;