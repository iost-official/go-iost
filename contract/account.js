class Account {
    constructor() {

    }
    init() {

    }
    _saveAccount(account) {
        storage.mapPut("account", account.id, JSON.stringify(account))
    }
    _loadAccount(id) {
        let a = storage.mapGet("account", id);
        return JSON.parse(a)
    }

    SignUp(id, owner, active) {
        let account = {};
        account.id = id;
        account.permissions = {};
        account.permissions.active = {
            name: "active",
            groups: [],
            users: [{
                id: active,
                is_key_pair: true,
            }],
            threshold: 1,
        };
        account.permissions.owner = {
            name: "owner",
            groups: [],
            users: [{
                id: owner,
                is_key_pair: true,
            }],
            threshold: 1,
        };
        this._saveAccount(account)
    }
    AddPermission(id, perm, thres) {
        let acc = this._loadAccount(id);
        acc.permissions[perm] = {
            name: perm,
            groups: [],
            users: [],
            threshold: thres,
        };
        this._saveAccount(acc)
    }
    DropPermission(id, perm) {
        let acc = this._loadAccount(id);
        acc.permissions[perm] = undefined;
        this._saveAccount(acc)
    }
    AssignPermission(id, perm, un, weight) {
        let acc = this._loadAccount(id);
        let len = un.indexOf("@");
        if (len < 0 && un.startsWith("IOST") ){
            acc.permissions[perm].users.push({
                id : un,
                is_key_pair: true,
                weight: weight
            });
        } else {
            acc.permissions[perm].users.push({
                id : un.substring(0, len),
                permission: un.substring(len, un.length()),
                is_key_pair: false,
                weight: weight
            });
        }

        this._saveAccount(acc)
    }
    AddGroup(id, grp) {
        let acc = this._loadAccount(id);
        acc.groups[grp] = {
            name : grp,
            users : [],
        };
        this._saveAccount(acc)
    }
    AssignGroup(id, group, un, weight) {
        let acc = this._loadAccount(id);
        let len = un.indexOf("@");
        if (len < 0 && un.startsWith("IOST")){
            acc.groups[group].users.push({
                id : un,
                is_key_pair: true,
                weight: weight
            });
        } else {
            acc.groups[group].users.push({
                id : un.substring(0, len),
                permission: un.substring(len, un.length()),
                is_key_pair: false,
                weight: weight
            });
        }

        this._saveAccount(acc)
    }
    AddPermissionToGroup(id, perm, group) {
        let acc = this._loadAccount(id);
        acc.permissions[perm].groups.push(group);
        this._saveAccount(acc)
    }
    DelPermissionInGroup(id, perm, group) {
        let acc = this._loadAccount(id);
        let index = acc.permissions[perm].groups.indexOf(group);
        if (index > -1) {
            acc.permissions[perm].groups.splice(index, 1);
        }
        this._saveAccount(acc)
    }
    DropGroup(id, group) {
        let acc = this._loadAccount(id);
        acc.permissions[group] = undefined;
        this._saveAccount(acc)
    }
}

module.exports = Account;