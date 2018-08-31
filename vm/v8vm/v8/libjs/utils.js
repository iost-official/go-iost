'use strict';

const emptyOptions = {};
function format(...args) {
    return formatWithOptions(emptyOptions, ...args);
}

function formatWithOptions(inspectOptions, f) {
    let i, tempStr;
    _native_log(JSON.stringify(f))
    if (typeof f !== 'string') {
        if (arguments.length === 1) return '';
        let res = '';
        for (i = 1; i < arguments.length - 1; i++) {
            _native_log(JSON.stringify(arguments[i]))
            res += inspect(arguments[i], inspectOptions);
            res += ' ';
        }
        res += inspect(arguments[i], inspectOptions);
        return res;
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
                    case 106: // 'j'
                        tempStr = tryStringify(arguments[a++]);
                        break;
                    case 100: // 'd'
                        tempStr = `${Number(arguments[a++])}`;
                        break;
                    case 79: // 'O'
                        tempStr = inspect(arguments[a++], inspectOptions);
                        break;
                    case 111: // 'o'
                    {
                        const opts = Object.assign({}, inspectOptions, {
                            showHidden: true,
                            showProxy: true,
                            depth: 4
                        });
                        tempStr = inspect(arguments[a++], opts);
                        break;
                    }
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
            str += ` ${inspect(x, inspectOptions)}`;
        }
    }
    return str;
}

/**
 * Echos the value of any input. Tries to print the value out
 * in the best way possible given the different types.
 *
 * @param {any} value The value to print out.
 * @param {Object} opts Optional options object that alters the output.
 */
/* Legacy: value, showHidden, depth, colors */
function inspect(value, opts) {
    // Default options
    const ctx = {
        seen: [],
        stylize: stylizeNoColor,
        showHidden: inspectDefaultOptions.showHidden,
        depth: inspectDefaultOptions.depth,
        colors: inspectDefaultOptions.colors,
        customInspect: inspectDefaultOptions.customInspect,
        showProxy: inspectDefaultOptions.showProxy,
        // TODO(BridgeAR): Deprecate `maxArrayLength` and replace it with
        // `maxEntries`.
        maxArrayLength: inspectDefaultOptions.maxArrayLength,
        breakLength: inspectDefaultOptions.breakLength,
        indentationLvl: 0,
        compact: inspectDefaultOptions.compact
    };
    // Legacy...
    if (arguments.length > 2) {
        if (arguments[2] !== undefined) {
            ctx.depth = arguments[2];
        }
        if (arguments.length > 3 && arguments[3] !== undefined) {
            ctx.colors = arguments[3];
        }
    }
    // Set user-specified options
    if (typeof opts === 'boolean') {
        ctx.showHidden = opts;
    } else if (opts) {
        const optKeys = Object.keys(opts);
        for (var i = 0; i < optKeys.length; i++) {
            ctx[optKeys[i]] = opts[optKeys[i]];
        }
    }
    if (ctx.colors) ctx.stylize = stylizeWithColor;
    if (ctx.maxArrayLength === null) ctx.maxArrayLength = Infinity;
    return formatValue(ctx, value, ctx.depth);
}

module.exports = format;