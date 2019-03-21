const activePermission = "active";
const totalSupply = 90000000000;
const blockContribRatio = new Float64("1.564349722223e-10");

class BonusContract {
    init() {
        this._initContribute();
        this._put("blockContrib", "3.28513441");
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
        if (block.time < lastTime + 604800000000000) {
            return;
        }
        const supply = new Float64(blockchain.callWithAuth("token.iost", "supply", ["iost"])[0]);
        const blockContrib = supply.multi(blockContribRatio).toFixed(8);
        this._put("blockContrib", blockContrib);
        this._put("lastTime", block.time);
    }

    // issueContribute to witness
    issueContribute(data) {
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

        blockchain.callWithAuth("token.iost", "destroy", [
            "contribute",
            account,
            amount.toFixed()
        ]);

        blockchain.withdraw(account, amount.toFixed(), "");
    }
}

module.exports = BonusContract;
