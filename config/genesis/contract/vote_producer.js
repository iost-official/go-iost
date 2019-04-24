const VOTE_THRESHOLD = "2100000";
const VOTE_LOCKTIME = 604800;
const VOTE_STAT_INTERVAL = 1200;
const SCORE_DECREASE_INTERVAL = 31104000;
const IOST_DECIMAL = 8;
const ADMIN_PERMISSION = "active";
const VOTE_PERMISSION = "vote";
const ACTIVE_PERMISSION = "active";
const WITHDRAW_PERMISSION = "operate";

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
        this._initVote();
    }

    _initVote() {
        const voteId = blockchain.callWithAuth("vote.iost", "newVote", [
            "vote_producer.iost",
            "vote for producer",
            {
                resultNumber: 2000,
                minVote: VOTE_THRESHOLD,
                options: [],
                anyOption: false,
                freezeTime: VOTE_LOCKTIME,
                canVote: false,
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

    _requireAuthList(accountList, permission) {
        for (const account of accountList) {
            if (blockchain.requireAuth(account, permission)) {
                return
            }
        }
        throw new Error("require auth failed.");
    }

    _getAccountList(account) {
        return [account, storage.get("adminID"), "operator"];
    }

    // call abi and parse result as JSON
    _call(contract, api, args) {
        const ret = blockchain.callWithAuth(contract, api, args);
        if (ret && Array.isArray(ret) && ret.length >= 1) {
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
        this._requireAuthList(this._getAccountList(account), VOTE_PERMISSION);
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
        blockchain.receipt(JSON.stringify([account, pubkey, loc, url, netId, isProducer]));
    }

    // apply remove account from producer list
    applyUnregister(account) {
        this._requireAuthList(this._getAccountList(account), VOTE_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        if (pro.status === STATUS_APPLY) {
            return;
        }
        if (pro.status !== STATUS_APPROVED) {
            throw new Error("producer not approved");
        }
        pro.status = STATUS_UNAPPLY;
        this._mapPut("producerTable", account, pro, blockchain.publisher());
        blockchain.receipt(JSON.stringify([account]));
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
        this._addToProducerMap(account, pro);
        blockchain.receipt(JSON.stringify([account]));
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
        pro.status = STATUS_UNAPPLY_APPROVED;
        this._mapPut("producerTable", account, pro);
        this._removeFromProducerMap(account, pro);
        blockchain.receipt(JSON.stringify([account]));
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
        this._removeFromProducerMap(account, pro);
        blockchain.receipt(JSON.stringify([account]));
    }

    unregister(account) {
        this._requireAuthList(this._getAccountList(account), VOTE_PERMISSION);
        const pro = this._mapGet("producerTable", account);
        if (pro && pro.status !== STATUS_APPLY && pro.status !== STATUS_UNAPPLY_APPROVED) {
            throw new Error("producer can not unregister");
        }
        if (this._inCurrentOrPendingList(account)) {
            throw new Error("producer in pending list or in current list, can't unregister");
        }
        const voteId = this._getVoteId();
        blockchain.callWithAuth("vote.iost", "removeOption", [
            voteId,
            account,
            true,
        ]);
        if (pro) {
            // will clear votes and score of the producer on stat
            this._doRemoveProducer(account, pro.pubkey);
        }
        blockchain.receipt(JSON.stringify([account]));
    }

    _inCurrentOrPendingList(account) {
        const producerKeyMap = this._get("producerKeyMap") || {};
        const pendingList =  this._get("pendingProducerList");
        for (const key of pendingList) {
            if (producerKeyMap[key] === account) {
                return true;
            }
        }
        const currentList = this._get("currentProducerList");
        for (const key of currentList) {
            if (producerKeyMap[key] === account) {
                return true;
            }
        }
        return false;
    }

    _doRemoveProducer(account, pubkey) {
        this._mapDel("producerTable", account);
        this._mapDel("producerKeyToId", pubkey);
        this._removeFromProducerMap(account);
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
        const pendingProducerList = this._get("pendingProducerList");
        const newProducerKeyMap = {};
        let inPending = false;
        for (const pubkey of pendingProducerList) {
            newProducerKeyMap[pubkey] = producerKeyMap[pubkey];
            if (producerKeyMap[pubkey] == account) {
                inPending = true;
            }
        }
        if (!inPending) {
            delete(producerMap[account]);
        } else if (pro !== undefined) {
            producerMap[account] = {
                pubkey : pro.pubkey,
                isProducer: pro.isProducer,
                status: pro.status,
                online: pro.online,
            };
        }
        for (const acc in producerMap) {
            newProducerKeyMap[producerMap[acc].pubkey] = acc;
        }
        this._put("producerMap", producerMap);
        this._put("producerKeyMap", newProducerKeyMap);
    }

    // update the information of a producer
    updateProducer(account, pubkey, loc, url, netId) {
        this._requireAuthList(this._getAccountList(account), VOTE_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        const publisher = blockchain.publisher();
        if (pro.pubkey !== pubkey) {
            if (storage.mapHas("producerKeyToId", pubkey)) {
                throw new Error("pubkey is used by another producer");
            }
            if (this._inCurrentOrPendingList(account)) {
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
        blockchain.receipt(JSON.stringify([account, pubkey, loc, url, netId]));
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
        this._requireAuthList(this._getAccountList(account), VOTE_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists, " + account);
        }
        const pro = this._mapGet("producerTable", account);
        pro.online = true;
        this._mapPut("producerTable", account, pro, blockchain.publisher());
        if (pro.status === STATUS_APPROVED || pro.status === STATUS_UNAPPLY) {
            this._addToProducerMap(account, pro);
        }
        blockchain.receipt(JSON.stringify([account]));
    }

    // producer log out as offline state
    logOutProducer(account) {
        this._requireAuthList(this._getAccountList(account), VOTE_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        pro.online = false;
        this._mapPut("producerTable", account, pro, blockchain.publisher());
        if (pro.status === STATUS_APPROVED || pro.status === STATUS_UNAPPLY) {
            this._addToProducerMap(account, pro);
        }
        blockchain.receipt(JSON.stringify([account]));
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

    _updateCandidateVars(account, amount, voteId, payer) {
        let votes = new Float64(this._call("vote.iost", "getOption", [
            voteId,
            account,
        ]).votes);

        if (amount.gt("0")) {
            if (votes.lt(VOTE_THRESHOLD)) {
                return;
            }

            if (votes.minus(amount).lt(VOTE_THRESHOLD)) {
                this._updateCandidateMask(account, votes, payer);
            } else {
                this._updateCandidateMask(account, amount, payer);
            }
        } else if (amount.lt("0")) {
            if (votes.minus(amount).lt(VOTE_THRESHOLD)){
                return;
            }

            if (votes.lt(VOTE_THRESHOLD)) {
                this._updateCandidateMask(account, votes.minus(amount).negated(), payer);
            } else {
                this._updateCandidateMask(account, amount, payer);
            }
        }
    }

    _fixAmount(amount) {
        amount = new Float64(new Float64(amount).toFixed(IOST_DECIMAL));
        if (amount.lte("0")) {
            throw new Error("amount must be positive");
        }
        return amount;
    }

    _checkSwitchOff() {
        return storage.get("switchOff") === "1";
    }

    switchOff(off) {
        storage.put("switchOff", off ? "1" : "0");
    }

    vote(voter, producer, amount) {
        this._requireAuth(voter, ACTIVE_PERMISSION);
        if (this._checkSwitchOff()) {
            throw new Error("can't vote for now");
        }

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
        blockchain.receipt(JSON.stringify([voter, producer, amount]));
    }

    unvote(voter, producer, amount) {
        this._requireAuth(voter, ACTIVE_PERMISSION);
        if (this._checkSwitchOff()) {
            throw new Error("can't unvote for now");
        }

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
        blockchain.receipt(JSON.stringify([voter, producer, amount]));
    }

    getVote(voter) {
        const voteId = this._getVoteId();
        return this._call("vote.iost", "getVote", [
            voteId,
            voter
        ]);
    }

    _topupVoterBonusInternal(account, amount, payer) {
        const voteId = this._getVoteId();
        let votes = new Float64(this._call("vote.iost", "getOption", [
            voteId,
            account,
        ]).votes);
        if (votes.lte("0")) {
            return false;
        }

        blockchain.deposit(payer, amount.toFixed(), "");

        let voterCoef = this._getVoterCoef(account);
        voterCoef = voterCoef.plus(amount.div(votes));
        this._mapPut(voterCoefTable, account, voterCoef.toFixed(), payer);
        return true;
    }

    topupVoterBonus(account, amount, payer) {
        if (this._checkSwitchOff()) {
            throw new Error("can't topup for now");
        }
        amount = this._fixAmount(amount);
        if (this._topupVoterBonusInternal(account, amount, payer)) {
            blockchain.receipt(JSON.stringify([account, amount, payer]));
            return true;
        }
        return false;
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
        blockchain.receipt(JSON.stringify([amount, payer]));
        return true;
    }

    _calVoterBonus(voter, updateMask) {
        let userVotes = this.getVote(voter);
        let earnings = new Float64(0);
        let receipt = {}
        for (const v of userVotes) {
            let voterCoef = this._getVoterCoef(v.option);
            let voterMask = this._getVoterMask(voter, v.option);
            let earning = voterCoef.multi(v.votes).minus(voterMask);
            earnings = earnings.plus(earning);
            if (updateMask) {
                voterMask = voterMask.plus(earning);
                this._mapPut(voterMaskPrefix + v.option, voter, voterMask.toFixed(), blockchain.publisher());
            }
            if (earning.gt("0")){
                receipt[v.option] = earning.toFixed(IOST_DECIMAL);
            }
        }
        let r = JSON.stringify(receipt)
        if ( r !== '{}' ) {
            blockchain.receipt(JSON.stringify(receipt))
        }

        return earnings;
    }

    getVoterBonus(voter) {
        return this._calVoterBonus(voter, false).toFixed(IOST_DECIMAL);
    }

    voterWithdraw(voter) {
        this._requireAuthList(this._getAccountList(voter), WITHDRAW_PERMISSION);
        if (this._checkSwitchOff()) {
            throw new Error("can't withdraw for now");
        }

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

        if (candKey.lt(VOTE_THRESHOLD)) {
            candKey = new Float64(0);
        }

        let candCoef = this._getCandCoef();
        let candMask = this._getCandMask(account);
        let earning = candCoef.multi(candKey).minus(candMask);
        if (updateMask) {
            candMask = candMask.plus(earning);
            this._mapPut(candidateMaskTable, account, candMask.toFixed(), blockchain.publisher());
        }
        return earning;
    }

    getCandidateBonus(account) {
        return this._calCandidateBonus(account, false).toFixed(IOST_DECIMAL);
    }

    candidateWithdraw(account) {
        this._requireAuthList(this._getAccountList(account), WITHDRAW_PERMISSION);
        if (this._checkSwitchOff()) {
            throw new Error("can't withdraw for now");
        }

        let earnings = this._calCandidateBonus(account, true);
        if (earnings.lte("0")) {
            return;
        }
        let halfEarning = earnings.div("2");
        let candidateBonus = halfEarning.toFixed(IOST_DECIMAL);
        let voterBonus = earnings.minus(candidateBonus).toFixed(IOST_DECIMAL);
        blockchain.withdraw(account, candidateBonus, "");

        this._topupVoterBonusInternal(account, new Float64(voterBonus), blockchain.contractName());
        blockchain.receipt(JSON.stringify([account, candidateBonus, voterBonus]));
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

    _getWitnessWatched() {
        const witnessWatched = this._get("witnessWatched") || {};
        for (const witness in witnessWatched) {
            if (witnessWatched[witness].bn + 144*VOTE_STAT_INTERVAL < block.number) {
                delete(witnessWatched[witness]);
            }
        }
        return witnessWatched;
    }

    _getWitnessPenality() {
        const witnessPenality = this._get("witnessPenality") || {};
        for (const witness in witnessPenality) {
            if (witnessPenality[witness] < block.number) {
                delete(witnessPenality[witness]);
            }
        }
        return witnessPenality;
    }

    _updateWitnessWatched(witnessWatched, account, produced) {
        if (produced) {
            delete(witnessWatched[account]);
            return;
        }
        if (!witnessWatched[account]) {
            witnessWatched[account] = {
                bn: block.number,
                count: 0,
            };
        }
        witnessWatched[account].count++;
    }

    _updateWitnessPenality(witnessPenality, witnessWatched, account) {
        if (witnessWatched[account] && witnessWatched[account].count >= 3) {
            witnessPenality[account] = block.number + 144 * VOTE_STAT_INTERVAL;
            delete(witnessWatched[account]);
        }
        return witnessPenality[account] || 0;
    }

    // calculate the vote result, modify pendingProducerList
    stat() {
        this._requireAuth("base.iost", ACTIVE_PERMISSION);
        const bn = block.number;
        if (bn % VOTE_STAT_INTERVAL !== 0) {
            return;
        }

        const voteId = this._getVoteId();
        const voteRes = this._call("vote.iost", "getResult", [voteId]);
        const preList = [];    // list of producers whose vote > threshold
        const witnessProduced = JSON.parse(storage.globalGet("base.iost", "witness_produced") || '{}');
        const witnessWatched = this._getWitnessWatched();
        const witnessPenality = this._getWitnessPenality();
        const pendingProducerList = this._get("pendingProducerList");
        const currentProducerList = this._get("currentProducerList");
        const producerMap = this._get("producerMap") || {};
        const producerKeyMap = this._get("producerKeyMap") || {};
        const validPendingMap = {};

        // update scores
        let scoreTotal = new BigNumber("0");
        let scoreCount = 0;
        let pendingIdMap = {};
        let scores = this._getScores();
        let newScores = {};
        for (const key of pendingProducerList) {
            const account = producerKeyMap[key];
            pendingIdMap[account] = true;
        }
        for (const res of voteRes) {
            const id = res.option;
            const pro = producerMap[id];
            if (!pro || !pro.isProducer || (pro.status !== STATUS_APPROVED && pro.status !== STATUS_UNAPPLY)) {
                continue;
            }
            const forbidUntil = witnessPenality[id] || 0;
            const isNormal = pro.online && forbidUntil < bn;
            const incScore = new BigNumber(isNormal ? res.votes : "0");
            const score = incScore.plus(scores[id] || "0");
            scoreTotal = scoreTotal.plus(score);
            scoreCount++;
            newScores[id]  = score.toFixed();

            if (!isNormal) {
                continue;
            }
            if (!pendingIdMap[id]) {
                preList.push({
                    key: pro.pubkey,
                    prior: 0,
                    score: score,
                });
            } else {
                validPendingMap[pro.pubkey] = true;
            }
        }

        // delete score if votes < threshold
        scores = newScores;

        // update pending list
        let oldPreList = [];
        const oldPreListToRemove = [];
        for (const key of pendingProducerList) {
            const account = producerKeyMap[key];
            const pro = producerMap[account];
            const score = new BigNumber(scores[account] || "0");
            this._updateWitnessWatched(witnessWatched, account, witnessProduced[key]);
            const forbidUntil = this._updateWitnessPenality(witnessPenality, witnessWatched, account);

            const pinfo = {
                key: pro.pubkey,
                prior: 0,
                score: new BigNumber("0"),
            };
            const isDead = currentProducerList.includes(key) && !witnessProduced[key];
            if (!pro.online || isDead || forbidUntil >= bn || !validPendingMap[pro.pubkey]) {
                oldPreListToRemove.push(pinfo);
                if (isDead) {
                    scores[account] = score.div("2").toFixed(IOST_DECIMAL);
                }
            } else {
                pinfo.prior = 1;
                pinfo.score = score;
                oldPreList.push(pinfo);
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

        oldPreList.splice(producerNumber);
        const newList = oldPreList;

        const currentList = pendingProducerList;
        const pendingList = newList.map(x => x.key);
        this._put("currentProducerList", currentList);
        this._put("pendingProducerList", pendingList);
        this._put("witnessWatched", witnessWatched);
        this._put("witnessPenality", witnessPenality);

        // decrease scores in producer list
        const scoreAvg = scoreTotal.div((scoreCount || producerNumber) * 10);
        for (const key of pendingList) {
            const account = producerKeyMap[key];
            const score = new BigNumber(scores[account] || "0").minus(scoreAvg);
            if (score.gte("0")) {
                scores[account] = score.toFixed(IOST_DECIMAL);
            } else {
                delete(scores[account]);
            }
        }

        if (bn % SCORE_DECREASE_INTERVAL === 0) {
            for (const acc in scores) {
                scores[acc] = new BigNumber(scores[acc]).div("2").toFixed(IOST_DECIMAL);
            }
        }

        this._putScores(scores);
    }
}

module.exports = VoteContract;
