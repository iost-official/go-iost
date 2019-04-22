class transfer {
    init(){
    }
    transfer(from, to, amount) {
        // let ret = blockchain.call("token.iost", "transfer", '["iost", "' + from + '","' + to + '","' + amount + '", ""]');
        blockchain.transfer(from, to, amount, "");
    }
    transferFreeze(from, to, amount) {
        blockchain.call("token.iost", "transferFreeze", '["iost", "' + from + '","' + to + '","' + amount + '", 1545385268000000000, ""]');
    }
    destroy(from, amount) {
        blockchain.call("token.iost", "destroy", '["iost", "' + from + '","' + amount + '"]');
    }
    transfer1(from, to, amount) {
        // let ret = blockchain.call("token.iost", "transfer", '["iost", "' + from + '","' + to + '","' + amount + '", ""]');
        blockchain.transfer(from, to, amount, "");
    }
    transfer2(from, to, amount) {
        // let ret = blockchain.call("token.iost", "transfer", '["iost", "' + from + '","' + to + '","' + amount + '", ""]');
        blockchain.transfer(from, to, amount, "");
    }
    transfer3(from, to, amount) {
        // let ret = blockchain.call("token.iost", "transfer", '["iost", "' + from + '","' + to + '","' + amount + '", ""]');
        blockchain.transfer(from, to, amount, "");
    }
    transfer4(from, to, amount) {
        // let ret = blockchain.call("token.iost", "transfer", '["iost", "' + from + '","' + to + '","' + amount + '", ""]');
        blockchain.transfer(from, to, amount, "");
    }
    transfermulti(from, from1, to, amount) {
        // let ret = blockchain.call("token.iost", "transfer", '["iost", "' + from + '","' + to + '","' + amount + '", ""]');
        blockchain.transfer(from, to, amount, "");
        blockchain.transfer(from1, to, amount, "");
    }
    withdraw(to, amount) {
        //let ret = blockchain.callWithAuth("token.iost", "transfer", '["iost", "' + blockchain.contractName() + '","' + to + '","' + amount + '", ""]');
        blockchain.withdraw(to, amount, "");
    }
    withdrawWithoutAuth(to, amount) {
        blockchain.call("token.iost", "transfer", '["iost", "' + blockchain.contractName() + '","' + to + '","' + amount + '", ""]');
    }
    withdrawWithoutAuthAfterWithAuth(to, amount) {
        blockchain.withdraw(to, amount, "");
        blockchain.call("token.iost", "transfer", '["iost", "' + blockchain.contractName() + '","' + to + '","' + amount + '", ""]');
    }
}

module.exports = transfer;
