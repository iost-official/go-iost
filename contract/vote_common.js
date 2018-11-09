const newVoteFee = "1000";
const descriptionMaxLength = 65536;
const optionMaxLength = 4096;
const resultMaxLength = 1024;
const iostDecimal = 8;

const adminPermission = "active";
const votePermission = "vote";

const preResultPrefix = "p-";
const optionPrefix = "v-";
const userVotePrefix = "u-";

class VoteCommonContract {
    constructor() {
        BigNumber.config({
            DECIMAL_PLACES: 50,
            POW_PRECISION: 50,
            ROUNDING_MODE: BigNumber.ROUND_DOWN
        })
    }

    init() {
        this._put("current_id", "0");
    }

    InitAdmin(adminID) {
        const bn = this._getBlockNumber();
        if(bn !== 0) {
            throw new Error("init out of genesis block")
        }
        this._put("adminID", adminID);
    }

    can_update(data) {
        const admin = this._get("adminID");
        this._requireAuth(admin, adminPermission);
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

    _getBlockNumber() {
        const bi = JSON.parse(BlockChain.blockInfo());
        if (!bi || bi === undefined || bi.number === undefined) {
            throw new Error("get block number failed. bi = " + bi);
        }
        return bi.number;
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
        return JSON.parse(val);
    }

    _mapPut(k, f, v) {
        storage.mapPut(k, f, JSON.stringify(v));
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

        let unvoteInterval = info.unvoteInterval;
        if (unvoteInterval === undefined) {
            unvoteInterval = 0;
        } else if (!Number.isInteger(unvoteInterval) || unvoteInterval < 0) {
            throw new Error("unvoteInterval not valid.");
        }

        const bn = this._getBlockNumber();

        if (bn > 0) {
            this._call("iost.token", "transfer", ["iost", owner, "iost.vote", newVoteFee]);
        }

        const voteId = this._nextId();

        this._mapPut("voteInfo", voteId, {
            description: description,
            resultNumber: resultNumber,
            minVote: minVote,
            anyOption: anyOption,
            unvoteInterval: unvoteInterval,
            deposit: bn > 0 ? newVoteFee : "0"
        });

        for (const option of options) {
            const initVotes = [
                "0",        // votes
                false,        // deleted
                -1,            // clearTime
            ]
            this._mapPut(optionPrefix + voteId, option, initVotes);
        }

        this._mapPut("owner", voteId, owner);

        return voteId;
    }

    AddOption(voteId, option, clearVote) {
        this._requireOwner(voteId);
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
                optionProp = ["0", false, this._getBlockNumber()];
            } else {
                optionProp[1] = false;
            }
        }

        this._mapPut(optionPrefix + voteId, option, optionProp);

        const info = this._mapGet("voteInfo", voteId);
        const votes = new Float64(optionProp[0]);
        if (votes.lt(new Float64(info.minVote))) {
            return;
        }

        this._mapPut(preResultPrefix + voteId, option, optionProp[0]);
    }

    RemoveOption(voteId, option, force) {
        this._requireOwner(voteId);
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
                if (votes.lt(new Float64(preVotes))) {
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
        this._mapPut(optionPrefix + voteId, option, optionProp);

        if (storage.mapHas(preResultPrefix + voteId, option)) {
            this._mapDel(preResultPrefix + voteId, option);
        }
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

        this._call("iost.token", "transfer", ["iost", account, "iost.vote", amount.number.toFixed()]);

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
            userVotes[option][0] = new Float64(userVotes[option][0]).plus(amount).number.toFixed();
            userVotes[option][1] = this._getBlockNumber();
        } else {
            userVotes[option] = this._clearUserVote(clearTime, [amount.number.toFixed(), this._getBlockNumber(), "0"]);
        }
        this._mapPut(userVotePrefix + voteId, account, userVotes);
        if (clearTime == this._getBlockNumber()) {
            // vote in clear block will do nothing.
            return;
        }

        const votes = new Float64(optionProp[0]).plus(amount);
        optionProp[0]  = votes.number.toFixed();
        this._mapPut(optionPrefix + voteId, option, optionProp);

        const info = this._mapGet("voteInfo", voteId);
        if (votes.lt(new Float64(info.minVote))) {
            return;
        }

        this._mapPut(preResultPrefix + voteId, option, optionProp[0]);
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
        if (info && userVotes[option][1] + info.unvoteInterval > this._getBlockNumber()) {
            throw new Error("unvoteInterval not reached.");
        }

        this._call("iost.token", "transfer", ["iost", "iost.vote", account, amount.number.toFixed()]);

        const leftVoteNum = votes.minus(amount);

        userVotes[option][0] = leftVoteNum.number.toFixed();

        const realUnvotes = new Float64(userVotes[option][2]).minus(amount);
        if (realUnvotes.gt(new Float64("0"))) {
            userVotes[option][2] = realUnvotes.number.toFixed();
            this._mapPut(userVotePrefix + voteId, account, userVotes);
            return;
        }

        userVotes[option][2] = "0";

        if (userVotes[option][0] === "0") {
            delete userVotes[option];
        }

        if (Object.keys(userVotes).length === 0) {
            this._mapDel(userVotePrefix + voteId, account);
        } else {
            this._mapPut(userVotePrefix + voteId, account, userVotes);
        }


        if (storage.mapHas(optionPrefix + voteId, option)) {
            optionProp[0] = new Float64(optionProp[0]).plus(realUnvotes).number.toFixed();
            this._mapPut(optionPrefix + voteId, option, optionProp);
        }

        if (storage.mapHas(preResultPrefix + voteId, option)) {
            let preResultVotes = this._mapGet(preResultPrefix + voteId, option);
            const votes = new Float64(preResultVotes).plus(realUnvotes);
            preResultVotes = votes.number.toFixed();

            if (votes.lt(new Float64(info.minVote))) {
                this._mapDel(preResultPrefix + voteId, option);
            } else {
                this._mapPut(preResultPrefix + voteId, option, preResultVotes);
            }
        }
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
            return new Float64(a.votes).lt(new Float64(b.votes));
        };
        preResult.sort(voteCmp);
        const info = this._mapGet("voteInfo", voteId);
        const result = preResult.slice(0, info.resultNumber);
        return result;
    }

    DelVote(voteId) {
        this._requireOwner(voteId);
        this._checkDel(voteId);

        const owner = this._mapGet("owner", voteId);
        const info = this._mapGet("voteInfo", voteId);

        const deposit = new Float64(info.deposit);
        if (!deposit.isZero()) {
            this._call("iost.token", "transfer", ["iost", owner, "iost.vote", deposit]);
        }
        this._delVote(voteId);
    }
}

module.exports = VoteCommonContract;
