class Test {
    init() {
        storage.put("num", JSON.stringify(9));
        storage.put("string", "hello");
        storage.put("bool",  JSON.stringify(true));
        storage.put("array", JSON.stringify([1,2,3]));
        storage.put("obj", JSON.stringify({"foo": "bar"}));

        storage.mapPut("map", "field1", "value");
        storage.mapPut("map", "field2", "v2")

    }

    read() {
        const num = JSON.parse(storage.get("num"));
        _native_log("num > "+num ); // 9

        const str = storage.get("string");
        _native_log("str > "+str ); //  "hello";

        const bool = JSON.parse(storage.get("bool"));
        _native_log("bool > "+bool? "true":"false" );

        const array = JSON.parse(storage.get("array"));
        _native_log("array > " + array.toString());

        const obj = JSON.parse(storage.get("obj"));
        _native_log("obj > " + obj.foo); // bar

        const map1 = storage.mapGet("map", "field1");
        _native_log("map > " + map1);

        const keys = storage.mapKeys("map");
        _native_log("keys > " + keys.toString());
    }

    change() {
        // this.array.push(7);
        // this.object["foo"] = "hahaha";
        // this["key"]["foo"] = "baz"
        // this.arrayobj[0]["foo"] = "oh bad" // error
        // this.objobj["i am"]["your"] = "grandpa"  // error

    }

    can_update(d) {
        return true
    }
}
module.exports = Test;
