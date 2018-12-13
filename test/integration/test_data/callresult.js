class Test {
    init() {
    }
    ret_eof() {
        storage.put("ret", "ab\x00c");
        const res = storage.get("ret");
        console.log("js result is " + res.replace("\x00", "\\x00"));
        return res + "d";
    }
    ret_obj() {
        return {
            toJSON: function() {
                throw new Error("error in JSON.stringfy");
            }
        };
    }
}

module.exports = Test;