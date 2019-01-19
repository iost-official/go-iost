const activePermission = "active";
const totalSupply = 90000000000;
const blockContribRadio = new Float64("9.6568764571e-11");

class BonusContract {
    init() {
        this._initContribute();
        this._put("blockContrib", "1.98779440");
        this._put("lastTime", block.time);
    }

    _initContribute() {
        blockchain.callWithAuth("token.iost", "create", [
            "contribute",
            "bonus.iost",
            totalSupply,
            {
                "can_transfer": false,
                "decimal": 8
            }
        ]);
    }

    initAdmin(adminID) {
        const bn = block.number;
        if(bn !== 0) {
            throw new Error("init out of genesis block");
        }
        this._put("adminID", adminID);
    }

    can_update(data) {
        const admin = this._get("adminID");
        this._requireAuth(admin, activePermission);
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

    _put(k, v, p) {
        storage.put(k, JSON.stringify(v), p);
    }

    _mapGet(k, f) {
        const val = storage.mapGet(k, f);
        if (val === "") {
            return null;
        }
        return JSON.parse(val);
    }

    _mapPut(k, f, v, p) {
        storage.mapPut(k, f, JSON.stringify(v), p);
    }

    _mapDel(k, f) {
        storage.mapDel(k, f);
    }

    _globalMapGet(c, k, f) {
        const val = storage.globalMapGet(c, k, f);
        if (val === "") {
            return null;
        }
        return JSON.parse(val);
    }

    _updateRate() {
        // update rate every 7 days
        const lastTime = this._get("lastTime");
        if (block.time < lastTime + 604800) {
            return;
        }
        const supply = new Float64(blockchain.callWithAuth("token.iost", "supply", ["iost"])[0]);
        const blockContrib = supply.multi(blockContribRadio).toFixed(8);
        this._put("blockContrib", blockContrib);
    }

    // issueContribute to witness
    issueContribute(data) {
        if (!data || !data.parent || !Array.isArray(data.parent)
            || data.parent.length !== 2 || !data.parent[0]) {
            return;
        }
        this._requireAuth("base.iost", activePermission);
        this._updateRate();
        let witness = data.parent[0];
        const blockContrib = this._get("blockContrib");
        // get account name of the witness
        const acc = this._globalMapGet("vote_producer.iost", "producerKeyToId", witness);
        if (acc) {
            witness = acc;
        }
        blockchain.callWithAuth("token.iost", "issue", [
            "contribute",
            witness,
            blockContrib
        ]);
    }

    // exchangeIOST with contribute
    exchangeIOST(account, amount) {
        this._requireAuth(account, activePermission);

        const lastExchangeTime = this._get(account) || 0;
        const currentTime = block.time;
        if (currentTime - lastExchangeTime < 86400000000000) {
            throw new Error("last exchange less than one day.");
        }

        const contribute = blockchain.callWithAuth("token.iost", "balanceOf", [
            "contribute",
            account
        ])[0];
        amount = new Float64(amount);
        if (amount.isZero()) {
            amount = new Float64(contribute);
        }

        if (amount.lte("0") || amount.gt(contribute)) {
            throw new Error("invalid amount: negative or greater than contribute");
        }

        const totalBonus = new Float64(blockchain.callWithAuth("token.iost", "balanceOf", [
            "iost",
            blockchain.contractName()
        ])[0]);

        if (amount.gt(totalBonus)) {
            throw new Error("left bonus not enough, please wait");
        }

        this._put(account, currentTime, account);

        blockchain.callWithAuth("token.iost", "destroy", [
            "contribute",
            account,
            amount.toFixed()
        ]);

        blockchain.withdraw(account, amount.toFixed(), "");
    }
}

module.exports = BonusContract;
