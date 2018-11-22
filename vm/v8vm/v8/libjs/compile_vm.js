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

    let cached = NativeModule.getCached(id);
    if (cached) {
        return cached.exports;
    }

    let nativeModule = new NativeModule(id);
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

    let fn = _native_run(source, this.filename);
    fn(this.exports, NativeModule.require, this, this.filename);

    this.loaded = true;
};

NativeModule.prototype.cache = function() {
    NativeModule._cache[this.id] = this;
};

let require = NativeModule.require;

let injectGas = require('inject_gas');

let _IOSTInstruction_counter = new IOSTInstruction;
