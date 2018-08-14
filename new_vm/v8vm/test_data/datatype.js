class Datatype {
    constructor() {
    }

    number(delta) {
        var a = 0;
        for (var i = 0; i < 10; i++) {
            a += delta;
            delta /= 10;
        }
        a -= a / 2;
        a /= 2;
        a *= 2;
        return a;
    }

    number_big(delta) {
        var a = new bigNumber(0);
        for (var i = 0; i < 10; i++) {
            a = a.plus(delta);
            delta /= 10;
        }
        a -= a / 2;
        a /= 2;
        a *= 2;
        return a;
    }

    number_op() {
        var a = new bigNumber(1);
        a <<= 2;
        a >>= 1;
        a **= 3;
        a %= 5;
        return a;
    }

    number_op2() {
        var a = new bigNumber(8);
        a |= 1;
        a &= 7;
        a ^= 3;
        return a == 2 ? 2 : 22;
    }

    number_strange() {
        var a = new bigNumber(8);
        return a / -0.0;
    }

    param() {
        return 3, 4;
    }

    param2(a) {
        return a;
    }

    bool() {
        var a = true;
        var b = true;
        return !a || !b && a;
    }

    string() {
        var a = "test";
        for (var i = 0; i < 10; i += 1) {
            a = a + a;
        }
        return a;
    }

    array() {
        var a = [0, 1, 2];
        a.push(3);
        return a;
    }

    object() {
        var a = {
            pos: 0,
            str: "test object"
        }
        return a;
    }
};

module.exports = Datatype;