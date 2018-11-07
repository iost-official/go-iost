const slotLength = 3;
const activePermission = "active";
const totalSupply = 9 * 1000 * 1000 * 1000 * 1000 * 1000 * 1000;

class BonusContract {
    constructor() {
        BigNumber.config({
            DECIMAL_PLACES:50,
            POW_PRECISION: 50,
            ROUNDING_MODE: BigNumber.ROUND_DOWN
        })
    }

    init() {
        this._initContribute();
    }

    _initContribute() {
        this._call("iost.token", "create", [
            "contribute",
            "iost.bonus",
            totalSupply,
            {
                "can_transfer": false,
                "default_rate": 1,
                "decimal": 0
            }
        ]);
        this._put("lastIssueBN", 0);
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
        this._requireAuth(admin, activePermission);
        return true;
    }

    _requireAuth(account, permission) {
        const ret = BlockChain.requireAuth(account, permission);
        if (ret !== true) {
            throw new Error("require auth failed. ret = " + ret);
        }
    }

    _call(contract, api, args) {
        const ret = JSON.parse(BlockChain.call(contract, api, JSON.stringify(args)));
        if (ret && Array.isArray(ret) && ret.length == 1) {
            return JSON.parse(ret[0]);
        }
        return ret;
    }

    _getBlockInfo() {
        const bi = JSON.parse(BlockChain.blockInfo());
        if (!bi || bi === undefined) {
            throw new Error("get block info failed. bi = " + bi);
        }
        return bi;
    }

    _getBlockNumber() {
        return this._getBlockInfo().number;
    }

    _getBlockTime() {
        return this._getBlockInfo().time;
    }

    _get(k) {
        const val = storage.get(k);
        if (val === "nil") {
            return null;
        }
        return JSON.parse(val);
    }

    _put(k, v) {
        const ret = storage.put(k, JSON.stringify(v));
        if (ret !== 0) {
            throw new Error("storage put failed. ret = " + ret);
        }
    }

    _mapGet(k, f) {
        return JSON.parse(storage.mapGet(k, f));
    }

    _mapPut(k, f, v) {
        const ret = storage.mapPut(k, f, JSON.stringify(v));
        if (ret !== 0) {
            throw new Error("storage map put failed. ret = " + ret);
        }
    }

    _mapDel(k, f) {
        const ret = storage.mapDel(k, f);
        if (ret !== 0) {
            throw new Error("storage map del failed. ret = " + ret);
        }
    }

    // IssueContribute to witness
    IssueContribute(gasUsed) {
        const bi = this._getBlockInfo();
        this._requireAuth(bi.witness, activePermission);

        const lastIssueBN = this._get("lastIssueBN");
        if (lastIssueBN === undefined) {
            throw new Error("lastIssueBN not set.");
        }
        if (bi.number == lastIssueBN) {
            throw new Error("contribute issued twice in this block.");
        }
        this._put("lastIssueBN", bi.number);

        gasUsed = new BigNumber(gasUsed);
        if (!gasUsed.isFinite()) {
            gasUsed = new BigNumber(0);
        }
        let blockContrib = new BigNumber("900");
        if (gasUsed.lte(1e8)) {
            blockContrib = blockContrib.plus(gasUsed.div(1e6));
        } else {
            blockContrib = new BigNumber("1000");
        }
        this._call("iost.token", "issue", [
            "contribute",
            bi.witness,
            blockContrib.toFixed(0)
        ]);

    }

    // ExchangeIOST with contribute
    ExchangeIOST(account, amount) {
        this._requireAuth(account, activePermission);

        const lastExchangeTime = this._get(account) || 0;
        const currentTime = this._getBlockTime();
        if (lastExchangeTime !== undefined
            && slotLength * (currentTime - lastExchangeTime) < 86400) {
            throw new Error("last exchange less than one day.");
        }

        const ownContribute = this._call("iost.token", "balanceOf", [
            "contribute",
            account
        ]);
        amount = new BigNumber(amount);

        if (amount.gt(ownContribute)) {
            throw new Error("contribute not enough. left contribute = " + ownContribute);
        }

        this._put(account, currentTime);

        this._call("iost.issue", "issueIOST", []);

        const totalBonus = new BigNumber(this._call("iost.token", "balanceOf", [
            "iost",
            "iost.bonus"
        ]));

        const totalContribute = new BigNumber(this._call("iost.token", "supply", ["contribute"]));
        const bonus = totalBonus.times(amount).div(totalContribute);

        this._call("iost.token", "destroy", [
            "contribute",
            account,
            amount.toFixed()
        ]);
        this._call("iost.token", "transfer", [
            "iost",
            "iost.bonus",
            account,
            bonus.toFixed()
        ]);
    }
}

module.exports = BonusContract;