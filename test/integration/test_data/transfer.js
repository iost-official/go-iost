class transfer {
    init(){
    }
    transfer(from, to, amount) {
        // let ret = blockchain.call("token.iost", "transfer", '["iost", "' + from + '","' + to + '","' + amount + '", ""]');
        blockchain.transfer(from, to, amount, "");
    }
    withdraw(to, amount) {
        //let ret = blockchain.callWithAuth("token.iost", "transfer", '["iost", "' + blockchain.contractName() + '","' + to + '","' + amount + '", ""]');
        blockchain.withdraw(to, amount, "");
    }
}

module.exports = transfer;
