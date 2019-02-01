'use strict';
class Int64Test {

    getPlus() {
        const number = new Int64("1234500000");
        const number2 = number.plus(1234);
        return number2.toString();
    }

    getMinus() {
        const number = new Int64("123400789");
        const number2 = number.minus(680);
        return number2.toString();
    }

    getMulti() {
        const number = new Int64("12345667");
        const number2 = number.multi(12);
        return number2.toString();
    }

    getDiv() {
        const number = new Int64("12345667");
        const number2 = number.div(12);
        return number2.toString();
    }

    getMod() {
        const number = new Int64("12345667");
        const number2 = number.mod(12);
        return number2.toString();
    }

    getPow(times) {
        const number = new Int64("1234");
        const number2 = number.pow(times);
        return number2.toString();
    }
}

module.exports = Int64Test;