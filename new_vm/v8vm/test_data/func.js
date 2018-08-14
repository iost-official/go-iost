class Func {
    constructor() {
    }

    func1() {
        this.func2()
    }

    func2() {
        this.func1()
    }

    func3(a) {
        return function(a) {
            return 9;
        }(a)
    }

    func4() {
        var a = ["i", "love", "iost"]
        var b = a.map(w => (w.length))
        return b.map(w => function () {
            return w + 1
        }());
    }
};

module.exports = Func;