'use strict'

function makeRequireFunction(mod) {
    const NativeModule = mod.constructor;

    function require(path) {
        return mod.require(path);
    }

    require.cache = NativeModule._cache;

    return require;
}

function NativeModule(id, parent) {
    this.id = id;
    this.exports = {};
    this.parent = parent;
}

NativeModule._load = function(request, parent) {
    _native_require(request);
    return module.exports
}

NativeModule.prototype.require = function (id) {
    if (typeof id !== 'string') {
        throw "id not string";
    }
    if (id === '') {
        throw "id empty string";
    }
    return NativeModule._load(id, this);
}

const module = new NativeModule();
const require = makeRequireFunction(module);
const exports = {};