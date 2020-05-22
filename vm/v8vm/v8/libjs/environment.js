'use strict';

(function () {
    // String
    const stringAllowedMethods = [
        'charAt',
        'charCodeAt',
        'length',
        'constructor',
        'toString',
        'valueOf',
        'concat',
        'includes',
        'endsWith',
        'indexOf',
        'lastIndexOf',
        'replace',
        'search',
        'split',
        'startsWith',
        'slice',
        'substring',
        'toLowerCase',
        'toUpperCase',
        'trim',
        'trimLeft',
        'trimRight',
        'repeat'
    ];

    const stringMethods = Object.getOwnPropertyNames(String.prototype);
    stringMethods.forEach((method) => {
        if (!stringAllowedMethods.includes(method)) {
            String.prototype[method] = null;
        }
    });

    const StringcharAt = String.prototype.charAt;
    String.prototype.charAt = function() {
        _IOSTInstruction_counter.incr(1);
        return StringcharAt.call(this, ...arguments);
    };

    const StringcharCodeAt = String.prototype.charCodeAt;
    String.prototype.charCodeAt = function() {
        _IOSTInstruction_counter.incr(1);
        return StringcharCodeAt.call(this, ...arguments);
    };

    const Stringconstructor = String.prototype.constructor;
    String.prototype.constructor = function() {
        _IOSTInstruction_counter.incr(1 + this.length * 0.1);
        return Stringconstructor.call(this, ...arguments);
    };

    const StringvalueOf = String.prototype.valueOf;
    String.prototype.valueOf = function() {
        _IOSTInstruction_counter.incr(1);
        return StringvalueOf.call(this, ...arguments);
    };

    const Stringconcat = String.prototype.concat;
    String.prototype.concat = function() {
        let res = Stringconcat.call(this, ...arguments);
        _IOSTInstruction_counter.incr(1 + res.length * 0.1);
        return res;
    };

    const Stringincludes = String.prototype.includes;
    String.prototype.includes = function() {
        let s = Stringconstructor(this);
        _IOSTInstruction_counter.incr(1 + s.length * 0.1);
        return Stringincludes.call(s, ...arguments);
    };

    const StringendsWith = String.prototype.endsWith;
    String.prototype.endsWith = function() {
        let s = Stringconstructor(this);
        _IOSTInstruction_counter.incr(1 + s.length * 0.1);
        return StringendsWith.call(s, ...arguments);
    };

    const StringindexOf = String.prototype.indexOf;
    String.prototype.indexOf = function() {
        let s = Stringconstructor(this);
        _IOSTInstruction_counter.incr(1 + s.length * 0.1);
        return StringindexOf.call(s, ...arguments);
    };

    const StringlastIndexOf = String.prototype.lastIndexOf;
    String.prototype.lastIndexOf = function() {
        let s = Stringconstructor(this);
        _IOSTInstruction_counter.incr(1 + s.length * 0.1);
        return StringlastIndexOf.call(s, ...arguments);
    };

    const Stringreplace = String.prototype.replace;
    String.prototype.replace = function() {
        let s = Stringconstructor(this);
        _IOSTInstruction_counter.incr(1 + s.length * 0.1);
        return Stringreplace.call(s, ...arguments);
    };

    const Stringsearch = String.prototype.search;
    String.prototype.search = function() {
        let s = Stringconstructor(this);
        _IOSTInstruction_counter.incr(1 + s.length * 0.1);
        return Stringsearch.call(s, ...arguments);
    };

    const Stringsplit = String.prototype.split;
    String.prototype.split = function() {
        let s = Stringconstructor(this);
        _IOSTInstruction_counter.incr(1 + s.length * 0.1);
        return Stringsplit.call(s, ...arguments);
    };

    const StringstartsWith = String.prototype.startsWith;
    String.prototype.startsWith = function() {
        let s = Stringconstructor(this);
        _IOSTInstruction_counter.incr(1 + s.length * 0.1);
        return StringstartsWith.call(s, ...arguments);
    };

    const Stringslice = String.prototype.slice;
    String.prototype.slice = function() {
        let s = Stringconstructor(this);
        _IOSTInstruction_counter.incr(1 + s.length * 0.1);
        return Stringslice.call(s, ...arguments);
    };

    const Stringsubstring = String.prototype.substring;
    String.prototype.substring = function() {
        let s = Stringconstructor(this);
        _IOSTInstruction_counter.incr(1 + s.length * 0.1);
        return Stringsubstring.call(s, ...arguments);
    };

    const StringtoLowerCase = String.prototype.toLowerCase;
    String.prototype.toLowerCase = function() {
        let s = Stringconstructor(this);
        _IOSTInstruction_counter.incr(1 + s.length * 0.1);
        return StringtoLowerCase.call(s, ...arguments);
    };

    const StringtoUpperCase = String.prototype.toUpperCase;
    String.prototype.toUpperCase = function() {
        let s = Stringconstructor(this);
        _IOSTInstruction_counter.incr(1 + s.length * 0.1);
        return StringtoUpperCase.call(s, ...arguments);
    };

    const Stringtrim = String.prototype.trim;
    String.prototype.trim = function() {
        let s = Stringconstructor(this);
        _IOSTInstruction_counter.incr(1 + s.length * 0.1);
        return Stringtrim.call(s, ...arguments);
    };

    const StringtrimLeft = String.prototype.trimLeft;
    String.prototype.trimLeft = function() {
        let s = Stringconstructor(this);
        _IOSTInstruction_counter.incr(1 + s.length * 0.1);
        return StringtrimLeft.call(s, ...arguments);
    };

    const StringtrimRight = String.prototype.trimRight;
    String.prototype.trimRight = function() {
        let s = Stringconstructor(this);
        _IOSTInstruction_counter.incr(1 + s.length * 0.1);
        return StringtrimRight.call(s, ...arguments);
    };

    const Stringrepeat = String.prototype.repeat;
    String.prototype.repeat = function() {
        let res = Stringrepeat.call(this, ...arguments);
        _IOSTInstruction_counter.incr(1 + res.length * 0.1);
        return res;
    };

    // Math
    const mathAllowedMethods = [
        'abs',
        'cbrt',
        'ceil',
        'floor',
        'log',
        'log10',
        'log1p',
        'max',
        'min',
        'pow',
        'round',
        'sqrt',
    ];
    const mathIgnoreMethods = [
        'E',
        'LN10',
        'LN2',
        'LOG10E',
        'LOG2E',
        'PI',
        'SQRT1_2',
        'SQRT2'
    ];

    const mathMethods = Object.getOwnPropertyNames(Math);
    mathMethods.forEach((method) => {
        if (!mathAllowedMethods.includes(method) && !mathIgnoreMethods.includes(method)) {
            Math[method] = null;
        }
    });

    const Mathabs = Math.abs;
    Math.abs = function() {
        _IOSTInstruction_counter.incr(2);
        return Mathabs.call(this, ...arguments);
    };

    const Mathcbrt = Math.cbrt;
    Math.cbrt = function() {
        _IOSTInstruction_counter.incr(2);
        return Mathcbrt.call(this, ...arguments);
    };

    const Mathceil = Math.ceil;
    Math.ceil = function() {
        _IOSTInstruction_counter.incr(2);
        return Mathceil.call(this, ...arguments);
    };

    const Mathfloor = Math.floor;
    Math.floor = function() {
        _IOSTInstruction_counter.incr(2);
        return Mathfloor.call(this, ...arguments);
    };

    const Mathlog = Math.log;
    Math.log = function() {
        _IOSTInstruction_counter.incr(2);
        return Mathlog.call(this, ...arguments);
    };

    const Mathlog10 = Math.log10;
    Math.log10 = function() {
        _IOSTInstruction_counter.incr(2);
        return Mathlog10.call(this, ...arguments);
    };

    const Mathlog1p = Math.log1p;
    Math.log1p = function() {
        _IOSTInstruction_counter.incr(2);
        return Mathlog1p.call(this, ...arguments);
    };

    const Mathmax = Math.max;
    Math.max = function() {
        _IOSTInstruction_counter.incr(2);
        return Mathmax.call(this, ...arguments);
    };

    const Mathmin = Math.min;
    Math.min = function() {
        _IOSTInstruction_counter.incr(2);
        return Mathmin.call(this, ...arguments);
    };

    const Mathpow = Math.pow;
    Math.pow = function() {
        _IOSTInstruction_counter.incr(2);
        return Mathpow.call(this, ...arguments);
    };

    const Mathround = Math.round;
    Math.round = function() {
        _IOSTInstruction_counter.incr(2);
        return Mathround.call(this, ...arguments);
    };

    const Mathsqrt = Math.sqrt;
    Math.sqrt = function() {
        _IOSTInstruction_counter.incr(2);
        return Mathsqrt.call(this, ...arguments);
    };

    // Array
    const arrayAllowedMethods = [
        'constructor',
        'toString',
        'concat',
        'every',
        'filter',
        'find',
        'findIndex',
        'forEach',
        'includes',
        'indexOf',
        'join',
        'keys',
        'lastIndexOf',
        'map',
        'pop',
        'push',
        'reverse',
        'shift',
        'slice',
        'sort',
        'splice',
        'unshift'
    ];
    const arrayMethods = Object.getOwnPropertyNames(Array.prototype);
    arrayMethods.forEach((method) => {
        if (!arrayAllowedMethods.includes(method)) {
            Array.prototype[method] = null;
        }
    });

    const Arrayconstructor = Array.prototype.constructor;
    Array.prototype.constructor = function() {
        let res = Arrayconstructor.call(this, ...arguments);
        _IOSTInstruction_counter.incr(1 + res.length * 0.2);
        return res;
    };

    const ArraytoString = Array.prototype.toString;
    Array.prototype.toString = function() {
        _IOSTInstruction_counter.incr(1 + this.length);
        return ArraytoString.call(this, ...arguments);
    };

    const Arrayconcat = Array.prototype.concat;
    Array.prototype.concat = function() {
        let res = Arrayconcat.call(this, ...arguments);
        _IOSTInstruction_counter.incr(1 + res.length);
        return res;
    };

    const Arrayevery = Array.prototype.every;
    Array.prototype.every = function() {
        _IOSTInstruction_counter.incr(1 + this.length);
        return Arrayevery.call(this, ...arguments);
    };

    const Arrayfilter = Array.prototype.filter;
    Array.prototype.filter = function() {
        _IOSTInstruction_counter.incr(1 + this.length * 10);
        return Arrayfilter.call(this, ...arguments);
    };

    const Arrayfind = Array.prototype.find;
    Array.prototype.find = function() {
        _IOSTInstruction_counter.incr(1 + this.length);
        return Arrayfind.call(this, ...arguments);
    };

    const ArrayfindIndex = Array.prototype.findIndex;
    Array.prototype.findIndex = function() {
        _IOSTInstruction_counter.incr(1 + this.length);
        return ArrayfindIndex.call(this, ...arguments);
    };

    const ArrayforEach = Array.prototype.forEach;
    Array.prototype.forEach = function() {
        _IOSTInstruction_counter.incr(1 + this.length);
        return ArrayforEach.call(this, ...arguments);
    };

    const Arrayincludes = Array.prototype.includes;
    Array.prototype.includes = function() {
        _IOSTInstruction_counter.incr(1 + this.length);
        return Arrayincludes.call(this, ...arguments);
    };

    const ArrayindexOf = Array.prototype.indexOf;
    Array.prototype.indexOf = function() {
        _IOSTInstruction_counter.incr(1 + this.length * 0.2);
        return ArrayindexOf.call(this, ...arguments);
    };

    const Arrayjoin = Array.prototype.join;
    Array.prototype.join = function() {
        _IOSTInstruction_counter.incr(1 + this.length);
        return Arrayjoin.call(this, ...arguments);
    };

    const Arraykeys = Array.prototype.keys;
    Array.prototype.keys = function() {
        _IOSTInstruction_counter.incr(1 + this.length * 0.2);
        return Arraykeys.call(this, ...arguments);
    };

    const ArraylastIndexOf = Array.prototype.lastIndexOf;
    Array.prototype.lastIndexOf = function() {
        _IOSTInstruction_counter.incr(1 + this.length * 0.2);
        return ArraylastIndexOf.call(this, ...arguments);
    };

    const Arraymap = Array.prototype.map;
    Array.prototype.map = function() {
        _IOSTInstruction_counter.incr(1 + this.length * 10);
        return Arraymap.call(this, ...arguments);
    };

    const Arraypop = Array.prototype.pop;
    Array.prototype.pop = function() {
        _IOSTInstruction_counter.incr(1);
        return Arraypop.call(this, ...arguments);
    };

    const Arraypush = Array.prototype.push;
    Array.prototype.push = function() {
        _IOSTInstruction_counter.incr(1);
        return Arraypush.call(this, ...arguments);
    };

    const Arrayreverse = Array.prototype.reverse;
    Array.prototype.reverse = function() {
        _IOSTInstruction_counter.incr(1 + this.length * 0.2);
        return Arrayreverse.call(this, ...arguments);
    };

    const Arrayshift = Array.prototype.shift;
    Array.prototype.shift = function() {
        _IOSTInstruction_counter.incr(1);
        return Arrayshift.call(this, ...arguments);
    };

    const Arrayslice = Array.prototype.slice;
    Array.prototype.slice = function() {
        _IOSTInstruction_counter.incr(1 + this.length * 10);
        return Arrayslice.call(this, ...arguments);
    };

    const Arraysort = Array.prototype.sort;
    Array.prototype.sort = function() {
        _IOSTInstruction_counter.incr(1 + this.length);
        return Arraysort.call(this, ...arguments);
    };

    const Arraysplice = Array.prototype.splice;
    Array.prototype.splice = function() {
        _IOSTInstruction_counter.incr(1 + this.length * 10);
        if (arguments.length >= 3) {
            _IOSTInstruction_counter.incr(arguments.length - 2);
        }
        return Arraysplice.call(this, ...arguments);
    };

    const Arrayunshift = Array.prototype.unshift;
    Array.prototype.unshift = function() {
        _IOSTInstruction_counter.incr(1);
        return Arrayunshift.call(this, ...arguments);
    };

    Array.from = null;
    Array.of = null;
    const OrigArray = Array;
    Array = OrigArray.prototype.constructor;
    Array.prototype = OrigArray.prototype;
    Array.isArray = OrigArray.isArray;


    // JSON
    const JSONparse = JSON.parse;
    JSON.parse = function () {
        if (arguments[0] === undefined || arguments[0] === null) {
            _IOSTInstruction_counter.incr(10);
        } else {
            _IOSTInstruction_counter.incr(10 + Stringconstructor(arguments[0]).length);
        }
        return JSONparse.call(this, ...arguments);
    };

    /*
    const JSONstringify = JSON.stringify;
    JSON.stringify = function () {
        const rs = JSONstringify.call(this, ...arguments);
        _IOSTInstruction_counter.incr(10 + rs.length);
        return rs;
    };
    */

    // Functions
    parseFloat = null;
    parseInt = null;
    decodeURI = null;
    decodeURIComponent = null;
    encodeURI = null;
    encodeURIComponent = null;
    escape = null;
    unescape = null;

    // Function
    const Functionconstructor = Function;
    Function.prototype.toString = function(){
        if (!(this instanceof Functionconstructor)) {
            throw new Error;
        }
        return `function ${this.name}() { [native code] }`;
    };
    Function = null;

    // Error
    Object.defineProperty(Error, "stackTraceLimit", {
        enumerable: false,
        configurable: false,
        writable: false,
        value: 0,
    });

    // Fundamental Objects
    Boolean = null;
    EvalError = null;
    RangeError = null;
    ReferenceError = null;
    SyntaxError = null;
    TypeError = null;
    URIError = null;

    // Numbers and dates
    Date = null;
    if (typeof Intl !== 'undefined') {
        Intl = null;
    }

    // Text processing
    RegExp = null;

    // Indexed collections
    Int8Array = null;
    Uint8Array = null;
    Uint8ClampedArray = null;
    Int16Array = null;
    Uint16Array = null;
    Int32Array = null;
    Uint32Array = null;
    Float32Array = null;
    Float64Array = null;

    // Keyed collections
    Map = null;
    Set = null;
    WeakMap = null;
    WeakSet = null;

    // Structured data
    ArrayBuffer = null;
    SharedArrayBuffer = null;
    Atomics = null;
    DataView = null;

    // Control abstraction objects
    Promise = null;

    // ReflectionSection
    Reflect = null;
    Proxy = null;

    // WebAssembly
    WebAssembly = null;

    // Native
    IOSTBlockchain = null;
    IOSTInstruction = null;
    IOSTStorage = null;
    _IOSTCrypto = null;
    _native_log = null;
    _native_run = null;
    _native_require = null;
    _cLog = null;

    // BigNumber
    const OrigBigNumber = BigNumber;
    const BigNumbertoString = BigNumber.prototype.toString;
    const BigNumberconstructor = BigNumber.prototype.constructor;
    BigNumber.prototype.constructor = function() {
        _IOSTInstruction_counter.incr(20);
        return BigNumberconstructor.call(this, ...arguments);
    };

    const BigNumberabs = BigNumber.prototype.abs;
    BigNumber.prototype.absoluteValue = BigNumber.prototype.abs = function() {
        _IOSTInstruction_counter.incr(20);
        return BigNumberabs.call(this, ...arguments);
    };

    const BigNumberdiv = BigNumber.prototype.div;
    BigNumber.prototype.dividedBy = BigNumber.prototype.div = function() {
        if (arguments[0] == null) {
            _IOSTInstruction_counter.incr(20 + BigNumbertoString.call(this).length * 10);
        } else {
            let argStr = (arguments[0] instanceof OrigBigNumber) ? BigNumbertoString.call(arguments[0]) : Stringconstructor(arguments[0]);
            _IOSTInstruction_counter.incr(20 + (BigNumbertoString.call(this).length + argStr.length) * 10);
        }
        return BigNumberdiv.call(this, ...arguments);
    };

    const BigNumberidiv = BigNumber.prototype.idiv;
    BigNumber.prototype.dividedToIntegerBy = BigNumber.prototype.idiv = function() {
        if (arguments[0] == null) {
            _IOSTInstruction_counter.incr(20 + BigNumbertoString.call(this).length * 10);
        } else {
            let argStr = (arguments[0] instanceof OrigBigNumber) ? BigNumbertoString.call(arguments[0]) : Stringconstructor(arguments[0]);
            _IOSTInstruction_counter.incr(20 + (BigNumbertoString.call(this).length + argStr.length) * 10);
        }
        return BigNumberidiv.call(this, ...arguments);
    };

    const BigNumberpow = BigNumber.prototype.pow;
    BigNumber.prototype.exponentiatedBy = BigNumber.prototype.pow = function () {
        if (arguments[0] == null) {
            _IOSTInstruction_counter.incr(20 + BigNumbertoString.call(this).length * 10);
        } else {
            let argStr = (arguments[0] instanceof OrigBigNumber) ? BigNumbertoString.call(arguments[0]) : Stringconstructor(arguments[0]);
            _IOSTInstruction_counter.incr(20 + BigNumbertoString.call(this).length * argStr.length * 10);
        }
        return BigNumberpow.call(this, ...arguments);
    };

    const BigNumberintegerValue = BigNumber.prototype.integerValue;
    BigNumber.prototype.integerValue = function() {
        _IOSTInstruction_counter.incr(20 + BigNumbertoString.call(this).length * 4);
        return BigNumberintegerValue.call(this, ...arguments);
    };

    const BigNumbereq = BigNumber.prototype.eq;
    BigNumber.prototype.isEqualTo = BigNumber.prototype.eq = function() {
        if (arguments[0] == null) {
            _IOSTInstruction_counter.incr(20 + BigNumbertoString.call(this).length * 2);
        } else {
            let argStr = (arguments[0] instanceof OrigBigNumber) ? BigNumbertoString.call(arguments[0]) : Stringconstructor(arguments[0]);
            _IOSTInstruction_counter.incr(20 + (BigNumbertoString.call(this).length + argStr.length) * 2);
        }
        return BigNumbereq.call(this, ...arguments);
    };

    const BigNumberisFinite = BigNumber.prototype.isFinite;
    BigNumber.prototype.isFinite = function() {
        _IOSTInstruction_counter.incr(20);
        return BigNumberisFinite.call(this, ...arguments);
    };

    const BigNumbergt = BigNumber.prototype.gt;
    BigNumber.prototype.isGreaterThan = BigNumber.prototype.gt = function() {
        if (arguments[0] == null) {
            _IOSTInstruction_counter.incr(20 + BigNumbertoString.call(this).length * 2);
        } else {
            let argStr = (arguments[0] instanceof OrigBigNumber) ? BigNumbertoString.call(arguments[0]) : Stringconstructor(arguments[0]);
            _IOSTInstruction_counter.incr(20 + (BigNumbertoString.call(this).length + argStr.length) * 2);
        }
        return BigNumbergt.call(this, ...arguments);
    };

    const BigNumbergte = BigNumber.prototype.gte;
    BigNumber.prototype.isGreaterThanOrEqualTo = BigNumber.prototype.gte = function() {
        if (arguments[0] == null) {
            _IOSTInstruction_counter.incr(20 + BigNumbertoString.call(this).length * 2);
        } else {
            let argStr = (arguments[0] instanceof OrigBigNumber) ? BigNumbertoString.call(arguments[0]) : Stringconstructor(arguments[0]);
            _IOSTInstruction_counter.incr(20 + (BigNumbertoString.call(this).length + argStr.length) * 2);
        }
        return BigNumbergte.call(this, ...arguments);
    };

    const BigNumberisInteger = BigNumber.prototype.isInteger;
    BigNumber.prototype.isInteger = function() {
        _IOSTInstruction_counter.incr(20);
        return BigNumberisInteger.call(this, ...arguments);
    };

    const BigNumberlt = BigNumber.prototype.lt;
    BigNumber.prototype.isLessThan = BigNumber.prototype.lt = function() {
        if (arguments[0] == null) {
            _IOSTInstruction_counter.incr(20 + BigNumbertoString.call(this).length * 2);
        } else {
            let argStr = (arguments[0] instanceof OrigBigNumber) ? BigNumbertoString.call(arguments[0]) : Stringconstructor(arguments[0]);
            _IOSTInstruction_counter.incr(20 + (BigNumbertoString.call(this).length + argStr.length) * 2);
        }
        return BigNumberlt.call(this, ...arguments);
    };

    const BigNumberlte = BigNumber.prototype.lte;
    BigNumber.prototype.isLessThanOrEqualTo = BigNumber.prototype.lte = function() {
        if (arguments[0] == null) {
            _IOSTInstruction_counter.incr(20 + BigNumbertoString.call(this).length * 2);
        } else {
            let argStr = (arguments[0] instanceof OrigBigNumber) ? BigNumbertoString.call(arguments[0]) : Stringconstructor(arguments[0]);
            _IOSTInstruction_counter.incr(20 + (BigNumbertoString.call(this).length + argStr.length) * 2);
        }
        return BigNumberlte.call(this, ...arguments);
    };

    const BigNumberisNaN = BigNumber.prototype.isNaN;
    BigNumber.prototype.isNaN = function() {
        _IOSTInstruction_counter.incr(20);
        return BigNumberisNaN.call(this, ...arguments);
    };

    const BigNumberisNegative = BigNumber.prototype.isNegative;
    BigNumber.prototype.isNegative = function() {
        _IOSTInstruction_counter.incr(20);
        return BigNumberisNegative.call(this, ...arguments);
    };

    const BigNumberisPositive = BigNumber.prototype.isPositive;
    BigNumber.prototype.isPositive = function() {
        _IOSTInstruction_counter.incr(20);
        return BigNumberisPositive.call(this, ...arguments);
    };

    const BigNumberisZero = BigNumber.prototype.isZero;
    BigNumber.prototype.isZero = function() {
        _IOSTInstruction_counter.incr(20);
        return BigNumberisZero.call(this, ...arguments);
    };

    const BigNumberminus = BigNumber.prototype.minus;
    BigNumber.prototype.minus = function() {
        if (arguments[0] == null) {
            _IOSTInstruction_counter.incr(20 + BigNumbertoString.call(this).length * 10);
        } else {
            let argStr = (arguments[0] instanceof OrigBigNumber) ? BigNumbertoString.call(arguments[0]) : Stringconstructor(arguments[0]);
            _IOSTInstruction_counter.incr(20 + (BigNumbertoString.call(this).length + argStr.length) * 10);
        }
        return BigNumberminus.call(this, ...arguments);
    };

    const BigNumbermod = BigNumber.prototype.mod;
    BigNumber.prototype.mod = function() {
        if (arguments[0] == null) {
            _IOSTInstruction_counter.incr(20 + BigNumbertoString.call(this).length * 10);
        } else {
            let argStr = (arguments[0] instanceof OrigBigNumber) ? BigNumbertoString.call(arguments[0]) : Stringconstructor(arguments[0]);
            _IOSTInstruction_counter.incr(20 + (BigNumbertoString.call(this).length + argStr.length) * 10);
        }
        return BigNumbermod.call(this, ...arguments);
    };

    const BigNumbertimes = BigNumber.prototype.times;
    BigNumber.prototype.multipliedBy = BigNumber.prototype.times = function() {
        if (arguments[0] == null) {
            _IOSTInstruction_counter.incr(20 + BigNumbertoString.call(this).length * 4);
        } else {
            let argStr = (arguments[0] instanceof OrigBigNumber) ? BigNumbertoString.call(arguments[0]) : Stringconstructor(arguments[0]);
            _IOSTInstruction_counter.incr(20 + (BigNumbertoString.call(this).length + argStr.length) * 4);
        }
        return BigNumbertimes.call(this, ...arguments);
    };

    const BigNumbernegated = BigNumber.prototype.negated;
    BigNumber.prototype.negated = function() {
        _IOSTInstruction_counter.incr(20);
        return BigNumbernegated.call(this, ...arguments);
    };

    const BigNumberplus = BigNumber.prototype.plus;
    BigNumber.prototype.plus = function() {
        if (arguments[0] == null) {
            _IOSTInstruction_counter.incr(20 + BigNumbertoString.call(this).length * 10);
        } else {
            let argStr = (arguments[0] instanceof OrigBigNumber) ? BigNumbertoString.call(arguments[0]) : Stringconstructor(arguments[0]);
            _IOSTInstruction_counter.incr(20 + (BigNumbertoString.call(this).length + argStr.length) * 10);
        }
        return BigNumberplus.call(this, ...arguments);
    };

    const BigNumbersqrt = BigNumber.prototype.sqrt;
    BigNumber.prototype.squareRoot = BigNumber.prototype.sqrt = function() {
        _IOSTInstruction_counter.incr(20 + BigNumbertoString.call(this).length * 4);
        return BigNumbersqrt.call(this, ...arguments);
    };

    const BigNumbertoFixed = BigNumber.prototype.toFixed;
    BigNumber.prototype.toFixed = function() {
        _IOSTInstruction_counter.incr(20 + BigNumbertoString.call(this).length * 4);
        return BigNumbertoFixed.call(this, ...arguments);
    };

    BigNumber.config({
        DECIMAL_PLACES: 50,
        POW_PRECISION: 50,
        ROUNDING_MODE: BigNumber.ROUND_DOWN
    });
    BigNumber = OrigBigNumber.prototype.constructor;
    BigNumber.prototype = OrigBigNumber.prototype;
    BigNumber.isBigNumber = OrigBigNumber.isBigNumber;
    BigNumber.maximum = BigNumber.max = OrigBigNumber.maximum;
    BigNumber.minimum = BigNumber.min = OrigBigNumber.minimum;
})();
