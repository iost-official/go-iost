'use strict';

const require = (function(){
    const inner_native_require = _native_require;
    const inner_native_run = _native_run;
    let NativeModule = function(id) {
        this.filename = id + '.js';
        this.id = id;
        this.exports = {};
        this.loaded = false;
    }

    NativeModule._cache = {};

    NativeModule.require = function (id) {
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

        const fn = inner_native_run(source, this.filename);
        fn(this.exports, NativeModule.require, this, this.filename);

        this.loaded = true;
    };

    NativeModule.prototype.cache = function() {
        NativeModule._cache[this.id] = this;
    };

    return NativeModule.require;
})();

// storage
const storage = require('storage');

// blockchain
const blockchain = require('blockchain');
