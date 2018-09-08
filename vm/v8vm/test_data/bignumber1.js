'use strict';
class BigNumber1{
    constructor() {
        let val = new BigNumber(0.00000000008);
        storage.put("val", JSON.stringify(val.plus(0.0000000000000029)));
    }
    getVal() {
        let val = storage.get("val");
        return new BigNumber(val);
    }
}

module.exports = BigNumber1;