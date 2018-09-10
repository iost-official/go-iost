class VoteContract {
    constructor() {
    }
    init() {
        this._put("producerRegisterFee", 1000 * 1000);
        this._put("preProducerThreshold", 2100 * 10000);
        this._put("voteLockTime", 200);
        this._put("voteStatInterval", 200);
        this._put("producerNumber", 7);
        this._put("currentProducerList", []);
        this._put("pendingProducerList", []);
        this._put("pendingBlockNumber", 0);
    }

    InitProducer(num, proStr) {
        if (num === 0) {
            throw new Error("producer list number empty.");
        }
        var pendingProducerList = JSON.parse(proStr);
        if (typeof pendingProducerList !== "object") {
            throw new Error("producer str invalid format. got ", pendingProducerList, ", type = ", typeof pendingProducerList);
        }
        if (pendingProducerList.length !== num) {
            throw new Error("producer list length mismatch with number.");
        }
        this._put("pendingProducerList", pendingProducerList);
        var producerNumber = pendingProducerList.length;
        this._put("producerNumber", producerNumber);

        for (var i = 0; i < producerNumber; i++) {
            var ret = BlockChain.deposit(pendingProducerList[i], this._get("producerRegisterFee"));
            if (ret != 0) {
                throw new Error("constructor deposit failed. ret = " + ret);
            }
            this._mapPut("producerTable", pendingProducerList[i], {
                "loc": "",
                "url": "",
                "netId": "",
                "online": true,
                "score": 0,
                "votes": 0
            });
        }
    }

	_requireAuth(account) {
		var ret = BlockChain.requireAuth(account);
		if (ret !== true) {
			throw new Error("require auth failed. ret = " + ret);
		}
	}

	_getBlockNumber() {
		var bi = JSON.parse(BlockChain.blockInfo());
		if (!bi || bi === undefined || bi.number === undefined) {
			throw new Error("get block number failed. bi = " + bi);
		}
		return bi.number;
	}

    _get(k) {
        // _native_log(k + "," + storage.get(k));
        return JSON.parse(storage.get(k));
    }
	_put(k, v) {
        // _native_log("_put " + k + "," + v);
        var ret = storage.put(k, JSON.stringify(v));
        if (ret != 0) {
            throw new Error("storage put failed. ret = " + ret);
        }
    }
    _mapGet(k, f) {
        return JSON.parse(storage.mapGet(k, f));
    }
    _mapPut(k, f, v) {
        var ret = storage.mapPut(k, f, JSON.stringify(v));
        if (ret != 0) {
            throw new Error("storage map put failed. ret = " + ret);
        }
    }

    _mapDel(k, f) {
        var ret = storage.mapDel(k, f);
        if (ret != 0) {
            throw new Error("storage map del failed. ret = " + ret);
        }
    }

	// register account as a producer, need to pledge token
    RegisterProducer(account, loc, url, netId) {
		this._requireAuth(account);
		if (storage.mapHas("producerTable", account)) {
			throw new Error("producer exists");
		}
		var ret = BlockChain.deposit(account, this._get("producerRegisterFee"));
		if (ret != 0) {
			throw new Error("register deposit failed. ret = " + ret);
		}
		this._mapPut(producerTable, account, {
			"loc": loc,
			"url": url,
			"netId": netId,
			"online": false,
			"score": 0,
			"votes": 0
		});
    }
	
	// update the information of a producer
    UpdateProducer(account, loc, url, netId) {
		this._requireAuth(account);
		if (!storage.mapHas("producerTable", account)) {
			throw new Error("producer not exists");
		}
		var pro = this._mapGet("producerTable", account);
		pro.loc = loc;
		pro.url = url;
		pro.netId = netId;
		this._mapPut("producerTable", account, pro);
    }

	// producer log in as online state
    LogInProducer(account) {
		this._requireAuth(account);
        if (!storage.mapHas("producerTable", account)) {
			throw new Error("producer not exists, " + account);
		}
        var pro = this._mapGet("producerTable", account);
		pro.online = true;
        this._mapPut("producerTable", account, pro);
    }

	// producer log out as offline state
    LogOutProducer(account) {
		this._requireAuth(account);
        if (!storage.mapHas("producerTable", account)) {
			throw new Error("producer not exists");
		}
		if (this._get("pendingProducerList").includes(account) ||
            this._get("currentProducerList").includes(account)) {
			throw new Error("producer in pending list or in current list, can't logout");
		}
        var pro = this._mapGet("producerTable", account);
		pro.online = false;
        this._mapPut("producerTable", account, pro);
    }

	// remove account from producer list
	UnregisterProducer(account) {
		this._requireAuth(account);
        if (!storage.mapHas("producerTable", account)) {
			throw new Error("producer not exists");
		}
        if (this._get("pendingProducerList").includes(account) ||
            this._get("currentProducerList").includes(account)) {
			throw new Error("producer in pending list or in current list, can't unregist");
		}
		// will clear votes and score of the producer

        this._mapDel("producerTable", account);
        this._mapDel("preProducerMap", account);

		var ret = BlockChain.withdraw(account, this._get("producerRegisterFee"));
		if (ret != 0) {
			throw new Error("withdraw failed. ret = " + ret);
		}
	}

	// vote, need to pledge token
	Vote(producer, voter, amount) {
		this._requireAuth(voter);

        if (!storage.mapHas("producerTable", producer)) {
			throw new Error("producer not exists");
		}

		var ret = BlockChain.deposit(voter, amount);
		if (ret != 0) {
			throw new Error("vote deposit failed. ret = " + ret);
		}

		var voteRes = {};
        if (storage.mapHas("voteTable", voter)) {
		    voteRes = this._mapGet("voteTable", voter);
		}
		// record the amount and time of the vote
		if (voteRes.hasOwnProperty(producer)) {
			voteRes[producer].amount += amount;
		} else {
            voteRes[producer] = {}
			voteRes[producer].amount = amount;
		}
		voteRes[producer].time = this._getBlockNumber();
        this._mapPut("voteTable", voter, voteRes);

		// if producer's votes >= preProducerThreshold, then insert into preProducer map
        var proRes = this._mapGet("producerTable", producer);
		proRes.votes += amount;
		if (proRes.votes - amount < this._get("preProducerThreshold ")&&
				proRes.votes >= this._get("preProducerThreshold")) {
		    this._mapPut("preProducerMap", producer, true);
		}
	}

	// unvote
	Unvote(producer, voter, amount) {
		this._requireAuth(voter);

		if (!storage.mapHas("voteTable", voter)) {
            throw new Error("producer not voted");
        }
        var voteRes = this._mapGet("voteTable", voter);
		if (!voteRes.hasOwnProperty(producer) ||
				voteRes[producer].amount < amount) {
			throw new Error("producer not voted or vote amount less than expected")
		}
		if (voteRes[producer].time + this._get("voteLockTime ")> this._getBlockNumber()) {
			throw new Error("vote still lockd")
		}
		voteRes[producer].amount -= amount;
		this._mapPut("voteTable", voter, voteRes);

		// if producer not exist, it's because producer has unregistered, do nothing
		if (storage.mapHas("producerTable", producer)) {
		    proRes = this._mapGet("producerTable", producer);
			var ori = proRes.votes;
			proRes.votes = Math.max(0, ori - amount);
			this._mapPut("producerTable", producer, proRes);

			// if producer's votes < preProducerThreshold, then delete from preProducer map
			if (ori >= this._get("preProducerThreshold ")&&
					proRes.votes < this._get("preProducerThreshold")) {
			    this._mapDel("preProducerMap", producer);
			}
		}

		var ret = BlockChain.withdraw(voter, amount);
		if (ret != 0) {
			throw new Error("withdraw failed. ret = " + ret);
		}

        var servi = Math.floor(amount * this._getBlockNumber() / this._get("voteLockTime"));
		ret = BlockChain.grantServi(voter, servi);
		if (ret != 0) {
		    throw new Error("grant servi failed. ret = " + ret);
        }
	}

	// calculate the vote result, modify pendingProducerList
	Stat() {
		// controll auth
		var bn = this._getBlockNumber();
		var pendingBlockNumber = this._get("pendingBlockNumber");
		if (bn % this._get("voteStatInterval ")!= 0 || bn <= pendingBlockNumber) {
			throw new Error("stat failed. block number mismatch. pending bn = " + pendingBlockNumber + ", bn = " + bn);
		}

		// add scores for preProducerMap
		var preList = [];	// list of producers whose vote > threshold
        var preProducerMapKeys = storage.mapKeys("preProducerMap");
		for (var i in preProducerMapKeys) {
		    var key = preProducerMapKeys[i];
		    var pro = this._mapGet("producerTable", key);
            // don't get score if in pending producer list or offline
		    if (!this._get("pendingProducerList").includes(key) && pro.votes >= this._get("preProducerThreshold ")&& pro.online === true) {
                preList.push({
                    "key": key,
                    "votes": pro.votes,
                    "score": pro.score
                });
            }
        }
		for (var i = 0; i < preList.length; i++) {
			var key = preList[i].key
			var delta = preList[i].votes - this._get("preProducerThreshold");
			var proRes = this._get("producerTable", key);
			proRes.score += delta;
            this._mapPut("producerTable", key, proRes);
			preList[i].score += delta;
		}

		// sort according to score in reversed order
		var scoreCmp = function(a, b) {
			return b.score - a.score;
		}
		preList.sort(scoreCmp);

		// update pending list
        var producerNumber = this._get("producerNumber");
		var replaceNum = Math.min(preList.length, Math.floor(producerNumber / 6));
		var oldPreList = [];
		var pendingProducerList = this._get("pendingProducerList");
        for (let key in pendingProducerList) {
		    var x = pendingProducerList[key];
			oldPreList.push({
				"key": x,
				"score": this._mapGet("producerTable", x).score
			});
		};

		// replace at most replaceNum producers
		for (var i = 0; i < replaceNum; i++) {
			oldPreList.push(preList[i]);
		}
		oldPreList.sort(scoreCmp);
		var newList = oldPreList.slice(0, producerNumber);

		var currentList = pendingProducerList;
		this._put("currentProducerList", currentList);
		this._put("pendingProducerList", newList.map(x => x.key));
		this._put("pendingBlockNumber", this._getBlockNumber());

		for (var i = 0; i < producerNumber; i++) {
			if (!pendingProducerList.includes(currentList[i])) {
                var proRes = this._mapGet("producerTable", currentList[i]);
                proRes.score = 0;
                this._mapPut("producerTable", currentList[i], 0);
			}
		}
	}

}

module.exports = VoteContract;
