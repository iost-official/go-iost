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

const module = new NativeModule();
const require = makeRequireFunction(module);
const exports = {};