'use strict';

class InputLen {
    mapPutLong() {
        storage.mapPut('k', 'f', 'x'.repeat(70000));
    }

    callLongArg() {
        blockchain.call('contract.iost', 'api', 'x'.repeat(70000));
    }

    putCJK() {
        storage.put('cjk', '世'.repeat(30000));
        return storage.get('cjk').length;
    }
};

module.exports = InputLen;
