const secondToNano = 1e9;
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
        this._call("token.iost", "create", [
            "contribute",
            "bonus.iost",
            totalSupply,
            {
                "can_transfer": false,
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
        const ret = BlockChain.callWithAuth(contract, api, JSON.stringify(args));
        if (ret && Array.isArray(ret) && ret.length === 1) {
            return ret[0] === "" ? "" : JSON.parse(ret[0]);
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
        return Math.floor(this._getBlockInfo().time / secondToNano);
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

    _mapGet(k, f) {
        const val = storage.mapGet(k, f);
        if (val === "") {
            return null;
        }
        return JSON.parse(val);
    }

    _mapPut(k, f, v) {
        storage.mapPut(k, f, JSON.stringify(v));
    }

    _mapDel(k, f) {
        storage.mapDel(k, f);
    }

    // IssueContribute to witness
    IssueContribute(data) {
        if (!data || !data.parent || !Array.isArray(data.parent)
            || data.parent.length !== 2 || !data.parent[0]) {
            return;
        }
        // TODO: change tests to enable requireAuth
        // this._requireAuth("base.iost", activePermission);
        const bi = this._getBlockInfo();

        const lastIssueBN = this._get("lastIssueBN");
        if (lastIssueBN === undefined) {
            throw new Error("lastIssueBN not set.");
        }
        if (bi.number === lastIssueBN) {
            throw new Error("contribute issued twice in this block.");
        }
        this._put("lastIssueBN", bi.number);

        const witness = data.parent[0];
        let gasUsage = new BigNumber(data.parent[1]);
        if (!gasUsage.isFinite()) {
            gasUsage = new BigNumber(0);
        }
        let blockContrib = new BigNumber("900");
        if (gasUsage.lte(1e8)) {
            blockContrib = blockContrib.plus(gasUsage.div(1e6));
        } else {
            blockContrib = new BigNumber("1000");
        }
        this._call("token.iost", "issue", [
            "contribute",
            witness,
            blockContrib.toFixed(0)
        ]);
    }

    // ExchangeIOST with contribute
    ExchangeIOST(account, amount) {
        this._requireAuth(account, activePermission);

        const lastExchangeTime = this._get(account) || 0;
        const currentTime = this._getBlockTime();
        if (currentTime - lastExchangeTime < 86400) {
            throw new Error("last exchange less than one day.");
        }

        const ownContribute = this._call("token.iost", "balanceOf", [
            "contribute",
            account
        ]);
        amount = new BigNumber(amount);

        if (amount.gt(ownContribute)) {
            throw new Error("contribute not enough. left contribute = " + ownContribute);
        }

        this._put(account, currentTime);

        this._call("issue.iost", "IssueIOST", []);

        const totalBonus = new BigNumber(this._call("token.iost", "balanceOf", [
            "iost",
            "bonus.iost"
        ]));

        const totalContribute = new BigNumber(this._call("token.iost", "supply", ["contribute"]));
        const bonus = totalBonus.times(amount).div(totalContribute);

        this._call("token.iost", "destroy", [
            "contribute",
            account,
            amount.toFixed()
        ]);
        this._call("token.iost", "transfer", [
            "iost",
            "bonus.iost",
            account,
            bonus.toFixed(),
            ""
        ]);
    }
}

module.exports = BonusContract;
