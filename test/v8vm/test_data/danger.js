'use strict';

const _this = this;
class Danger {
    bigArray() {
        return new Array(1000000);
    }

    tooBigArray() {
        return new Array(1000000000);
    }

    visitUndefined() {
        let a = undefined;
        a.c = 1
    }

    throw() {
        throw("test throw")
    }

    nativerun() {
        const src = `
        const arr = new Array(10000000);
        let cnt = 0;
        for (let i=0; i < arr.length; i++) {
            cnt += 1;
        }
        cnt;
        `;
        const fn = "test.js";
        return _native_run(src, fn);
    }

    putlong() {
        storage.put('x', 'x'.repeat(65537));
    }

    objadd() {
        const s = "x".repeat(1000);
        const sObj = {
            toString() {
                return s;
            }
        };
        let r = ""
        const rObj = {
            valueOf() {
                return r;
            }
        };
        for (let i = 0; i < 100; i++) {
            r = rObj + sObj;
        }
        return r.length;
    }
};

module.exports = function () {
    return undefined;
};
module.exports = Danger;
