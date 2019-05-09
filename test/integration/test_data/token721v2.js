class Test {
    init() {
    }

    approve(from, to, tokenId) {
        blockchain.call("token721.iost", "approve", ["iost", from, to, tokenId]);
    }

    approveWithAuth(from, to, tokenId) {
        blockchain.callWithAuth("token721.iost", "approve", ["iost", from, to, tokenId]);
    }

    transfer(from, to, tokenId) {
        blockchain.call("token721.iost", "transfer", ["iost", from, to, tokenId]);
    }

    transferWithAuth(from, to, tokenId) {
        blockchain.callWithAuth("token721.iost", "transfer", ["iost", from, to, tokenId]);
    }
}

module.exports = Test;