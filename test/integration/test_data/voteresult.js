class voteresult {
    init(){
    }

    _call(contract, api, args) {
        const ret = JSON.parse(blockchain.call(contract, api, JSON.stringify(args)));
        if (ret && Array.isArray(ret) && ret.length == 1) {
            return JSON.parse(ret[0]);
        }
        return ret;
    }

    getResult(voteId) {
        let ret = this._call("vote.iost", "getResult",  [voteId]);
        storage.mapPut("vote_result", voteId, JSON.stringify(ret));
    }
}

module.exports = voteresult;
