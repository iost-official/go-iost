'use strict'

function Module(id, parent) {
    this.id = id;
    this.exports = {};
    this.parent = parent;

    this.filename = null;
    this.loaded = false;
    this.exited = false;
    this.children = [];
}
module.exports = Module;

Module._cache = {};

Module.runMain = function() {
    // Load the main module--the command line argument.
    Module._load('_native_main', null, true);
};

Module._load = function(request, parent, isMain) {
    var filename = request;

    var module = new Module(filename, parent);

    if (isMain) {
        module.id = '.'
    }
    Module._cache[filename] = module;
    try {
        module.load(filename);
    } catch(err) {
        delete Module._cache[filename];
        throw err;
    }
};

Module.prototype.load = function(filename) {
    this.filename = filename;

    var extension = '.js';
    Module._extensions[extension](this, filename);
    this.loaded = true;
};

Module.prototype.require = function(path) {
    return Module._load(path, this);
};

Module.prototype._compile = function(content, filename) {
    var self = this;

    function require(path) {
        return self.require(path);
    }
    require.path = Module._cache;
};

Module._extensions['.js'] = function(module, filename) {
    var content = _native_readFile(filename);
    module._compile(content, filename);
};