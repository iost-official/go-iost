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
        if (!arg instanceof Int64 || arg.constructor !== arg.constructor) {
            throw new Error('Int64 argument: ' + arg + ' is not Int64 type');
        }
        if (typeof arg === 'undefined' || arg == null) {
            throw new Error('Int64 argument: ' + arg + ' is empty');
        }
        arg._validate();
    }

    plus(n) {
        this._checkArgument(n);
        let rs = this.number.plus(n.number);
        return new this(rs);
    }

    minus(n) {
        this._checkArgument(n);
        let rs = this.number.minus(n.number);
        return new this(rs);
    }

    multi(n) {
        this._checkArgument(n);
        let rs = this.number.times(n.number);
        return new this(rs);
    }

    div(n) {
        this._checkArgument(n);
        let rs = this.number.idiv(n.number);
        return new this(rs);
    }

    mod(n) {
        this._checkArgument(n);
        let rs = this.number.mod(n);
        return new this(rs);
    }

    pow(n) {
        this._checkArgument(n);
        let rs = this.number.pow(n.number);
        return new this(rs);
    }

    sqrt() {
        let rs = this.number.sqrt();
        return new this(rs);
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