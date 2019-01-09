class Base {
    init() {
    }

    stat() {
        blockchain.callWithAuth("vote_producer.iost", "stat", '[]');
    }

    issueContribute(data) {
        blockchain.callWithAuth("bonus.iost", "issueContribute", JSON.stringify([data]));
    }
}

module.exports = Base;
