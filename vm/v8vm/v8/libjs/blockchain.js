let BlockChain = (function () {
    let bc = new IOSTBlockchain;
    return {
        // transfer IOSToken
        transfer: function (from, to, amount) {
            if (!(amount instanceof Int64)) {
                amount = new Int64(amount);
            }
            return
        },
        // withdraw IOSToken
        withdraw: function (to, amount) {
            if (!(amount instanceof Int64)) {
                amount = new Int64(amount);
            }
            return
        },
        // deposit IOSToken
        deposit: function (from, amount) {
            if (!(amount instanceof Int64)) {
                amount = new Int64(amount);
            }
            return
        },
        // get blockInfo
        blockInfo: function () {
            return bc.blockInfo();
        },
        // get transactionInfo
        txInfo: function () {
            return bc.txInfo();
        },
        // get transactionInfo
        contextInfo: function () {
            return bc.contextInfo();
        },
        // call contract's api using args
        call: function (contract, api, args) {
            return bc.call(contract, api, args);
        },
        // call contract's api using args with auth
        callWithAuth: function (contract, api, args) {
            return bc.callWithAuth(contract, api, args);
        },
        // check account's permission
        requireAuth: function (accountID, permission) {
            return bc.requireAuth(accountID, permission);
        },
        // generate receipt
        receipt: function (content) {
            return bc.receipt(content);
        },
        // post event
        receipt: function (content) {
            return bc.event(content);
        },
    }
})();

module.exports = BlockChain;