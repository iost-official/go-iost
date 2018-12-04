'use strict';

class LibOp {
    constructor() {
    }

    doStringToString(num) {
        for (let i in Array.apply(null, { length: num })) {
            ("redbluegreen").toString()
        }
    }

    doStringValueOf(num) {
        for (let i in Array.apply(null, { length: num })) {
            ("redbluegreen").valueOf()
        }
    }

    doStringConcat(num) {
        for (let i in Array.apply(null, { length: num })) {
            ("redbluegreen").concat("redbluegreen")
        }
    }

    doStringIncludes(num) {
        for (let i in Array.apply(null, { length: num })) {
            ("redbluegreen").includes("red")
        }
    }

    doStringEndsWith(num) {
        for (let i in Array.apply(null, { length: num })) {
            ("redbluegreen").endsWith("red")
        }
    }

    doStringIndexOf(num) {
        for (let i in Array.apply(null, { length: num })) {
            ("redbluegreen").indexOf("red")
        }
    }

    doStringLastIndexOf(num) {
        for (let i in Array.apply(null, { length: num })) {
            ("redbluegreen").lastIndexOf("red")
        }
    }

    doStringReplace(num) {
        for (let i in Array.apply(null, { length: num })) {
            ("redbluegreen").replace("red", "blue")
        }
    }

    // Array
    doArrayToString(num) {
        for (let i in Array.apply(null, { length: num })) {
            (["red", "blue", "green"]).toString()
        }
    }

    doArrayValueOf(num) {
        for (let i in Array.apply(null, { length: num })) {
            (["red", "blue", "green"]).valueOf()
        }
    }

    doArrayConcat(num) {
        for (let i in Array.apply(null, { length: num })) {
            (["red", "blue", "green"]).concat(["red", "blue", "green"])
        }
    }

    doArrayIncludes(num) {
        for (let i in Array.apply(null, { length: num })) {
            (["red", "blue", "green"]).includes("red")
        }
    }


};

module.exports = LibOp;