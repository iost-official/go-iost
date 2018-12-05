class Contract {
    init() {

    }
    f1() {
        let a = "22";
        let b = "22";
        let c = a + b;
    }
    f2() {
        let a = "999999999";
        let b = "999999999";
        let c = a + b;
    }
    f3() {
        let a = "999999999";
        let b = "999999999";
        let c = a === b;
    }
    f4() {
        let a = "999999999";
        let b = "999999999";
        let c = a += b;
    }
}

module.exports = Contract;