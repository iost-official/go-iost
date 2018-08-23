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
    }

    main() {
        this.num = 10;
        this.string = "hello world";
        this.bool = false;
        this.array = [1, 2, 3, 4, 5, 6];
        this.object = {"foo": "bar", "abc": 123};
        this.arrayobj = [{"foo": "bar"}, {"abc": 123}, {"ok": true}];
        this.objobj = {"i am": {"your": "father"}}
    }

    change() {
        this.array.push(7);
        this.object["foo"] = "hahaha"
        // this.arrayobj[0]["foo"] = "oh bad" // error
        // this.objobj["i am"]["your"] = "grandpa"  // error

    }

    can_update(d) {
        return true
    }
}
module.exports = Test;
