const secondToNano = 1e9;
const iostIssueRate = new BigNumber("1.0000000028119105");
const activePermission = "active";

class IssueContract {
    constructor() {
    }

    init() {
        this._put("FoundationAccount", "");
    }

    _initIOST(config, witnessInfo) {
        this._call("token.iost", "create", [
            "iost",
            "issue.iost",
            config.IOSTTotalSupply,
            {
                "can_transfer": true,
                "decimal": config.IOSTDecimal
            }
        ]);
        for (const info of witnessInfo) {
            this._call("token.iost", "issue", [
                "iost",
                info.ID,
                new BigNumber(info.Balance).toFixed()
            ]);
        }
        this._put("IOSTDecimal", config.IOSTDecimal);
        this._put("IOSTLastIssueTime", this._getBlockTime());
    }

    /**
     * genesisConfig = {
     *      FoundationAccount string
     *      IOSTTotalSupply   int64
     *      IOSTDecimal       int64
     * }
     * witnessInfo = [{
     *      ID      string
     *      Owner   string
     *      Active  string
     *      Balance int64
     * }]
     */
    InitGenesis(adminID, genesisConfig, witnessInfo) {
        const bn = block.number;
        if(bn !== 0) {
            throw new Error("init out of genesis block")
        }
        this._put("adminID", adminID);
        this._put("FoundationAccount", genesisConfig.FoundationAccount);

        this._initIOST(genesisConfig, witnessInfo);
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

    _getBlockTime() {
        return Math.floor(block.time / secondToNano);
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

    _issueIOST(account, amount) {
        const amountStr = ((typeof amount === "string") ? amount : amount.toFixed(this._get("IOSTDecimal")));
        const args = ["iost", account, amountStr];
        console.log("issueiost", args)
        this._call("token.iost", "issue", args);
    }

    IssueIOSTTo(account, amount) {
        const whitelist = ["auth.iost"];
        let auth = false;
        for (const c of whitelist) {
            if (blockchain.requireAuth(c, "active")) {
                auth = true;
                break
            }
        }
        if (!auth) {
            throw new Error("issue iost permission denied")
        }
        this._issueIOST(account, amount)
    }

    // IssueIOST to bonus.iost and iost foundation
    IssueIOST() {
        const lastIssueTime = this._get("IOSTLastIssueTime");
        if (lastIssueTime === 0 || lastIssueTime === undefined) {
            throw new Error("IOSTLastIssueTime not set.");
        }
        const currentTime = this._getBlockTime();
        const gap = Math.floor((currentTime - lastIssueTime) / 3);
        if (gap <= 0) {
            return;
        }

        const foundationAcc = this._get("FoundationAccount");
        if (!foundationAcc) {
            throw new Error("FoundationAccount not set.");
        }

        this._put("IOSTLastIssueTime", currentTime);

        const supply = new Float64(this._call("token.iost", "supply", ["iost"]));
        const issueAmount = supply.multi(iostIssueRate.pow(gap).minus(1));
        const bonus = issueAmount.multi(0.33);
        this._issueIOST("bonus.iost", bonus);
        this._issueIOST(foundationAcc, issueAmount.minus(bonus));
    }
}

module.exports = IssueContract;
