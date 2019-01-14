const VPContract = "vote_producer.iost";
const BonusContract = "bonus.iost";
const IssueContract = "issue.iost";
const TokenContract = "token.iost";

class VoteChecker {
    init() {
    }

    _get(k) {
        const val = storage.get(k);
        if (val === "") {
            return null;
        }
        return JSON.parse(val);
    }

	_put(k, v, p) {
        storage.put(k, JSON.stringify(v), p);
    }

    _call(contract, api, args) {
        const ret = blockchain.callWithAuth(contract, api, JSON.stringify(args));
        if (ret && Array.isArray(ret) && ret.length === 1) {
            return ret[0] === "" ? "" : JSON.parse(ret[0]);
        }
        return ret;
    }

    vote(from, to, amount) {
        let voteInfo1 =  this._call(VPContract, "getProducer", [to]);
        blockchain.callWithAuth(VPContract, "vote", [from, to, amount]);
        let voteInfo2 =  this._call(VPContract, "getProducer", [to]);
        let data = this._get("vote") || {};
        let f = data[from] || {};
        f[to] = {
            vote: new Float64(amount).plus(f[to] || 0).toFixed(8),
            VPContract: new Float64(voteInfo2.voteInfo.votes).minus(voteInfo1.voteInfo.votes).toFixed(8),
        };
        data[from] = f;
        this._put("vote", data);
        return data;
    }

    unvote(from, to, amount) {
        let voteInfo1 =  this._call(VPContract, "getProducer", [to]);
        blockchain.callWithAuth(VPContract, "unvote", [from, to, amount]);
        let voteInfo2 =  this._call(VPContract, "getProducer", [to]);
        let data = this._get("unvote") || {};
        let f = data[from] || {};
        f[to] = {
            vote: new Float64(amount).plus(f[to] || 0).toFixed(8),
            VPContract: new Float64(voteInfo2.voteInfo.votes).minus(voteInfo1.voteInfo.votes).toFixed(8),
        };
        data[from] = f;
        this._put("unvote", data);
        return data;
    }

    issueIOST() {
        let total1 = blockchain.callWithAuth(TokenContract, "supply", ["iost"])[0];
        blockchain.callWithAuth(IssueContract, "issueIOST", []);
        let total2 = blockchain.callWithAuth(TokenContract, "supply", ["iost"])[0];
        let data = this._get("issueIOST") || [];
        data.push(new Float64(total2).minus(total1).toFixed(8));
        this._put("issueIOST", data);
        return data;
    }

    exchangeIOST() {
        let publisher = blockchain.publisher();
        let balance10 = blockchain.callWithAuth(TokenContract, "balanceOf", ["iost", BonusContract])[0];
        let balance11 = blockchain.callWithAuth(TokenContract, "balanceOf", ["iost", publisher])[0];
        blockchain.callWithAuth(BonusContract, "exchangeIOST", [publisher, "0"]);
        let balance20 = blockchain.callWithAuth(TokenContract, "balanceOf", ["iost", BonusContract])[0];
        let balance21 = blockchain.callWithAuth(TokenContract, "balanceOf", ["iost", publisher])[0];
        let data = this._get("exchangeIOST") || {};
        data[publisher] = {
            BonusContract: new Float64(balance20).minus(balance10).toFixed(8),
            publisher: new Float64(balance21).minus(balance11).toFixed(8),
        }
        this._put("exchangeIOST", data);
        return data;
    }

    candidateWithdraw() {
        let publisher = blockchain.publisher();
        let balance10 = blockchain.callWithAuth(TokenContract, "balanceOf", ["iost", VPContract])[0];
        let balance11 = blockchain.callWithAuth(TokenContract, "balanceOf", ["iost", publisher])[0];
        let bonus = blockchain.callWithAuth(VPContract, "getCandidateBonus", [publisher])[0];
        blockchain.callWithAuth(VPContract, "candidateWithdraw", [publisher]);
        let balance20 = blockchain.callWithAuth(TokenContract, "balanceOf", ["iost", VPContract])[0];
        let balance21 = blockchain.callWithAuth(TokenContract, "balanceOf", ["iost", publisher])[0];
        let voteInfo =  this._call(VPContract, "getProducer", [publisher]);
        let vote = this._get("vote");
        let unvote = this._get("unvote");
        let votes = {};
        for (let a in vote) {
            let v = (vote[a][publisher] || {})["vote"] || "0";
            votes[a] = new Float64(v).plus(votes[a] || "0").toFixed(8);
        }
        for (let a in unvote) {
            let v = (unvote[a][publisher] || {})["vote"] || "0";
            votes[a] = new Float64(votes[a] || "0").minus(v).toFixed(8);
        }
        let data = this._get("candidateWithdraw") || {};
        data[publisher] = {
            bonus: bonus,
            VPContract: new Float64(balance20).minus(balance10).toFixed(8),
            publisher: new Float64(balance21).minus(balance11).toFixed(8),
            votes: votes,
            totalVotes: voteInfo.voteInfo.votes,
        }
        this._put("candidateWithdraw", data);
        return data;
    }

    topupVoterBonus(account, amount) {
        let publisher = blockchain.publisher();
        let balance10 = blockchain.callWithAuth(TokenContract, "balanceOf", ["iost", VPContract])[0];
        let balance11 = blockchain.callWithAuth(TokenContract, "balanceOf", ["iost", publisher])[0];
        blockchain.callWithAuth(VPContract, "topupVoterBonus", [account, amount, publisher]);
        let balance20 = blockchain.callWithAuth(TokenContract, "balanceOf", ["iost", VPContract])[0];
        let balance21 = blockchain.callWithAuth(TokenContract, "balanceOf", ["iost", publisher])[0];
        let voteInfo =  this._call(VPContract, "getProducer", [account]);
        let vote = this._get("vote");
        let unvote = this._get("unvote");
        let votes = {};
        for (let a in vote) {
            let v = (vote[a][account] || {}).vote || "0";
            votes[a] = new Float64(v).plus(votes[a] || "0").toFixed(8);
        }
        for (let a in unvote) {
            let v = (unvote[a][account] || {}).vote || "0";
            votes[a] = new Float64(votes[a] || "0").minus(v).toFixed(8);
        }
        let data = this._get("topupVoterBonus") || {};
        data[account] = {
            VPContract: new Float64(balance20).minus(balance10).toFixed(8),
            publisher: new Float64(balance21).minus(balance11).toFixed(8),
            votes: votes,
            totalVotes: voteInfo.voteInfo.votes,
        }
        this._put("topupVoterBonus", data);
        return data;
    }

    voterWithdraw() {
        let publisher = blockchain.publisher();
        let balance10 = blockchain.callWithAuth(TokenContract, "balanceOf", ["iost", VPContract])[0];
        let balance11 = blockchain.callWithAuth(TokenContract, "balanceOf", ["iost", publisher])[0];
        blockchain.callWithAuth(VPContract, "voterWithdraw", [publisher]);
        let balance20 = blockchain.callWithAuth(TokenContract, "balanceOf", ["iost", VPContract])[0];
        let balance21 = blockchain.callWithAuth(TokenContract, "balanceOf", ["iost", publisher])[0];
        let data = this._get("voterWithdraw") || {};
        data[publisher] = {
            VPContract: new Float64(balance20).minus(balance10).toFixed(8),
            publisher: new Float64(balance21).minus(balance11).toFixed(8),
        }
        this._put("voterWithdraw", data);
        return data;
    }

    checkResult() {
        let data =  {
            vote: this._get("vote"),
            unvote: this._get("unvote"),
            issueIOST: this._get("issueIOST"),
            exchangeIOST: this._get("exchangeIOST"),
            topupVoterBonus: this._get("topupVoterBonus"),
            candidateWithdraw: this._get("candidateWithdraw"),
            voterWithdraw: this._get("voterWithdraw"),
            checkResult: {},
        }

        for (let voter in data.voterWithdraw) {
            let voterBonus = new Float64("0");
            for (let pro in data.candidateWithdraw) {
                let v = data.candidateWithdraw[pro].votes[voter];
                let totalb = new Float64(data.candidateWithdraw[pro].bonus).minus(data.candidateWithdraw[pro].publisher);
                let b = new Float64(v).div(data.candidateWithdraw[pro].totalVotes).multi(totalb);
                voterBonus = voterBonus.plus(b);
            }
            for (let pro in data.topupVoterBonus) {
                let v = data.topupVoterBonus[pro].votes[voter];
                let b = new Float64(v).div(data.topupVoterBonus[pro].totalVotes).multi(data.topupVoterBonus[pro].VPContract);
                voterBonus = voterBonus.plus(b);
            }
            if (voterBonus.toFixed(8) !== data.voterWithdraw[voter].publisher) {
                throw new Error("voter bonus not right: " + voterBonus.toFixed(8) + " != " + data.voterWithdraw[voter].publisher);
            }
            data.checkResult[voter] = voterBonus.toFixed(8);
        }
        return data;
    }
}

module.exports = VoteChecker;