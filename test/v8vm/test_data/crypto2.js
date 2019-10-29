'use strict';
class crypto1 {
    sha3(msg) {
        return IOSTCrypto.sha3(msg);
    }
    sha3Hex(msg) {
        return IOSTCrypto.sha3Hex(msg);
    }
    ripemd160Hex(msg) {
        return IOSTCrypto.ripemd160Hex(msg);
    }
    verify(algo, msg, sig, pubkey) {
        return IOSTCrypto.verify(algo, msg, sig, pubkey);
    }

}

module.exports = crypto1;