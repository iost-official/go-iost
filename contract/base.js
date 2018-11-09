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
        return JSON.parse(storage.get(k));
    }
    _put(k, v) {
        storage.put(k, JSON.stringify(v));
    }

    _vote() {
        BlockChain.call("iost.vote_producer", "Stat", `[]`);
    }

    // The first contract executed
    Exec() {
        const bn = this._getBlockNumber();
        const execBlockNumber = this._get("execBlockNumber");
        if (bn === execBlockNumber){
            return true
        }
        this._put("execBlockNumber", bn);

        this._vote();
    }

}

module.exports = Base