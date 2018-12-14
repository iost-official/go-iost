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
            String.prototype.constructor("yellow")
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
            Array.prototype.constructor(["red", "blue", "green"])
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

    // Math
    doMathAbs(num) {
        for (let i = 0; i < num; i++) {
            Math.abs(-1)
        }
    }

    doMathCbrt(num) {
        for (let i = 0; i < num; i++) {
            Math.cbrt(2.5)
        }
    }

    doMathCeil(num) {
        for (let i = 0; i < num; i++) {
            Math.ceil(-1.5)
        }
    }

    doMathFloor(num) {
        for (let i = 0; i < num; i++) {
            Math.floor(1.5)
        }
    }

    doMathLog(num) {
        for (let i = 0; i < num; i++) {
            Math.log(5)
        }
    }

    doMathLog10(num) {
        for (let i = 0; i < num; i++) {
            Math.log10(1234)
        }
    }

    doMathLog1p(num) {
        for (let i = 0; i < num; i++) {
            Math.log1p(0.7)
        }
    }

    doMathMax(num) {
        for (let i = 0; i < num; i++) {
            Math.max(3, 10, 5)
        }
    }

    doMathMin(num) {
        for (let i = 0; i < num; i++) {
            Math.min(3, 10, 5)
        }
    }

    doMathPow(num) {
        for (let i = 0; i < num; i++) {
            Math.pow(3, 15.5)
        }
    }

    doMathRound(num) {
        for (let i = 0; i < num; i++) {
            Math.round(2.7)
        }
    }

    doMathSqrt(num) {
        for (let i = 0; i < num; i++) {
            Math.sqrt(3.4)
        }
    }

    // BigNumber
    doBigNumberConstructor(num) {
        for (let i = 0; i < num; i++) {
            BigNumber.prototype.constructor("99999999999999999999999999999")
        }
    }

    doBigNumberAbs(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").abs()
        }
    }

    doBigNumberDiv(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").div("99999")
        }
    }

    doBigNumberIdiv(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").idiv("99999")
        }
    }

    doBigNumberPow(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").pow("99999")
        }
    }

    doBigNumberIntegerValue(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").integerValue()
        }
    }

    doBigNumberEq(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").eq("99999")
        }
    }

    doBigNumberIsFinite(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").isFinite()
        }
    }

    doBigNumberGt(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").gt("99999")
        }
    }

    doBigNumberGte(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").gte("99999")
        }
    }

    doBigNumberIsInteger(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").isInteger()
        }
    }

    doBigNumberLt(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").lt("99999")
        }
    }

    doBigNumberLte(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").lte("99999")
        }
    }

    doBigNumberIsNaN(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").isNaN()
        }
    }

    doBigNumberIsNegative(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").isNegative()
        }
    }

    doBigNumberIsPositive(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").isPositive()
        }
    }

    doBigNumberIsZero(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").isZero()
        }
    }

    doBigNumberMinus(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").minus("99999")
        }
    }

    doBigNumberMod(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").mod("99999")
        }
    }

    doBigNumberTimes(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").times("99999")
        }
    }

    doBigNumberNegated(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").negated()
        }
    }

    doBigNumberPlus(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").plus("99999")
        }
    }

    doBigNumberSqrt(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").sqrt()
        }
    }

    doBigNumberToFixed(num) {
        for (let i = 0; i < num; i++) {
            new BigNumber("-99999999999999999999999999999").toFixed(5)
        }
    }

};

module.exports = LibOp;