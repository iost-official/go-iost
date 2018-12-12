'use strict';

function NativeModule(id) {
    this.filename = id + '.js';
    this.id = id;
    this.exports = {};
    this.loaded = false;
}

NativeModule._cache = {};

NativeModule.require = function (id) {
    if (id === '_native_module') {
        return NativeModule;
    }

    const cached = NativeModule.getCached(id);
    if (cached) {
        return cached.exports;
    }

    const nativeModule = new NativeModule(id);
    nativeModule.compile();
    nativeModule.cache();

    return nativeModule.exports;
};

NativeModule.getCached = function(id) {
    return NativeModule._cache[id];
};

NativeModule.getSource = function(id) {
    return _native_require(id);
};

NativeModule.wrap = function(script) {
    return NativeModule.wrapper[0] + script + NativeModule.wrapper[1];
};

NativeModule.wrapper = [
    '(function (exports, require, module, __filename, __dirname) {\n',
    '\n});'
];

NativeModule.prototype.compile = function () {
    let source = NativeModule.getSource(this.id);
    source = NativeModule.wrap(source);

    const fn = _native_run(source, this.filename);
    fn(this.exports, NativeModule.require, this, this.filename);

    this.loaded = true;
};

NativeModule.prototype.cache = function() {
    NativeModule._cache[this.id] = this;
};

const require = NativeModule.require;

// storage
const storage = require('storage');

// blockchain
const blockchain = require('blockchain');

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
        (op === "+" || op === "==" || op === "!=" || op === "===" || op === "!==" || op === "<" || op === "<=" || op === ">" || op === ">=")) {
        _IOSTInstruction_counter.incr(left === null || left === undefined || left.toString().length <= 0 ? 0 : left.toString().length);
        _IOSTInstruction_counter.incr(right === null || right === undefined || right.toString().length <= 0 ? 0 : right.toString().length);
    }
    _IOSTInstruction_counter.incr(3);
    switch (op) {
        case '+':
            return left + right;
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
    _IOSTInstruction_counter.incr(8);
    let res = new String("");
    for (let i = 0; i < strings.length - 1; i++) {
        _IOSTInstruction_counter.incr(23);
        res = res.concat(strings[i], keys[i]);
    }
    _IOSTInstruction_counter.incr(26);
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
