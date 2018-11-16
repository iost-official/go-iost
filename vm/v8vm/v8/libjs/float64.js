'use strict';

const MaxFloat64 = new BigNumber('9223372036854775807');
const MinFloat64 = new BigNumber('-9223372036854775808');

class Float64 {
    constructor(n, base) {
        this.number = new BigNumber(n, base);
        this._validate();
    }

    _validate() {
        if (this.number.gt(MaxFloat64)) {
            throw new Error('Float64: ' + this.number + ' overflow float64');
        }

        if (this.number.lt(MinFloat64)) {
            throw new Error('Float64: ' + this.number + ' underflow float64');
        }
    }

    // Check is argument float64
    _checkArgument(arg) {
        if (typeof arg === 'undefined' || arg == null) {
            throw new Error('float64 argument: ' + arg + ' is empty');
        }

        if (!(arg instanceof Float64) || arg.constructor !== this.constructor) {
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
    negated(n) {
        let arg = this._checkArgument(n);
        arg.number = arg.number.negated();
        return arg;
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

module.exports = Float64;
