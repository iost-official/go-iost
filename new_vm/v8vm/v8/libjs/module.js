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

NativeModule._cache = Object.create(null);

NativeModule.wrap = function (script) {
    return NativeModule.wrapper[0] + script + NativeModule.wrapper[1]
};

NativeModule.wrapper = [
    '(function (exports, require, module, __filename, __dirname) { ',
    '\n});'
];

NativeModule.prototype.require = function(id) {
    if (typeof id !== 'string') {
        throw 'require id not string'
    }
    if (id === '') {
        throw 'require id is empty'
    }
    return NativeModule._load(id, this)
}

NativeModule._load = function(request, parent) {
    var nativeObj = _native_require(request);
    // return nativeObj;
}

NativeModule.prototype._compile = function (content, filename) {
    var wrapper = NativeModule.wrap(content);

    var require = function (path) {

    }
}

// module.exports = {
//     NativeModule,
//     makeRequireFunction,
// };