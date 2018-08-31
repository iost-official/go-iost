class VoteContract {
    constructor() {
		this.producerRegisterFee = 1000 * 10000;
		this.producerNumber = 7;
		this.preProducerThreshold = 2100 * 10000;
		this.preProducerMap = {};
		this.voteLockTime = 200;

		this.currentProducerList = [];
		this.pendingProducerList = ["a", "b", "c", "d", "e", "f", "g"];
		this.pendingBlockNumber = 0;
		this.producerTable = {}
		this.voteTable = {}

		for (var i = 0; i < this.producerNumber; i++) {
			// todo init producer list pledge token
			// var ret = BlockChain.deposit(this.pendingProducerList[i], this.producerRegisterFee);
			// if (ret != 0) {
			// 	throw new Error("deposit failed. ret = " + ret);
			// }
			this.producerTable[this.pendingProducerList[i]] = {
				"loc": "",
				"url": "",
				"netId": "",
				"online": true,
				"score": 0,
				"votes": 0
			}
		}
    }

	_requireAuth(account) {
		// var ret = BlockChain.requireAuth(account);
		// if (ret !== 0) {
		// 	throw new Error("require auth failed. ret = " + ret);
		// }
	}

	_getBlockNumber() {
		var bi = JSON.parse(BlockChain.blockInfo());
		if (!bi || bi === undefined || bi.number === undefined) {
			throw new Error("get block number failed. bi = " + bi);
		}
		return bi.number;
	}

	// register account as a producer, need to pledge token
    RegisterProducer(account, loc, url, netId) {
		this._requireAuth(account);
		if (this.producerTable.hasOwnProperty(account)) {
			throw new Error("producer exists");
		}
		var ret = BlockChain.deposit(account, this.producerRegisterFee);
		if (ret != 0) {
			throw new Error("deposit failed. ret = " + ret);
		}
		this.producerTable[account] = {
			"loc": loc,
			"url": url,
			"netId": netId,
			"online": false,
			"score": 0,
			"votes": 0
		};
    }
	
	// update the information of a producer
    UpdateProducer(account, loc, url, netId) {
		this._requireAuth(account);
		if (!this.producerTable.hasOwnProperty(account)) {
			throw new Error("producer not exists");
		}
		var pro = this.producerTable[account];
		pro.loc = loc;
		pro.url = url;
		pro.netId = netId;
		this.producerTable[account]	= pro;
    }

	// producer log in as online state
    LogInProducer(account) {
		this._requireAuth(account);
		if (this.producerTable.hasOwnProperty(account) === false) {
			throw new Error("producer not exists, " + account);
		}
		var pro = this.producerTable[account];
		pro.online = true;
		this.producerTable[account] = pro;
    }

	// producer log out as offline state
    LogOutProducer(account) {
		this._requireAuth(account);
		if (this.producerTable.hasOwnProperty(account) === false) {
			throw new Error("producer not exists");
		}
		if (this.pendingProducerList.includes(account) || this.currentProducerList.includes(account)) {
			throw new Error("producer in pending list or in current list, can't logout");
		}
		var pro = this.producerTable[account];
		pro.online = false;
		this.producerTable[account] = pro;
    }

	// remove account from producer list
	UnregisterProducer(account) {
		this._requireAuth(account);
		if (!this.producerTable.hasOwnProperty(account)) {
			throw new Error("producer not exists");
		}
		if (this.pendingProducerList.includes(account) || this.currentProducerList.includes(account)) {
			throw new Error("producer in pending list or in current list, can't unregist");
		}
		// will clear votes and score of the producer
		this.producerTable[account] = undefined;
		this.preProducerMap[account] = undefined;

		var ret = BlockChain.withdraw(account, this.producerRegisterFee);
		if (ret != 0) {
			throw new Error("withdraw failed. ret = " + ret);
		}
	}

	// vote, need to pledge token
	Vote(producer, voter, amount) {
		this._requireAuth(voter);

		if (!this.producerTable.hasOwnProperty(producer)) {
			throw new Error("producer not exists");
		}

		var ret = BlockChain.deposit(voter, amount);
		if (ret != 0) {
			throw new Error("deposit failed. ret = " + ret);
		}

		if (!this.voteTable.hasOwnProperty(voter)) {
			this.voteTable[voter] = {}
		}
		// record the amount and time of the vote
		if (this.voteTable[voter].hasOwnProperty(producer)) {
			this.voteTable[voter][producer].amount += amount;
		} else {
			this.voteTable[voter][producer] = {};
			this.voteTable[voter][producer].amount = amount;
		}
		this.voteTable[voter][producer].time = this._getBlockNumber();

		// if producer's votes >= preProducerThreshold, then insert into preProducer map
		this.producerTable[producer].votes += amount;
		if (this.producerTable[producer].votes - amount < this.preProducerThreshold &&
				this.producerTable[producer].votes >= this.preProducerThreshold) {
			this.preProducerMap[producer] = true
		}
	}

	// unvote
	Unvote(producer, voter, amount) {
		this._requireAuth(voter);

		if (!this.voteTable.hasOwnProperty(voter) || !this.voteTable[voter].hasOwnProperty(producer) || 
				this.voteTable[voter][producer].amount < amount) {
			throw new Error("producer not voted or vote amount less than expected")
		}
		if (this.voteTable[voter][producer].time + this.voteLockTime > this._getBlockNumber()) {
			throw new Error("vote still lockd")
		}

		this.voteTable[voter][producer].amount -= amount;

		// if producer not exist, it's because producer has unregistered, do nothing
		if (this.producerTable.hasOwnProperty(producer)) {
			var ori = this.producerTable[producer].votes;
			this.producerTable[producer].votes = Math.max(0, this.producerTable[producer].votes - amount);
			// if producer's votes < preProducerThreshold, then delete from preProducer map
			if (ori >= this.preProducerThreshold && 
					this.producerTable[producer].votes < this.preProducerThreshold) {
				// todo cut down score?
				this.preProducerMap[producer] = undefined;
			}
		}

		var ret = BlockChain.withdraw(voter, amount);
		if (ret != 0) {
			throw new Error("withdraw failed. ret = " + ret);
		}

		// todo calc servi
	}

	// calculate the vote result, modify pendingProducerList
	Stat() {
		//todo require auth

		// add scores for preProducerMap
		var preList = [];	// list of producers whose vote > threshold
		var that = this;
		Object.keys(this.preProducerMap).forEach(function(key){
			var pro = that.producerTable[key]
			if (!that.pendingProducerList.includes(key) && pro.votes >= that.preProducerThreshold && pro.online === true) {
				preList.push({
					"key": key,
					"votes": pro.votes,
					"score": pro.score
				});
			}
		});
		for (var i = 0; i < preList.length; i++) {
			var key = preList[i].key
			// don't get score if in pending producer list
			if (this.pendingProducerList.includes(key)) {
				continue;
			}
			var delta = preList[i].votes - this.preProducerThreshold;
			this.producerTable[key].score += delta;
			preList[i].score += delta;
		}
		_native_log("pre list len = " + preList.length)

		// sort according to score in reversed order
		var scoreCmp = function(a, b) {
			return b.score - a.score;
		}
		preList.sort(scoreCmp);

		// update pending list
		var replaceNum = Math.min(preList.length, Math.floor(this.producerNumber / 6));
		var oldPreList = this.pendingProducerList.map(function(x){
			return {
				"key": x,
				"score": that.producerTable[x].score
			};
		});
		oldPreList.sort(scoreCmp);

		// replace at most replaceNum producers
		for (var i = 0; i < replaceNum; i++) {
			oldPreList.push(preList[i]);
		}
		oldPreList.sort(scoreCmp);
		var newList = oldPreList.slice(0, this.producerNumber);

		this.currentProducerList = this.pendingProducerList;
		this.pendingProducerList = newList.map(x => x.key);
		this.pendingBlockNumber = this._getBlockNumber();

		for (var i = 0; i < this.producerNumber; i++) {
			if (!this.pendingProducerList.includes(this.currentProducerList[i])) {
				this.producerTable[this.currentProducerList[i]].score = 0;
			}
		}
	}

}

module.exports = VoteContract;
