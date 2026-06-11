//  JSON.stringify wrapper with gas estimation for QuickJS
(function () {
    "use strict";

    const nativeStringify = JSON.stringify;
    const TypeErrorRef = TypeError;

    // Iterative walk to avoid WASM stack overflow on deeply nested objects.
    function estimateGas(value) {
        let gas = 0;
        let stack = [value];

        while (stack.length > 0) {
            let v = stack.pop();
            gas += 56;
            switch (typeof v) {
            case "string":
                gas += 12 + 2 * v.length + v.length;
                break;
            case "number":
                gas += (isFinite(v) ? String(v) : "null").length;
                break;
            case "boolean":
            case "null":
                gas += String(v).length;
                break;
            case "object":
                if (!v) {
                    gas += 4;
                    break;
                }
                if (Array.isArray(v)) {
                    let len = v.length;
                    gas += 16 + 24 * len;
                    for (let i = len - 1; i >= 0; i--) {
                        stack.push(v[i]);
                    }
                    gas += len === 0 ? 2 : 2 + len;
                    break;
                }
                let keys = Object.keys(v);
                gas += 16 + 16 * keys.length;
                for (let i = keys.length - 1; i >= 0; i--) {
                    let k = keys[i];
                    gas += 12 + 2 * k.length + k.length;
                    stack.push(v[k]);
                }
                gas += 8 * keys.length;
                gas += keys.length === 0 ? 2 : 2 + keys.length;
                break;
            }
            // undefined / symbol / function fall through, no extra gas
        }

        return gas;
    }

    JSON.stringify = function (value, replacer, space) {
        let gas = 24; // wrapper entry

        if (typeof space === "number") {
            gas += 2 * Math.abs(space);
        } else if (typeof space === "string") {
            // indent string cost is not explicitly charged in old json.js
        }

        let rs;
        try {
            rs = nativeStringify(value, replacer, space);
        } catch (e) {
            const msg = (e && e.message) ? e.message : String(e);
            if (msg.toLowerCase().indexOf("circular") !== -1) {
                throw new TypeErrorRef("Converting circular structure to JSON");
            }
            throw e;
        }

        if (typeof rs === "string") {
            try {
                gas += estimateGas(value);
            } catch (e) {
                throw new Error("estimateGas failed: " + (e && e.message ? e.message : e));
            }
            gas += 3 * rs.length + 4;
        }

        __IOST_internal_gas_acc += gas;
        return rs;
    };
}());
