class Contract {
    constructor() {
        this.maxUserNumber = 10;
        this.userNumber = 0;
        this.totalCoins = 0;
        this.lastLuckyBlock = -1;
        this.round = 0;
        this["results"]= [];
        this.clearUserValue()
    }
    clearUserValue() {
        this["table0"] = {};
        this["table1"] = {};
        this["table2"] = {};
        this["table3"] = {};
        this["table4"] = {};
        this["table5"] = {};
        this["table6"] = {};
        this["table7"] = {};
        this["table8"] = {};
        this["table9"] = {}
    }
    bet(account, luckyNumber, coins) {
        if (coins < 1 || coins > 5) {
            return "bet coins should be >=1 and <= 5"
        }
        if (luckyNumber < 0 && luckyNumber > 9) {
            return "bet coins should be >=1 and <= 5"
        }

        BlockChain.deposit(account, coins);


        if (this["table"+luckyNumber][account] === undefined) {
            this["table"+luckyNumber][account] = coins
        } else {
            this["table"+luckyNumber][account] += coins
        }

        this.userNumber ++;
        this.totalCoins += coins;

        if (this.userNumber >= this.maxUserNumber) {
            let bi = JSON.parse(BlockChain.blockInfo());
            let bn = bi.number;
            let ph = bi.parent_hash;
            if ( true /*this.lastLuckyBlock < 0 || bn - this.lastLuckyBlock >= 16 || bn > this.lastLuckyBlock && ph[ph.length-1] % 16 === 0*/) {
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

        let winTable = this["table"+ln];
        _native_log("winTab is" + (winTable));

        for (let [key, value] of winTable) {
            totalVal += value;
            kNum ++
        }

        let result = {
            number: this.lastLuckyBlock,
            user_number: this.userNumber,
            k_number: kNum,
            total_coins : tc,
            rewards : []
        };

        if (kNum > 0) {
            let unit = tc / totalVal;
            for (let [key, value] of winTable) {
                BlockChain.withdraw(key, value * unit);
                result.rewards.push({"key":key, "value": value * unit})
            }
        }
        let results =  this["results"];
        results.push(result);
        this["results"] = results
    }
}

module.exports = Contract;