class transfer {
    init(){
    }
    transfer(from, to, amount) {
        let ret = BlockChain.call("iost.token", "transfer", '["iost", "' + from + '","' + to + '","' + amount + '", ""]');
    }
    withdraw(to, amount) {
        let ret = BlockChain.callWithAuth("iost.token", "transfer", '["iost", "Contracttransfer","' + to + '","' + amount + '", ""]');
    }
}

module.exports = transfer;