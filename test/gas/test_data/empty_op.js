'use strict';

class EmptyOp {
    doStartUp() {
    }

    doEmpty(num) {
        for (let i = 0; i < num; i++) {
        }
    }
};

module.exports = EmptyOp;