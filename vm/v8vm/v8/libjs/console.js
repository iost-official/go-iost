'use strict';

class Console {
    constructor() {
        this._cLog = _cLog;
    }
    _format(...args) {
        let formatStr = format(...args);
        return formatStr;
    }

    debug(...args) {
        this._cLog('Debug', this._format(...args));
    }

    info(...args) {
        this._cLog('Info', this._format(...args));
    }

    warn(...args) {
        this._cLog('Warn', this._format(...args));
    }

    error(...args) {
        this._cLog('Error', this._format(...args));
    }

    log(...args) {
        this.info(...args)
    }
}