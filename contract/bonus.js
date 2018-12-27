const secondToNano = 1e9;
const activePermission = "active";
const totalSupply = 9 * 1000 * 1000 * 1000 * 1000 * 1000 * 1000;
const blockContribRadio = new Float64("9.6568764571e-11");

class BonusContract {
    constructor() {
    }

    init() {
        this._initContribute();
        this._put("blockContrib", "1.98779440");
        this._put("lastTime", block.time);
    }

    _initContribute() {
        this._call("token.iost", "create", [
            "contribute",
            "bonus.iost",
            totalSupply,
            {
                "can_transfer": false,
                "decimal": 0
            }
        ]);
    }

    InitAdmin(adminID) {
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

    _call(contract, api, args) {
        const ret = blockchain.callWithAuth(contract, api, JSON.stringify(args));
        if (ret && Array.isArray(ret) && ret.length === 1) {
            return ret[0] === "" ? "" : JSON.parse(ret[0]);
        }
        return ret;
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
        const supply = new Float64(this._call("token.iost", "supply", ["iost"]));
        const blockContrib = supply.multi(blockContribRadio).toFixed(8);
        this._put("blockContrib", blockContrib);
    }

    // IssueContribute to witness
    IssueContribute(data) {
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
        this._call("token.iost", "issue", [
            "contribute",
            witness,
            blockContrib
        ]);
    }

    // ExchangeIOST with contribute
    ExchangeIOST(account, amount) {
        this._requireAuth(account, activePermission);

        const lastExchangeTime = this._get(account) || 0;
        const currentTime = block.time;
        if (currentTime - lastExchangeTime < 86400000000000) {
            throw new Error("last exchange less than one day.");
        }

        const contribute = this._call("token.iost", "balanceOf", [
            "contribute",
            account
        ]);
        amount = new Float64(amount);
        if (amount.isZero()) {
            amount = new Float64(contribute);
        }

        if (!amount.isPositive() || amount.gt(contribute)) {
            throw new Error("invalid amount: negative or greater than contribute");
        }

        const totalBonus = new Float64(this._call("token.iost", "balanceOf", [
            "iost",
            blockchain.contractName()
        ]));

        if (amount.gt(totalBonus)) {
            throw new Error("left bonus not enough, please wait");
        }

        this._put(account, currentTime, account);

        this._call("token.iost", "destroy", [
            "contribute",
            account,
            amount.toFixed()
        ]);
        const voterBonus = amount.div(2);

        blockchain.withdraw(account, amount.minus(voterBonus).toFixed(), "");
        const succ = this._call("vote_producer.iost", "TopupVoterBonus", [
            account,
            voterBonus.toFixed(),
            blockchain.contractName()
        ]);
        if (!succ) {
            // transfer voteBonus to account if topup failed
            blockchain.withdraw(account, voterBonus.toFixed(), "");
        }
    }
}

module.exports = BonusContract;
