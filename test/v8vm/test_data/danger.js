'use strict';
class Danger {
    constructor() {
    }

    bigArray() {
        return new Array(1000000000)
    }

    visitUndefined() {
        let a = undefined;
        a.c = 1
    }

    throw() {
        throw("test throw")
    }
};

module.exports = function () {
    return undefined;
};
module.exports = Danger;
