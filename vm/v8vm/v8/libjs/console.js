'use strict';

class Console {
    _format(...args) {
        let formatStr = format(...args);
        return formatStr;
    }

    debug(...args) {
        _cLog('Debug', this._format(...args));
    }

    info(...args) {
        _cLog('Info', this._format(...args));
    }

    warn(...args) {
        _cLog('Warn', this._format(...args));
    }

    error(...args) {
        _cLog('Error', this._format(...args));
    }

    fatal(...args) {
        _cLog('Fatal', this._format(...args));
    }

    log(...args) {
        this.info(...args)
    }
}