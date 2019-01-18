const PRE_PRODUCER_THRESHOLD = "2100000";
const PARTNER_THRESHOLD = "2100000";
const VOTE_LOCKTIME = 604800;
const VOTE_STAT_INTERVAL = 2000;
const SCORE_DECREASE_INTERVAL = 51840000;
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

class VoteContract {
    init() {
        this._put("currentProducerList", []);
        this._put("pendingProducerList", []);
        this._put("pendingBlockNumber", 0);
        this._initVote();
    }

    _initVote() {
        const voteId = blockchain.callWithAuth("vote.iost", "newVote", [
            "vote_producer.iost",
            "vote for producer",
            {
                resultNumber: 2000,
                minVote: PRE_PRODUCER_THRESHOLD,
                options: [],
                anyOption: false,
                freezeTime: VOTE_LOCKTIME
            }
        ])[0];
        storage.put("voteId", voteId);
    }

    initProducer(proID, proPubkey) {
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
        blockchain.callWithAuth("vote.iost", "addOption", [
            voteId,
            proID,
            false
        ]);

        const pro = {
            "pubkey" : proPubkey,
            "loc": "",
            "url": "",
            "netId": "",
            "isProducer": true,
            "status": STATUS_APPROVED,
            "online": true,
        }; 
        this._mapPut("producerTable", proID, pro, proID);
        this._mapPut("producerKeyToId", proPubkey, proID, proID);
        this._addToProducerMap(proID, pro);
    }

    initAdmin(adminID) {
        const bn = block.number;
        if(bn !== 0) {
            throw new Error("init out of genesis block")
        }
        storage.put("adminID", adminID);
    }

    can_update(data) {
        const admin = storage.get("adminID");
        this._requireAuth(admin, ADMIN_PERMISSION);
        return true;
    }

    _requireAuth(account, permission) {
        const ret = blockchain.requireAuth(account, permission);
        if (ret !== true) {
            throw new Error("require auth failed. ret = " + ret);
        }
    }

    // call abi and parse result as JSON
    _call(contract, api, args) {
        const ret = blockchain.callWithAuth(contract, api, args);
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
        return storage.get("voteId");
    }

    // register account as a producer
    applyRegister(account, pubkey, loc, url, netId, isProducer) {
        this._requireAuth(account, VOTE_PERMISSION);
        if (storage.mapHas("producerTable", account)) {
            throw new Error("producer exists");
        }
        if (storage.mapHas("producerKeyToId", pubkey)) {
            throw new Error("pubkey is used by another producer");
        }

        const publisher = blockchain.publisher();
        this._mapPut("producerTable", account, {
            "pubkey" : pubkey,
            "loc": loc,
            "url": url,
            "netId": netId,
            "isProducer": isProducer,
            "status": STATUS_APPLY,
            "online": false,
        }, publisher);
        this._mapPut("producerKeyToId", pubkey, account, publisher);

        const voteId = this._getVoteId();
        blockchain.callWithAuth("vote.iost", "addOption", [
            voteId,
            account,
            false
        ]);
    }

    // apply remove account from producer list
    applyUnregister(account) {
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
        this._mapPut("producerTable", account, pro, blockchain.publisher());
    }

    // approve account as a producer
    approveRegister(account) {
        const admin = storage.get("adminID");
        this._requireAuth(admin, ADMIN_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        pro.status = STATUS_APPROVED;
        this._mapPut("producerTable", account, pro);
        this._removeFromWaitList(admin, account);
        this._initCandidateVars(admin, account, this._getVoteId(), pro);
        this._addToProducerMap(account, pro);
    }

    // approve remove account from producer list
    approveUnregister(account) {
        const admin = storage.get("adminID");
        this._requireAuth(admin, ADMIN_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        if (pro.status !== STATUS_UNAPPLY) {
            throw new Error("producer not unapplied");
        }
        // will clear votes and score of the producer on stat
        this._clearCandidateVars(admin, account, this._getVoteId(), pro);
        pro.status = STATUS_UNAPPLY_APPROVED;
        this._mapPut("producerTable", account, pro);
        this._tryRemoveProducer(admin, account, pro);
        this._removeFromProducerMap(account, pro);
    }

    // force approve remove account from producer list
    forceUnregister(account) {
        const admin = storage.get("adminID");
        this._requireAuth(admin, ADMIN_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        // will clear votes and score of the producer on stat
        pro.status = STATUS_UNAPPLY_APPROVED;
        this._mapPut("producerTable", account, pro);
        this._tryRemoveProducer(admin, account, pro);
        this._clearCandidateVars(admin, account, this._getVoteId());
        this._removeFromProducerMap(account, pro);
    }

    unregister(account) {
        this._requireAuth(account, VOTE_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        if (pro.status !== STATUS_UNAPPLY_APPROVED) {
            throw new Error("producer can not unregister");
        }
        const voteId = this._getVoteId();
        blockchain.callWithAuth("vote.iost", "removeOption", [
            voteId,
            account,
            true,
        ]);
        // will clear votes and score of the producer on stat
        this._doRemoveProducer(account, pro.pubkey, true);
    }

    _tryRemoveProducer(admin, account, pro) {
        const currentList = this._get("currentProducerList");
        const pendingList = this._get("pendingProducerList");
        if (currentList.includes(pro.pubkey) || pendingList.includes(pro.pubkey)) {
            this._waitRemoveProducer(admin, account);
        } else {
            let scores = this._getScores();
            if (scores[account] !== undefined) {
                delete(scores[account]);
                this._putScores(scores);
            }
        }
    }

    _removeFromWaitList(admin, account) {
        let waitList = this._get("waitingRemoveList") || [];
        const idx = waitList.indexOf(account);
        if (idx !== -1) {
            waitList.splice(idx, 1);
            this._put("waitingRemoveList", waitList, admin);
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

    _addToProducerMap(account, pro) {
        const producerMap = this._get("producerMap") || {};
        const producerKeyMap = this._get("producerKeyMap") || {};
        producerMap[account] = {
            pubkey : pro.pubkey,
            isProducer: pro.isProducer,
            status: pro.status,
            online: pro.online,
        };
        producerKeyMap[pro.pubkey] = account;
        this._put("producerMap", producerMap);
        this._put("producerKeyMap", producerKeyMap);
    }

    _removeFromProducerMap(account, pro) {
        const producerMap = this._get("producerMap") || {};
        const producerKeyMap = this._get("producerKeyMap") || {};
        delete(producerMap[account]);
        this._put("producerMap", producerMap);

        const pendingProducerList = this._get("pendingProducerList");
        const newProducerKeyMap = {};
        for (const pubkey of pendingProducerList) {
            newProducerKeyMap[pubkey] = producerKeyMap[pubkey];
        }
        for (const acc in producerMap) {
            newProducerKeyMap[producerMap[acc].pubkey] = acc;
        }
        this._put("producerKeyMap", newProducerKeyMap);
    }

    // update the information of a producer
    updateProducer(account, pubkey, loc, url, netId) {
        this._requireAuth(account, VOTE_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        const publisher = blockchain.publisher();
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
            this._mapPut("producerKeyToId", pubkey, account, publisher);
        }
        pro.pubkey = pubkey;
        pro.loc = loc;
        pro.url = url;
        pro.netId = netId;
        this._mapPut("producerTable", account, pro, publisher);
        if (pro.status === STATUS_APPROVED || pro.status === STATUS_UNAPPLY) {
            this._addToProducerMap(account, pro);
        }
    }

    getProducer(account) {
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        if (pro.status === STATUS_APPROVED || pro.status === STATUS_UNAPPLY) {
            const voteId = this._getVoteId();
            pro["voteInfo"] = this._call("vote.iost", "getOption", [
                voteId,
                account
            ]);
        }
        return pro;
    }

    // producer log in as online state
    logInProducer(account) {
        this._requireAuth(account, VOTE_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists, " + account);
        }
        const pro = this._mapGet("producerTable", account);
        pro.online = true;
        this._mapPut("producerTable", account, pro, blockchain.publisher());
        if (pro.status === STATUS_APPROVED || pro.status === STATUS_UNAPPLY) {
            this._addToProducerMap(account, pro);
        }
    }

    // producer log out as offline state
    logOutProducer(account) {
        this._requireAuth(account, VOTE_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        if (this._get("pendingProducerList").includes(account) ||
            this._get("currentProducerList").includes(account)) {
            throw new Error("producer in pending list or in current list, can't logout");
        }
        const pro = this._mapGet("producerTable", account);
        pro.online = false;
        this._mapPut("producerTable", account, pro, blockchain.publisher());
        if (pro.status === STATUS_APPROVED || pro.status === STATUS_UNAPPLY) {
            this._addToProducerMap(account, pro);
        }
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

    _updateVoterMask(voter, producer, amount, payer) {
        let voterCoef = this._getVoterCoef(producer);
        let voterMask = this._getVoterMask(voter, producer);
        voterMask = voterMask.plus(voterCoef.multi(amount));
        this._mapPut(voterMaskPrefix + producer, voter, voterMask.toFixed(), payer);
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

    _updateCandidateMask(account, key, payer) {
        let allKey = this._getCandidateAllKey().plus(key);
        this._put(candidateAllKey, allKey.toFixed());

        let candCoef = this._getCandCoef();
        let candMask = this._getCandMask(account);
        candMask = candMask.plus(candCoef.multi(key));
        this._mapPut(candidateMaskTable, account, candMask.toFixed(), payer);
    }

    _updateCandidateVars(account, amount, voteId, payer, votes, pro) {
        if (typeof pro === 'undefined') {
            pro = this._mapGet("producerTable", account);
        }

        if (pro.status !== STATUS_APPROVED && pro.status !== STATUS_UNAPPLY) {
            return;
        }

        if (typeof votes === 'undefined') {
            votes = new Float64(this._call("vote.iost", "getOption", [
                voteId,
               account,
            ]).votes);
        }

        let threshold = PRE_PRODUCER_THRESHOLD;
        if (!pro.isProducer) {
            threshold = PARTNER_THRESHOLD;
        }

        if (amount.gt("0")) {
            if (votes.lt(threshold)) {
                return;
            }

            if (votes.minus(amount).lt(threshold)) {
                this._updateCandidateMask(account, votes, payer);
            } else {
                this._updateCandidateMask(account, amount, payer);
            }
        } else if (amount.lt("0")) {
            if (votes.lt(threshold)) {
                this._updateCandidateMask(account, votes.negated(), payer);
            } else {
                this._updateCandidateMask(account, amount, payer);
            }
        }
    }

    _clearCandidateVars(admin, account, voteId, pro) {
        let votes = new Float64(this._call("vote.iost", "getOption", [
           voteId,
           account,
        ]).votes);
        if (votes && votes.isPositive()) {
            this._updateCandidateVars(account, votes.negated(), voteId, admin, votes, pro);
        }
    }

    _initCandidateVars(admin, account, voteId, pro) {
        let votes = new Float64(this._call("vote.iost", "getOption", [
           voteId,
           account,
        ]).votes);
        if (votes && votes.isPositive()) {
            this._updateCandidateVars(account, votes, voteId, admin, votes);
        }
    }

    _fixAmount(amount) {
        amount = new Float64(new Float64(amount).toFixed(IOST_DECIMAL));
        if (amount.lte("0")) {
            throw new Error("amount must be positive");
        }
        return amount;
    }

    voteFor(payer, voter, producer, amount) {
        this._requireAuth(payer, ACTIVE_PERMISSION);

        if (!storage.mapHas("producerTable", producer)) {
            throw new Error("producer not exists");
        }

        amount = this._fixAmount(amount);

        const voteId = this._getVoteId();
        blockchain.callWithAuth("vote.iost", "voteFor", [
            voteId,
            payer,
            voter,
            producer,
            amount.toFixed(),
        ]);

        this._updateVoterMask(voter, producer, amount, payer);
        this._updateCandidateVars(producer, amount, voteId, payer);
    }

    vote(voter, producer, amount) {
        this._requireAuth(voter, ACTIVE_PERMISSION);

        if (!storage.mapHas("producerTable", producer)) {
            throw new Error("producer not exists");
        }

        amount = this._fixAmount(amount);

        const voteId = this._getVoteId();
        blockchain.callWithAuth("vote.iost", "vote", [
            voteId,
            voter,
            producer,
            amount.toFixed(),
        ]);

        this._updateVoterMask(voter, producer, amount, voter);
        this._updateCandidateVars(producer, amount, voteId, voter);
    }

    unvote(voter, producer, amount) {
        this._requireAuth(voter, ACTIVE_PERMISSION);

        amount = this._fixAmount(amount);

        const voteId = this._getVoteId();
        blockchain.callWithAuth("vote.iost", "unvote", [
            voteId,
            voter,
            producer,
            amount.toFixed(),
        ]);

        this._updateVoterMask(voter, producer, amount.negated(), voter);
        this._updateCandidateVars(producer, amount.negated(), voteId, voter);
    }

    getVote(voter) {
        const voteId = this._getVoteId();
        return this._call("vote.iost", "getVote", [
            voteId,
            voter
        ]);
    }

    topupVoterBonus(account, amount, payer) {
        const voteId = this._getVoteId();
        let votes = new Float64(this._call("vote.iost", "getOption", [
            voteId,
            account,
        ]).votes);
        if (votes.lte("0")) {
            return false;
        }

        amount = this._fixAmount(amount);

        blockchain.deposit(payer, amount.toFixed(), "");

        let voterCoef = this._getVoterCoef(account);
        voterCoef = voterCoef.plus(amount.div(votes));
        this._mapPut(voterCoefTable, account, voterCoef.toFixed(), payer);
        return true;
    }

    topupCandidateBonus(amount, payer) {
        let allKey = this._getCandidateAllKey();
        if (allKey.lte("0")) {
            return false;
        }

        amount = this._fixAmount(amount);

        blockchain.deposit(payer, amount.toFixed(), "");

        let candCoef = this._getCandCoef();
        candCoef = candCoef.plus(amount.div(allKey));
        this._put(candidateCoef, candCoef.toFixed(), payer);
        return true;
    }

    _calVoterBonus(voter, updateMask) {
        let userVotes = this.getVote(voter);
        let earnings = new Float64(0);
        for (const v of userVotes) {
            let voterCoef = this._getVoterCoef(v.option);
            let voterMask = this._getVoterMask(voter, v.option);
            let earning = voterCoef.multi(v.votes).minus(voterMask);
            earnings = earnings.plus(earning);
            if (updateMask) {
                voterMask = voterMask.plus(earning);
                this._mapPut(voterMaskPrefix + v.option, voter, voterMask.toFixed(), voter);
            }
        }
        return earnings;
    }

    getVoterBonus(voter) {
        return this._calVoterBonus(voter, false).toFixed(IOST_DECIMAL);
    }

    voterWithdraw(voter) {
        this._requireAuth(voter, ACTIVE_PERMISSION);

        let earnings = this._calVoterBonus(voter, true);
        if (earnings.lte("0")) {
            return;
        }
        blockchain.withdraw(voter, earnings.toFixed(IOST_DECIMAL), "");
    }

    _calCandidateBonus(account, updateMask) {
        const voteId = this._getVoteId();
        let candKey = new Float64(this._call("vote.iost", "getOption", [
            voteId,
            account,
        ]).votes);

        if (candKey.lt(PRE_PRODUCER_THRESHOLD)) {
            candKey = new Float64(0);
        }

        let candCoef = this._getCandCoef();
        let candMask = this._getCandMask(account);
        let earning = candCoef.multi(candKey).minus(candMask);
        if (updateMask) {
            candMask = candMask.plus(earning);
            this._mapPut(candidateMaskTable, account, candMask.toFixed(), account);
        }
        return earning;
    }

    getCandidateBonus(account) {
        return this._calCandidateBonus(account, false).toFixed(IOST_DECIMAL);
    }

    candidateWithdraw(account) {
        this._requireAuth(account, ACTIVE_PERMISSION);

        let earnings = this._calCandidateBonus(account, true);
        if (earnings.lte("0")) {
            return;
        }
        let halfEarning = earnings.div("2");
        blockchain.withdraw(account, halfEarning.toFixed(IOST_DECIMAL), "");

        this.topupVoterBonus(account, earnings.minus(halfEarning.toFixed(IOST_DECIMAL)).toFixed(IOST_DECIMAL), blockchain.contractName());
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
    stat() {
        this._requireAuth("base.iost", ACTIVE_PERMISSION);
        const bn = block.number;
        const pendingBlockNumber = this._get("pendingBlockNumber");
        if (bn % VOTE_STAT_INTERVAL !== 0 || bn <= pendingBlockNumber) {
            return;
        }

        const voteId = this._getVoteId();
        const voteRes = this._call("vote.iost", "getResult", [voteId]);
        const preList = [];    // list of producers whose vote > threshold
        const waitingRemoveList = this._get("waitingRemoveList") || [];
        let scores = this._getScores();
        const pendingProducerList = this._get("pendingProducerList");
        const producerMap = this._get("producerMap") || {};
        const producerKeyMap = this._get("producerKeyMap") || {};
        const validPendingMap = {};

        // update scores
        let scoreTotal = new Float64("0");
        let scoreCount = 0;
        for (const res of voteRes) {
            const id = res.option;
            const pro = producerMap[id];
            if (!pro || !pro.online || !pro.isProducer || (pro.status !== STATUS_APPROVED && pro.status !== STATUS_UNAPPLY)) {
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
            } else {
                validPendingMap[pro.pubkey] = true;
            }
        }

        // update pending list
        let oldPreList = [];
        const oldPreListToRemove = [];
        let minScore = new Float64(MaxFloat64);
        for (const key of pendingProducerList) {
            const account = producerKeyMap[key];
            const score = new Float64(scores[account] || "0");
            if (waitingRemoveList.includes(account) || !validPendingMap[key]) {
                oldPreListToRemove.push({
                    "account": account,
                    "key": key,
                    "prior": 0,
                    "score": score
                });
                minScore = new Float64(0);
                delete(scores[account]);
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
            const scoreAvg = scoreTotal.div(scoreCount*10);
            for (const key of pendingList) {
                const account = producerKeyMap[key];
                const score = new Float64(scores[account] || "0").minus(scoreAvg);
                if (score.gte("0")) {
                    scores[account] = score.toFixed(IOST_DECIMAL);
                }
            }
        } else {
            for (const key of pendingList) {
                const account = producerKeyMap[key];
                delete(scores[account]);
            }
        }

        for (const removed of removedList) {
            if (!waitingRemoveList.includes(removed.account)) {
                continue;
            }
            delete(scores[removed.account]);
        }
        const newWaitingRemoveList = waitingRemoveList.filter(function(value, index, arr) {
            return !removedList.includes(value);
        });
        this._put("waitingRemoveList", newWaitingRemoveList);

        if (bn % SCORE_DECREASE_INTERVAL === 0) {
            for (const acc in scores) {
                scores[acc] = new Float64(scores[acc]).div("2").toFixed(IOST_DECIMAL);
            }
        }

        this._putScores(scores);
    }
}

module.exports = VoteContract;
