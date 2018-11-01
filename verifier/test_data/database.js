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
        if (num !== 9) {
            throw num
        }

        const str = storage.get("string");
        if (str !== "hello") {
            throw str
        }

        const bool = JSON.parse(storage.get("bool"));
        if (!bool) {
            throw bool
        }

        const array = JSON.parse(storage.get("array"));
        if (array[0] !== 1) {
            throw array
        }

        const obj = JSON.parse(storage.get("obj"));
        if (obj.foo !== "bar") {
            throw obj
        }

        const map1 = storage.mapGet("map", "field1");
        if (map1 !== "value") {
            throw map1
        }
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
