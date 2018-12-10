class Test {
    init() {
        //Execute once when contract is packed into a block
    }

    constructor() {
        //Execute everytime the contract class is called
    }

    transfer(from, to, amount) {
        //Function called by other
        blockchain.transfer(from, to, amount, "")

    }

};
module.exports = Test;
