class transfer {
    init(){
    }
    transfer(from, to, amount) {
        let ret = blockchain.call("Contracttransfer", "transfer", '["' + from + '","' + to + '","' + amount + '"]');
    }
}

module.exports = transfer;
