'use strict';

class BaseOp {
    constructor() {
    }

    doCall(num) {
        function F(a) {
            return F
        }
        let a = F
        for (let i = 0; i < num; i++) {
            a = a("justtest")
        }
    }

    doNew(num) {
        for (let i = 0; i < num; i++) {
            let a = "justtest"
        }
    }

    doThrow(num) {
        for (let i = 0; i < num; i++) {
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
        for (let i = 0; i < num; i++) {
            a.next()
        }
    }

    doMember(num) {
        let a = {
            justtest: "justtest"
        }
        for (let i = 0; i < num; i++) {
            a.justtest
        }
    }

    doMeta(num) {
        for (let i = 0; i < num; i++) {
            new.target
        }
    }

    doAssignment(num) {
        let a = "justtest"
        for (let i = 0; i < num; i++) {
            a = "justtest"
        }
    }

    doPlus(num) {
        let a = 9999
        for (let i = 0; i < num; i++) {
            a++;
        }
    }

    doAdd(num) {
        for (let i = 0; i < num; i++) {
            9999 + 8888;
        }
    }

    doSub(num) {
        for (let i = 0; i < num; i++) {
            9999 - 8888;
        }
    }

    doMutiple(num) {
        for (let i = 0; i < num; i++) {
            9999 * 8888;
        }
    }

    doDiv(num) {
        for (let i = 0; i < num; i++) {
            9999 / 8888;
        }
    }

    doNot(num) {
        for (let i = 0; i < num; i++) {
            ~1;
        }
    }

    doAnd(num) {
        for (let i = 0; i < num; i++) {
            1 && 0;
        }
    }

    doConditional(num) {
        for (let i = 0; i < num; i++) {
            true?true:true;
        }
    }
};

module.exports = BaseOp;