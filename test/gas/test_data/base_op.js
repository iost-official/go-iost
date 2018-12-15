'use strict';

class BaseOp {
    constructor() {
    }

    doThrowStatement(num) {
        for (let i = 0; i < num; i++) {
            try {
                throw "justtest";
            } catch (err) {
            }
        }
    }

    doCallExpression(num) {
        function f(a) {
            return
        }
        for (let i = 0; i < num; i++) {
            f("justtest")
        }
    }

    doTemplateLiteral(num) {
        let person = "iost"
        let age = "18"
        for (let i = 0; i < num; i++) {
            `that ${ person } is a ${ age }`
        }
    }

    doTaggedTemplateExpression(num) {
        function myTag(s, p, a) {
            return s
        }
        let person = "iost"
        let age = "18"
        for (let i = 0; i < num; i++) {
            myTag`that ${ person } is a ${ age }`
        }
    }

    doNewExpression(num) {
        for (let i = 0; i < num; i++) {
            new Number(10)
        }
    }

    doYieldExpression(num) {
        function *F() {
            while (true) {
                yield "justtest"
            }
        }
        let f = F()
        for (let i = 0; i < num; i++) {
            f.next()
        }
    }

    doMemberExpression(num) {
        let a = {
            justtest: "justtest"
        }
        for (let i = 0; i < num; i++) {
            a.justtest
        }
    }

    doMetaProperty(num) {
        for (let i = 0; i < num; i++) {
            new.target
        }
    }

    doAssignmentExpression(num) {
        let a = "justtest"
        for (let i = 0; i < num; i++) {
            a = "justtest"
        }
    }

    doUpdateExpression(num) {
        let a = 9999
        for (let i = 0; i < num; i++) {
            a++;
        }
    }

    doBinaryExpressionAdd(num) {
        for (let i = 0; i < num; i++) {
            9999 + 8888;
        }
    }

    doBinaryExpressionSub(num) {
        for (let i = 0; i < num; i++) {
            9999 - 8888;
        }
    }

    doBinaryExpressionMutiple(num) {
        for (let i = 0; i < num; i++) {
            9999 * 8888;
        }
    }

    doBinaryExpressionDiv(num) {
        for (let i = 0; i < num; i++) {
            9999 / 8888;
        }
    }

    doUnaryExpressionNot(num) {
        for (let i = 0; i < num; i++) {
            ~1;
        }
    }

    doLogicalExpressionAnd(num) {
        for (let i = 0; i < num; i++) {
            1 && 0;
        }
    }

    doConditionalExpression(num) {
        for (let i = 0; i < num; i++) {
            true?true:true;
        }
    }
};

module.exports = BaseOp;