class Test {
    init() {
        storage.put("num", JSON.stringify(9))
    }

    number() {
        const snum = storage.get("num");
        const num = JSON.parse(snum);
        return this._num(num)
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
