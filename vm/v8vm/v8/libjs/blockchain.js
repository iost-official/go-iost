let BlockChain = (function () {
    let bc = new IOSTBlockchain;
    // get contractName
    let contractName = function () {
        let ctxInfo = JSON.parse(bc.contextInfo());
        return ctxInfo["contract_name"];
    };
    // transfer IOSToken
    let transfer = function (from, to, amount) {
        if (!(amount instanceof Float64)) {
            amount = new Float64(amount);
        }
        return bc.call("iost.token", "transfer", "[\"iost\", \"" + from + "\",\"" + to + "\",\"" + amount.toString() + "\"]");
    };
    return {
        // transfer IOSToken
        transfer: transfer,
        // withdraw IOSToken
        withdraw: function (to, amount) {
            return transfer(contractName(), to, amount);
        },
        // deposit IOSToken
        deposit: function (from, amount) {
            return transfer(from, contractName(), amount);
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
        // get contractName
        contractName: contractName,
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