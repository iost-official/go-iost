'use strict';
class Datatype {
    constructor() {
    }

    number(delta) {
        let a = 0;
        for (let i = 0; i < 10; i++) {
            a += delta;
            delta /= 10;
        }
        a -= a / 2;
        a /= 2;
        a *= 2;
        return a;
    }

    number_big(delta) {
        let a = new BigNumber(0);
        for (let i = 0; i < 10; i++) {
            a = a.plus(delta);
            delta /= 10;
        }
        a -= a / 2;
        a /= 2;
        a *= 2;
        return a;
    }

    number_op() {
        let a = new BigNumber(1);
        a <<= 2;
        a >>= 1;
        a **= 3;
        a %= 5;
        return a;
    }

    number_op2() {
        let a = new BigNumber(8);
        a |= 1;
        a &= 7;
        a ^= 3;
        return a == 2 ? 2 : 22;
    }

    number_strange() {
        let a = new BigNumber(8);
        return a / -0.0;
    }

    param() {
        return 3, 4;
    }

    param2(a) {
        return a;
    }

    param3(a) {
        a.push({a:3});
        return a;
    }

    bool() {
        let a = true;
        let b = true;
        return !a || !b && a;
    }

    string() {
        let a = "test";
        for (let i = 0; i < 10; i += 1) {
            a = a + a;
        }
        return a;
    }

    array() {
        let a = [0, 1, 2];
        a.push(3);
        return a;
    }

    object() {
        let a = {
            pos: 0,
            str: "test object"
        }
        return a;
    }
};

module.exports = Datatype;