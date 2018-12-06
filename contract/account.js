class Account {
    constructor() {

    }
    init() {

    }
    _saveAccount(account, payer) {
        if (payer === undefined) {
            payer = account.id
        }
        storage.mapPut("auth", account.id, JSON.stringify(account), payer);
    }
    _loadAccount(id) {
        let a = storage.mapGet("auth", id);
        return JSON.parse(a);
    }
    static _find(items, name) {
        for (let i = 0; i < items.length(); i ++) {
            if (items[i].id === name ) {
                return i
            }
        }
        return -1
    }

    _hasAccount(id) {
        return storage.mapHas("auth", id);
    }

    _ra(id) {
        if (!BlockChain.requireAuth(id, "owner")) {
            throw new Error("require auth failed");
        }
    }

    _checkIdValid(id) {
        if (block.number === 0) {
            return
        }
        if (id.length < 5 || id.length > 11) {
            throw new Error("id invalid. id length should be between 5,11 > " + id)
        }
        for (let i in id) {
            let ch = id[i];
            if (!(ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' || ch === '_')) {
                throw new Error("id invalid. id contains invalid character > " + ch);
            }
        }
    }

    SignUp(id, owner, active) {
        if (this._hasAccount(id)) {
            throw new Error("id existed > " + id);
        }
        this._checkIdValid(id);
        let account = {};
        account.id = id;
        account.permissions = {};
        account.permissions.active = {
            name: "active",
            groups: [],
            items: [{
                id: active,
                is_key_pair: true,
                weight: 1,
            }],
            threshold: 1,
        };
        account.permissions.owner = {
            name: "owner",
            groups: [],
            items: [{
                id: owner,
                is_key_pair: true,
                weight: 1,
            }],
            threshold: 1,
        };
        this._saveAccount(account, BlockChain.publisher());
    }
    AddPermission(id, perm, thres) {
        this._ra(id);
        let acc = this._loadAccount(id);
        if (acc.permissions[perm] !== undefined) {
            throw new Error("permission already exist");
        }
        acc.permissions[perm] = {
            name: perm,
            groups: [],
            items: [],
            threshold: thres,
        };
        this._saveAccount(acc);
    }
    DropPermission(id, perm) {
        this._ra(id);
        let acc = this._loadAccount(id);
        acc.permissions[perm] = undefined;
        this._saveAccount(acc);
    }
    AssignPermission(id, perm, un, weight) {
        this._ra(id);
        let acc = this._loadAccount(id);
        const index = Account._find(acc.permissions[perm].items, un);
        if (index < 0) {
            const len = un.indexOf("@");
            if (len < 0 && un.startsWith("IOST") ){
                acc.permissions[perm].items.push({
                    id : un,
                    is_key_pair: true,
                    weight: weight
                });
            } else {
                acc.permissions[perm].items.push({
                    id : un.substring(0, len),
                    permission: un.substring(len, un.length()),
                    is_key_pair: false,
                    weight: weight
                });
            }
        } else {
            acc.permissions[perm].items[index].weight = weight
        }
        this._saveAccount(acc);
    }
    RevokePermission(id, perm, un) {
        this._ra(id);
        let acc = this._loadAccount(id);
        const index = Account._find(acc.permissions[perm].items, un);
        if (index < 0) {
            throw new Error("item not found");
        } else {
            acc.permissions[perm].items.splice(index, 1)
        }
        this._saveAccount(acc);
    }
    AddGroup(id, grp) {
        this._ra(id);
        let acc = this._loadAccount(id);
        if (acc.groups[grp] !== undefined) {
            throw new Error("group already exist");
        }
        acc.groups[grp] = {
            name : grp,
            items : [],
        };
        this._saveAccount(acc);
    }
    DropGroup(id, group) {
        this._ra(id);
        let acc = this._loadAccount(id);
        acc.permissions[group] = undefined;
        this._saveAccount(acc);
    }
    AssignGroup(id, group, un, weight) {
        this._ra(id);
        let acc = this._loadAccount(id);
        const index = Account._find(acc.permissions[group].items, un);
        if (index < 0) {
            let len = un.indexOf("@");
            if (len < 0 && un.startsWith("IOST")) {
                acc.groups[group].items.push({
                    id: un,
                    is_key_pair: true,
                    weight: weight
                });
            } else {
                acc.groups[group].items.push({
                    id: un.substring(0, len),
                    permission: un.substring(len, un.length()),
                    is_key_pair: false,
                    weight: weight
                });
            }
        } else {
            acc.groups[group].items[index].weight = weight
        }

        this._saveAccount(acc);
    }
    RevokeGroup(id, grp, un) {
        this._ra(id);
        let acc = this._loadAccount(id);
        const index = Account._find(acc.permissions[grp].items, un);
        if (index < 0) {
            throw new Error("item not found");
        } else {
            acc.permissions[grp].items.splice(index, 1)
        }
        this._saveAccount(acc);
    }
    AssignPermissionToGroup(id, perm, group) {
        this._ra(id);
        let acc = this._loadAccount(id);
        acc.permissions[perm].groups.push(group);
        this._saveAccount(acc);
    }
    RevokePermissionInGroup(id, perm, group) {
        this._ra(id);
        let acc = this._loadAccount(id);
        let index = acc.permissions[perm].groups.indexOf(group);
        if (index > -1) {
            acc.permissions[perm].groups.splice(index, 1);
        }
        this._saveAccount(acc);
    }
}

module.exports = Account;
