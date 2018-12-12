const newVoteFee = "1000";
const descriptionMaxLength = 65536;
const optionMaxLength = 4096;
const resultMaxLength = 1024;
const iostDecimal = 8;

const adminPermission = "active";
const votePermission = "vote";

const preResultPrefix = "p_";
const optionPrefix = "v_";
const userVotePrefix = "u_";

class VoteCommonContract {
    constructor() {
        BigNumber.config({
            DECIMAL_PLACES: 50,
            POW_PRECISION: 50,
            ROUNDING_MODE: BigNumber.ROUND_DOWN
        });
    }

    init() {
        this._put("current_id", "0");
    }

    InitAdmin(adminID) {
        const bn = block.number;
        if(bn !== 0) {
            throw new Error("init out of genesis block");
        }
        this._put("adminID", adminID);
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

    _call(contract, api, args) {
        const ret = blockchain.callWithAuth(contract, api, JSON.stringify(args));
        if (ret && Array.isArray(ret) && ret.length === 1) {
            return ret[0] === "" ? "" : JSON.parse(ret[0]);
        }
        return ret;
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

    _delVote(voteId) {
        this._mapPut("voteInfo", voteId, false);
        const optionKeys = storage.mapKeys(optionPrefix + voteId);
        for (const key of optionKeys) {
            storage.mapDel(optionPrefix + voteId, key);
        }
        const preResultKeys = storage.mapKeys(preResultPrefix + voteId);
        for (const key of preResultKeys) {
            storage.mapDel(preResultPrefix + voteId, key);
        }
    }

    _checkDel(voteId) {
        const info = this._mapGet("voteInfo", voteId);
        if (info === false) {
            throw new Error("vote has been deleted.");
        }
    }

    NewVote(owner, description, info) {
        this._requireAuth(owner, adminPermission);

        if (description.length > descriptionMaxLength) {
            throw new Error("description too long. max length is " + descriptionMaxLength + " Byte.");
        }

        const resultNumber = info.resultNumber;
        if (!resultNumber || !Number.isInteger(resultNumber) || resultNumber <= 0 || resultNumber > 100) {
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
            this._call("token.iost", "transfer", ["iost", owner, "vote.iost", newVoteFee, ""]);
        }

        const voteId = this._nextId();

        this._mapPut("voteInfo", voteId, {
            description: description,
            resultNumber: resultNumber,
            minVote: minVote,
            anyOption: anyOption,
            freezeTime: freezeTime,
            deposit: bn > 0 ? newVoteFee : "0"
        }, owner);

        for (const option of options) {
            const initVotes = [
                "0",        // votes
                false,      // deleted
                -1,         // clearTime
            ];
            this._mapPut(optionPrefix + voteId, option, initVotes, owner);
        }

        this._mapPut("owner", voteId, owner, owner);

        return voteId;
    }

    AddOption(voteId, option, clearVote) {
        const owner = this._requireOwner(voteId);
        this._checkDel(voteId);

        if (option.length > resultMaxLength) {
            throw new Error("option too long. max length is 1024 Byte.");
        }

        const options = storage.mapKeys(optionPrefix + voteId);

        if (options.length >= optionMaxLength) {
            throw new Error("options is full.");
        }

        let optionProp = ["0", false, -1];
        if (storage.mapHas(optionPrefix + voteId, option)) {
            optionProp = this._mapGet(optionPrefix + voteId, option);
            if (optionProp[1] === false) {
                throw new Error("option already exist.");
            }
            if (clearVote === true) {
                optionProp = ["0", false, block.number];
            } else {
                optionProp[1] = false;
            }
        }

        this._mapPut(optionPrefix + voteId, option, optionProp, owner);

        const info = this._mapGet("voteInfo", voteId);
        const votes = new Float64(optionProp[0]);
        if (votes.lt(info.minVote)) {
            return;
        }

        this._mapPut(preResultPrefix + voteId, option, optionProp[0], owner);
    }

    RemoveOption(voteId, option, force) {
        const owner = this._requireOwner(voteId);
        this._checkDel(voteId);

        if (!storage.mapHas(optionPrefix + voteId, option)) {
            throw new Error("option not exist");
        }

        const info = this._mapGet("voteInfo", voteId);
        const optionProp = this._mapGet(optionPrefix + voteId, option);

        if (!force && storage.mapHas(preResultPrefix + voteId, option)) {
            let order = 0;
            const votes = new Float64(optionProp[0]);
            const preResultKeys = storage.mapKeys(preResultPrefix + voteId);
            for (const key of preResultKeys) {
                const preVotes = this._mapGet(preResultPrefix + voteId, key);
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

        optionProp[1] = true;
        this._mapPut(optionPrefix + voteId, option, optionProp, owner);

        if (storage.mapHas(preResultPrefix + voteId, option)) {
            this._mapDel(preResultPrefix + voteId, option);
        }
    }

    GetOption(voteId, option) {
        this._checkVote(voteId);
        this._checkDel(voteId);

        if (!storage.mapHas(optionPrefix + voteId, option)) {
            throw new Error("option not exist");
        }

        const optionProp = this._mapGet(optionPrefix + voteId, option);
        return {
            votes: optionProp[0],
            deleted: optionProp[1],
            clearTime: optionProp[2]
        };
    }

    _clearUserVote(clearTime, userVote) {
        if (clearTime >= userVote[1]) {
            userVote[2] = userVote[0];
        }
        return userVote;
    }

    _fixAmount(amount) {
        return new Float64(new BigNumber(amount).toFixed(iostDecimal));
    }

    Vote(voteId, account, option, amount) {
        this._checkVote(voteId);
        this._checkDel(voteId);
        this._requireAuth(account, votePermission);

        amount = this._fixAmount(amount);

        this._call("token.iost", "transfer", ["iost", account, "vote.iost", amount.toFixed(), ""]);

        if (!storage.mapHas(optionPrefix + voteId, option)) {
            throw new Error("option does not exist");
        }

        const optionProp = this._mapGet(optionPrefix + voteId, option);
        if (optionProp[1] === true) {
            throw new Error("option is removed.");
        }

        const userVotes = this._mapGet(userVotePrefix + voteId, account) || {};
        const clearTime = optionProp[2];

        if (userVotes.hasOwnProperty(option)) {
            userVotes[option] = this._clearUserVote(clearTime, userVotes[option]);
            userVotes[option][0] = new Float64(userVotes[option][0]).plus(amount).toFixed();
            userVotes[option][1] = block.number;
        } else {
            userVotes[option] = this._clearUserVote(clearTime, [amount.toFixed(), block.number, "0"]);
        }
        this._mapPut(userVotePrefix + voteId, account, userVotes, account);
        if (clearTime === block.number) {
            // vote in clear block will do nothing.
            return;
        }

        const votes = new Float64(optionProp[0]).plus(amount);
        optionProp[0]  = votes.toFixed();
        this._mapPut(optionPrefix + voteId, option, optionProp, account);

        const info = this._mapGet("voteInfo", voteId);
        if (votes.lt(info.minVote)) {
            return;
        }

        this._mapPut(preResultPrefix + voteId, option, optionProp[0], account);
    }

    Unvote(voteId, account, option, amount) {
        this._checkVote(voteId);
        this._requireAuth(account, votePermission);

        amount = this._fixAmount(amount);
        if (!storage.mapHas(userVotePrefix + voteId, account)) {
            throw new Error("account didn't vote.");
        }
        let userVotes = this._mapGet(userVotePrefix + voteId, account);
        if (!userVotes[option]) {
            throw new Error("account didn't vote for this option.");
        }

        const optionProp = this._mapGet(optionPrefix + voteId, option);
        let clearTime = -1;
        if (optionProp && optionProp.length === 3) {
            clearTime = optionProp[2];
        }

        userVotes[option] = this._clearUserVote(clearTime, userVotes[option]);
        const votes = new Float64(userVotes[option][0]);
        if (votes.lt(amount)) {
            throw new Error("amount too large. max amount = " + votes);
        }
        const info = this._mapGet("voteInfo", voteId);
        let freezeTime = tx.time;
        if (info !== false) {
            freezeTime += info.freezeTime*1e9;
        }
        this._call("token.iost", "transferFreeze", ["iost", "vote.iost", account, amount.toFixed(), freezeTime, ""]);

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
            this._mapDel(userVotePrefix + voteId, account);
        } else {
            this._mapPut(userVotePrefix + voteId, account, userVotes, account);
        }


        if (storage.mapHas(optionPrefix + voteId, option)) {
            optionProp[0] = new Float64(optionProp[0]).plus(realUnvotes).toFixed();
            this._mapPut(optionPrefix + voteId, option, optionProp, account);
        }

        if (storage.mapHas(preResultPrefix + voteId, option)) {
            let preResultVotes = this._mapGet(preResultPrefix + voteId, option);
            const votes = new Float64(preResultVotes).plus(realUnvotes);
            preResultVotes = votes.toFixed();

            if (votes.lt(info.minVote)) {
                this._mapDel(preResultPrefix + voteId, option);
            } else {
                this._mapPut(preResultPrefix + voteId, option, preResultVotes, account);
            }
        }
    }

    GetVote(voteId, account) {
        this._checkVote(voteId);

        let userVotes = this._mapGet(userVotePrefix + voteId, account);
        if (!userVotes) {
            return {};
        }
        let votes = [];
        for (const k in userVotes) {
            if (!userVotes.hasOwnProperty(k)) {
                continue;
            }
            votes.push({
                option: k,
                votes: new Float64(userVotes[k][0]).minus(userVotes[k][2]).toFixed(),
                voteTime: userVotes[k][1],
                clearedVotes: userVotes[k][2]
            });
        }
        return votes;
    }

    GetResult(voteId) {
        this._checkVote(voteId);
        const preResultKeys = storage.mapKeys(preResultPrefix + voteId);
        let preResult = [];
        for (const key of preResultKeys) {
            preResult.push({
                option: key,
                votes: this._mapGet(preResultPrefix + voteId, key)
            });
        }
        // sort according to votes in reversed order
        const voteCmp = function(a, b) {
            return new Float64(a.votes).lt(b.votes);
        };
        preResult.sort(voteCmp);
        const info = this._mapGet("voteInfo", voteId);
        return preResult.slice(0, info.resultNumber);
    }

    DelVote(voteId) {
        this._requireOwner(voteId);
        this._checkDel(voteId);

        const owner = this._mapGet("owner", voteId);
        const info = this._mapGet("voteInfo", voteId);

        const deposit = new Float64(info.deposit);
        if (!deposit.isZero()) {
            this._call("token.iost", "transfer", ["iost", owner, "vote.iost", deposit, ""]);
        }
        this._delVote(voteId);
    }
}

module.exports = VoteCommonContract;
