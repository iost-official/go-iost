class transfer {
    init(){
    }
    transfer(from, to, amount) {
        let ret = BlockChain.call("Contracttransfer", "transfer", '["' + from + '","' + to + '","' + amount + '"]');
    }
}

module.exports = transfer;