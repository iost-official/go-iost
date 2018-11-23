class transfer {
    init(){
    }
    transfer(from, to, amount) {
        // let ret = BlockChain.call("token.iost", "transfer", '["iost", "' + from + '","' + to + '","' + amount + '", ""]');
        BlockChain.transfer(from, to, amount, "");
    }
    withdraw(to, amount) {
        //let ret = BlockChain.callWithAuth("token.iost", "transfer", '["iost", "' + BlockChain.contractName() + '","' + to + '","' + amount + '", ""]');
        BlockChain.withdraw(to, amount, "");
    }
}

module.exports = transfer;
