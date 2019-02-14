'use strict';

function Console() {
    if (!(this instanceof Console)) {
        return new Console();
    }
}

(function(){
    const inner_cLog = _cLog;
    const inner_format = format;

    Console.prototype.debug = function(...args) {
        inner_cLog('Debug', inner_format(...args));
    }

    Console.prototype.info = function(...args) {
        inner_cLog('Info', inner_format(...args));
    }

    Console.prototype.warn = function(...args) {
        inner_cLog('Warn', inner_format(...args));
    }

    Console.prototype.error = function(...args) {
        inner_cLog('Error', inner_format(...args));
    }

    Console.prototype.log = function(...args) {
        this.info(...args)
    }
})();