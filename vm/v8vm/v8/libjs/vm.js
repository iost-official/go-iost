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
const BlockChain = require('blockchain');

// other helper functions
// var BigNumber = require('bignumber');
// var Int64 = require('int64');

// var injectGas = require('inject_gas');

const _IOSTInstruction_counter = new IOSTInstruction;

// + - * / % **, | & ^ >> >>> <<, || &&, == != === !== > >= < <=, instanceOf in
const _IOSTBinaryOp = function(left, right, op) {
    if ((typeof left === "string" || typeof right === "string") &&
        (op === "+" || op === "==" || op === "!=" || op === "===" || op === "!==" || op === "<" || op === "<=" || op === ">" || op === ">=")) {
        _IOSTInstruction_counter.incr(left === null || left === undefined ? 0 : left.toString().length);
        _IOSTInstruction_counter.incr(right === null || right === undefined ? 0 : right.toString().length);
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

const console = new Console;
