const producerRegisterFee = "200000000";
const preProducerThreshold = "210000000";
const voteLockTime = 864000;
const voteStatInterval = 200;
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
        const voteId = this._call("iost.vote", "NewVote", [
            "iost.vote_producer",
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

    InitProducer(proID) {
        const bn = this._getBlockNumber();
        if(bn !== 0) {
            throw new Error("init out of genesis block")
        }

        let pendingProducerList = this._get("pendingProducerList");
        pendingProducerList.push(proID);
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

        this._call("iost.token", "transfer", ["iost", proID, "iost.vote_producer", producerRegisterFee]);

        const voteId = this._getVoteId();
        this._call("iost.vote", "AddOption", [
            voteId,
            proID,
            false
        ]);
        this._mapPut("producerTable", proID, {
            "loc": "",
            "url": "",
            "netId": "",
            "online": true,
            "registerFee": producerRegisterFee,
            "score": "0"
        });
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
        const ret = BlockChain.requireAuth(account, permission);
        if (ret !== true) {
            throw new Error("require auth failed. ret = " + ret);
        }
    }

    _call(contract, api, args) {
        const ret = JSON.parse(BlockChain.call(contract, api, JSON.stringify(args)));
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
        return JSON.parse(storage.get(k));
    }

    _put(k, v) {
        const ret = storage.put(k, JSON.stringify(v));
        if (ret !== 0) {
            throw new Error("storage put failed. ret = " + ret);
        }
    }

    _mapGet(k, f) {
        return JSON.parse(storage.mapGet(k, f));
    }

    _mapPut(k, f, v) {
        const ret = storage.mapPut(k, f, JSON.stringify(v));
        if (ret !== 0) {
            throw new Error("storage map put failed. ret = " + ret);
        }
    }

    _mapDel(k, f) {
        const ret = storage.mapDel(k, f);
        if (ret !== 0) {
            throw new Error("storage map del failed. ret = " + ret);
        }
    }

    _getVoteId() {
        return this._get("voteId");
    }

    // register account as a producer, need to pledge token
    RegisterProducer(account, loc, url, netId) {
        this._requireAuth(account, producerPermission);
        if (storage.mapHas("producerTable", account)) {
            throw new Error("producer exists");
        }

        this._call("iost.token", "transfer", ["iost", account, "iost.vote_producer", producerRegisterFee]);

        const voteId = this._getVoteId();
        this._call("iost.vote", "AddOption", [
            voteId,
            account,
            false
        ]);

        this._mapPut("producerTable", account, {
            "loc": loc,
            "url": url,
            "netId": netId,
            "online": false,
            "registerFee": producerRegisterFee,
            "score": "0"
        });
    }

    // update the information of a producer
    UpdateProducer(account, loc, url, netId) {
        this._requireAuth(account, producerPermission);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists");
        }
        const pro = this._mapGet("producerTable", account);
        pro.loc = loc;
        pro.url = url;
        pro.netId = netId;
        this._mapPut("producerTable", account, pro);
    }

    // producer log in as online state
    LogInProducer(account) {
        this._requireAuth(account, producerPermission);
        if (!storage.mapHas("producerTable", account)) {
            throw new Error("producer not exists, " + account);
        }
        const pro = this._mapGet("producerTable", account);
        pro.online = true;
        this._mapPut("producerTable", account, pro);
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
        this._mapPut("producerTable", account, pro);
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
        this._call("iost.vote", "RemoveOption", [
            voteId,
            account,
            true,
        ]);
        // will clear votes and score of the producer

        const pro = this._mapGet("producerTable", account);

        this._mapDel("producerTable", account);
        this._mapDel("preProducerMap", account);

        this._call("iost.token", "transfer", ["iost", "iost.vote_producer", account, pro.registerFee]);
        /*
        const ret = BlockChain.withdraw(account, pro.registerFee);
        if (ret != 0) {
            throw new Error("withdraw failed. ret = " + ret);
        }*/
    }

    // vote, need to pledge token
    Vote(producer, voter, amount) {
        this._requireAuth(voter, votePermission);

        if (!storage.mapHas("producerTable", producer)) {
            throw new Error("producer not exists");
        }

        const voteId = this._getVoteId();
        this._call("iost.vote", "Vote", [
            voteId,
            voter,
            producer,
            amount,
        ]);
    }

    // unvote
    Unvote(producer, voter, amount) {
        this._requireAuth(voter, votePermission);
        const voteId = this._getVoteId();
        this._call("iost.vote", "Unvote", [
            voteId,
            voter,
            producer,
            amount,
        ]);
    }

    // calculate the vote result, modify pendingProducerList
    Stat() {
        // controll auth
        const bn = this._getBlockNumber();
        const pendingBlockNumber = this._get("pendingBlockNumber");
        if (bn % voteStatInterval!== 0 || bn <= pendingBlockNumber) {
            throw new Error("stat failed. block number mismatch. pending bn = " + pendingBlockNumber + ", bn = " + bn);
        }

        const voteId = this._getVoteId();
        const voteRes = this._call("iost.vote", "GetResult", [voteId]);
        const preList = [];    // list of producers whose vote > threshold

        const pendingProducerList = this._get("pendingProducerList");

        const ppThreshold = new Float64(preProducerThreshold);
        for (const res of voteRes) {
            const key = res.option;
            const pro = this._mapGet("producerTable", key);
            // don't get score if in pending producer list or offline
            const votes = new Float64(res.votes);
            if (!pendingProducerList.includes(key) &&
                !votes.lt(ppThreshold) &&
                pro.online === true) {
                preList.push({
                    "key": key,
                    "prior": 0,
                    "votes": votes,
                    "score": pro.score
                });
            }
        }
        for (let i = 0; i < preList.length; i++) {
            const key = preList[i].key;
            const delta = preList[i].votes.minus(ppThreshold);
            const proRes = this._mapGet("producerTable", key);
            preList[i].score = delta.plus(new Float64(proRes.score));

            proRes.score = preList[i].score.number.toFixed();
            this._mapPut("producerTable", key, proRes);
        }

        // sort according to score in reversed order
        const scoreCmp = function(a, b) {
            if (!a.score.eq(b.score)) {
                return a.score.lt(b.score) ? 1 : -1;
            } else if (b.prior != a.prior) {
                return b.prior - a.prior;
            } else {
                return b.key < a.key ? 1 : -1;
            }
        };
        preList.sort(scoreCmp);

        // update pending list
        const producerNumber = this._get("producerNumber");
        const replaceNum = Math.min(preList.length, Math.floor(producerNumber / 6));
        const oldPreList = [];
        for (let key in pendingProducerList) {
            const x = pendingProducerList[key];
            oldPreList.push({
                "key": x,
                "prior": 1,
                "score": new Float64(this._mapGet("producerTable", x).score)
            });
        }

        // replace at most replaceNum producers
        for (let i = 0; i < replaceNum; i++) {
            oldPreList.push(preList[i]);
        }
        oldPreList.sort(scoreCmp);
        const newList = oldPreList.slice(0, producerNumber);

        const currentList = pendingProducerList;
        const pendingList = newList.map(x => x.key);
        this._put("currentProducerList", currentList);
        this._put("pendingProducerList", pendingList);
        this._put("pendingBlockNumber", this._getBlockNumber());

        for (const key of currentList) {
            if (!pendingList.includes(key)) {
                const proRes = this._mapGet("producerTable", key);
                proRes.score = "0";
                this._mapPut("producerTable", key, proRes);
            }
        }
    }

}

module.exports = VoteContract;
