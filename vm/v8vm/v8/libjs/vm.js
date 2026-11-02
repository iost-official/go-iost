
// other helper functions
// var BigNumber = require('bignumber');
// var Int64 = require('int64');

// var injectGas = require('inject_gas');

// crypto
const IOSTCrypto = new _IOSTCrypto;

const _IOSTInstruction_counter = new IOSTInstruction;

// + - * / % **, | & ^ >> >>> <<, || &&, == != === !== > >= < <=, instanceOf in
const _IOSTBinaryOp = function(left, right, op) {
    if ((typeof left === "string" || typeof right === "string") &&
        (op === "==" || op === "!=" || op === "===" || op === "!==" || op === "<" || op === "<=" || op === ">" || op === ">=")) {
        _IOSTInstruction_counter.incr(left === null || left === undefined || left.toString().length <= 0 ? 0 : left.toString().length);
        _IOSTInstruction_counter.incr(right === null || right === undefined || right.toString().length <= 0 ? 0 : right.toString().length);
    }
    _IOSTInstruction_counter.incr(3);
    switch (op) {
        case '+':
            const ret = left + right;
            if (typeof ret === "string") {
                _IOSTInstruction_counter.incr(ret.length);
            }
            return ret;
        case '-':
            return left - right;
        case '*':
            return left * right;
        case '/':
            return left / right;
        case '%':
            return left % right;
        case '**':
            return left ** right;
        case '|':
            return left | right;
        case '&':
            return left & right;
        case '^':
            return left ^ right;
        case '>>':
            return left >> right;
        case '>>>':
            return left >>> right;
        case '<<':
            return left << right;
        case '==':
            return left === right;
        case '!=':
            return left !== right;
        case '===':
            return left === right;
        case '!==':
            return left !== right;
        case '>':
            return left > right;
        case '>=':
            return left >= right;
        case '<':
            return left < right;
        case '<=':
            return left <= right;
    }
};

const _IOSTTemplateTag = function(strings, ...keys) {
    _IOSTInstruction_counter.incr(4);
    let res = new String("");
    for (let i = 0; i < strings.length - 1; i++) {
        _IOSTInstruction_counter.incr(8);
        res = res.concat(strings[i], keys[i]);
    }
    res = res.concat(strings[strings.length - 1]);
    return res.toString();
};

const _IOSTSpreadElement = function (args) {
    if (args !== undefined && args !== null && args.length > 0) {
        _IOSTInstruction_counter.incr(args.length);
    }
    return args;
}
const console = new Console;
