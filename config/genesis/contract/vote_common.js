const newVoteFee = "1000";
const descriptionMaxLength = 65536;
const optionMaxNum = 65536;
const resultMaxLength = 2048;
const iostDecimal = 8;

const adminPermission = "active";
const votePermission = "active";

const optionPrefix = "v_";
const userVotePrefix = "u_";

const TRUE = 1;
const FALSE = 0;

class VoteCommonContract {
    init() {
        storage.put("current_id", "0");
    }

    initAdmin(adminID) {
        const bn = block.number;
        if(bn !== 0) {
            throw new Error("init out of genesis block");
        }
        storage.put("adminID", adminID);
        this._put("fundID", [adminID]);
    }

    can_update(data) {
        const admin = storage.get("adminID");
        this._requireAuth(admin, adminPermission);
        return true;
    }

    _requireAuth(account, permission) {
        const ret = blockchain.requireAuth(account, permission);
        if (ret !== true) {
            throw new Error("require auth failed. ret = " + ret);
        }
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

    _nextId() {
        const id = new Int64(storage.get("current_id")).plus(1).toString();
        storage.put("current_id", id);
        return id;
    }

    _checkVote(voteId) {
        if (!storage.mapHas("voteInfo", voteId)) {
            throw new Error("vote does not exist.");
        }
    }

    _requireOwner(voteId) {
        this._checkVote(voteId);
        if (!storage.mapHas("owner", voteId)) {
            throw new Error("check owner failed.");
        }
        const owner = this._mapGet("owner", voteId);
        this._requireAuth(owner, adminPermission);
        return owner;
    }

    _delVote(voteId, info) {
        info.deleted = TRUE;
        this._mapPut("voteInfo", voteId, info);
        storage.mapDel("preResult", voteId);
        storage.mapDel("preOrder", voteId);
    }

    _checkDel(voteId) {
        const info = this._mapGet("voteInfo", voteId);
        if (info.deleted === TRUE) {
            throw new Error("vote has been deleted.");
        }
    }

    _addToPreResult(info, voteId, option, votes) {
        const preResult = this._mapGet("preResult", voteId);
        if (preResult === null || info === null) {
            return;
        }
        let preOrder = this._mapGet("preOrder", voteId) || [];
        preResult[option] = votes;

        const idx = preOrder.indexOf(option);
        if (idx !== -1) {
            preOrder.splice(idx, 1);
        }
        const v = new Float64(votes);
        const findPos = function(start, end) {
            if (end - start <= 1) {
                if (start === end) {
                    return start;
                }
                if (v.gt(preResult[preOrder[start]])) {
                    return start;
                } else {
                    return end;
                }
            }
            if (v.lte(preResult[preOrder[end-1]])) {
                return end;
            }
            const m = Math.floor((start + end) / 2);
            if (v.gt(preResult[preOrder[m]])) {
                return findPos(start, m);
            } else {
                return findPos(m, end);
            }
        }

        const pos = findPos(0, preOrder.length);
        preOrder.splice(pos, 0, option);

        if (preOrder.length > info.resultNumber) {
            const deleted = preOrder.splice(info.resultNumber, preOrder.length - info.resultNumber);
            for (const o of deleted) {
                delete(preResult[o]);
            }
        }
        this._mapPut("preResult", voteId, preResult);
        this._mapPut("preOrder", voteId, preOrder);
    }

    _removeFromPreResult(voteId, option, preResult) {
        if (preResult === undefined) {
            preResult = this._mapGet("preResult", voteId);
        }
        if (preResult && preResult.hasOwnProperty(option)) {
            delete(preResult[option]);
            this._mapPut("preResult", voteId, preResult);

            let preOrder = this._mapGet("preOrder", voteId) || [];
            const idx = preOrder.indexOf(option);
            if (idx !== -1) {
                preOrder.splice(idx, 1);
                this._mapPut("preOrder", voteId, preOrder);
            }
        }
    }

    newVote(owner, description, info) {
        this._requireAuth(owner, adminPermission);

        if (description.length > descriptionMaxLength) {
            throw new Error("description too long. max length is " + descriptionMaxLength + " Byte.");
        }

        const resultNumber = info.resultNumber;
        if (!resultNumber || !Number.isInteger(resultNumber) || resultNumber <= 0 || resultNumber > 2000) {
            throw new Error("resultNumber not valid.");
        }

        const minVote = info.minVote;
        if (!minVote || new Float64(minVote).lte(0)) {
            throw new Error("minVote not valid.");
        }

        const options = info.options;
        if (!options || !Array.isArray(options) || options.length > optionMaxNum) {
            throw new Error("options not valid.");
        }

        for (const option of options) {
            if (typeof option !== "string") {
                throw new Error("options has none string value.");
            }
            if (option.length > resultMaxLength) {
                throw new Error("option too long. max length is " + resultMaxLength + " Byte.");
            }
        }

        const anyOption = !!info.anyOption; // default to false
        if (anyOption) {
            throw new Error("anyOption not implemented");
        }

        let freezeTime = info.freezeTime;
        if (freezeTime === undefined) {
            freezeTime = 0;
        } else if (!Number.isInteger(freezeTime) || freezeTime < 0) {
            throw new Error("freezeTime not valid.");
        }

        const bn = block.number;

        if (bn > 0) {
            blockchain.callWithAuth("token.iost", "transfer", ["iost", owner, "vote.iost", newVoteFee, ""]);
        }

        const voteId = this._nextId();
        const voteInfo = {
            deleted: FALSE,
            description: description,
            resultNumber: resultNumber,
            minVote: minVote,
            anyOption: anyOption,
            freezeTime: freezeTime,
            deposit: bn > 0 ? newVoteFee : "0",
            optionNum: options.length,
        };
        for (const option of options) {
            const optionInfo = {
                votes: "0",
                deleted: FALSE,
                clearTime: -1
            }
            this._mapPut(optionPrefix + voteId, option, optionInfo, owner);
        }

        this._mapPut("voteInfo", voteId, voteInfo, owner);
        this._mapPut("owner", voteId, owner, owner);
        this._mapPut("preResult", voteId, {}); // pay by system

        return voteId;
    }

    addOption(voteId, option, clearVote) {
        const owner = this._requireOwner(voteId);
        this._checkDel(voteId);

        if (option.length > resultMaxLength) {
            throw new Error("option too long. max length is 1024 Byte.");
        }

        const info = this._mapGet("voteInfo", voteId);
        if (info.optionNum >= optionMaxNum) {
            throw new Error("options is full.");
        }

        if (storage.mapHas(optionPrefix + voteId, option)) {
            const optionInfo = this._mapGet(optionPrefix + voteId, option);
            if (optionInfo.deleted === FALSE) {
                throw new Error("option already exist.");
            }
            optionInfo.deleted = FALSE;
            if (clearVote === true) {
                optionInfo.clearTime = block.number;
                optionInfo.votes = "0";
            } else {
                const votes = new Float64(optionInfo.votes);
                if (votes.gte(info.minVote)) {
                    this._addToPreResult(info, voteId, option, optionInfo.votes);
                }
            }
            this._mapPut(optionPrefix + voteId, option, optionInfo, owner);
        } else {
            const optionInfo = {
                votes: "0",
                deleted: FALSE,
                clearTime: -1
            }
            this._mapPut(optionPrefix + voteId, option, optionInfo, owner);
        }
        info.optionNum++;
        this._mapPut("voteInfo", voteId, info, owner);
    }

    removeOption(voteId, option, force) {
        const owner = this._requireOwner(voteId);
        this._checkVote(voteId);

        if (!storage.mapHas(optionPrefix + voteId, option)) {
            throw new Error("option not exist");
        }

        const info = this._mapGet("voteInfo", voteId);
        const preResult = this._mapGet("preResult", voteId);
        if (!force && preResult && preResult[option] !== undefined) {
            let order = 0;
            const votes = new Float64(preResult[option]);
            for (const o in preResult) {
                if (o === option) {
                    continue;
                }
                if (votes.lt(preResult[o])) {
                    order++;
                }
                if (order >= info.resultNumber) {
                    break;
                }
            }
            if (order < info.resultNumber) {
                throw new Error("option in result, can't be removed.");
            }
        }

        const optionInfo = this._mapGet(optionPrefix + voteId, option);
        if (info.deleted === TRUE || new Float64(optionInfo.votes).isZero()) {
            storage.mapDel(optionPrefix + voteId, option);
        } else {
            optionInfo.deleted = TRUE;
            this._mapPut(optionPrefix + voteId, option, optionInfo, owner);
        }

        this._removeFromPreResult(voteId, option, preResult);

        info.optionNum--;
        this._mapPut("voteInfo", voteId, info, owner);
    }

    getOption(voteId, option) {
        this._checkVote(voteId);
        this._checkDel(voteId);

        if (!storage.mapHas(optionPrefix + voteId, option)) {
            throw new Error("option not exist");
        }

        return this._mapGet(optionPrefix + voteId, option);
    }

    _clearUserVote(clearTime, userVote) {
        if (clearTime >= userVote[1]) {
            userVote[2] = userVote[0];
        }
        return userVote;
    }

    _fixAmount(amount) {
        amount = new Float64(new Float64(amount).toFixed(iostDecimal));
        if (amount.lte("0")) {
            throw new Error("amount must be positive");
        }
        return amount;
    }

    _checkVoteAuth(account, payer) {
        if (account === payer) {
            this._requireAuth(payer, votePermission);
        } else {
            this._requireAuth(payer, votePermission);
            const fundIDs = this._get("fundID");
            if (!fundIDs.includes(payer)) {
                throw new Error("payer is not allowed to call voteFor.");
            }
        }
    }

    voteFor(voteId, payer, account, option, amount) {
        this._checkVote(voteId);
        this._checkDel(voteId);
        this._checkVoteAuth(account, payer);

        if (!storage.mapHas(optionPrefix + voteId, option)) {
            throw new Error("option does not exist");
        }

        amount = this._fixAmount(amount);
        blockchain.deposit(payer, amount.toFixed(), "");

        const optionInfo = this._mapGet(optionPrefix + voteId, option);
        if (optionInfo.deleted === TRUE) {
            throw new Error("option is removed.");
        }

        const userVotes = this._mapGet(userVotePrefix + voteId, account) || {};
        const clearTime = optionInfo.clearTime;

        if (userVotes.hasOwnProperty(option)) {
            userVotes[option] = this._clearUserVote(clearTime, userVotes[option]);
            userVotes[option][0] = new Float64(userVotes[option][0]).plus(amount).toFixed();
            userVotes[option][1] = block.number;
        } else {
            userVotes[option] = this._clearUserVote(clearTime, [amount.toFixed(), block.number, "0"]);
        }
        this._mapPut(userVotePrefix + voteId, account, userVotes, payer);
        if (clearTime === block.number) {
            // vote in clear block will do nothing.
            return;
        }

        const info = this._mapGet("voteInfo", voteId);
        const votes = amount.plus(optionInfo.votes);
        optionInfo.votes = votes.toFixed();
        this._mapPut(optionPrefix + voteId, option, optionInfo, payer);

        if (votes.gte(info.minVote)) {
            this._addToPreResult(info, voteId, option, optionInfo.votes);
        }
    }

    vote(voteId, account, option, amount) {
        this.voteFor(voteId, account, account, option, amount);
    }

    unvote(voteId, account, option, amount) {
        this._checkVote(voteId);
        this._requireAuth(account, votePermission);

        if (!storage.mapHas(userVotePrefix + voteId, account)) {
            throw new Error("account didn't vote.");
        }
        let userVotes = this._mapGet(userVotePrefix + voteId, account);
        if (!userVotes[option]) {
            throw new Error("account didn't vote for this option.");
        }

        amount = this._fixAmount(amount);
        const optionInfo = this._mapGet(optionPrefix + voteId, option);
        const clearTime = optionInfo ? optionInfo.clearTime : -1;
 
        userVotes[option] = this._clearUserVote(clearTime, userVotes[option]);
        const votes = new Float64(userVotes[option][0]);
        if (votes.lt(amount)) {
            throw new Error("amount too large. max amount = " + votes.toFixed());
        }
        let freezeTime = tx.time;

        const info = this._mapGet("voteInfo", voteId);
        if (info.deleted === FALSE) {
            freezeTime += info.freezeTime*1e9;
        }
        blockchain.callWithAuth("token.iost", "transferFreeze", ["iost", "vote.iost", account, amount.toFixed(), freezeTime, ""]);

        const leftVoteNum = votes.minus(amount);
        userVotes[option][0] = leftVoteNum.toFixed();

        const realUnvotes = new Float64(userVotes[option][2]).minus(amount);
        if (realUnvotes.gt("0")) {
            userVotes[option][2] = realUnvotes.toFixed();
            this._mapPut(userVotePrefix + voteId, account, userVotes, account);
            return;
        }

        userVotes[option][2] = "0";

        if (userVotes[option][0] === "0") {
            delete userVotes[option];
        }

        if (Object.keys(userVotes).length === 0) {
            storage.mapDel(userVotePrefix + voteId, account);
        } else {
            this._mapPut(userVotePrefix + voteId, account, userVotes, account);
        }

        if (storage.mapHas(optionPrefix + voteId, option)) {
            const optionVotes = new Float64(optionInfo.votes);
            const leftVotes = optionVotes.plus(realUnvotes);
            optionInfo.votes = leftVotes.toFixed();
            if (info.deleted === TRUE && leftVotes.isZero()) {
                storage.mapDel(optionPrefix + voteId, option);
            } else {
                this._mapPut(optionPrefix + voteId, option, optionInfo, account);
            }
            if (leftVotes.lt(info.minVote)) {
                this._removeFromPreResult(voteId, option);
            } else {
                this._addToPreResult(info, voteId, option, optionInfo.votes);
            }
        }
    }

    getVote(voteId, account) {
        this._checkVote(voteId);

        let userVotes = this._mapGet(userVotePrefix + voteId, account);
        if (!userVotes) {
            return [];
        }
        let votes = [];
        for (const o in userVotes) {
            votes.push({
                option: o,
                votes: new Float64(userVotes[o][0]).minus(userVotes[o][2]).toFixed(),
                voteTime: userVotes[o][1],
                clearedVotes: userVotes[o][2]
            });
        }
        return votes;
    }

    getResult(voteId) {
        this._checkVote(voteId);
        const preResult = this._mapGet("preResult", voteId);
        let preOrder = this._mapGet("preOrder", voteId) || [];
        let result = [];
        if (!preResult || !preOrder) {
            return result;
        }
        const info = this._mapGet("voteInfo", voteId);
        // pre sorted result
        preOrder = preOrder.slice(0, info.resultNumber);
        for (const o of preOrder) {
            result.push({
                option: o,
                votes: preResult[o],
            });
        }
        return result;
    }

    delVote(voteId) {
        this._requireOwner(voteId);
        this._checkDel(voteId);

        const owner = this._mapGet("owner", voteId);
        const info = this._mapGet("voteInfo", voteId);

        const deposit = new Float64(info.deposit);
        if (!deposit.isZero()) {
            blockchain.withdraw(owner, deposit.toFixed(), "");
        }
        this._delVote(voteId, info);
    }
}

module.exports = VoteCommonContract;
