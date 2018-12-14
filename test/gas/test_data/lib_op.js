'use strict';

class LibOp {
    constructor() {
    }

    // String
    doStringCharAt(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").charAt(3)
        }
    }

    doStringCharCodeAt(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").charCodeAt(3)
        }
    }

    doStringLength(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").length
        }
    }

    doStringConstructor(num) {
        for (let i = 0; i < num; i++) {
            String.constructor("yellow")
        }
    }

    doStringToString(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").toString()
        }
    }

    doStringValueOf(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").valueOf()
        }
    }

    doStringConcat(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").concat("redbluegreen")
        }
    }

    doStringIncludes(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").includes("red")
        }
    }

    doStringEndsWith(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").endsWith("red")
        }
    }

    doStringIndexOf(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").indexOf("red")
        }
    }

    doStringLastIndexOf(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").lastIndexOf("red")
        }
    }

    doStringReplace(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").replace("red", "blue")
        }
    }

    doStringSearch(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").search("red")
        }
    }

    doStringSplit(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").split("red")
        }
    }
    
    doStringStartsWith(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").startsWith("red")
        }
    }

    doStringSlice(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").slice()
        }
    }

    doStringSubstring(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").substring(3)
        }
    }

    doStringToLowerCase(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").toLowerCase()
        }
    }

    doStringToUpperCase(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").toUpperCase()
        }
    }

    doStringTrim(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").trim()
        }
    }

    doStringTrimLeft(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").trimLeft()
        }
    }

    doStringTrimRight(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").trimRight()
        }
    }

    doStringRepeat(num) {
        for (let i = 0; i < num; i++) {
            ("redbluegreen").repeat(10)
        }
    }

    // Array
    doArrayConstructor(num) {
        for (let i = 0; i < num; i++) {
            Array.constructor(["red", "blue", "green"])
        }
    }

    doArrayToString(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).toString()
        }
    }

    doArrayConcat(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).concat(["red", "blue", "green"])
        }
    }

    doArrayEvery(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).every(function (x) { return true; })
        }
    }
    
    doArrayFilter(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).filter(function (x) { return true; })
        }
    }

    doArrayFind(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).find(function (x) { return true; })
        }
    }

    doArrayFindIndex(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).findIndex(function (x) { return true; })
        }
    }

    doArrayForEach(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).forEach(function (x) { return; })
        }
    }

    doArrayIncludes(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).includes("red")
        }
    }

    doArrayIndexOf(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).indexOf("red")
        }
    }

    doArrayJoin(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).join()
        }
    }

    doArrayKeys(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).keys()
        }
    }

    doArrayLastIndexOf(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).lastIndexOf("red")
        }
    }

    doArrayMap(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).map(function (x) { return x; })
        }
    }

    doArrayPop(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).pop()
        }
    }

    doArrayPush(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).push("yellow")
        }
    }

    doArrayReverse(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).reverse()
        }
    }

    doArrayShift(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).shift()
        }
    }

    doArraySlice(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).slice()
        }
    }

    doArraySort(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).sort()
        }
    }

    doArraySplice(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).splice(0)
        }
    }

    doArrayUnshift(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).unshift("yellow")
        }
    }

    // JSON
    doJSONParse(num) {
        for (let i = 0; i < num; i++) {
            JSON.parse(`
            {
                "array": ["red", "blue", "green"],
                "string": "",
                "float": 100
            }
            `)
        }
    }

    doJSONStringify(num) {
        for (let i = 0; i < num; i++) {
            JSON.stringify(
                {
                    array: ["red", "blue", "green"],
                    string: "",
                    float: 100
                }
            )
        }
    }
};

module.exports = LibOp;