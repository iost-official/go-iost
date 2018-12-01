'use strict';
class Gas {
    constructor() {
    }
    assignment0() {
        [[[[[[[[[[[[[[[[[[[{a=b[0]}]]]]]]]]]]]]]]]]]]]=[0];
    }
    assignment1() {
        // todo should calc instruction according to length
        let a = "testabcddddddddd";
        let b = [1,2,3,4,5,6];
    }
    assignment2(N) {
        // todo should forbid Array.from
        Array.from({length: N}, (val, index) => index);
    }
    assignment3(N) {
        // can't reach Arrayconcat to get back original prototype
        let a = new Array(N);
        Array.prototype.concat = function () {
            return Arrayconcat.call(this, ...arguments);
        };
        a = a.concat(a);
    }
    counter0(N) {
        const a = new IOSTInstruction;
        return _IOSTInstruction_counter.prototype.constructor;
    }
};
module.exports = Gas;
