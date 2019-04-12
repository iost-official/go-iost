class Test {
    ops() {
        if (new BigNumber("-10").abs().toFixed() !== "10") {
            throw "abs error";
        }
        if (new BigNumber("10").div(3).toFixed(3) !== "3.333") {
            throw "div error";
        }
        if (new BigNumber("10").div(new BigNumber(3)).toFixed(3) !== "3.333") {
            throw "div error";
        }
        if (new BigNumber("10").idiv(3).toFixed() !== "3") {
            throw "idiv error";
        }
        if (new BigNumber("10").idiv(new BigNumber(3)).toFixed() !== "3") {
            throw "idiv error";
        }
        if (new BigNumber("10").pow(3).toFixed() !== "1000") {
            throw "pow error";
        }
        if (new BigNumber("10.2").integerValue().toFixed() !== "10") {
            throw "integerValue error";
        }
        if (new BigNumber("10").eq(10) !== true) {
            throw "eq error";
        }
        if (new BigNumber("10").eq(new BigNumber(10)) !== true) {
            throw "eq error";
        }
        if (new BigNumber("10").eq(100) !== false) {
            throw "eq error";
        }
        if (new BigNumber("10").eq(new BigNumber(100)) !== false) {
            throw "eq error";
        }
        if (new BigNumber("10").isFinite() !== true) {
            throw "isFinite error";
        }
        if (new BigNumber(Infinity).isFinite() !== false) {
            throw "isFinite error";
        }
        if (new BigNumber("10").gt(9) !== true) {
            throw "gt error";
        }
        if (new BigNumber("10").gt(new BigNumber(9)) !== true) {
            throw "gt error";
        }
        if (new BigNumber("10").gt(90) !== false) {
            throw "gt error";
        }
        if (new BigNumber("10").gt(new BigNumber(90)) !== false) {
            throw "gt error";
        }
        if (new BigNumber("10").gte(10) !== true) {
            throw "gte error";
        }
        if (new BigNumber("10").gte(new BigNumber(10)) !== true) {
            throw "gte error";
        }
        if (new BigNumber("10").gt(90) !== false) {
            throw "gte error";
        }
        if (new BigNumber("10").gt(new BigNumber(90)) !== false) {
            throw "gte error";
        }
        if (new BigNumber("10.0").isInteger() !== true) {
            throw "isInteger error";
        }
        if (new BigNumber("10.1").isInteger() !== false) {
            throw "isInteger error";
        }
        if (new BigNumber("10").lt(11) !== true) {
            throw "lt error";
        }
        if (new BigNumber("10").lt(new BigNumber(11)) !== true) {
            throw "lt error";
        }
        if (new BigNumber("10").lt(1) !== false) {
            throw "lt error";
        }
        if (new BigNumber("10").lt(new BigNumber(1)) !== false) {
            throw "lt error";
        }
        if (new BigNumber("10").lte(10) !== true) {
            throw "lte error";
        }
        if (new BigNumber("10").lte(new BigNumber(10)) !== true) {
            throw "lte error";
        }
        if (new BigNumber("10").lt(1) !== false) {
            throw "lte error";
        }
        if (new BigNumber("10").lt(new BigNumber(1)) !== false) {
            throw "lte error";
        }
        if (new BigNumber("10").isNaN() !== false) {
            throw "isNaN error";
        }
        if (new BigNumber(NaN).isNaN() !== true) {
            throw "isNaN error";
        }
        if (new BigNumber("10").isNegative() !== false) {
            throw "isNegative error";
        }
        if (new BigNumber("-10").isNegative() !== true) {
            throw "isNegative error";
        }
        if (new BigNumber("10").isPositive() !== true) {
            throw "isPositive error";
        }
        if (new BigNumber("-10").isPositive() !== false) {
            throw "isPositive error";
        }
        if (new BigNumber("10").isZero() !== false) {
            throw "isZero error";
        }
        if (new BigNumber(0).isZero() !== true) {
            throw "isZero error";
        }
        if (new BigNumber("10").minus(3).toFixed() !== "7") {
            throw "minus error";
        }
        if (new BigNumber("10").minus(new BigNumber(3)).toFixed() !== "7") {
            throw "minus error";
        }
        if (new BigNumber("10").mod(3).toFixed() !== "1") {
            throw "mod error";
        }
        if (new BigNumber("10").mod(new BigNumber(3)).toFixed() !== "1") {
            throw "mod error";
        }
        if (new BigNumber("10").times(3).toFixed() !== "30") {
            throw "times error";
        }
        if (new BigNumber("10").times(new BigNumber(3)).toFixed() !== "30") {
            throw "times error";
        }
        if (new BigNumber("10").negated().toFixed() !== "-10") {
            throw "negated error";
        }
        if (new BigNumber("-10").negated().toFixed() !== "10") {
            throw "negated error";
        }
        if (new BigNumber("10").plus(3).toFixed() !== "13") {
            throw "plus error";
        }
        if (new BigNumber("10").plus(new BigNumber(3)).toFixed() !== "13") {
            throw "plus error";
        }
        if (new BigNumber("10").sqrt().toFixed(3) !== "3.162") {
            throw "sqrt error";
        }
    }
}

module.exports = Test;