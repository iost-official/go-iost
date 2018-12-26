'use strict';

const MaxInt64 = new BigNumber('9223372036854775807');
const MinInt64 = new BigNumber('-9223372036854775808');

class Int64 {
    constructor(n, base) {
        this.number = new BigNumber(n, base);
        this._validate();
    }

    // Check is int64 (Interger that greater than MinInt64, less than MaxInt64)
    _validate() {
        if (!this.number.isInteger()) {
            throw new Error('Int64: ' + this.number + ' is not an integer');
        }

        if (this.number.gt(MaxInt64)) {
            throw new Error('Int64: ' + this.number + ' overflow int64');
        }

        if (this.number.lt(MinInt64)) {
            throw new Error('Int64: ' + this.number + ' underflow int64');
        }
    }

    // Check is argument int64
    _checkArgument(arg) {
        if (typeof arg === 'undefined' || arg == null) {
            throw new Error('Int64 argument: ' + arg + ' is empty');
        }

        if (!(arg instanceof Int64) || arg.constructor !== this.constructor) {
            arg = new this.constructor(arg);
        }

        arg._validate();

        return arg
    }

    // plus n
    plus(n) {
        let arg = this._checkArgument(n);
        let rs = this.number.plus(arg.number);
        return new this.constructor(rs);
    }

    // minus n
    minus(n) {
        let arg = this._checkArgument(n);
        let rs = this.number.minus(arg.number);
        return new this.constructor(rs);
    }

    // Multi n
    multi(n) {
        let arg = this._checkArgument(n);
        let rs = this.number.times(arg.number);
        return new this.constructor(rs);
    }

    // Div n
    div(n) {
        let arg = this._checkArgument(n);
        let rs = this.number.idiv(arg.number);
        return new this.constructor(rs);
    }

    // Mod n
    mod(n) {
        let arg = this._checkArgument(n);
        let rs = this.number.mod(arg.number);
        return new this.constructor(rs);
    }

    // shift n
    shift(n) {
        let rs = this.number.shiftedBy(n);
        return new this.constructor(rs);
    }

    // Power n
    pow(n) {
        let arg = this._checkArgument(n);
        let rs = this.number.pow(arg.number);
        return new this.constructor(rs);
    }

    // Check equal n
    eq(n) {
        let arg = this._checkArgument(n);
        return this.number.eq(arg.number);
    }

    // Check greater than n
    gt(n) {
        let arg = this._checkArgument(n);
        return this.number.gt(arg.number);
    }

    // Check greater than or equal to n
    gte(n) {
        let arg = this._checkArgument(n);
        return this.number.gte(arg.number);
    }

    // Check less than n
    lt(n) {
        let arg = this._checkArgument(n);
        return this.number.lt(arg.number);
    }

    // Check less than or equal to n
    lte(n) {
        let arg = this._checkArgument(n);
        return this.number.lte(arg.number);
    }

    // negated
    negated() {
        return this.number.negated();
    }

    // Check is Zero
    isZero() {
        return this.number.isZero();
    }

    // Check is Positive
    isPositive() {
        return this.number.isPositive();
    }

    // Check is Negative
    isNegative() {
        return this.number.isNegative();
    }

    // convert to String
    toString() {
        return this.number.toString();
    }

    // to fixed
    toFixed(n) {
        return this.number.toFixed(n);
    }
}

module.exports = Int64;