class Str{
    ops() {
        let str = "abcdefg";
        if (str.charAt(2) !== "c") {
            throw "charAt error";
        }
        if (str.charCodeAt(2) !== 99) {
            throw 'charCodeAt error';
        }
        if (str.length !== 7) {
            throw "length error";
        }
        if (str.toString() !== str) {
            throw 'toString error';
        }
        if (str.valueOf() !== str) {
            throw 'valueOf error';
        }
        if (str.concat("hi") !== str + "hi") {
            throw 'concat error';
        }
        if (str.includes("cde") !== true) {
            throw 'includes error';
        }
        if (str.includes("hi") !== false) {
            throw 'includes error';
        }
        if (str.endsWith("efg") !== true) {
            throw 'endsWith error';
        }
        if (str.endsWith("abc") !== false) {
            throw 'endsWith error';
        }
        if (str.indexOf("c") !== 2) {
            throw 'indexOf error';
        }
        if (str.lastIndexOf("c") !== 2) {
            throw 'lastIndexOf error';
        }
        if (str.replace("c", "b") !== "abbdefg") {
            throw 'replace error';
        }
        if (str.search("[abc]c") !== 1) {
            throw 'search error';
        }
        if (str.split("c")[0] !== "ab") {
            throw 'split error';
        }
        if (str.startsWith("abc") !== true) {
            throw 'startsWith error';
        }
        if (str.startsWith("efg") !== false) {
            throw 'startsWith error';
        }
        if (str.slice(1, 3) !== "bc") {
            throw 'slice error';
        }
        if (str.substring(1, 3) !== "bc") {
            throw 'substring error';
        }
        if (str.toUpperCase() !== "ABCDEFG") {
            throw 'toUpperCase error';
        }
        if (str.toUpperCase().toLowerCase() !== str) {
            throw 'toLowerCase error';
        }
        if ((" " + str + " ").trim() !== str) {
            throw 'trim error';
        }
        if ((" " + str + " ").trimLeft() !== (str + " ")) {
            throw 'trimLeft error';
        }
        if ((" " + str + " ").trimRight() !== (" " + str)) {
            throw 'trimRight error';
        }
        if (str.repeat(3) !== (str + str + str)) {
            throw 'repeat error';
        }
    }
}

module.exports = Str;