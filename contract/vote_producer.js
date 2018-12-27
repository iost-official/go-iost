const PRE_PRODUCER_THRESHOLD = "10500000";
const VOTE_LOCKTIME = 604800;
const VOTE_STAT_INTERVAL = 2000;
const IOST_DECIMAL = 8;
const ADMIN_PERMISSION = "active";
const VOTE_PERMISSION = "vote";
const ACTIVE_PERMISSION = "active";

const STATUS_APPLY = 0;
const STATUS_APPROVED = 1;
const STATUS_UNAPPLY = 2;
const STATUS_UNAPPLY_APPROVED = 3;

const voterMaskPrefix = "v_";
const voterCoefTable = "voterCoef";

const candidateMaskTable = "candMask";
const candidateCoef = "candCoef";
const candidateAllKey = "candAllKey";

const voteThreshold = new Float64(21 * 1000 * 1000 * 1000 * 0.0005);

class VoteContract {
    init() {
        this._put("currentProducerList", []);
        this._put("pendingProducerList", []);
        this._put("pendingBlockNumber", 0);
        this._initVote();
    }

    _initVote() {
        const voteId = this._call("vote.iost", "NewVote", [
            "vote_producer.iost",
            "vote for producer",
            {
                resultNumber: 100,
                minVote: PRE_PRODUCER_THRESHOLD,
                options: [],
                anyOption: false,
                freezeTime: VOTE_LOCKTIME
            }
        ]);
        this._put("voteId", JSON.stringify(voteId));
    }

    InitProducer(proID, proPubkey) {
        const bn = block.number;
        if(bn !== 0) {
            throw new Error("init out of genesis block");
        }
        if (storage.mapHas("producerKeyToId", proPubkey)) {
            throw new Error("pubkey is used by another producer");
        }

        let pendingProducerList = this._get("pendingProducerList");
        pendingProducerList.push(proPubkey);
        const keyCmp = function(a, b) {
            if (b < a) {
                return 1;
            } else {
                return -1;
            }
        };
        pendingProducerList.sort(keyCmp);
        this._put("pendingProducerList", pendingProducerList);

        const producerNumber = pendingProducerList.length;
        this._put("producerNumber", producerNumber);

        const voteId = this._getVoteId();
        this._call("vote.iost", "AddOption", [
            voteId,
            proID,
            false
        ]);

        this._mapPut("producerTable", proID, {
            "pubkey" : proPubkey,
            "loc": "",
            "url": "",
            "netId": "",
            "status": STATUS_APPROVED,
            "online": true,
        }, proID);
        this._mapPut("producerKeyToId", proPubkey, proID, proID);
    }

    InitAdmin(adminID) {
        const bn = block.number;
        if(bn !== 0) {
            throw new Error("init out of genesis block")
        }
        this._put("adminID", adminID);
    }

    can_update(data) {
        const admin = this._get("adminID");
        this._requireAuth(admin, ADMIN_PERMISSION);
        return true;
    }

    _requireAuth(account, permission) {
        blockchain.requireAuth(account, permission);
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

    _getVoteId() {
        return this._get("voteId");
    }

    // register account as a producer
    ApplyRegister(account, pubkey, loc, url, netId) {
        this._requireAuth(account, VOTE_PERMISSION);
        if (storage.mapHas("producerTable", account)) {
            throw new Error("producer exists");
        }
        if (storage.mapHas("producerKeyToId", pubkey)) {
            throw new Error("pubkey is used by another producer");
        }

        this._mapPut("producerTable", account, {
            "pubkey" : pubkey,
            "loc": loc,
            "url": url,
            "netId": netId,
            "status": STATUS_APPLY,
            "online": false,
        }, account);
        this._mapPut("producerKeyToId", pubkey, account, account);

        const voteId = this._getVoteId();
        this._call("vote.iost", "AddOption", [
            voteId,
            account,
            false
        ]);
    }

    // apply remove account from producer list
    ApplyUnregister(account) {
        this._requireAuth(account, VOTE_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        if (pro.status === STATUS_APPLY) {
            this._doRemoveProducer(account, pro.pubkey, false);
            return;
        }
        if (pro.status !== STATUS_APPROVED) {
            throw new Error("producer not approved");
        }
        pro.status = STATUS_UNAPPLY;
        this._mapPut("producerTable", account, pro, account);
    }

    // approve account as a producer
    ApproveRegister(account) {
        const admin = this._get("adminID");
        this._requireAuth(admin, ADMIN_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        pro.status = STATUS_APPROVED;
        this._mapPut("producerTable", account, pro);
    }

    // approve remove account from producer list
    ApproveUnregister(account) {
        const admin = this._get("adminID");
        this._requireAuth(admin, ADMIN_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        if (pro.status !== STATUS_UNAPPLY) {
            throw new Error("producer not unapplied");
        }
        // will clear votes and score of the producer on stat
        pro.status = STATUS_UNAPPLY_APPROVED;
        this._mapPut("producerTable", account, pro);
        this._tryRemoveProducer(admin, account, pro);
    }

    // force approve remove account from producer list
    ForceUnregister(account) {
        const admin = this._get("adminID");
        this._requireAuth(admin, ADMIN_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        // will clear votes and score of the producer on stat
        pro.status = STATUS_UNAPPLY_APPROVED;
        this._mapPut("producerTable", account, pro);
        this._tryRemoveProducer(admin, account, pro);
    }

    Unregister(account) {
        this._requireAuth(account, VOTE_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        if (pro.status !== STATUS_UNAPPLY_APPROVED) {
            throw new Error("producer can not unregister");
        }
        const voteId = this._getVoteId();
        this._call("vote.iost", "RemoveOption", [
            voteId,
            account,
            true,
        ]);
        // will clear votes and score of the producer on stat
        pro.status = STATUS_UNAPPLY_APPROVED;
        this._mapPut("producerTable", account, pro);
        this._doRemoveProducer(admin, account, pro);
    }

    _tryRemoveProducer(admin, account, pro) {
        const currentList = this._get("currentProducerList");
        const pendingList = this._get("pendingProducerList");
        if (currentList.includes(pro.pubkey) || pendingList.includes(pro.pubkey)) {
            this._waitRemoveProducer(admin, account);
        }
    }

    _waitRemoveProducer(admin, account) {
        let waitList = this._get("waitingRemoveList") || [];
        if (!waitList.includes(account)) {
            waitList.push(account);
            this._put("waitingRemoveList", waitList, admin);
            let scores = this._getScores();
            if (scores[account] !== undefined) {
                scores[account] = "0";
                this._putScores(scores);
            }
        }
    }

    _doRemoveProducer(account, pubkey, deleteScore = true) {
        this._mapDel("producerTable", account);
        this._mapDel("producerKeyToId", pubkey);
        if (!deleteScore) {
            return;
        }
        let scores = this._getScores();
        if (scores[account] !== undefined) {
            delete(scores[account]);
            this._putScores(scores);
        }
    }

    // update the information of a producer
    UpdateProducer(account, pubkey, loc, url, netId) {
        this._requireAuth(account, VOTE_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        if (pro.pubkey !== pubkey) {
            if (storage.mapHas("producerKeyToId", pubkey)) {
                throw new Error("pubkey is used by another producer");
            }
            const currentList = this._get("currentProducerList");
            const pendingList = this._get("pendingProducerList");
            if (currentList.includes(pro.pubkey) || pendingList.includes(pro.pubkey)) {
                throw new Error("account in producerList, can't change pubkey");
            }

            this._mapDel("producerKeyToId", pro.pubkey, account);
            this._mapPut("producerKeyToId", pubkey, account, account);
        }
        pro.pubkey = pubkey;
        pro.loc = loc;
        pro.url = url;
        pro.netId = netId;
        this._mapPut("producerTable", account, pro, account);
    }

    GetProducer(account) {
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        if (pro.status === STATUS_APPROVED || pro.status === STATUS_UNAPPLY) {
            const voteId = this._getVoteId();
            pro["voteInfo"] = this._call("vote.iost", "GetOption", [
                voteId,
                account
            ]);
        }
        return pro;
    }

    // producer log in as online state
    LogInProducer(account) {
        this._requireAuth(account, VOTE_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists, " + account);
        }
        const pro = this._mapGet("producerTable", account);
        if (pro.status === STATUS_APPLY) {
            throw new Error("producer not approved");
        }
        pro.online = true;
        this._mapPut("producerTable", account, pro, account);
    }

    // producer log out as offline state
    LogOutProducer(account) {
        this._requireAuth(account, VOTE_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        if (pro.status === STATUS_APPLY) {
            throw new Error("producer not approved");
        }
        if (this._get("pendingProducerList").includes(account) ||
            this._get("currentProducerList").includes(account)) {
            throw new Error("producer in pending list or in current list, can't logout");
        }
        const pro = this._mapGet("producerTable", account);
        pro.online = false;
        this._mapPut("producerTable", account, pro, account);
    }

    _getVoterCoef(producer) {
        let voterCoef = this._mapGet(voterCoefTable, producer);
        if (!voterCoef) {
            voterCoef = "0";
        }
        return new Float64(voterCoef);
    }

    _getVoterMask(voter, producer) {
        let voterMask = this._mapGet(voterMaskPrefix + producer, voter);
        if (!voterMask) {
            voterMask = "0";
        }
        return new Float64(voterMask);
    }

    _updateVoterMask(voter, producer, amount) {
        let voterCoef = this._getVoterCoef(producer);
        let voterMask = this._getVoterMask(voter, producer);
        voterMask = voterMask.plus(voterCoef.multi(amount));
        this._mapPut(voterMaskPrefix + producer, voter, voterMask.toFixed(), producer);
    }

    _getCandidateAllKey() {
        let k = this._get(candidateAllKey);
        if (!k) {
            k = "0";
        }
        return new Float64(k);
    }

    _getCandCoef() {
        let candCoef = this._get(candidateCoef);
        if (!candCoef) {
            candCoef = "0";
        }
        return new Float64(candCoef);
    }

    _getCandMask(account) {
        let candMask = this._mapGet(candidateMaskTable, account)
        if (!candMask) {
            candMask = "0";
        }
        return new Float64(candMask);
    }

    _updateCandidateMask(account, key) {
        let allKey = this._getCandidateAllKey().plus(key);
        this._put(candidateAllKey, allKey.toFixed()); // payer?

        let candCoef = this._getCandCoef();
        let candMask = this._getCandMask(account);
        candMask = candMask.plus(candCoef.multi(key));
        this._mapPut(candidateMaskTable, account, candMask.toFixed(), account);
    }

    _updateCandidateVars(account, amount, voteId) {
        let votes = new Float64(this._call("vote.iost", "GetOption", [
           voteId,
           account,
        ]).votes);

        if (amount.isPositive()) {
            if (votes.lt(voteThreshold)) {
                return;
            }

            if (votes.minus(amount).lt(voteThreshold)) {
                this._updateCandidateMask(account, votes)
            } else {
                this._updateCandidateMask(account, amount)
            }

        } else if (amount.isNegative()) {
            if (votes.minus(amount).lt(voteThreshold)) {
                return;
            }

            if (votes.lt(voteThreshold)) {
                this._updateCandidateMask(account, votes.negated())
            } else {
                this._updateCandidateMask(account, amount)
            }
        }

    }

    VoteFor(payer, voter, producer, amount) {
        this._requireAuth(payer, ACTIVE_PERMISSION);

        if (!storage.mapHas("producerTable", producer)) {
            throw new Error("producer not exists");
        }

        const voteId = this._getVoteId();
        this._call("vote.iost", "VoteFor", [
            voteId,
            payer,
            voter,
            producer,
            amount,
        ]);

        this._updateVoterMask(voter, producer, new Float64(amount));
        this._updateCandidateVars(producer, new Float64(amount), voteId);
    }

    Vote(voter, producer, amount) {
        this._requireAuth(voter, ACTIVE_PERMISSION);

        if (!storage.mapHas("producerTable", producer)) {
            throw new Error("producer not exists");
        }

        const voteId = this._getVoteId();
        this._call("vote.iost", "Vote", [
            voteId,
            voter,
            producer,
            amount,
        ]);

        this._updateVoterMask(voter, producer, new Float64(amount));
        this._updateCandidateVars(producer, new Float64(amount), voteId);
    }

    Unvote(voter, producer, amount) {
        this._requireAuth(voter, VOTE_PERMISSION);

        const voteId = this._getVoteId();
        this._call("vote.iost", "Unvote", [
            voteId,
            voter,
            producer,
            amount,
        ]);

        this._updateVoterMask(voter, producer, new Float64(amount).negated());
        this._updateCandidateVars(producer, new Float64(amount).negated(), voteId);
    }

    GetVote(voter) {
        const voteId = this._getVoteId();
        return this._call("vote.iost", "GetVote", [
            voteId,
            voter
        ]);
    }

    TopupVoterBonus(account, amount) {
        this._requireAuth(account, ACTIVE_PERMISSION);
        const voteId = this._getVoteId();
        let votes = new Float64(this._call("vote.iost", "GetOption", [
           voteId,
           account,
        ]).votes);
        if (!votes.isPositive()) {
            return false;
        }

        let voterCoef = this._getVoterCoef(account);
        voterCoef = voterCoef.plus(new Float64(amount).div(votes));
        this._mapPut(voterCoefTable, account, voterCoef.toFixed(), account);
        return true;
    }

    TopupCandidateBonus(amount) {
        // TODO requireAuth?
        let allKey = this._getCandidateAllKey();
        if (!allKey.isPositive()) {
            return false;
        }

        let candCoef = this._getCandCoef();
        candCoef = candCoef.plus(new Float64(amount).div(allKey));
        this._put(candidateCoef, candCoef.toFixed());
        return true;
    }

    _calVoterBonus(voter, updateMask) {
        let userVotes = this.GetVote(voter);
        let earnings = new Float64(0);
        for (const v in userVotes) {
           let voterCoef = this._getVoterCoef(v.option);
           let voterMask = this._getVoterMask(voter, v.option);
           let earning = voterCoef.multi(new Float64(v.votes)).minus(voterMask);
           earnings = earnings.plus(earning);
           if (updateMask) {
              voterMask = voterMask.plus(earning);
              this._mapPut(voterMaskPrefix + v.option, voter, voterMask.toFixed(), v.option);
           }
        }
        return earnings;
    }

    GetVoterBonus(voter) {
        return this._calVoterBonus(voter, false).toFixed();
    }

    VoterWithdraw(voter) {
        this._requireAuth(voter, ACTIVE_PERMISSION);

        let earnings = this._calVoterBonus(voter, true);
        // TODO transfer earnings to voter
    }

    _calCandidateBonus(account, updateMask) {
        const voteId = this._getVoteId();
        let candKey = new Float64(this._call("vote.iost", "GetOption", [
           voteId,
           account,
        ]).votes);

        if (candKey.lt(voteThreshold)) {
            candKey = new Float64(0);
        }

        let candCoef = this._getCandCoef();
        let candMask = this._getCandMask(account);
        let earning = candCoef.plus(candKey).minus(candMask);
        if (updateMask) {
            candMask = candMask.plus(earning);
            this._mapGet(candidateMaskTable, account, candMask.toFixed(), account);
        }
        return earning;
     }

    GetCandidateBonus(account) {
        return this._calCandidateBonus(account, false).toFixed();
    }

    CandidateWithdraw(account) {
        this._requireAuth(account, ACTIVE_PERMISSION);

        let earnings = this._calCandidateBonus(account, true);
        let halfEarning = earnings.div(new Float64("0.5"));
        // TODO: transfer half of earnings to account

        this.TopupVoterBonus(account, halfEarning);
        // TODO: earnings - halfEarning - halfEarning?
    }

    _getScores() {
        const scores = this._get("producerScores");
        if (!scores) {
            return {};
        }
        return scores;
    }

    _putScores(scores) {
        this._put("producerScores", scores);
    }

    // calculate the vote result, modify pendingProducerList
    Stat() {
        this._requireAuth("base.iost", ACTIVE_PERMISSION);
        const bn = block.number;
        const pendingBlockNumber = this._get("pendingBlockNumber");
        if (bn % VOTE_STAT_INTERVAL !== 0 || bn <= pendingBlockNumber) {
            return;
        }

        const voteId = this._getVoteId();
        const voteRes = this._call("vote.iost", "GetResult", [voteId]);
        const preList = [];    // list of producers whose vote > threshold
        const waitingRemoveList = this._get("waitingRemoveList") || [];
        let scores = this._getScores();
        const pendingProducerList = this._get("pendingProducerList");

        // update scores
        let scoreTotal = new Float64("0");
        let scoreCount = 0;
        for (const res of voteRes) {
            const id = res.option;
            const pro = this._mapGet("producerTable", id);
            if (pro.online === false || (pro.status !== STATUS_APPROVED && pro.status !== STATUS_UNAPPLY)) {
                continue;
            }
            const score = new Float64(res.votes).plus(scores[id] || "0");
            scoreTotal = scoreTotal.plus(score);
            scoreCount++;
            scores[id]  = score.toFixed();
            if (!pendingProducerList.includes(pro.pubkey)) {
                preList.push({
                    "id" : id,
                    "key": pro.pubkey,
                    "prior": 0,
                    "score": score,
                });
            }
        }

        // update pending list
        let oldPreList = [];
        const oldPreListToRemove = [];
        let minScore = new Float64(MaxFloat64);
        for (const key of pendingProducerList) {
            const account = this._mapGet("producerKeyToId", key);
            const score = new Float64(scores[account] || "0");
            if (waitingRemoveList.includes(account)) {
                oldPreListToRemove.push({
                    "account": account,
                    "key": key,
                    "prior": 0,
                    "score": score
                });
                minScore = new Float64(0);
            } else {
                oldPreList.push({
                    "account": account,
                    "key": key,
                    "prior": 1,
                    "score": score
                });
                if (score.lt(minScore)) {
                    minScore = score;
                }
            }
        }

        // sort according to score in reversed order
        const scoreCmp = function(a, b) {
            if (!a.score.eq(b.score)) {
                return a.score.lt(b.score) ? 1 : -1;
            } else if (b.prior !== a.prior) {
                return b.prior - a.prior;
            } else {
                return b.key < a.key ? 1 : -1;
            }
        };

        // replace producerNumber producers
        const producerNumber = this._get("producerNumber");
        oldPreList = [...oldPreList, ...preList];
        oldPreList.sort(scoreCmp);
        oldPreList = [...oldPreList, ...oldPreListToRemove];

        const removedList = oldPreList.splice(producerNumber);
        const newList = oldPreList;

        const currentList = pendingProducerList;
        const pendingList = newList.map(x => x.key);
        this._put("currentProducerList", currentList);
        this._put("pendingProducerList", pendingList);
        this._put("pendingBlockNumber", block.number);

        if (scoreCount > 0) {
            const scoreAvg = scoreTotal.div(scoreCount);
            for (const key of pendingList) {
                const account = this._mapGet("producerKeyToId", key);
                scores[account] = new Float64(scores[account] || "0").minus(scoreAvg).toFixed(IOST_DECIMAL);
            }
        } else {
            for (const key of pendingList) {
                const account = this._mapGet("producerKeyToId", key);
                scores[account] = "0";
            }
        }

        for (const removed of removedList) {
            if (!waitingRemoveList.includes(removed.account)) {
                continue;
            }
            delete(scores[removed.account]);
            this._doRemoveProducer(removed.account, removed.key, false);
        }
        const newWaitingRemoveList = waitingRemoveList.filter(function(value, index, arr) {
            return !removedList.includes(value);
        });
        this._put("waitingRemoveList", newWaitingRemoveList);

        this._putScores(scores);
    }
}

module.exports = VoteContract;
