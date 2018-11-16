class Contract {
    init() {

    }
    add2() {
        let a = "22";
        let b = "22";
        let c = a + b;
    }
    add9() {
        let a = "999999999";
        let b = "999999999";
        let c = a + b;
    }
    equal9() {
        let a = "999999999";
        let b = "999999999";
        let c = a === b;
    }
    superadd9() {
        let a = "999999999";
        let b = "999999999";
        let c = a += b;
    }
}

module.exports = Contract;