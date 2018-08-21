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
        for (let i = 0; i <= 9; i ++) {
            this.tables[i] = {}
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

        if (this.tables[luckyNumber][account] === undefined) {
            this.tables[luckyNumber][account] = coins
        } else {
            this.tables[luckyNumber][account] += coins
        }

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

        for (let [key, value] of this.tables[ln]) {
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
            for (let [key, value] of this.tables[ln]) {
                BlockChain.withdraw(key, value * unit);
                result.rewards.push({"key":key, "value": value * unit})
            }
        }
        this.results.push(result)
    }
}

module.exports = Contract;