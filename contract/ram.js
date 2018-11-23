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
        return "ram";
    }

    initContractName(contractName) {
        const bn = this._getBlockNumber();
        if(bn !== 0) {
            throw new Error("init out of genesis block");
        }
        this._put("contractName", contractName);
    }
    initAdmin(adminID) {
        const bn = this._getBlockNumber();
        if(bn !== 0) {
            throw new Error("init out of genesis block");
        }
        this._put("adminID", adminID);
    }
    _getContractName() {
        return this._get("contractName");
    }
    _getAdminID() {
        return this._get("adminID");
    }
    _changeLeftSpace(delta) {
        this._put("leftSpace", this._getLeftSpace() + delta);
    }
    _getLeftSpace() {
        if (this._get("leftSpace") === null) {
            throw new Error("no leftSpace key");
        }
        const result = this._get("leftSpace");
        if (result < 0) {
            throw new Error("internal error: leftSpace is negative");
        }
        return result;
    }
    _changeUsedSpace(delta) {
        this._put("usedSpace", this._getUsedSpace() + delta);
    }
    _getUsedSpace() {
        if (this._get("usedSpace") === null) {
            throw new Error("no usedSpace key");
        }
        const result = this._get("usedSpace");
        if (result < 0) {
            throw new Error("internal error: usedSpace is negative");
        }
        return result;
    }
    _getBalance() {
        if (this._get("balance") === null) {
            throw new Error("no balance key");
        }
        const result = this._get("balance");
        if (result < 0) {
            throw new Error("internal error: balance is negative");
        }
        return result;
    }
    _changeBalance(delta) {
        this._put("balance", this._getBalance() + delta);
    }


    // initialTotal: 128 * 1024 * 1024 * 1024
    // increaseInterval: 24 * 3600 / 3
    // increaseAmount: 188272539 = Math.round(64 * 1024 * 1024 * 1024 / 365)
    issue(initialTotal, increaseInterval, increaseAmount, reserve) {
        const bn = this._getBlockNumber();
        if(bn !== 0) {
            throw new Error("init out of genesis block");
        }
        const veryLarge = 100 * 64 * 1024 * 1024 * 1024;
        let data = [this._getTokenName(), this._getContractName(), veryLarge, {"decimal":0}];
        BlockChain.callWithAuth("token.iost", "create", JSON.stringify(data));
        data = [this._getTokenName(), this._getContractName(), (initialTotal).toString()];
        BlockChain.callWithAuth("token.iost", "issue", JSON.stringify(data));
        this._put("lastUpdateBlockTime", this._getBlockTime());
        this._put("increaseInterval", increaseInterval);
        this._put("increaseAmount", increaseAmount);
        this._put("leftSpace", initialTotal - reserve);
        this._put("balance", 0);
        this._put("usedSpace", reserve);
    }

    _price(action, amount) {
        const priceCoefficient = 30; // when RAM is empty, every KiB worth `priceCoefficient` IOST
        const leftSpace = this._getLeftSpace();
        if (action === "buy") {
            if (this._getBlockNumber() === 0) {
                return priceCoefficient * amount;
            }
            if (leftSpace <= amount) {
                throw new Error("buy amount is too much. left space is not enough " + leftSpace.toString() + " is less than " + amount.toString());
            }
            const price = Math.ceil(priceCoefficient * 128 * 1024 * 1024 * Math.log1p(amount / (leftSpace - amount)));
            return price;
        } else if (action === "sell") {
            //const price = Math.floor(priceCoefficient * 128 * 1024 * 1024 * Math.log1p(amount / leftSpace));
            const price = Math.floor(amount / this._getUsedSpace() * this._getBalance());
            return price;
        }
        throw new Error("invalid action");

    }

    _checkIssue() {
        const t = this._getBlockTime();
        const nextUpdateTime = this._get("lastUpdateBlockTime") + this._get("increaseInterval") * 1000 * 1000 * 1000;
        if (t < nextUpdateTime) {
            return;
        }
        const increaseAmount = this._get("increaseAmount");
        const data = [this._getTokenName(), this._getContractName(), increaseAmount.toString()];
        BlockChain.callWithAuth("token.iost", "issue", JSON.stringify(data));
        this._put("lastUpdateBlockTime", t);
        this._changeLeftSpace(increaseAmount);
    }

    buy(payer, account, amount) {
        this._requireAuth(payer, transferPermission);
        this._checkIssue();
        const rawPrice = this._price("buy", amount);
        const feeRate = 0.01;
        const fee = Math.ceil(feeRate * rawPrice);
        const price = rawPrice + fee;
        BlockChain.callWithAuth("token.iost", "transfer", JSON.stringify(["iost", payer, this._getContractName(), price.toString(), ""]));
        const data = [this._getTokenName(), this._getContractName(), account, (amount).toString(), ""];
        BlockChain.callWithAuth("token.iost", "transfer", JSON.stringify(data));
        this._changeLeftSpace(-amount);
        this._changeBalance(rawPrice);
        this._changeUsedSpace(amount);
        return price;
    }

    sell(account, receiver, amount) {
        this._requireAuth(account, transferPermission);
        const data = [this._getTokenName(), account, this._getContractName(), (amount).toString(), ""];
        BlockChain.callWithAuth("token.iost", "transfer", JSON.stringify(data));
        const price = this._price("sell", amount);
        BlockChain.callWithAuth("token.iost", "transfer", JSON.stringify(["iost", this._getContractName(), receiver, price.toString(), ""]));
        this._changeLeftSpace(amount);
        this._changeBalance(-price);
        this._changeUsedSpace(-amount);
        return price;
    }
}

module.exports = RAMContract;
