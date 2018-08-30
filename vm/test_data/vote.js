class VoteContract {
    constructor() {
		this.producerRegisterFee = 1000 * 10000;
		this.producerNumber = 7;
		this.preProducerThreshold = 2100 * 10000;
		this.preProducerMap = {};
		this.voteLockTime = 200;

		this.currentProducerList = [];
		this.currentProducerList[0] = "";
		this.currentProducerList[this.producerNumber - 1] = "";
		this.pendingProducerList = ["a", "b", "c", "d", "e", "f", "g"];

		// todo init producer list pledge token
		for (var i = 0; i < this.producerNumber; i++) {
			this["producer-" + this.pendingProducerList[i]] = {
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
		var ret = BlockChain.require_auth(account);
		if (ret !== 0) {
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

	// register account as a producer, need to pledge token
    RegisterProducer(account, loc, url, netId) {
		this._requireAuth(account);
		if (this.hasOwnProperty("producer-" + account)) {
			throw new Error("producer exists");
		}
		var ret = BlockChain.deposit(account, this.producerRegisterFee);
		if (ret != 0) {
			throw new Error("deposit failed. ret = " + ret);
		}
		this["producer-" + account] = {
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
		if (!this.hasOwnProperty("producer-" + account)) {
			throw new Error("producer not exists");
		}
		this["producer-" + account].loc = loc;
		this["producer-" + account].url = url;
		this["producer-" + account].netId = netId;
    }

	// producer log in as online state
    LogInProducer(account) {
		this._requireAuth(account);
		if (!this.hasOwnProperty("producer-" + account)) {
			throw new Error("producer not exists");
		}
		this["producer-" + account].online = true;
    }

	// producer log out as offline state
    LogOutProducer(account) {
		this._requireAuth(account);
		if (!this.hasOwnProperty("producer-" + account)) {
			throw new Error("producer not exists");
		}
		if (this.pendingProducerList.includes(account) || this.currentProducerList.includes(account)) {
			throw new Error("producer in pending list or in current list, can't logout");
		}
		this["producer-" + account].online = false;
    }

	// remove account from producer list
	UnregisterProducer(account) {
		this._requireAuth(account);
		if (!this.hasOwnProperty("producer-" + account)) {
			throw new Error("producer not exists");
		}
		if (this.pendingProducerList.includes(account) || this.currentProducerList.includes(account)) {
			throw new Error("producer in pending list or in current list, can't unregist");
		}
		// will clear votes and score of the producer
		delete this["producer-" + account];
		delete this.preProducerMap(account);

		var ret = BlockChain.withdraw(account, this.producerRegisterFee);
		if (ret != 0) {
			throw new Error("withdraw failed. ret = " + ret);
		}
	}

	// vote, need to pledge token
	Vote(producer, voter, amount) {
		this._requireAuth(voter);

		if (!this.hasOwnProperty("producer-" + producer)) {
			throw new Error("producer not exists");
		}

		var ret = BlockChain.deposit(voter, amount);
		if (ret != 0) {
			throw new Error("deposit failed. ret = " + ret);
		}

		// record the amount and time of the vote
		if (this["vote-" + voter].hasOwnProperty(producer)) {
			this["vote-" + voter][producer].amount += amount;
		} else {
			this["vote-" + voter][producer] = {};
			this["vote-" + voter][producer].amount = amount;
		}
		this["vote-" + voter][producer].time = _getBlockNumber();

		// if producer's votes >= preProducerThreshold, then insert into preProducer map
		this["producer-" + producer].votes += amount;
		if (this["producer-" + producer].votes - amount < this.preProducerThreshold &&
				this["producer-" + producer].votes >= this.preProducerThreshold) {
			this.preProducerMap[producer] = true
		}
	}

	// unvote
	Unvote(producer, voter, amount) {
		this._requireAuth(voter);

		if (!this["vote-" + voter].hasOwnProperty(producer) || this["vote-" + voter][producer].amount < amount) {
			throw new Error("producer not voted or vote amount less than expected")
		}
		if (this["vote-" + voter][producer].time + this.voteLockTime > _getBlockNumber()) {
			throw new Error("vote still lockd")
		}

		this["vote-" + voter][producer].amount -= amount;

		// if producer not exist, it's because producer has unregistered, do nothing
		if (this.hasOwnProperty("producer-" + producer)) {
			this["producer-" + producer].votes -= amount;
			// if producer's votes < preProducerThreshold, then delete from preProducer map
			if (this["producer-" + producer].votes + amount >= this.preProducerThreshold && 
					this["producer-" + producer].votes < this.preProducerThreshold) {
				// todo cut down score?
				delete this.preProducerMap[producer];
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
		Object.keys(this.preProducerMap).forEach(function(key){
			var pro = this["producer-" + key]
			if (pro.votes >= this.preProducerThreshold && pro.online === true) {
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
			delta = preList[i].votes - this.preProducerThreshold;
			this["producer-" + key].score += delta;
			preList[i].score += delta;
		}

		// sort according to score in reversed order
		scoreCmp = function(a, b) {
			return b.score - a.score;
		}
		preList.sort(scoreCmp);

		// update pending list
		var replaceNum = Math.min(preList.length, Math.floor(this.producerNumber / 6));
		oldPreList = this.pendingProducerList.map(function(x){
			return {
				"key": x,
				"score": this["producer-" + x].score
			};
		});
		oldPreList.sort(scoreCmp);

		// replace at most replaceNum producers
		for (var i = 0; i < replaceNum; i++) {
			oldPreList.push(preList[i]);
		}
		oldPreList.sort(scoreCmp);
		newList = oldPreList.slice(0, this.producerNumber);

		this.currentProducerList = this.pendingProducerList;
		this.pendingProducerList = newList.map(x => x.key);

		for (var i = 0; i < this.producerNumber; i++) {
			if (!this.pendingProducerList.includes(this.currentProducerList[i])) {
				this["producer-" + this.currentProducerList[i]].score = 0;
			}
		}
	}

}

module.exports = VoteContract;
