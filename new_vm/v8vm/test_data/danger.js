class Datatype {
    constructor() {
    }

    bigArray() {
        return new Array(1000000000)
    }

    visitUndefined() {
        var a = undefined
        a.c = 1
    }

    throw() {
        throw("test throw")
    }
};

module.exports = Datatype;