'use strict';

const emptyOptions = {};
function format(...args) {
    return formatWithOptions(emptyOptions, ...args);
}

function formatWithOptions(inspectOptions, f) {
    let i, tempStr;
    if (typeof f !== 'string') {
        let args = Array.from(arguments);
        args.shift();
        if (args.length === 1) {
            return JSON.stringify(args[0])
        }
        return JSON.stringify(args);
    }

    if (arguments.length === 2) return f;

    let str = '';
    let a = 2;
    let lastPos = 0;
    for (i = 0; i < f.length - 1; i++) {
        if (f.charCodeAt(i) === 37) { // '%'
            const nextChar = f.charCodeAt(++i);
            if (a !== arguments.length) {
                switch (nextChar) {
                    case 115: // 's'
                        tempStr = String(arguments[a++]);
                        break;
                    case 100: // 'd'
                        let curr = arguments[a++];
                        if (curr instanceof Int64 || curr instanceof BigNumber) {
                            tempStr = `${curr.toString()}`
                        } else {
                            tempStr = `${Number(arguments[a++])}`;
                        }
                        break;
                    case 118: // 'v'
                        tempStr = JSON.stringify(arguments[a++]);
                        break;
                    case 105: // 'i'
                        tempStr = `${parseInt(arguments[a++])}`;
                        break;
                    case 102: // 'f'
                        tempStr = `${parseFloat(arguments[a++])}`;
                        break;
                    case 37: // '%'
                        str += f.slice(lastPos, i);
                        lastPos = i + 1;
                        continue;
                    default: // any other character is not a correct placeholder
                        continue;
                }
                if (lastPos !== i - 1)
                    str += f.slice(lastPos, i - 1);
                str += tempStr;
                lastPos = i + 1;
            } else if (nextChar === 37) {
                str += f.slice(lastPos, i);
                lastPos = i + 1;
            }
        }
    }
    if (lastPos === 0)
        str = f;
    else if (lastPos < f.length)
        str += f.slice(lastPos);
    while (a < arguments.length) {
        const x = arguments[a++];
        if ((typeof x !== 'object' && typeof x !== 'symbol') || x === null) {
            str += ` ${x}`;
        } else {
            str += ` ${JSON.stringify(x)}`;
        }
    }
    return str;
}