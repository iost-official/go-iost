class Contract {
    constructor() {
        this.maxUserNumber = 10;
        this.userNumber = 0;
        this.totalCoins = 0;
        this.lastLuckyBlock = -1;
        this.round = 0;
        this.results = [];
        this.tables = [];
        this.clearUserValue()
    }
    clearUserValue() {
        this.tables = [];
    }
    bet(account, luckyNumber, coins) {
        if (coins < 1 || coins > 5) {
            return "bet coins should be >=1 and <= 5"
        }
        if (luckyNumber < 0 && luckyNumber > 9) {
            return "bet coins should be >=1 and <= 5"
        }

        BlockChain.deposit(account, coins);

        if (this.tables[luckyNumber] === undefined) {
            this.tables[luckyNumber] = [];
        }
        
        let isExist = false;
        
        this.tables[luckyNumber].forEach(function (record) {
            if (record.account === account) {
                record.coins += coins;
                isExist = true;
            }
        });

        if (!isExist) {
            this.tables[luckyNumber].push({ account:account, coins : coins})
        }

        // if (this.tables[luckyNumber] === undefined) {
        //     this.tables[luckyNumber] = {};
        // }
        //
        // if (this.tables[luckyNumber].get(account) === undefined) {
        //     this.tables[luckyNumber].set(account, coins);
        // } else {
        //     let c = this.tables[luckyNumber].get(account);
        //     this.tables[luckyNumber].set(account, c + coins);
        // }

        this.userNumber ++;
        this.totalCoins += coins;

        if (this.userNumber >= this.maxUserNumber) {
            let bi = JSON.parse(BlockChain.blockInfo());
            let bn = bi.number;
            let ph = bi.parent_hash;
            if ( this.lastLuckyBlock < 0 || bn - this.lastLuckyBlock >= 16 || bn > this.lastLuckyBlock && ph[ph.length-1] % 16 === 0) {
                this.lastLuckyBlock = bn;

                this.getReward(bn);
                this.userNumber = 0;
                this.totalCoins = 0
            }
        }
    }

    getReward() {
        let ln = this.lastLuckyBlock % 10;
        this.round ++;

        let tc = this.totalCoins * 0.95;
        let totalVal = 0;
        let kNum = 0;

        _native_log("lucky number is "+ln);


        // for (let [key, value] of this.tables[ln]) {
        //     totalVal += value;
        //     kNum ++
        // }

        this.tables[ln].forEach(function (record) {
            totalVal += record.coins;
            kNum ++;
        });

        let result = {
            number: this.lastLuckyBlock,
            user_number: this.userNumber,
            k_number: kNum,
            total_coins : tc,
            rewards : []
        };

        if (kNum > 0) {
            let unit = tc / totalVal;
            // for (let [key, value] of winTable) {
            //     BlockChain.withdraw(key, value * unit);
            //     result.rewards.push({"account":key, "reward": value * unit})
            // }
            this.tables[ln].forEach(function (record) {
                let reward = record.coins * unit;
                BlockChain.withdraw(record.account, reward);
                result.rewards.push({"account":record.account, "reward": reward})
            });
        }
        let results =  this.results;
        results.push(result);
        this.results = results
    }
}

module.exports = Contract;