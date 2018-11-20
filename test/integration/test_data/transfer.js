class transfer {
    init(){
    }
    transfer(from, to, amount) {
        let ret = BlockChain.call("token.iost", "transfer", '["iost", "' + from + '","' + to + '","' + amount + '", ""]');
    }
    withdraw(to, amount) {
        let ret = BlockChain.callWithAuth("token.iost", "transfer", '["iost", "Contracttransfer","' + to + '","' + amount + '", ""]');
    }
}

module.exports = transfer;
