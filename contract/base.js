const producerPermission = "active";
const voteStatInterval = 200;

class Base {
    constructor() {
    }
    init() {
        this._put("execBlockNumber", 0);
    }

    InitAdmin(adminID) {
        const bn = this._getBlockNumber();
        if(bn !== 0) {
            throw new Error("init out of genesis block")
        }
        this._put("adminID", adminID);
    }

    can_update(data) {
        const admin = this._get("adminID");
        this._requireAuth(admin, producerPermission);
        return true;
    }

    _requireAuth(account, permission) {
        BlockChain.requireAuth(account, permission);
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

    _saveBlockInfo() {
        let json = storage.get("current_block_info");
        storage.mapPut("chain_info", block.parentHash, JSON.stringify(json));
        storage.put("current_block_info", JSON.stringify(block))
    }

    // The first contract executed
    Exec(data) {
        this._saveBlockInfo();
        const bn = this._getBlockNumber();
        const execBlockNumber = this._get("execBlockNumber");
        if (bn === execBlockNumber){
            return true
        }
        this._put("execBlockNumber", bn);

        if (bn%voteStatInterval === 0){
            this._vote();
        }
        this._bonus(data);
    }

}

module.exports = Base;
