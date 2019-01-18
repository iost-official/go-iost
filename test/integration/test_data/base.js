class Base {
    init() {
    }

    initWitness(lst) {
        const map = {};
        for (const witness of lst) {
            map[witness] = 1;
        }
        storage.put("witness_produced", JSON.stringify(map));
    }

    stat() {
        blockchain.callWithAuth("vote_producer.iost", "stat", '[]');
    }

    issueContribute(data) {
        blockchain.callWithAuth("bonus.iost", "issueContribute", JSON.stringify([data]));
    }
}

module.exports = Base;
