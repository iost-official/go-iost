class Test {
    init() {
    }
    ret_eof() {
        storage.put("ret", "ab\x00c");
        const res = storage.get("ret");
        console.log("js result is " + res.replace("\x00", "\\x00"));
        return res + "d";
    }
}

module.exports = Test;