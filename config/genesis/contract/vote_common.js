const newVoteFee = "1000";
const descriptionMaxLength = 65536;
const optionMaxLength = 4096;
const resultMaxLength = 1024;
const iostDecimal = 8;

const adminPermission = "active";
const votePermission = "vote";

const optionPrefix = "v_";
const userVotePrefix = "u_";

const TRUE = 1;
const FALSE = 0;

class VoteCommonContract {
    constructor() {
    }

    init() {
        this._put("current_id", "0");
    }

    initAdmin(adminID) {
        const bn = block.number;
        if(bn !== 0) {
            throw new Error("init out of genesis block");
        }
        this._put("adminID", adminID);
    }

    InitFundIDs(ids) {
        const bn = block.number;
        if(bn !== 0) {
            throw new Error("init out of genesis block");
        }
        this._put("fundID", ids);
    }

    can_update(data) {
        const admin = this._get("adminID");
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

    _mapDel(k, f) {
        storage.mapDel(k, f);
    }

    _nextId() {
        const id = new BigNumber(this._get("current_id")).plus(1).toFixed();
        this._put("current_id", id);
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
        for (const id in info.preResult) {
            storage.mapDel(optionPrefix + voteId, id);
        }
    }

    _checkDel(voteId) {
        const info = this._mapGet("voteInfo", voteId);
        if (info.deleted === TRUE) {
            throw new Error("vote has been deleted.");
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
        if (!minVote || new BigNumber(minVote).lte(0)) {
            throw new Error("minVote not valid.");
        }

        const options = info.options;
        if (!options || !Array.isArray(options) || options.length > optionMaxLength) {
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
            optionId: 0,
            optionNum: 0,
            options: {},
            preResult: {}
        };
        for (const option of options) {
            const id = String(voteInfo.optionId++);
            voteInfo.options[option] = {
                id: id,
                deleted: FALSE,
                clearTime: -1
            }
            voteInfo.preResult[id] = FALSE;
            this._mapPut(optionPrefix + voteId, id, "0", owner);
        }
        voteInfo.optionNum = voteInfo.optionId;

        this._mapPut("voteInfo", voteId, voteInfo, owner);
        this._mapPut("owner", voteId, owner, owner);
        this._mapPut("preResult", voteId, [], owner);

        return voteId;
    }

    addOption(voteId, option, clearVote) {
        const owner = this._requireOwner(voteId);
        this._checkDel(voteId);

        if (option.length > resultMaxLength) {
            throw new Error("option too long. max length is 1024 Byte.");
        }

        const info = this._mapGet("voteInfo", voteId);
        if (info.optionNum >= optionMaxLength) {
            throw new Error("options is full.");
        }

        if (info.options.hasOwnProperty(option)) {
            if (info.options[option].deleted === FALSE) {
                throw new Error("option already exist.");
            }
            info.options[option].deleted = FALSE;
            if (clearVote === true) {
                info.options[option].clearTime = block.number;
                this._mapPut(optionPrefix + voteId, info.options[option].id, "0", owner);
            } else {
                const id = info.options[option].id;
                const votes = new Float64(this._mapGet(optionPrefix + voteId, id));
                if (votes.gte(info.minVote)) {
                    info.preResult[id] = TRUE;
                }
            }
        } else {
            const id = String(info.optionId++);
            info.options[option] = {
                id: id,
                deleted: FALSE,
                clearTime: -1
            }
            info.preResult[id] = FALSE;
            this._mapPut(optionPrefix + voteId, id, "0", owner);
        }
        this._mapPut("voteInfo", voteId, info, owner);
    }

    removeOption(voteId, option, force) {
        const owner = this._requireOwner(voteId);
        this._checkDel(voteId);

        const info = this._mapGet("voteInfo", voteId);
        if (!info.options[option]) {
            throw new Error("option not exist");
        }

        const id = info.options[option].id;
        if (!force && info.preResult[id] === TRUE) {
            let order = 0;
            const votes = new Float64(this._mapGet(optionPrefix + voteId, id));
            for (const id in info.preResult) {
                if (info.preResult[id] === FALSE) {
                    continue;
                }
                const preVotes = this._mapGet(optionPrefix + voteId, id);
                if (votes.lt(preVotes)) {
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

        info.options[option].deleted = TRUE;
        info.preResult[id] = FALSE;
        this._mapPut("voteInfo", voteId, info, owner);
    }

    getOption(voteId, option) {
        this._checkVote(voteId);
        this._checkDel(voteId);

        const info = this._mapGet("voteInfo", voteId);
        if (!info.options[option]) {
            throw new Error("option not exist");
        }

        const votes = this._mapGet(optionPrefix + voteId, info.options[option].id);
        return {
            votes: votes,
            deleted: info.options[option].deleted,
            clearTime: info.options[option].clearTime
        };
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

        amount = this._fixAmount(amount);
        const info = this._mapGet("voteInfo", voteId);
        if (!info.options[option]) {
            throw new Error("option does not exist");
        }

        blockchain.deposit(payer, amount.toFixed(), "");

        const id = info.options[option].id;
        if (info.options[option].deleted === TRUE) {
            throw new Error("option is removed.");
        }

        const userVotes = this._mapGet(userVotePrefix + voteId, account) || {};
        const clearTime = info.options[option].clearTime;

        if (userVotes.hasOwnProperty(id)) {
            userVotes[id] = this._clearUserVote(clearTime, userVotes[id]);
            userVotes[id][0] = new Float64(userVotes[id][0]).plus(amount).toFixed();
            userVotes[id][1] = block.number;
        } else {
            userVotes[id] = this._clearUserVote(clearTime, [amount.toFixed(), block.number, "0"]);
        }
        this._mapPut(userVotePrefix + voteId, account, userVotes, payer);
        if (clearTime === block.number) {
            // vote in clear block will do nothing.
            return;
        }

        const votes = new Float64(this._mapGet(optionPrefix + voteId, id)).plus(amount);
        this._mapPut(optionPrefix + voteId, id, votes.toFixed(), payer);

        if (votes.gte(info.minVote)) {
            info.preResult[id] = TRUE;
        }

        this._mapPut("voteInfo", voteId, info);
    }

    vote(voteId, account, option, amount) {
        this.voteFor(voteId, account, account, option, amount);
    }

    unvote(voteId, account, option, amount) {
        this._checkVote(voteId);
        this._requireAuth(account, votePermission);

        amount = this._fixAmount(amount);
        if (!storage.mapHas(userVotePrefix + voteId, account)) {
            throw new Error("account didn't vote.");
        }
        let userVotes = this._mapGet(userVotePrefix + voteId, account);
        const info = this._mapGet("voteInfo", voteId);
        const id = info.options[option].id;
        if (!userVotes[id]) {
            throw new Error("account didn't vote for this option.");
        }

        const clearTime = info.options[option].clearTime;
        userVotes[id] = this._clearUserVote(clearTime, userVotes[id]);
        const votes = new Float64(userVotes[id][0]);
        if (votes.lt(amount)) {
            throw new Error("amount too large. max amount = " + votes.toFixed());
        }
        let freezeTime = tx.time;
        if (info.deleted === FALSE) {
            freezeTime += info.freezeTime*1e9;
        }
        blockchain.callWithAuth("token.iost", "transferFreeze", ["iost", "vote.iost", account, amount.toFixed(), freezeTime, ""]);

        const leftVoteNum = votes.minus(amount);
        userVotes[id][0] = leftVoteNum.toFixed();

        const realUnvotes = new Float64(userVotes[id][2]).minus(amount);
        if (realUnvotes.gt("0")) {
            userVotes[id][2] = realUnvotes.toFixed();
            this._mapPut(userVotePrefix + voteId, account, userVotes, account);
            return;
        }

        userVotes[id][2] = "0";

        if (userVotes[id][0] === "0") {
            delete userVotes[id];
        }

        if (Object.keys(userVotes).length === 0) {
            this._mapDel(userVotePrefix + voteId, account);
        } else {
            this._mapPut(userVotePrefix + voteId, account, userVotes, account);
        }

        if (storage.mapHas(optionPrefix + voteId, id)) {
            const optionVotes = new Float64(this._mapGet(optionPrefix + voteId, id));
            const leftVotes = optionVotes.plus(realUnvotes);
            this._mapPut(optionPrefix + voteId, id, leftVotes.toFixed(), account);
            if (info.preResult[id] === TRUE) {
                if (leftVotes.lt(info.minVote)) {
                    info.preResult[id] = FALSE;
                    this._mapPut("voteInfo", voteId, info);
                }
            }
        }
    }

    getVote(voteId, account) {
        this._checkVote(voteId);

        let userVotes = this._mapGet(userVotePrefix + voteId, account);
        if (!userVotes) {
            return {};
        }
        const info = this._mapGet("voteInfo", voteId);
        let id2option = {};
        for (const option in info.options) {
            id2option[info.options[option].id] = option;
        }
        let votes = [];
        for (const k in userVotes) {
            if (!userVotes.hasOwnProperty(k)) {
                continue;
            }
            votes.push({
                option: id2option[k],
                votes: new Float64(userVotes[k][0]).minus(userVotes[k][2]).toFixed(),
                voteTime: userVotes[k][1],
                clearedVotes: userVotes[k][2]
            });
        }
        return votes;
    }

    getResult(voteId) {
        this._checkVote(voteId);
        const info = this._mapGet("voteInfo", voteId);
        let preResult = [];
        for (const o in info.options) {
            const id = info.options[o].id;
            if (info.preResult[id] === FALSE) {
                continue;
            }
            preResult.push({
                option: o,
                votes: this._mapGet(optionPrefix + voteId, id)
            });
        }
        // sort according to votes in reversed order
        const voteCmp = function(a, b) {
            return new Float64(a.votes).lt(b.votes);
        };
        preResult.sort(voteCmp);
        return preResult.slice(0, info.resultNumber);
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
