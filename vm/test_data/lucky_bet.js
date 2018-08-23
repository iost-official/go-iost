class Contract {
    constructor() {
        this.maxUserNumber = 100;
        this.userNumber = 0;
        this.totalCoins = 0;
        this.lastLuckyBlock = -1;
        this.round = 0;
        this.results = [];
        this.tables = [];
        this.clearUserValue()
    }
    clearUserValue() {
        this.tables = {
            "0":{},
            "1":{},
            "2":{},
            "3":{},
            "4":{},
            "5":{},
            "6":{},
            "7":{},
            "8":{},
            "9":{}
        }
    }
    bet(account, luckyNumber, coins) {
        if (coins < 1 || coins > 5) {
            return "bet coins should be >=1 and <= 5"
        }
        if (luckyNumber < 0 && luckyNumber > 9) {
            return "bet coins should be >=1 and <= 5"
        }

        BlockChain.deposit(account, coins);

        let betTables = this.tables;
        _native_log("here"+betTables);
        let betTable = betTables[luckyNumber+""];


        if (betTable[account] === undefined) {
            betTable[account] = coins
        } else {
            betTable[account] += coins
        }

        this.tables = betTables;

        this.userNumber ++;
        this.totalCoins += coins;

        if (this.userNumber >= this.maxUserNumber) {
            let bi = JSON.parse(BlockChain.blockInfo());
            let bn = bi.number;
            let ph = bi.parent_hash;
            if (this.lastLuckyBlock <0 || bn - this.lastLuckyBlock >= 16 || bn > this.lastLuckyBlock && ph % 16 === 0) {
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

        let winTable = this.tables[ln];
        for (let [key, value] of winTable) {
            totalVal ++;
            kNum ++
        }

        let result = {
            number: this.lastLuckyBlock,
            user_number: this.userNumber,
            k_number: kNum,
            total_coins : tc,
            rewards : []
        };

        if (kNum >0) {
            let unit = tc / totalVal;
            for (let [key, value] of winTable) {
                BlockChain.withdraw(key, value * unit);
                result.rewards.push({"key":key, "value": value * unit})
            }
        }
        this.results.push(result)
    }
}

module.exports = Contract;