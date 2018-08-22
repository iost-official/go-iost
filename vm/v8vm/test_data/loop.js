class Datatype {
    constructor() {
    }

    for() {
        for (;;) {}
    }

    for2() {
        for (var i = 0; i >= 0; i++) {}
    }

    for3() {
        var i = 0;
        for (i = 1; i >= 0; i++)
            if (i == 10)
                break
            else
                continue
        return i
    }

    forin() {
        var a = [0,1,2,3]
        var r = 0
        for (var i in a) {
            r += Number(i)
            r += a[i]
        }
        return r
    }

    forof() {
        var a = [0,1,2,3]
        var r = 0
        for (let i of a)
            r += i
        return r
    }

    while() {
        while (1) {}
    }

    dowhile() {
        do {
            var i = 9;
        }
        while (true)
    }
};

module.exports = Datatype;