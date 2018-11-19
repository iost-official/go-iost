const transferPermission = "transfer";
const updatePermission = "active";

class RAMContract {
    constructor() {
    }

    init() {
    }

    _requireAuth(account, permission) {
        const ret = BlockChain.requireAuth(account, permission);
        if (ret !== true) {
            throw new Error("require auth failed. ret = " + ret);
        }
    }

    can_update(data) {
        const admin = this._get("adminID");
        this._requireAuth(admin, updatePermission);
        return true;
    }

    _getBlockNumber() {
        const bi = JSON.parse(BlockChain.blockInfo());
        if (!bi || bi === undefined || bi.number === undefined) {
            throw new Error("get block number failed. bi = " + bi);
        }
        return bi.number;
    }

    _getBlockTime() {
        const bi = JSON.parse(BlockChain.blockInfo());
        if (!bi || bi === undefined || bi.number === undefined) {
            throw new Error("get block time failed. bi = " + bi);
        }
        return bi.time;
    }

    _get(k) {
        var raw = storage.get(k);
        if (raw === null || raw === "") {
            return null;
        }
        return JSON.parse(raw);
    }
    _put(k, v) {
        storage.put(k, JSON.stringify(v));
    }
    _mapGet(k, f) {
        return JSON.parse(storage.mapGet(k, f));
    }
    _mapPut(k, f, v) {
        storage.mapPut(k, f, JSON.stringify(v));
    }

    _mapDel(k, f) {
        const ret = storage.mapDel(k, f);
        if (ret !== 0) {
            throw new Error("storage map del failed. ret = " + ret);
        }
    }

    _getTokenName() {
        return "ram"
    }

    initContractName(contractName) {
        const bn = this._getBlockNumber();
        if(bn !== 0) {
            throw new Error("init out of genesis block")
        }
        this._put("contractName", contractName);
    }
    initAdmin(adminID) {
        const bn = this._getBlockNumber();
        if(bn !== 0) {
            throw new Error("init out of genesis block")
        }
        this._put("adminID", adminID);
    }
    _getContractName() {
        return this._get("contractName")
    }
    _getAdminID() {
        return this._get("adminID")
    }

    _getLeftSpace() {
        if (this._get("leftSpace") === null) {
            throw "no leftSpace key";
        }
        return this._get("leftSpace");
    }
    _changeLeftSpace(delta) {
        this._put("leftSpace", this._getLeftSpace() + delta);
    }


    // initialTotal: 128 * 1024 * 1024 * 1024
    // increaseInterval: 24 * 3600 / 3
    // increaseAmount: 188272539 = Math.round(64 * 1024 * 1024 * 1024 / 365)
    issue(initialTotal, increaseInterval, increaseAmount) {
        const bn = this._getBlockNumber();
        if(bn !== 0) {
            throw new Error("init out of genesis block")
        }
        var veryLarge = 100 * 64 * 1024 * 1024 * 1024;
        let data = [this._getTokenName(), this._getContractName(), veryLarge, {"decimal":0}];
        BlockChain.callWithAuth("iost.token", "create", JSON.stringify(data));
        data = [this._getTokenName(), this._getContractName(), (initialTotal).toString()];
        BlockChain.callWithAuth("iost.token", "issue", JSON.stringify(data));
        this._put("lastUpdateBlockTime", this._getBlockTime());
        this._put("increaseInterval", increaseInterval);
        this._put("increaseAmount", increaseAmount);
        this._put("leftSpace", initialTotal);
    }

    _price(action, amount) {
        //return amount * 1; // TODO not use log/exp, implement a price function
        const priceCoefficient = 1024; // when RAM is empty, every KiB worth `priceCoefficient` IOST
        const feeRate = 0.01;
        const leftSpace = this._getLeftSpace();
        if (action === "buy") {
            if (this._getBlockNumber() === 0) {
                return priceCoefficient * amount;
            }
            if (leftSpace <= amount) {
                throw new Error("buy amount is too much. left space is not enough " + leftSpace.toString() + " is less than " + amount.toString())
            }
            return Math.ceil((1 + feeRate) * priceCoefficient * 1024 * 1024 * 128 * Math.log1p(amount / (leftSpace - amount)))
        } else if (action === "sell") {
            return Math.floor(priceCoefficient * 1024 * 1024 * 128 * Math.log1p(amount / leftSpace))
        }
        throw new Error("invalid action")

    }

    _checkIssue() {
        const t = this._getBlockTime();
        const nextUpdateTime = this._get("lastUpdateBlockTime") + this._get("increaseInterval") * 1000 * 1000 * 1000;
        if (t < nextUpdateTime) {
            return
        }
        const increaseAmount = this._get("increaseAmount");
        const data = [this._getTokenName(), this._getContractName(), increaseAmount.toString()];
        BlockChain.callWithAuth("iost.token", "issue", JSON.stringify(data));
        this._put("lastUpdateBlockTime", t);
        this._changeLeftSpace(increaseAmount)
    }

    buy(payer, account, amount) {
        this._requireAuth(payer, transferPermission);
        this._checkIssue();
        const price = this._price("buy", amount);
        BlockChain.callWithAuth("iost.token", "transfer", JSON.stringify(["iost", payer, this._getContractName(), price.toString(), ""]));
        const data = [this._getTokenName(), this._getContractName(), account, (amount).toString(), ""];
        BlockChain.callWithAuth("iost.token", "transfer", JSON.stringify(data));
        this._changeLeftSpace(-amount)
    }

    sell(account, receiver, amount) {
        this._requireAuth(account, transferPermission);
        const data = [this._getTokenName(), account, this._getContractName(), (amount).toString(), ""];
        BlockChain.callWithAuth("iost.token", "transfer", JSON.stringify(data));
        const price = this._price("sell", amount);
        BlockChain.callWithAuth("iost.token", "transfer", JSON.stringify(["iost", this._getContractName(), receiver, price.toString(), ""]));
        this._changeLeftSpace(amount)
    }
}

module.exports = RAMContract;
