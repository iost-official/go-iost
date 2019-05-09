class Test {
    init(){
    }

    caller() {
        return blockchain.caller();
    }

    caller2() {
        return blockchain.call(blockchain.contractName(), "caller", "[]")[0];
    }
    
    transfer(to) {
        blockchain.transfer(tx.publisher, to, "0.00000001", "");
    }

    requireAuth() {
        let r0 = blockchain.requireAuth("u1", "p1");
        let r1 = blockchain.requireAuth("u1", "active");
        let r2 = blockchain.requireAuth("u1", "owner");
        let r3 = blockchain.requireAuth("u2", "p20");
        let r4 = blockchain.requireAuth("u2", "p21");
        let r5 = blockchain.requireAuth("u2", "active");
        let r6 = blockchain.requireAuth("u2", "owner");
        return [r0,r1,r2,r3,r4,r5,r6];
    }
}

module.exports = Test;
