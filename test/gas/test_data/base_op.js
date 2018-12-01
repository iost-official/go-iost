'use strict';

class BaseOp {
    constructor() {
    }

    doEmpty(num) {
        for (let i in Array.apply(null, { length: num })) {
        }
    }

    doCall(num) {
        function F(a) {
            return F
        }
        let a = F
        for (let i in Array.apply(null, { length: num })) {
            a = a("justtest")
        }
    }

    doNew(num) {
        for (let i in Array.apply(null, { length: num })) {
            let a = "justtest"
        }
    }

    doThrow(num) {
        for (let i in Array.apply(null, { length: num })) {
            try {
                throw "justtest";
            } catch (err) {
            }
        }
    }

    doYield(num) {
        function *F() {
            while (true) {
                yield F
            }
        }
        let a = F()
        for (let i in Array.apply(null, { length: num })) {
            a.next()
        }
    }

    doMember(num) {
        let a = {
            justtest: "justtest"
        }
        for (let i in Array.apply(null, { length: num })) {
            a.justtest
        }
    }

    doMeta(num) {
        for (let i in Array.apply(null, { length: num })) {
            new.target
        }
    }

    doAssignment(num) {
        let a = "justtest"
        for (let i in Array.apply(null, { length: num })) {
            a = "justtest"
        }
    }

    doPlus(num) {
        let a = 9999
        for (let i in Array.apply(null, { length: num })) {
            a++;
        }
    }

    doAdd(num) {
        for (let i in Array.apply(null, { length: num })) {
            9999 + 8888;
        }
    }

    doSub(num) {
        for (let i in Array.apply(null, { length: num })) {
            9999 - 8888;
        }
    }

    doMutiple(num) {
        for (let i in Array.apply(null, { length: num })) {
            9999 * 8888;
        }
    }

    doDiv(num) {
        for (let i in Array.apply(null, { length: num })) {
            9999 / 8888;
        }
    }

    doNot(num) {
        for (let i in Array.apply(null, { length: num })) {
            ~1;
        }
    }

    doAnd(num) {
        for (let i in Array.apply(null, { length: num })) {
            1 && 0;
        }
    }

    doConditional(num) {
        for (let i in Array.apply(null, { length: num })) {
            true?true:true;
        }
    }
};

module.exports = BaseOp;