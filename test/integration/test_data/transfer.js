class transfer {
    init(){
    }
    transfer(from, to, amount) {
        let ret = BlockChain.call("iost.token", "transfer", '["iost", "' + from + '","' + to + '","' + amount + '"]');
    }
}

module.exports = transfer;