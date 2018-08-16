class Test {
    constructor() {
		this.num = 9;
    }

    number() {
        return this._num(this.num);
    }

    _num(d) {
        return d;
    }
};
module.exports = Test;
