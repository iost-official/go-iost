const defaultPermission = "active";
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

    _get(k) {
        var raw = storage.get(k);
        if (raw == "nil") {
            return null;
        }
        return JSON.parse(raw);
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
        if (this._get("leftSpace") == null) {
            return 0
        }
        return this._get("_leftSpace");
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
        BlockChain.call("iost.token", "create", JSON.stringify(data));
        data = [this._getTokenName(), this._getContractName(), (initialTotal).toString()];
        BlockChain.call("iost.token", "issue", JSON.stringify(data));
        this._put("lastUpdateBlockNumber", bn);
        this._put("increaseInterval", increaseInterval);
        this._put("increaseAmount", increaseAmount);
    }

    _price(action, amount) {
        return amount * 1; // TODO not use log/exp, implement a price function
        /*
        const priceCoefficient = 1; // when RAM is empty, every KB worth `priceCoefficient` IOST
        const feeRate = 0.01;
        const leftSpace = this._getLeftSpace();
        if (action == "buy") {
            if (leftSpace <= amount) {
                throw new Error("buy amount is too much")
            }
            return Math.ceil((1 + feeRate) * priceCoefficient * 1024 * 1024 * 128 * Math.log1p(amount / (leftSpace - amount)))
        } else if (action == "sell") {
            return Math.floor(priceCoefficient * 1024 * 1024 * 128 * Math.log1p(amount / leftSpace))
        }
        throw new Error("invalid action")
        */
    }

    checkIssue() {
        const bn = this._getBlockNumber();
        if (bn < this._get("lastUpdateBlockNumber") + this._get("increaseInterval")) {
            return
        }
        const data = [this._getTokenName(), this._getContractName(), this._get("increaseAmount").toString()];
        let ret = BlockChain.call("iost.token", "issue", JSON.stringify(data));
        if (ret != 0) {
            throw "issue err " + ret
        }
        this._put("lastUpdateBlockNumber", bn);
    }

    buy(account, amount) {
        this._requireAuth(account, defaultPermission);
        this.checkIssue();
        const price = this._price("buy", amount);
        let ret = BlockChain.deposit(account, price.toString());
        if (ret != 0) {
            throw "deposit err " + ret
        }
        const data = [this._getTokenName(), this._getContractName(), account, (amount).toString()];
        ret = BlockChain.call("iost.token", "transfer", JSON.stringify(data));
        if (ret != "[]") {
            throw "transfer err " + ret
        }
        this._changeLeftSpace(-amount)
    }

    sell(account, amount) {
        this._requireAuth(account, defaultPermission);
        const data = [this._getTokenName(), account, this._getContractName(), (amount).toString()];
        let ret = BlockChain.call("iost.token", "transfer", JSON.stringify(data));
        if (ret != "[]") {
            throw "transfer err " + ret
        }
        const price = this._price("sell", amount);
        ret = BlockChain.withdraw(account, price.toString());
        if (ret != 0) {
            throw "withdraw err " + ret
        }
        this._changeLeftSpace(amount)
    }
}

module.exports = RAMContract;
