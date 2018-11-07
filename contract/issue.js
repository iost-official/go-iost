const slotLength = 3;
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
                "default_rate": 1,
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
                "default_rate": 1,
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
        return JSON.parse(storage.get(k));
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

    // IssueIOST to iost.bonus and iost foundation
    IssueIOST() {
        const lastIssueTime = this._get("IOSTLastIssueTime");
        if (lastIssueTime === 0 || lastIssueTime === undefined) {
            throw new Error("IOSTLastIssueTime not set.");
        }
        const currentTime = this._getBlockTime();
        const gap = currentTime - lastIssueTime;
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
        if (gap * slotLength < 86400 /* one day */) {
            return;
        }
        this._put("RAMLastIssueTime", currentTime);
        const issueAmount = 6538 * gap;
        this._call("iost.token", "issue", [
            "ram",
            "iost.pledge",
            JSON.stringify(issueAmount)
        ]);
    }
}

module.exports = IssueContract;