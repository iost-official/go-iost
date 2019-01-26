'use strict';
class float64Test {
    number() {
        var a = new Float64("11.11")
        return a.toString();
    }

    getMinus() {
        const number = new Float64("12345.6789");
        const number2 = number.minus(680);
        return number2.toString();
    }

    getMulti() {
        const number = new Float64("12345.6789");
        const number2 = number.multi(12);
        return number2.toString();
    }

    getDiv() {
        const number = new Float64("12345.6789");
        const number2 = number.div(12);
        return number2.toString();
    }

    getPow(times) {
        const number = new Float64("12345.6789");
        const number2 = number.pow(times);
        return number2.toString();
    }


}

module.exports = float64Test;
