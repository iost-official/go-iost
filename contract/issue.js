const secondToNano = 1e9;
const iostIssueRate = new BigNumber("1.0000000028119105");
const activePermission = "active";
const iostTotalSupply = 90 * 1000 * 1000 * 1000;
const ramTotalSupply = 9 * 1000 * 1000 * 1000 * 1000 * 1000 * 1000;

class IssueContract {
    constructor() {
        BigNumber.config({
            DECIMAL_PLACES: 50,
            POW_PRECISION: 50,
            ROUNDING_MODE: BigNumber.ROUND_DOWN
        })
    }

    init() {
        this._put("foundationAcc", "")
    }

    _initIOST(config) {
        this._call("iost.token", "create", [
            "iost",
            "iost.issue",
            iostTotalSupply,
            {
                "can_transfer": true,
                "decimal": config.iostDecimal
            }
        ]);
        for (const witnessInfo of config.iostWitnessInfo) {
            this._call("iost.token", "issue", [
                "iost",
                witnessInfo.ID,
                new BigNumber(witnessInfo.Balance)
            ]);
        }
        this._put("IOSTDecimal", config.iostDecimal);
        this._put("IOSTLastIssueTime", this._getBlockTime());
    }

    _initRAM(config) {
        this._call("iost.token", "create", [
            "ram",
            "iost.issue",
            ramTotalSupply,
            {
                "can_transfer": false,
                "decimal": 0
            }
        ]);
        this._call("iost.token", "issue", [
            "ram",
            "iost.pledge",
            new BigNumber(config.ramGenesisAmount)
        ]);
        this._put("RAMLastIssueTime", this._getBlockTime());
    }

    /**
     * genesisConfig = {
     *      iostWitnessInfo: {
     *          ID: xxx,
     *          Wwner: xxx,
     *          Active: xxx,
     *          Balance: 210xxx
     *      },
     *      iostDecimal: 8,
     *      foundationAcc: "IOSTfQFocqDn7VrKV7vvPqhAQGyeFU9XMYo5SNn5yQbdbzC75wM7C",
     *      ramGenesisAmount: 128 * 1024 * 1024 * 1024,
     * }
     */
    InitGenesis(adminID, genesisConfig) {
        const bn = this._getBlockNumber();
        if(bn !== 0) {
            throw new Error("init out of genesis block")
        }
        this._put("adminID", adminID);
        this._put("foundationAcc", genesisConfig.foundationAcc);

        this._initIOST(genesisConfig);
        this._initRAM(genesisConfig);
    }

    can_update(data) {
        const admin = this._get("adminID");
        this._requireAuth(admin, activePermission);
        return true;
    }

    _requireAuth(account, permission) {
        BlockChain.requireAuth(account, permission);
    }

    _call(contract, api, args) {
        const ret = JSON.parse(BlockChain.callWithAuth(contract, api, JSON.stringify(args)));
        if (ret && Array.isArray(ret) && ret.length == 1) {
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

    // IssueIOST to iost.bonus and iost foundation
    IssueIOST() {
        const lastIssueTime = this._get("IOSTLastIssueTime");
        if (lastIssueTime === 0 || lastIssueTime === undefined) {
            throw new Error("IOSTLastIssueTime not set.");
        }
        const currentTime = this._getBlockTime();
        const gap = Math.floor((currentTime - lastIssueTime) / 3);
        if (gap <= 0) {
            return
        }

        const foundationAcc = this._get("foundationAcc");
        const decimal = this._get("IOSTDecimal");
        if (!foundationAcc) {
            throw new Error("foundationAcc not set.");
        }

        this._put("IOSTLastIssueTime", currentTime);

        const supply = new Float64(this._call("iost.token", "supply", ["iost"]));
        const issueAmount = supply.multi(iostIssueRate.pow(gap).minus(1));
        const bonus = issueAmount.multi(0.33);
        this._call("iost.token", "issue", [
            "iost",
            "iost.bonus",
            bonus.number.toFixed(decimal)
        ]);
        this._call("iost.token", "issue", [
            "iost",
            foundationAcc,
            issueAmount.minus(bonus).number.toFixed(decimal)
        ]);
    }

    // IssueRAM to iost.pledge
    IssueRAM() {
        // this._requireAuth("iost.pledge", activePermission);
        const lastIssueTime = this._get("RAMLastIssueTime");
        if (lastIssueTime === 0 || lastIssueTime === undefined) {
            throw new Error("RAMLastIssueTime not set.");
        }
        const currentTime = this._getBlockTime();
        const gap = currentTime - lastIssueTime;
        if (gap < 86400 /* one day */) {
            return;
        }
        this._put("RAMLastIssueTime", currentTime);
        const issueAmount = 2179 * gap;
        this._call("iost.token", "issue", [
            "ram",
            "iost.pledge",
            JSON.stringify(issueAmount)
        ]);
    }
}

module.exports = IssueContract;