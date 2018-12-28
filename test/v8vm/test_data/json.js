class Test {
    stringify10() {
        const arr = {
            a: [1,2,3],
            b: [1,2,3],
            c: [1,2,3],
            d: [1,2,3],
            e: ["abc","abc","abc"],
            f: ["abc","abc","abc"],
            g: ["abc","abc","abc"],
            h: ["abcdefghijklmn","abcdefghijklmn","abcdefghijklmn"],
            i: ["abcdefghijklmn","abcdefghijklmn","abcdefghijklmn"],
        };
        return arr;
    }
    stringify11() {
        const arr = {
            a: [1,2,3],
            b: [1,2,3],
            c: [1,2,3],
            d: [1,2,3],
            e: ["abc","abc","abc"],
            f: ["abc","abc","abc"],
            g: ["abc","abc","abc"],
            h: ["abcdefghijklmn","abcdefghijklmn","abcdefghijklmn"],
            i: ["abcdefghijklmn","abcdefghijklmn","abcdefghijklmn"],
        };
        let res = [];
        for (let k in arr) {
            let r = [];
            for (let i of arr[k]) {
                if (Number.isInteger(i) || typeof i === "string") {
                    r.push('"' + i.valueOf() + '"');
                }
            }
            res.push('"' + k +'":[' + r.join(',') + ']');
        }
        return '{' + res.join(',') + '}';
    }

    stringify20() {
        let obj = {};
        for (let i = 0; i < 3000; i++) {
            obj["a" + String(i)] = "\n";
        }
        return obj;
    }
    stringify21() {
        let obj = {};
        for (let i = 0; i < 3000; i++) {
            obj["a" + String(i)] = "\n";
        }
        let res = [];
        for (let i in obj) {
            if (typeof obj[i] === "string") {
                if (obj[i] === "\n") {
                    res.push("\"" + i + '":"' + '\\n' + '"');
                }
            }
        }
        return "{" + res.join(",") + "}";
    }

    stringify30() {
        let a = "a";
        for (let i = 0; i < 1000; i++) {
            a = {
                a : a
            };
        }
        return a;
    }
    stringify31() {
        let a = "a";
        for (let i = 0; i < 1000; i++) {
            a = {
                a : a
            };
        }
        let str = function(a) {
            if (typeof a === "object") {
                return '{"a":' + str(a['a']) + '}';
            } else {
                return '"' + a + '"';
            }
        }
        return str(a);
    }

    stringify40() {
        let a = {};
        a.a = a;
        return a;
    }

    stringify50() {
        function replacer(key, value) {
            // Filtering out properties
            if (typeof value === 'string') {
                return undefined;
            }
            return value;
        }
        let foo = {foundation: 'Mozilla', model: 'box', week: 45, transport: 'car', month: 7};
        return JSON.stringify(foo, replacer);
    }
    stringify51() {
        let foo = {foundation: 'Mozilla', model: 'box', week: 45, transport: 'car', month: 7};
        return JSON.stringify(foo, ['week', 'month']);
    }

    stringify60() {
        let a = {a: {b:{c:""}}};
        let s1 = JSON.stringify(a);
        let s2 = JSON.stringify(a);
        return s1 + " " + s2;
    }
}

module.exports = Test;