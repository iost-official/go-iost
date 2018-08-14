class Func {
    constructor() {
    }

    func1() {
        this.func2()
    }

    func2() {
        this.func1()
    }
};

module.exports = Func;