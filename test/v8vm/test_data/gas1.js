'use strict';
function* range(N) {
    let i = 0;
    while (i < N) {
        yield i++;
    }
}

function raw(strings, ...keys) {
    return strings.join(',') + keys.join(',');
}

class Gas {
    constructor() {
    }
    init() {
    }
    assignment0() {
        //[[[[[[[[[[[[[[[[[[[{a=b[0]}]]]]]]]]]]]]]]]]]]]=[0];
    }
    assignment1() {
        let a = "testabcddddddddd";
        let d = a;
        d[0] = "a";
        return d;
    }

    assignment11() {
        let b = [1,2,3,4,5,6];
        let c = b;
        c[0] = 0;
        return b;
    }

    assignment111() {
        let b = {"p0": "p0", "p1": "p1"};
        let c = b;
        c.p0 = "ppp";
        return b;
    }

    assignment2(N) {
        // should forbid Array.from
        //Array.from({length: N}, (val, index) => index);
    }
    assignment3(N) {
        // can't reach Arrayconcat to get back original prototype
        let a = new Array(N);
        Array.prototype.concat = function () {
            return Arrayconcat.call(this, ...arguments);
        };
        a = a.concat(a);
    }
    // deconstruct assignment is not allowed now
    /*
    assignment4(N) {
        let a = new Array();
        a[N] = 1;
        // should handle array deconstruct assignment
        let [c,d,,...e] = a;
        a[4] = 9;
        return e;
    }
    assignment44(N) {
        let a = "abcdefg";
        let [c,d,,...e] = a;
        return e;
    }
    assignment444(N) {
        let a = {"p0": "p0", "p1": "p1"};
        let {p1} = a;
        a.p1 = "ppp";
        return p1;
    }
    assignment4444(N) {
        let a = [0, 1, [2, 3, 4]];
        //let [c,d,e] = a;
        a[2][0] = 5;
        return a;
    }
    */
    counter0(N) {
        let names = Object.getOwnPropertyNames("_IOSTInstruction" + "_counter");
        let methods = new Array();
        names.forEach((method) => {
            methods = methods.concat(method);
        });
        methods = methods.concat(methods.concat.toString());
        return methods;
    }
    yield0(N) {
        let g = range(N);
        let num = 0;
        while (!g.next().done) {
            num++;
        }
        return num;
    }
    yield1(N) {
        function *g(x) {
            /*
            (x = yield) => {}
            */
        }
        let ag = g(N);
        while (!ag.next().done) {}
    }
    library0(N) {
        let a = new esprima;
    }
    library1(N) {
        let a = require("inject_gas");
        console.log(Object.getOwnPropertyNames(a).toString());
        console.log(a.toString());
        console.log(a("123"));
        console.log(new a("123"));
    }
    eval0(N) {
        let x = 0;
        //eval.call(null, 'x = 0; for (let i = 0; i < N; i++) {x += 1;}');

        //(1, eval)('x = 0; for (let i = 0; i < N; i++) {x += 1;}');

        var xeval = eval;
        xeval.call(null, 'x = 0; for (let i = 0; i < N; i++) {x += 1;}');
        return x;
    }
    function0(N) {
        let g = new Function('let a = 1; for (let i = 0; i < N; i++) {a += 1;} return a;')
    }
    function1(N) {
        function f() {}
        let f0 = new f.constructor("let a = 0; for (let i = 0; i < 100; i++) a += 2; return a;");
        return f0();
    }
    literal0(N) {
        function raw(strings, ...keys) {
            return strings + keys
        }
        let x = 0;
        return raw`token ${`nested ${`deeply ${function(){ for(let i = 0; i < N; i++) x+=1; return x}()} blah`}`}`
    }
    literal1(N) {
        // return '\10';
        // return '\x1'
    }

    templateString(N, data) {
        for (let i = 0; i < N; i += 2) {
            data = raw`${data}456`;
            data = `${data}123${data}`;
        }
        return data;
    }

    length0(a) {
        String.prototype.toString = function () {
            return {length: -100};
        };
        a.concat("a");
    }
};
module.exports = Gas;