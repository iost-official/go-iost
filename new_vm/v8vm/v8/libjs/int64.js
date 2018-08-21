'use strict';

const MaxInt64 = new BigNumber('9223372036854775807');
const MinInt64 = new BigNumber('-9223372036854775808');

class Int64 {
    constructor(n, base) {
        this.number = new BigNumber(n, base);
        this._validate();
    }

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

    _checkArgument(arg) {
        if (typeof arg === 'undefined' || arg == null) {
            throw new Error('Int64 argument: ' + arg + ' is empty');
        }

        if (!(arg instanceof Int64) || arg.constructor !== arg.constructor) {
            arg = new this.constructor(arg);
        }

        arg._validate();

        return arg
    }

    plus(n) {
        let arg = this._checkArgument(n);
        let rs = this.number.plus(arg.number);
        return new this.constructor(rs);
    }

    minus(n) {
        let arg = this._checkArgument(n);
        let rs = this.number.minus(arg.number);
        return new this.constructor(rs);
    }

    multi(n) {
        let arg = this._checkArgument(n);
        let rs = this.number.times(arg.number);
        return new this.constructor(rs);
    }

    div(n) {
        let arg = this._checkArgument(n);
        let rs = this.number.idiv(arg.number);
        return new this.constructor(rs);
    }

    mod(n) {
        let arg = this._checkArgument(n);
        let rs = this.number.mod(arg.number);
        return new this.constructor(rs);
    }

    pow(n) {
        let arg = this._checkArgument(n);
        let rs = this.number.pow(arg.number);
        return new this.constructor(rs);
    }

    sqrt() {
        let rs = this.number.sqrt();
        rs = rs.integerValue();
        return new this.constructor(rs);
    }

    eq(n) {
        this._checkArgument(n);
        return this.number.eq(n.number);
    }

    gt(n) {
        this._checkArgument(n);
        return this.number.gt(n.number);
    }

    lt(n) {
        this._checkArgument(n);
        return this.number.lt(n.number);
    }

    isZero() {
        return this.number.isZero();
    }


    toString() {
        return this.number.toString();
    }
}

module.exports = Int64;