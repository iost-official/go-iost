class Int64Test {
    constructor() {

    }

    getPlus() {
        this.number = new Int64("1234500000");
        this.number = this.number.plus(1234);
        return this.number.toString();
    }
}

module.exports = Int64Test;