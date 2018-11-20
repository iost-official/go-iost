class Base {
    constructor() {
    }
    init() {
        this._put("execBlockNumber", 0);
    }

    _getBlockNumber() {
        const bi = JSON.parse(BlockChain.blockInfo());
        if (!bi || bi === undefined || bi.number === undefined) {
            throw new Error("get block number failed. bi = " + bi);
        }
        return bi.number;
    }
    _get(k) {
        const val = storage.get(k);
        if (val === "") {
            return null;
        }
        return JSON.parse(val);
    }

    _put(k, v) {
        storage.put(k, JSON.stringify(v));
    }

    _vote() {
        BlockChain.call("vote_producer.iost", "Stat", `[]`);
    }

    _bonus(data) {
        BlockChain.call("bonus.iost", "IssueContribute", JSON.stringify([data]));
    }

    // The first contract executed
    Exec(data) {
        const bn = this._getBlockNumber();
        const execBlockNumber = this._get("execBlockNumber");
        if (bn === execBlockNumber){
            return true
        }
        this._put("execBlockNumber", bn);

        this._vote();
        this._bonus(data);
    }

}

module.exports = Base;
