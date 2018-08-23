class Test {
    constructor() {
        this.num = 9;
        this.string = "hello";
        this.bool = true;
        this.nil = null;
        this.array = [1, 2, 3];
        this.object = {"foo": "bar"};
        this.arrayobj = [{"foo": "bar"}, {"foo": "bar"}, {"foo": "bar"}];
        this.objobj = {"you": {"killed": "my father"}}
        this["key"] = {"foo": "bar"}
    }

    read() {
        _native_log("num > "+this.num );//  9;
        _native_log("str > "+this.string );//  "hello";
        _native_log("boo > "+this.bool );//  true;
        _native_log("arr > "+this.array );//  [1, 2, 3]
        _native_log("obj > "+this.object );//  {"foo": "bar"};
        _native_log("key > "+this["key"])
        // _native_log("aio > "+this.arrayobj );//  [{"foo": "bar"}, {"abc": 123}, {"ok": true}];
        // _native_log("oio > "+this.objobj )//  {"i am": {"your": "father"}}
    }

    change() {
        this.array.push(7);
        this.object["foo"] = "hahaha";
        this["key"]["foo"] = "baz"
        // this.arrayobj[0]["foo"] = "oh bad" // error
        // this.objobj["i am"]["your"] = "grandpa"  // error

    }

    can_update(d) {
        return true
    }
}
module.exports = Test;
