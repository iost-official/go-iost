const producerRegisterFee = "200000000";
const preProducerThreshold = "210000000";
const voteLockTime = 864000;
const voteStatInterval = 200;
const iostDecimal = 8;
const scoreDecreaseRate = new Float64("0.999995");
const producerPermission = "active";
const votePermission = "vote";

class VoteContract {
    constructor() {
    }

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
                minVote: preProducerThreshold,
                options: [],
                anyOption: false,
                unvoteInterval: voteLockTime
            }
        ]);
        this._put("voteId", JSON.stringify(voteId));
    }

    InitProducer(proID, proPubkey) {
        const bn = this._getBlockNumber();
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

        this._call("token.iost", "transfer", ["iost", proID, "vote_producer.iost", producerRegisterFee, ""]);

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
            "online": true,
            "registerFee": producerRegisterFee,
        }, proID);
        this._mapPut("producerKeyToId", proPubkey, proID, proID);
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
        this._requireAuth(admin, producerPermission);
        return true;
    }

    _requireAuth(account, permission) {
        BlockChain.requireAuth(account, permission);
    }

    _call(contract, api, args) {
        const ret = BlockChain.callWithAuth(contract, api, JSON.stringify(args));
        if (ret && Array.isArray(ret) && ret.length === 1) {
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

    // register account as a producer, need to pledge token
    RegisterProducer(account, pubkey, loc, url, netId) {

        this._requireAuth(account, producerPermission);
        if (storage.mapHas("producerTable", account)) {
            throw new Error("producer exists");
        }
        if (storage.mapHas("producerKeyToId", pubkey)) {
            throw new Error("pubkey is used by another producer");
        }

        this._call("token.iost", "transfer", ["iost", account, "vote_producer.iost", producerRegisterFee, ""]);

        const voteId = this._getVoteId();
        this._call("vote.iost", "AddOption", [
            voteId,
            account,
            false
        ]);

        this._mapPut("producerTable", account, {
            "pubkey" : pubkey,
            "loc": loc,
            "url": url,
            "netId": netId,
            "online": false,
            "registerFee": producerRegisterFee,
        }, account);
        this._mapPut("producerKeyToId", pubkey, account, account);
    }

    // update the information of a producer
    UpdateProducer(account, pubkey, loc, url, netId) {
        this._requireAuth(account, producerPermission);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        if (pro.pubkey !== pubkey) {
            if (storage.mapHas("producerKeyToId", pubkey)) {
                throw new Error("pubkey is used by another producer");
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
        const voteId = this._getVoteId();
        pro["voteInfo"] = this._call("vote.iost", "GetOption", [
            voteId,
            account
        ]);
        return pro;
    }

    // producer log in as online state
    LogInProducer(account) {
        this._requireAuth(account, producerPermission);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists, " + account);
        }
        const pro = this._mapGet("producerTable", account);
        pro.online = true;
        this._mapPut("producerTable", account, pro, account);
    }

    // producer log out as offline state
    LogOutProducer(account) {
        this._requireAuth(account, producerPermission);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        if (this._get("pendingProducerList").includes(account) ||
            this._get("currentProducerList").includes(account)) {
            throw new Error("producer in pending list or in current list, can't logout");
        }
        const pro = this._mapGet("producerTable", account);
        pro.online = false;
        this._mapPut("producerTable", account, pro, account);
    }

    // remove account from producer list
    UnregisterProducer(account) {
        this._requireAuth(account, producerPermission);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        if (this._get("pendingProducerList").includes(account) ||
            this._get("currentProducerList").includes(account)) {
            throw new Error("producer in pending list or in current list, can't unregist");
        }
        const voteId = this._getVoteId();
        this._call("vote.iost", "RemoveOption", [
            voteId,
            account,
            true,
        ]);
        // will clear votes and score of the producer

        const pro = this._mapGet("producerTable", account);
        this._mapDel("producerTable", account);
        this._mapDel("preProducerMap", account);
        this._mapDel("producerKeyToId", pro.pubkey, account);

        this._call("token.iost", "transfer", ["iost", "vote_producer.iost", account, pro.registerFee, ""]);
    }

    // vote, need to pledge token
    Vote(voter, producer, amount) {
        this._requireAuth(voter, votePermission);

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
        this._requireAuth(voter, votePermission);
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
        // controll auth
        const bn = this._getBlockNumber();
        const pendingBlockNumber = this._get("pendingBlockNumber");
        if (bn % voteStatInterval !== 0 || bn <= pendingBlockNumber) {
            return;
        }

        const voteId = this._getVoteId();
        const voteRes = this._call("vote.iost", "GetResult", [voteId]);
        const preList = [];    // list of producers whose vote > threshold
        let scores = this._getScores();

        const pendingProducerList = this._get("pendingProducerList");

        const ppThreshold = new Float64(preProducerThreshold);
        for (const res of voteRes) {
            const id = res.option;
            const pro = this._mapGet("producerTable", id);
            // don't get score if in pending producer list or offline
            const votes = new Float64(res.votes);
            if (!pendingProducerList.includes(pro.pubkey) &&
                !votes.lt(ppThreshold) &&
                pro.online === true) {
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
            const delta = preList[i].votes.minus(ppThreshold);
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
        const producerNumber = this._get("producerNumber");
        const replaceNum = Math.min(preList.length, Math.floor(producerNumber / 6));
        const maxInsertPlace = Math.floor(producerNumber * 2 / 3);
        const oldPreList = [];
        let minScore = new Float64(MaxFloat64);
        for (const key of pendingProducerList) {
            const account = this._mapGet("producerKeyToId", key);
            const score = new Float64(scores[account] ? scores[account] : "0");
            oldPreList.push({
                "key": key,
                "prior": 1,
                "score": score
            });
            if (score.lt(minScore)) {
                minScore = score;
            }
        }

        // replace at most replaceNum producers
        for (let i = replaceNum - 1; i >= 0; i--) {
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
        const newList = oldPreList.slice(0, producerNumber);

        const currentList = pendingProducerList;
        const pendingList = newList.map(x => x.key);
        this._put("currentProducerList", currentList);
        this._put("pendingProducerList", pendingList);
        this._put("pendingBlockNumber", this._getBlockNumber());

        for (const key of currentList) {
            if (!pendingList.includes(key)) {
                const account = this._mapGet("producerKeyToId", key);
                scores[account] = "0";
            }
        }

        for (const key of pendingList) {
            const account = this._mapGet("producerKeyToId", key);
            const origScore = scores[account] ? scores[account] : "0";
            scores[account] = new Float64(origScore).multi(scoreDecreaseRate).toFixed(iostDecimal);
        }
        this._putScores(scores);
    }
}

module.exports = VoteContract;
