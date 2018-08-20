class Test {
    constructor() {
        this.num = 9;
    }

    number() {
        return this._num(this.num)
    }

    _num(d) {
        return d;
    }

    can_update(d) {
        return true
    }

    can_destroy() {
        return true
    }
};
module.exports = Test;
