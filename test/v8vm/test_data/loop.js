'use strict';
class Datatype {
    for() {
        for (;;) {}
    }

    for2() {
        for (let i = 0; i >= 0; i++) {}
    }

    for3() {
        let i = 0;
        for (i = 1; i >= 0; i++)
            if (i === 10)
                break;
        return i
    }

    forin() {
        let a = [0,1,2,3];
        let r = 0;
        for (let i in a) {
            r += Number(i);
            r += a[i]
        }
        return r
    }

    forof() {
        let a = [0,1,2,3];
        let r = 0;
        for (let i of a)
            r += i
        return r
    }

    while() {
        while (1) {}
    }

    dowhile() {
        do {
            let i = 9;
        }
        while (true)
    }
};

module.exports = Datatype;