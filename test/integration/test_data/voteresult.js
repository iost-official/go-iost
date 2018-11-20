class voteresult {
    init(){
    }
    
    _call(contract, api, args) {
        const ret = JSON.parse(BlockChain.call(contract, api, JSON.stringify(args)));
        if (ret && Array.isArray(ret) && ret.length == 1) {
            return JSON.parse(ret[0]);
        }
        return ret;
    }

    GetResult(voteId) {
        let ret = this._call("vote.iost", "GetResult",  [voteId]);
        storage.mapPut("vote-result", voteId, JSON.stringify(ret));
    }
}

module.exports = voteresult;
