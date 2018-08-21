var BlockChain = (function () {
    var bc = new IOSTBlockchain;
    return {
        transfer: function (from, to, amount) {
            if (!amount instanceof BigNumber) {
                amount = new BigNumber(amount);
            }
            return bc.transfer(from, to, amount);
        },
        withdraw: function (to, amount) {
            if (!amount instanceof BigNumber) {
                amount = new BigNumber(amount);
            }
            return bc.withdraw(to, amount);
        },
        deposit: function (from, amount) {
            if (!amount instanceof BigNumber) {
                amount = new BigNumber(amount);
            }
            return bc.deposit(from, amount);
        },
        topUp: function (contract, from, amount) {
            if (!amount instanceof BigNumber) {
                amount = new BigNumber(amount);
            }
            return bc.topUp(contract, from, amount);
        },
        countermand: function (contract, to, amount) {
            if (!amount instanceof BigNumber) {
                amount = new BigNumber(amount);
            }
            return bc.countermand(contract, to, amount);
        },
        blockInfo: function () {
            return bc.blockInfo();
        },
        txInfo: function () {
            return bc.txInfo();
        },
        call: function (contract, api, args) {
            return bc.call(contract, api, args);
        }
    }
})();

module.exports = BlockChain;