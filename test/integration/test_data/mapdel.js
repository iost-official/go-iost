class Test {
    init() {}

    keys() {
        storage.mapPut("k", "s", "0");
        storage.mapPut("k", "abcd", "1");
        storage.mapPut("k", "abc", "2");
        storage.mapPut("k", "ab", "3");
        storage.mapDel("k", "ab");
        storage.mapPut("k", "ab", "4");
        return storage.mapKeys("k");
    }

    del() {
        storage.mapDel("k", "ab");
        return storage.mapKeys("k");
    }
}

module.exports = Test;