class DexContract {
    init(){}
    _requireAuth(account, permission) {
        const ret = blockchain.requireAuth(account, permission);
        if (ret !== true) {
            throw new Error("require auth failed. ret = " + ret);
        }
    }
    can_update(data) {
        const owner = storage.globalMapGet("system.iost", "contract_owner", blockchain.contractName());
        return blockchain.requireAuth(owner, "active");
    }
    _new_order_id() {
        return (Number(storage.get("order_id_cur")) + 1).toString();
    }
    _get_order(orderID) {
        if (!storage.mapHas("orders", orderID)) {
            return null
        }
        return JSON.parse(storage.mapGet("orders", orderID));
    }
    _put_order(orderID, orderJSON) {
        storage.mapPut("orders", orderID, JSON.stringify(orderJSON));
        storage.put("order_id_cur", orderID.toString());
    }
    _del_order(orderID) {
        storage.mapDel("orders", orderID);
    }
    place_order(sellType, sellAmount, buyType, buyAmount) {
        let orderID = this._new_order_id();
        blockchain.callWithAuth("token.iost", "transfer",
            JSON.stringify([sellType, blockchain.publisher(), blockchain.contractName(), sellAmount, "put" + orderID]));
        this._put_order(orderID, {"sell_type": sellType,
            "sell_amount": sellAmount,
            "buy_type": buyType,
            "buy_amount": buyAmount,
            "seller": blockchain.publisher()
        });
        return orderID;
    }
    cancel_order(orderID) {
        let orderJSON = this._get_order(orderID);
        if (orderJSON === null) {
            throw new Error("not a valid order");
        }
        this._requireAuth(orderJSON["seller"], "active");
        blockchain.callWithAuth("token.iost", "transfer",
            JSON.stringify([orderJSON["sell_type"], blockchain.contractName(), orderJSON["seller"], orderJSON["sell_amount"], "cancel" + orderID]));
        this._del_order(orderID);
    }
    take_order(orderID) {
        let orderJSON = this._get_order(orderID);
        if (orderJSON === null) {
            throw new Error("not a valid order");
        }
        blockchain.callWithAuth("token.iost", "transfer",
            JSON.stringify([orderJSON["buy_type"], blockchain.publisher(), orderJSON["seller"], orderJSON["buy_amount"], "take" + orderID]));
        blockchain.callWithAuth("token.iost", "transfer",
            JSON.stringify([orderJSON["sell_type"], blockchain.contractName(), blockchain.publisher(), orderJSON["sell_amount"], "finish" + orderID]));
        this._del_order(orderID);
    }
}

module.exports = DexContract;