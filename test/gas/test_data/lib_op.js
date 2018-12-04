'use strict';

class LibOp {
    constructor() {
    }

    // String
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
            ("redbluegreen").startsWith("red")
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
    doArrayIsArray(num) {
        for (let i = 0; i < num; i++) {
            Array.isArray(["red", "blue", "green"])
        }
    }

    doArrayOf(num) {
        for (let i = 0; i < num; i++) {
            Array.of(["red", "blue", "green"])
        }
    }

    doArrayConcat(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).concat(["red", "blue", "green"])
        }
    }

    doArrayEvery(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).every(function (x) { return x == "read"; })
        }
    }
    
    doArrayFilter(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).filter(function (x) { return x == "read"; })
        }
    }

    doArrayFind(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).find(function (x) { return x == "read"; })
        }
    }

    doArrayFindIndex(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).findIndex(function (x) { return x == "read"; })
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
            (["red", "blue", "green"]).IndexOf("red")
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
            (["red", "blue", "green"]).map(function (x) { return x == "read"; })
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

    doArrayToString(num) {
        for (let i = 0; i < num; i++) {
            (["red", "blue", "green"]).toString()
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