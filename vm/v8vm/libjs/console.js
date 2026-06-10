'use strict';

class Console {
    constructor() {
    }
    _format(...args) {
        let formatStr = format(...args);
        return formatStr;
    }
}

(function(){
    let _log = _cLog;
    let P = Console.prototype;
    P.debug = function(...args) {
        _log('Debug', this._format(...args));
    }
    P.info = function(...args) {
        _log('Info', this._format(...args));
    }
    P.warn = function(...args) {
        _log('Warn', this._format(...args));
    }
    P.error = function(...args) {
        _log('Error', this._format(...args));
    }
    P.log = function(...args) {
        this.info(...args)
    }
})();