class BigNumber1{
    constructor() {
        this.val = new BigNumber(0.00000000008);
        this.val = this.val.plus(0.0000000000000029);
    }
    getVal() {
        return this.val.toString(10);
    }
}

module.exports = BigNumber1;