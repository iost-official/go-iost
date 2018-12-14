const transferPermission = "transfer";
const updatePermission = "active";

class RAMContract {
    constructor() {
    }

    init() {
    }

    _requireAuth(account, permission) {
        const ret = blockchain.requireAuth(account, permission);
        if (ret !== true) {
            throw new Error("require auth failed. ret = " + ret);
        }
    }

    can_update(data) {
        const admin = this._get("adminID");
        this._requireAuth(admin, updatePermission);
        return true;
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

    initAdmin(adminID) {
        const bn = block.number;
        if(bn !== 0) {
            throw new Error("init out of genesis block");
        }
        this._put("adminID", adminID);
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
        const bn = block.number;
        if(bn !== 0) {
            throw new Error("init out of genesis block");
        }
        const veryLarge = 100 * 64 * 1024 * 1024 * 1024;
        const tokenInfo = {"decimal":0, "fullName":"IOST system ram"}
        let data = [this._getTokenName(), blockchain.contractName(), veryLarge, tokenInfo];
        blockchain.callWithAuth("token.iost", "create", JSON.stringify(data));
        data = [this._getTokenName(), blockchain.contractName(), (initialTotal).toString()];
        blockchain.callWithAuth("token.iost", "issue", JSON.stringify(data));
        this._put("lastUpdateBlockTime", block.time);
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
            if (block.number === 0) {
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
        const t = block.time;
        const nextUpdateTime = this._get("lastUpdateBlockTime") + this._get("increaseInterval") * 1000 * 1000 * 1000;
        if (t < nextUpdateTime) {
            return;
        }
        const increaseAmount = this._get("increaseAmount");
        const data = [this._getTokenName(), blockchain.contractName(), increaseAmount.toString()];
        blockchain.callWithAuth("token.iost", "issue", JSON.stringify(data));
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
        blockchain.callWithAuth("token.iost", "transfer", JSON.stringify(["iost", payer, blockchain.contractName(), price.toString(), ""]));
        const data = [this._getTokenName(), blockchain.contractName(), account, (amount).toString(), ""];
        blockchain.callWithAuth("token.iost", "transfer", JSON.stringify(data));
        this._changeLeftSpace(-amount);
        this._changeBalance(rawPrice);
        this._changeUsedSpace(amount);
        return price;
    }

    sell(account, receiver, amount) {
        this._requireAuth(account, transferPermission);
        const data = [this._getTokenName(), account, blockchain.contractName(), (amount).toString(), ""];
        blockchain.callWithAuth("token.iost", "transfer", JSON.stringify(data));
        const price = this._price("sell", amount);
        blockchain.callWithAuth("token.iost", "transfer", JSON.stringify(["iost", blockchain.contractName(), receiver, price.toString(), ""]));
        this._changeLeftSpace(amount);
        this._changeBalance(-price);
        this._changeUsedSpace(-amount);
        return price;
    }
}

module.exports = RAMContract;
