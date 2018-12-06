class transfer {
    init(){
    }
    transfer(from, to, amount) {
        const a = new Int64(amount);
        blockchain.transfer(from, to, a)
    }
}

module.exports = transfer;
