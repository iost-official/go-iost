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
        this._checkArgument(n);
        return this.number.eq(n.number);
    }

    // Check greater than n
    gt(n) {
        this._checkArgument(n);
        return this.number.gt(n.number);
    }

    // Check less than n
    lt(n) {
        this._checkArgument(n);
        return this.number.lt(n.number);
    }

    // Check is Zero
    isZero() {
        return this.number.isZero();
    }

    // convert to String
    toString() {
        return this.number.toString();
    }
}

module.exports = Float64;
