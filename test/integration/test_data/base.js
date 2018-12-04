class Base {
    init() {
    }

    Stat() {
        BlockChain.callWithAuth("vote_producer.iost", "Stat", '[]');
    }

    IssueContribute(data) {
        BlockChain.callWithAuth("bonus.iost", "IssueContribute", JSON.stringify([data]));
    }
}

module.exports = Base;