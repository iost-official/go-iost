//  modified from json2.js (https://github.com/douglascrockford/JSON-js)
(function () {
    "use strict";

    const rx_escapable = /[\\"\u0000-\u001f\u007f-\u009f\u00ad\u0600-\u0604\u070f\u17b4\u17b5\u200c-\u200f\u2028-\u202f\u2060-\u206f\ufeff\ufff0-\uffff]/g;

    function f(n) {
        _IOSTInstruction_counter.incr(12);
        return (n < 10)
            ? "0" + n
            : n;
    }

    if (typeof Boolean.prototype.toJSON !== "function") {
        Boolean.prototype.toJSON = Boolean.prototype.valueOf;
    }
    if (typeof Number.prototype.toJSON !== "function") {
        Number.prototype.toJSON = Number.prototype.valueOf;
    }
    if (typeof String.prototype.toJSON !== "function") {
        String.prototype.toJSON = String.prototype.valueOf;
    }
    if (typeof Date.prototype.toJSON !== "function") {
        Date.prototype.toJSON = function () {
            return isFinite(this.valueOf())
                ? (
                    this.getUTCFullYear()
                    + "-"
                    + f(this.getUTCMonth() + 1)
                    + "-"
                    + f(this.getUTCDate())
                    + "T"
                    + f(this.getUTCHours())
                    + ":"
                    + f(this.getUTCMinutes())
                    + ":"
                    + f(this.getUTCSeconds())
                    + "Z"
                )
                : null;
        };
    }

    let gap;
    let indent;
    let meta;
    let rep;
    const dup = Symbol();

    function quote(string) {
        _IOSTInstruction_counter.incr(12 + 2 * string.length);
        rx_escapable.lastIndex = 0;
        return rx_escapable.test(string)
            ? "\"" + string.replace(rx_escapable, function (a) {
                let c = meta[a];
                return typeof c === "string"
                    ? c
                    : "\\u" + ("0000" + a.charCodeAt(0).toString(16)).slice(-4);
            }) + "\""
            : "\"" + string + "\"";
    }

    function str(key, holder) {
        _IOSTInstruction_counter.incr(56);
        let i;          // The loop counter.
        let k;          // The member key.
        let v;          // The member value.
        let length;
        let mind = gap;
        let partial;
        let value = holder[key];

        if (
            value
            && typeof value === "object"
            && typeof value.toJSON === "function"
        ) {
            _IOSTInstruction_counter.incr(8);
            value = value.toJSON(key);
        }

        if (typeof rep === "function") {
            _IOSTInstruction_counter.incr(8);
            value = rep.call(holder, key, value);
        }

        switch (typeof value) {
        case "string":
            v = quote(value);
            _IOSTInstruction_counter.incr(v.length);
            return v;
        case "number":
            v = (isFinite(value))
                ? String(value)
                : "null";
            _IOSTInstruction_counter.incr(v.length);
            return v;
        case "boolean":
        case "null":
            v = String(value);
            _IOSTInstruction_counter.incr(v.length);
            return v;
        case "object":
            if (!value) {
                _IOSTInstruction_counter.incr(4);
                return "null";
            }
            if (value.hasOwnProperty(dup)) {
                throw new Error("Converting circular structure to JSON");
            }
            Object.defineProperty(value, dup, {configurable: true});
            gap += indent;
            partial = [];
            if (Array.isArray(value)) {
                length = value.length;
                _IOSTInstruction_counter.incr(16 + 24 * length);
                for (i = 0; i < length; i += 1) {
                    partial[i] = str(i, value) || "null";
                }
                v = partial.length === 0
                    ? "[]"
                    : gap
                        ? (
                            "[\n"
                            + gap
                            + partial.join(",\n" + gap)
                            + "\n"
                            + mind
                            + "]"
                        )
                        : "[" + partial.join(",") + "]";
                gap = mind;
                _IOSTInstruction_counter.incr(v.length);
                delete(value[dup]);
                return v;
            }

            if (rep && typeof rep === "object") {
                length = rep.length;
                _IOSTInstruction_counter.incr(16 + 16 * length);
                for (i = 0; i < length; i += 1) {
                    if (typeof rep[i] === "string") {
                        k = rep[i];
                        v = str(k, value);
                        if (v) {
                            partial.push(quote(k) + (
                                (gap)
                                    ? ": "
                                    : ":"
                            ) + v);
                        }
                    }
                }
            } else {
                const keys = Object.keys(value);
                _IOSTInstruction_counter.incr(16 + 16 * keys.length);
                for (k of keys) {
                    v = str(k, value);
                    if (v) {
                        partial.push(quote(k) + (
                            (gap)
                                ? ": "
                                : ":"
                        ) + v);
                    }
                }
            }

            _IOSTInstruction_counter.incr(8 * partial.length);
            v = partial.length === 0
                ? "{}"
                : gap
                    ? "{\n" + gap + partial.join(",\n" + gap) + "\n" + mind + "}"
                    : "{" + partial.join(",") + "}";
            gap = mind;
            _IOSTInstruction_counter.incr(v.length);
            delete(value[dup]);
            return v;
        }
    }

    meta = {    // table of character substitutions
        "\b": "\\b",
        "\t": "\\t",
        "\n": "\\n",
        "\f": "\\f",
        "\r": "\\r",
        "\"": "\\\"",
        "\\": "\\\\"
    };
    JSON.stringify = function (value, replacer, space) {
        _IOSTInstruction_counter.incr(24);
        let i;
        gap = "";
        indent = "";

        if (typeof space === "number") {
            _IOSTInstruction_counter.incr(2 * Math.abs(space));
            for (i = 0; i < space; i += 1) {
                indent += " ";
            }
        } else if (typeof space === "string") {
            indent = space;
        }

        rep = replacer;
        if (replacer && typeof replacer !== "function" && (
            typeof replacer !== "object"
            || typeof replacer.length !== "number"
        )) {
            throw new Error("JSON.stringify");
        }

        const rs = str("", {"": value});
        if (typeof rs == "string") {
            _IOSTInstruction_counter.incr(2 * rs.length);
        }
        return rs;
    };
}());