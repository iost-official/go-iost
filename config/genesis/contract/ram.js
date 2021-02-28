const transferPermission = "transfer";
const updatePermission = "active";

class RAMContract {
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
        if (bn !== 0) {
            throw new Error("init out of genesis block");
        }
        const veryLarge = 100 * 64 * 1024 * 1024 * 1024;
        const tokenInfo = {"decimal": 0, "fullName": "IOST system ram", "onlyIssuerCanTransfer": true};
        let data = [this._getTokenName(), blockchain.contractName(), veryLarge, tokenInfo];
        blockchain.callWithAuth("token.iost", "create", data);
        data = [this._getTokenName(), blockchain.contractName(), (initialTotal).toString()];
        blockchain.callWithAuth("token.iost", "issue", data);
        this._put("lastUpdateBlockTime", block.time);
        this._put("increaseInterval", increaseInterval);
        this._put("increaseAmount", increaseAmount);
        this._put("leftSpace", initialTotal);
        this._put("usedSpace", 0);
        this._put("balance", initialTotal * this._getInitialPrice() / this._getF());
        if (reserve !== 0) {
            const reservePrice = this._price("buy", reserve);
            this._changeLeftSpace(-reserve);
            this._changeBalance(reservePrice);
            this._changeUsedSpace(reserve);
        }
    }
    _getInitialPrice() {
        return 0.005; // when RAM is empty, every byte worth `priceCoefficient` IOST
    }
    _getF() {
        return 1.0;
    }
    _price(action, amount) {
        const leftSpace = this._getLeftSpace();
        if (action === "buy") {
            if (leftSpace <= amount) {
                throw new Error("buy amount is too much. left space is not enough " + leftSpace.toString() + " is less than " + amount.toString());
            }
            return this._getBalance() * (Math.pow(leftSpace / (leftSpace - amount), this._getF()) - 1.0);
        } else if (action === "sell") {
            //const price = Math.floor(priceCoefficient * 128 * 1024 * 1024 * Math.log1p(amount / leftSpace));
            return -this._price("buy", -amount);
        }
        throw new Error("invalid action");

    }

    _round(f) {
        return Math.round(f * 100)/100;
    }

    _checkIssue() {
        const increaseInterval = this._get("increaseInterval");
        const slotNum = Math.floor(block.time/1e9/increaseInterval) - Math.floor(this._get("lastUpdateBlockTime")/1e9/increaseInterval);
        if (slotNum > 0) {
            const increaseAmount = this._get("increaseAmount") * slotNum;
            const data = [this._getTokenName(), blockchain.contractName(), increaseAmount.toString()];
            blockchain.callWithAuth("token.iost", "issue", data);
            this._put("lastUpdateBlockTime", block.time);
            this._changeLeftSpace(increaseAmount);
        }
    }

    _getAccountSelfRAM(acc) {
        const result = this._get("SR" + acc);
        if (result === null) {
            return 0
        }
        return result
    }
    _changeAccountSelfRAM(acc, delta) {
        this._put("SR" + acc, this._getAccountSelfRAM(acc) + delta)
    }
    _getAccountTotalRAM(acc) {
        const result = this._get("TR" + acc);
        if (result === null) {
            return 0
        }
        return result
    }
    _changeAccountTotalRAM(acc, delta) {
        this._put("TR" + acc, this._getAccountTotalRAM(acc) + delta)
    }

    _handle_fee(acc, fee) {
        let destroyAmount = fee;
        if (destroyAmount.toFixed(2) !== "0.00") {
            blockchain.callWithAuth("token.iost", "transfer", ["iost", blockchain.contractName(), "deadaddr", destroyAmount.toFixed(2), ""]);
        }
    }

    buy(payer, account, amount) {
        if (amount < 10) {
            throw new Error("minimum ram amount for trading is 10 byte");
        }
        this._requireAuth(payer, transferPermission);
        this._checkIssue();
        const rawPrice = this._round(this._price("buy", amount));
        const feeRate = 0.02;
        let fee = this._round(feeRate * rawPrice);
        if (fee < 0.01) {
            fee = 0.01;
        }
        const price = rawPrice + fee;
        blockchain.callWithAuth("token.iost", "transfer", ["iost", payer, blockchain.contractName(), price.toFixed(2), ""]);
        this._handle_fee(payer, fee);
        const data = [this._getTokenName(), blockchain.contractName(), account, amount.toString(), ""];
        blockchain.callWithAuth("token.iost", "transfer", data);
        this._changeLeftSpace(-amount);
        this._changeBalance(rawPrice);
        this._changeUsedSpace(amount);
        this._changeAccountSelfRAM(account, amount);
        this._changeAccountTotalRAM(account, amount);
        return price.toFixed(2);
    }

    sell(account, receiver, amount) {
        if (amount < 10) {
            throw new Error("minimum ram amount for trading is 10 byte");
        }
        this._requireAuth(account, transferPermission);
        if (this._getAccountSelfRAM(account) < amount) {
            throw new Error("self ram amount " + this._getAccountSelfRAM(account) + ", not enough for sell");
        }
        const data = [this._getTokenName(), account, blockchain.contractName(), amount.toString(), ""];
        blockchain.callWithAuth("token.iost", "transfer", data);
        const price = this._round(this._price("sell", amount));
        blockchain.callWithAuth("token.iost", "transfer", ["iost", blockchain.contractName(), receiver, price.toFixed(2), ""]);
        this._changeLeftSpace(amount);
        this._changeBalance(-price);
        this._changeUsedSpace(-amount);
        this._changeAccountSelfRAM(account, -amount);
        this._changeAccountTotalRAM(account, -amount);
        return price.toFixed(2);
    }

    lend(from, to, amount) {
        if (amount < 10) {
            throw new Error("minimum ram amount for trading is 10 byte");
        }
        this._requireAuth(from, transferPermission);
        if (this._getAccountSelfRAM(from) < amount) {
            throw new Error("self ram amount " + this._getAccountSelfRAM(from) + ", not enough for lend");
        }
        const data = [this._getTokenName(), from, to, amount.toString(), ""];
        blockchain.callWithAuth("token.iost", "transfer", data);
        this._changeAccountSelfRAM(from, -amount);
        this._changeAccountTotalRAM(from, -amount);
        this._changeAccountTotalRAM(to, amount);
    }
}

module.exports = RAMContract;
