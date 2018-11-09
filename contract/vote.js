const softFloatRate = 1;
const producerRegisterFee = 1000 * 1000 * softFloatRate;
const preProducerThreshold = 2100 * 10000;
const voteLockTime = 200;
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

        BlockChain.deposit(proID, producerRegisterFee);
        this._mapPut("producerTable", proID, {
            "loc": "",
            "url": "",
            "netId": "",
            "online": true,
            "score": 0,
            "votes": 0
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
        storage.put(k, JSON.stringify(v));
    }
    _mapGet(k, f) {
        return JSON.parse(storage.mapGet(k, f));
    }
    _mapPut(k, f, v) {
        storage.mapPut(k, f, JSON.stringify(v));
    }

    _mapDel(k, f) {
        storage.mapDel(k, f);
    }

	// register account as a producer, need to pledge token
    RegisterProducer(account, loc, url, netId) {
		this._requireAuth(account, producerPermission);
		if (storage.mapHas("producerTable", account)) {
			throw new Error("producer exists");
		}
		BlockChain.deposit(account, producerRegisterFee);
		this._mapPut("producerTable", account, {
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
		// will clear votes and score of the producer

        this._mapDel("producerTable", account);
        this._mapDel("preProducerMap", account);

		const ret = BlockChain.withdraw(account, producerRegisterFee);
		if (ret != 0) {
			throw new Error("withdraw failed. ret = " + ret);
		}
	}

	// vote, need to pledge token
	Vote(producer, voter, amount) {
		this._requireAuth(voter, votePermission);
		amount = Math.floor(amount);

        if (!storage.mapHas("producerTable", producer)) {
			throw new Error("producer not exists");
		}

		BlockChain.deposit(voter, amount * softFloatRate);

		let voteRes = {};
        if (storage.mapHas("voteTable", voter)) {
		    voteRes = this._mapGet("voteTable", voter);
		}
		// record the amount and time of the vote
		if (voteRes.hasOwnProperty(producer)) {
			voteRes[producer].amount += amount;
		} else {
            voteRes[producer] = {};
			voteRes[producer].amount = amount;
		}
		voteRes[producer].time = this._getBlockNumber();
        this._mapPut("voteTable", voter, voteRes);

		// if producer's votes >= preProducerThreshold, then insert into preProducer map
        const proRes = this._mapGet("producerTable", producer);
		proRes.votes += amount;
		if (proRes.votes - amount <  preProducerThreshold &&
				proRes.votes >= preProducerThreshold) {
		    this._mapPut("preProducerMap", producer, true);
		}
        this._mapPut("producerTable", producer, proRes)
	}

	// unvote
	Unvote(producer, voter, amount) {
        amount = Math.floor(amount);
		this._requireAuth(voter, votePermission);

		if (!storage.mapHas("voteTable", voter)) {
            throw new Error("producer not voted");
        }
        const voteRes = this._mapGet("voteTable", voter);
		if (!voteRes.hasOwnProperty(producer)) {
            throw new Error("producer not voted")
        }
        if (voteRes[producer].amount < amount) {
			throw new Error("vote amount less than expected")
		}
		if (voteRes[producer].time + voteLockTime> this._getBlockNumber()) {
			throw new Error("vote still locked")
		}
		voteRes[producer].amount -= amount;
		this._mapPut("voteTable", voter, voteRes);

		// if producer not exist, it's because producer has unregistered, do nothing
		if (storage.mapHas("producerTable", producer)) {
		    const proRes = this._mapGet("producerTable", producer);
			const ori = proRes.votes;
			proRes.votes = Math.max(0, ori - amount);
			this._mapPut("producerTable", producer, proRes);

			// if producer's votes < preProducerThreshold, then delete from preProducer map
			if (ori >= preProducerThreshold &&
					proRes.votes < preProducerThreshold) {
			    this._mapDel("preProducerMap", producer);
			}
		}

		const ret = BlockChain.withdraw(voter, amount * softFloatRate);
		if (ret !== 0) {
			throw new Error("withdraw failed. ret = " + ret);
		}

        const servi = Math.floor(amount * this._getBlockNumber() / voteLockTime);
		const ret2 = BlockChain.grantServi(voter, servi);
		if (ret2 !== 0) {
		    throw new Error("grant servi failed. ret = " + ret2);
        }
	}

	// calculate the vote result, modify pendingProducerList
	Stat() {
		// controll auth
		const bn = this._getBlockNumber();
		const pendingBlockNumber = this._get("pendingBlockNumber");
		if (bn % voteStatInterval!== 0 || bn <= pendingBlockNumber) {
			throw new Error("stat failed. block number mismatch. pending bn = " + pendingBlockNumber + ", bn = " + bn);
		}

		// add scores for preProducerMap
		const preList = [];	// list of producers whose vote > threshold
        const preProducerMapKeys = storage.mapKeys("preProducerMap");

        const pendingProducerList = this._get("pendingProducerList");

		for (let i in preProducerMapKeys) {
		    const key = preProducerMapKeys[i];
		    const pro = this._mapGet("producerTable", key);
            // don't get score if in pending producer list or offline
		    if (!pendingProducerList.includes(key) &&
                pro.votes >= preProducerThreshold &&
                pro.online === true) {
                preList.push({
                    "key": key,
                    "prior": 0,
                    "votes": pro.votes,
                    "score": pro.score
                });
            }
        }
        for (let i = 0; i < preList.length; i++) {
			const key = preList[i].key;
			const delta = preList[i].votes - preProducerThreshold;
            const proRes = this._mapGet("producerTable", key);

            proRes.score += delta;
            this._mapPut("producerTable", key, proRes);
			preList[i].score += delta;
		}

		// sort according to score in reversed order
		const scoreCmp = function(a, b) {
			if (b.score != a.score) {
			    return b.score - a.score;
			} else if (b.prior != a.prior) {
			    return b.prior - a.prior;
			} else if (b.key < a.key) {
			    return 1;
			} else {
			    return -1;
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
				"score": this._mapGet("producerTable", x).score
			});
		}

		// replace at most replaceNum producers
		for (let i = 0; i < replaceNum; i++) {
			oldPreList.push(preList[i]);
		}
		oldPreList.sort(scoreCmp);
		const newList = oldPreList.slice(0, producerNumber);

		const currentList = pendingProducerList;
		this._put("currentProducerList", currentList);
		this._put("pendingProducerList", newList.map(x => x.key));
		this._put("pendingBlockNumber", this._getBlockNumber());

		for (let i = 0; i < producerNumber; i++) {
			if (!pendingProducerList.includes(currentList[i])) {
                const proRes = this._mapGet("producerTable", currentList[i]);
                proRes.score = 0;
                this._mapPut("producerTable", currentList[i], 0);
			}
		}
	}

}

module.exports = VoteContract;
