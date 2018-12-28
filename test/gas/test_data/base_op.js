'use strict';

class BaseOp {
    constructor() {
    }

    doThrowStatement(num) {
        for (let i = 0; i < num; i++) {
            try {
                throw 123456;
            } catch (err) {
            }
        }
    }

    doCallExpression(num) {
        function f(a, b, c, d) {
            return
        }
        for (let i = 0; i < num; i++) {
            f(10, 11, 12, 13)
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
                yield 123456
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
        let a = 123456
        for (let i = 0; i < num; i++) {
            a = 123456
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

    doSpreadElement(num) {
        function f(a, b, c, d) {
            return
        }
        for (let i = 0; i < num; i++) {
            f(...[10, 11, 12, 13])
        }
    }

    doObjectExpression(num) {
        for (let i = 0; i < num; i++) {
            let a = { name: "Jack", age: 10, 5: true }
        }
    }

    doArrayExpression(num) {
        for (let i = 0; i < num; i++) {
            [3, 5, 100]
        }
    }

    doFunctionExpression(num) {
        for (let i = 0; i < num; i++) {
            let a = function(width, height) { return width * height }
        }
    }

    doArrowFunctionExpression(num) {
        for (let i = 0; i < num; i++) {
            material => material.length
        }
    }

    doClassDeclaration(num) {
        for (let i = 0; i < num; i++) {
            class Polygon { }
        }
    }

    doFunctionDeclaration(num) {
        for (let i = 0; i < num; i++) {
            function a(width, height) { return width * height }
        }
    }

    doVariableDeclarator(num) {
        for (let i = 0; i < num; i++) {
            let a = 1
        }
    }

    doVariableDeclaratorWithoutInit(num) {
        for (let i = 0; i < num; i++) {
            let a
        }
    }

    doMethodDefinition(num) {
        for (let i = 0; i < num; i++) {
            let obj = {
                foo() {
                    return 'bar';
                }
            }
        }
    }

    doStringLiteral(num) {
        for (let i = 0; i < num; i++) {
            "justtest"
        }
    }

    doForStatement(num) {
        for (let i = 0; i < num; i++) {
            for (let j = 0; j < 20; j++) { }
        }
    }

    doForInStatement(num) {
        for (let i = 0; i < num; i++) {
            for (let j in [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19]) { }
        }
    }

    doForOfStatement(num) {
        for (let i = 0; i < num; i++) {
            for (let j of [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19]) { }
        }
    }

    doWhileStatement(num) {
        for (let i = 0; i < num; i++) {
            let j = 0
            while (j < 20) { j++ }
        }
    }

    doDoWhileStatement(num) {
        for (let i = 0; i < num; i++) {
            let j = 0
            do { j++ } while (j < 20)
        }
    }
};

module.exports = BaseOp;