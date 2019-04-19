class Test {
    init() {
    }

    inner() {
        return blockchain.requireAuth(blockchain.contractName(), "active");
    }

    outer() {
        return blockchain.callWithAuth(blockchain.contractName(), "inner", []);
    }
}

module.exports = Test;