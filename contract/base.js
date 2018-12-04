const producerPermission = "active";
const voteStatInterval = 200;

class Base {
    constructor() {
    }
    init() {
        this._put("execBlockNumber", 0);
    }

    InitAdmin(adminID) {
        const bn = block.number;
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
        const ret = BlockChain.requireAuth(account, permission);
        if (ret !== true) {
            throw new Error("require auth failed. ret = " + ret);
        }
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
        BlockChain.callWithAuth("vote_producer.iost", "Stat", `[]`);
    }

    _bonus(data) {
        BlockChain.callWithAuth("bonus.iost", "IssueContribute", JSON.stringify([data]));
    }

    _saveBlockInfo() {
        let json = storage.get("current_block_info");
        storage.put("chain_info_" + block.parentHash, JSON.stringify(json));
        storage.put("current_block_info", JSON.stringify(block))
    }

    // The first contract executed
    Exec(data) {
        this._saveBlockInfo();
        const bn = block.number;
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
