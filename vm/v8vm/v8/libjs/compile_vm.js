'use strict';

const require = (function(){
    const inner_native_require = _native_require;
    const inner_native_run = _native_run;

    function NativeModule(id) {
        this.filename = id + '.js';
        this.id = id;
        this.exports = {};
        this.loaded = false;
    }

    NativeModule._cache = {};

    NativeModule.require = function (id) {
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
        return inner_native_require(id);
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

        let fn = inner_native_run(source, this.filename);
        fn(this.exports, NativeModule.require, this, this.filename);

        this.loaded = true;
    };

    NativeModule.prototype.cache = function() {
        NativeModule._cache[this.id] = this;
    };

    return NativeModule.require;
})();

let injectGas = require('inject_gas');

let _IOSTInstruction_counter = new IOSTInstruction;
