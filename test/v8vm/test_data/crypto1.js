'use strict';
class crypto1 {
    constructor() {
    }

    sha3(msg) {
        return IOSTCrypto.sha3(msg);
    }

}

module.exports = crypto1;