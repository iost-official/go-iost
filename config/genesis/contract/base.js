const producerPermission = "active";
const voteStatInterval = 1200;
const issueInterval = 172800;

class Base {
    init() {
        this._put("execBlockNumber", 0);
    }

    initAdmin(adminID) {
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
        const ret = blockchain.requireAuth(account, permission);
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
        blockchain.callWithAuth("vote_producer.iost", "stat", `[]`);
    }

    _bonus(data) {
        blockchain.callWithAuth("bonus.iost", "issueContribute", [data]);
    }

    _issue() {
        blockchain.callWithAuth("issue.iost", "issueIOST", `[]`);
    }

    _saveBlockInfo() {
        let json = storage.get("current_block_info");
        storage.put("chain_info_" + block.parentHash, JSON.stringify(json));
        storage.put("current_block_info", JSON.stringify(block))
    }

    _saveWitnessInfo() {
        let map = JSON.parse(storage.get("witness_produced") || '{}');
        map[block.witness] = (map[block.witness] || 0) + 1;
        storage.put("witness_produced", JSON.stringify(map));
    }

    _clearWitnessInfo() {
        storage.del("witness_produced");
    }

    // The first contract executed
    exec(data) {
        const bn = block.number;
        const execBlockNumber = this._get("execBlockNumber");
        if (bn === execBlockNumber){
            return true
        }
        this._put("execBlockNumber", bn);

        this._saveBlockInfo();
        this._saveWitnessInfo();
        if (bn%voteStatInterval === 0 && data.parent[2] === false){
            this._vote();
            this._clearWitnessInfo();
        }
        if (bn%issueInterval === 0) {
            this._issue();
        }
        this._bonus(data);
    }

}

module.exports = Base;
