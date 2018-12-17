class Base {
    init() {
    }

    Stat() {
        blockchain.callWithAuth("vote_producer.iost", "Stat", '[]');
    }

    IssueContribute(data) {
        blockchain.callWithAuth("bonus.iost", "IssueContribute", JSON.stringify([data]));
    }
}

module.exports = Base;
