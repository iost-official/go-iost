var BlockChain = (function () {
    var bc = new IOSTBlockchain;
    return {
        transfer: function (from, to, amount) {
            if (!(amount instanceof Int64)) {
                amount = new Int64(amount);
            }
            return bc.transfer(from, to, amount.toString());
        },
        withdraw: function (to, amount) {
            if (!(amount instanceof Int64)) {
                amount = new Int64(amount);
            }
            return bc.withdraw(to, amount.toString());
        },
        deposit: function (from, amount) {
            if (!(amount instanceof Int64)) {
                amount = new Int64(amount);
            }
            return bc.deposit(from, amount.toString());
        },
        topUp: function (contract, from, amount) {
            if (!(amount instanceof Int64)) {
                amount = new Int64(amount);
            }
            return bc.topUp(contract, from, amount.toString());
        },
        countermand: function (contract, to, amount) {
            if (!(amount instanceof Int64)) {
                amount = new Int64(amount);
            }
            return bc.countermand(contract, to, amount.toString());
        },
        blockInfo: function () {
            return bc.blockInfo();
        },
        txInfo: function () {
            return bc.txInfo();
        },
        call: function (contract, api, args) {
            return bc.call(contract, api, args);
        },
        callWithReceipt: function (contract, api, args) {
            return bc.callWithReceipt(contract, api, args);
        },
        requireAuth: function (account) {
            return 0;
        }
    }
})();

module.exports = BlockChain;