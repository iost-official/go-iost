const PRE_PRODUCER_THRESHOLD = "1";
const VOTE_LOCKTIME = 2592000;
const VOTE_STAT_INTERVAL = 200;
const IOST_DECIMAL = 8;
const SCORE_DECREASE_RATE = new Float64("0.9");
const ADMIN_PERMISSION = "active";
const PRODUCER_PERMISSION = "active";
const VOTE_PERMISSION = "active";
const STAT_PERMISSION = "active";

const STATUS_APPLY = 0;
const STATUS_APPROVED = 1;
const STATUS_UNAPPLY = 2;
const STATUS_UNAPPLY_APPROVED = 3;

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
        this._requireAuth(account, PRODUCER_PERMISSION);
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
    }

    // apply remove account from producer list
    ApplyUnregister(account) {
        this._requireAuth(account, PRODUCER_PERMISSION);
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

        const voteId = this._getVoteId();
        this._call("vote.iost", "AddOption", [
            voteId,
            account,
            false
        ]);
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
        const voteId = this._getVoteId();
        this._call("vote.iost", "RemoveOption", [
            voteId,
            account,
            true,
        ]);
        // will clear votes and score of the producer on stat
        pro.status = STATUS_UNAPPLY_APPROVED;
        this._mapPut("producerTable", account, pro);
        this._tryRemoveProducer(admin, account, pro);
    }

    // force approve remove account from producer list
    Unregister(account) {
        const admin = this._get("adminID");
        this._requireAuth(admin, ADMIN_PERMISSION);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        const voteId = this._getVoteId();
        this._call("vote.iost", "RemoveOption", [
            voteId,
            account,
            true,
        ]);
        // will clear votes and score of the producer on stat
        pro.status = STATUS_UNAPPLY_APPROVED;
        this._mapPut("producerTable", account, pro);
        this._tryRemoveProducer(admin, account, pro);
    }

    _tryRemoveProducer(admin, account, pro) {
        const currentList = this._get("currentProducerList");
        const pendingList = this._get("pendingProducerList");
        if (currentList.includes(pro.pubkey) || pendingList.includes(pro.pubkey)) {
            this._waitRemoveProducer(admin, account);
        } else {
            this._doRemoveProducer(account, pro,pubkey);
        }
    }

    _waitRemoveProducer(admin, account) {
        let waitList = this._get("waitingRemoveList") || [];
        if (!waitList.includes(account)) {
            waitList.push(account);
            this._put("waitingRemoveList", waitList, admin);
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
        this._requireAuth(account, PRODUCER_PERMISSION);
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
        this._requireAuth(account, PRODUCER_PERMISSION);
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
        this._requireAuth(account, PRODUCER_PERMISSION);
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

    // vote, need to pledge token
    Vote(voter, producer, amount) {
        this._requireAuth(voter, VOTE_PERMISSION);

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
    }

    // unvote
    Unvote(voter, producer, amount) {
        this._requireAuth(voter, VOTE_PERMISSION);
        const voteId = this._getVoteId();
        this._call("vote.iost", "Unvote", [
            voteId,
            voter,
            producer,
            amount,
        ]);
    }

    GetVote(voter) {
        const voteId = this._getVoteId();
        return this._call("vote.iost", "GetVote", [
            voteId,
            voter
        ]);
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
        this._requireAuth("base.iost", STAT_PERMISSION);
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
        const ppThreshold = new Float64(PRE_PRODUCER_THRESHOLD);
        for (const res of voteRes) {
            const id = res.option;
            const pro = this._mapGet("producerTable", id);
            // don't get score if in pending producer list or offline
            const votes = new Float64(res.votes);
            if (!pendingProducerList.includes(pro.pubkey) && !votes.lt(ppThreshold) &&
                pro.online === true && (pro.status === STATUS_APPROVED || pro.status === STATUS_UNAPPLY)) {
                preList.push({
                    "id" : id,
                    "key": pro.pubkey,
                    "prior": 0,
                    "votes": votes,
                    "score": scores[id] ? scores[id] : "0",
                });
            }
        }
        for (let i = 0; i < preList.length; i++) {
            const id = preList[i].id;
            const delta = preList[i].votes;
            const origScore = scores[id] ? scores[id] : "0";
            preList[i].score = delta.plus(origScore);
            scores[id] = preList[i].score.toFixed();
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
        preList.sort(scoreCmp);

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

        const producerNumber = this._get("producerNumber");
        const replaceNum = Math.floor(producerNumber / 6);
        const realReplaceNum = Math.min(preList.length, Math.max(replaceNum, oldPreListToRemove.length));
        const maxInsertPlace = Math.min(oldPreList.length, Math.floor(producerNumber * 2 / 3));

        // replace realReplaceNum producers
        oldPreList = [...oldPreList, ...oldPreListToRemove];
        for (let i = realReplaceNum - 1; i >= 0; i--) {
            const preProducer = preList[i];
            if (!minScore.lt(preProducer.score)) {
                continue;
            }
            let insertPlace = maxInsertPlace;
            for (let j = maxInsertPlace - 1; j >= 0 ; j--) {
                if (scoreCmp(preProducer, oldPreList[j]) < 0) {
                    insertPlace = j;
                } else {
                    break;
                }
            }
            oldPreList.splice(insertPlace, 0, preProducer);
        }
        const removedList = oldPreList.splice(producerNumber);
        const newList = oldPreList;

        const currentList = pendingProducerList;
        const pendingList = newList.map(x => x.key);
        this._put("currentProducerList", currentList);
        this._put("pendingProducerList", pendingList);
        this._put("pendingBlockNumber", block.number);

        for (const key of currentList) {
            if (!pendingList.includes(key)) {
                const account = this._mapGet("producerKeyToId", key);
                scores[account] = "0";
            }
        }

        for (const key of pendingList) {
            const account = this._mapGet("producerKeyToId", key);
            const origScore = scores[account] ? scores[account] : "0";
            scores[account] = new Float64(origScore).multi(SCORE_DECREASE_RATE).toFixed(IOST_DECIMAL);
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
