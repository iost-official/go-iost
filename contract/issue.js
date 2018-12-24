const iostIssueRate = new Float64("0.0296");
const oneYearNano = new Float64("31536000000000000");
const activePermission = "active";

class IssueContract {
    init() {
        storage.put("FoundationAccount", "");
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
        storage.put("IOSTDecimal", new Int64(config.IOSTDecimal).toFixed());
        storage.put("IOSTLastIssueTime", this._getBlockTime().toFixed());
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
        storage.put("adminID", adminID);
        storage.put("FoundationAccount", genesisConfig.FoundationAccount);

        this._initIOST(genesisConfig, witnessInfo);
    }

    can_update(data) {
        const admin = storage.get("adminID");
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
        return new Float64(block.time);
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

    // IssueIOST to bonus.iost and iost foundation
    IssueIOST() {
        const lastIssueTime = storage.get("IOSTLastIssueTime");
        if (lastIssueTime === null || lastIssueTime === 0 || lastIssueTime === undefined) {
            throw new Error("IOSTLastIssueTime not set.");
        }
        const currentTime = this._getBlockTime();
        const gap = currentTime.minus(lastIssueTime);
        if (gap.lte(0)) {
            return;
        }

        const foundationAcc = storage.get("FoundationAccount");
        const decimal = JSON.parse(storage.get("IOSTDecimal"));
        if (!foundationAcc) {
            throw new Error("FoundationAccount not set.");
        }

        storage.put("IOSTLastIssueTime", currentTime.toFixed());

        const supply = new Float64(this._call("token.iost", "supply", ["iost"]));
        const issueAmount = supply.multi(iostIssueRate).multi(gap).div(oneYearNano);
        const bonus = issueAmount.multi("0.66");
        this._call("token.iost", "issue", [
            "iost",
            blockchain.contractName(),
            bonus.toFixed(decimal)
        ]);
        this._call("token.iost", "issue", [
            "iost",
            foundationAcc,
            issueAmount.minus(bonus).minus(bonus).toFixed(decimal)
        ]);
        const producerBonus = bonus.div("2");
        this._call("bonus.iost", "Topup", [
            producerBonus.toFixed(decimal),             // producer bonus
            bonus.minus(producerBonus).toFixed(decimal) // vote bonus
        ]);
    }
}

module.exports = IssueContract;
