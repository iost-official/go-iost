class Int64Test {
    constructor() {

    }

    getPlus() {
        this.number = new Int64("1234500000");
        this.number = this.number.plus(1234);
        return this.number.toString();
    }

    getMinus() {
        this.number = new Int64("123400789");
        this.number = this.number.minus(680);
        return this.number.toString();
    }

    getMulti() {
        this.number = new Int64("12345667");
        this.number = this.number.multi(12);
        return this.number.toString();
    }

    getDiv() {
        this.number = new Int64("12345667");
        this.number = this.number.div(12);
        return this.number.toString();
    }

    getMod() {
        this.number = new Int64("12345667");
        this.number = this.number.mod(12);
        return this.number.toString();
    }

    getPow(times) {
        this.number = new Int64("1234");
        this.number = this.number.pow(times);
        return this.number.toString();
    }

    getSqrt() {
        this.number = new Int64("1234789");
        this.number = this.number.sqrt();
        return this.number.toString();
    }
}

module.exports = Int64Test;