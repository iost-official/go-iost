'use strict';

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
    _IOSTInstruction_counter.incr(this.toString().length);
    return Stringconstructor.call(this, ...arguments);
};

const StringvalueOf = String.prototype.valueOf;
String.prototype.valueOf = function() {
    _IOSTInstruction_counter.incr(1);
    return StringvalueOf.call(this, ...arguments);
};

const Stringconcat = String.prototype.concat;
String.prototype.concat = function() {
    _IOSTInstruction_counter.incr(this.toString().length);
    return Stringconcat.call(this, ...arguments);
};

const Stringincludes = String.prototype.includes;
String.prototype.includes = function() {
    _IOSTInstruction_counter.incr(this.toString().length);
    return Stringincludes.call(this, ...arguments);
};

const StringendsWith = String.prototype.endsWith;
String.prototype.endsWith = function() {
    _IOSTInstruction_counter.incr(this.toString().length);
    return StringendsWith.call(this, ...arguments);
};

const StringindexOf = String.prototype.indexOf;
String.prototype.indexOf = function() {
    _IOSTInstruction_counter.incr(this.toString().length);
    return StringindexOf.call(this, ...arguments);
};

const StringlastIndexOf = String.prototype.lastIndexOf;
String.prototype.lastIndexOf = function() {
    _IOSTInstruction_counter.incr(this.toString().length);
    return StringlastIndexOf.call(this, ...arguments);
};

const Stringreplace = String.prototype.replace;
String.prototype.replace = function() {
    _IOSTInstruction_counter.incr(this.toString().length);
    return Stringreplace.call(this, ...arguments);
};

const Stringsearch = String.prototype.search;
String.prototype.search = function() {
    _IOSTInstruction_counter.incr(this.toString().length);
    return Stringsearch.call(this, ...arguments);
};

const Stringsplit = String.prototype.split;
String.prototype.split = function() {
    _IOSTInstruction_counter.incr(this.toString().length);
    return Stringsplit.call(this, ...arguments);
};

const StringstartsWith = String.prototype.startsWith;
String.prototype.startsWith = function() {
    _IOSTInstruction_counter.incr(this.toString().length);
    return StringstartsWith.call(this, ...arguments);
};

const Stringslice = String.prototype.slice;
String.prototype.slice = function() {
    _IOSTInstruction_counter.incr(this.toString().length);
    return Stringslice.call(this, ...arguments);
};

const StringtoLowerCase = String.prototype.toLowerCase;
String.prototype.toLowerCase = function() {
    _IOSTInstruction_counter.incr(this.toString().length);
    return StringtoLowerCase.call(this, ...arguments);
};

const StringtoUpperCase = String.prototype.toUpperCase;
String.prototype.toUpperCase = function() {
    _IOSTInstruction_counter.incr(this.toString().length);
    return StringtoUpperCase.call(this, ...arguments);
};

const Stringtrim = String.prototype.trim;
String.prototype.trim = function() {
    _IOSTInstruction_counter.incr(this.toString().length);
    return Stringtrim.call(this, ...arguments);
};

const StringtrimLeft = String.prototype.trimLeft;
String.prototype.trimLeft = function() {
    _IOSTInstruction_counter.incr(this.toString().length);
    return StringtrimLeft.call(this, ...arguments);
};

const StringtrimRight = String.prototype.trimRight;
String.prototype.trimRight = function() {
    _IOSTInstruction_counter.incr(this.toString().length);
    return StringtrimRight.call(this, ...arguments);
};

const Stringrepeat = String.prototype.repeat;
String.prototype.repeat = function() {
    _IOSTInstruction_counter.incr(this.toString().length);
    return Stringrepeat.call(this, ...arguments);
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
    _IOSTInstruction_counter.incr(this.length);
    return Arrayconstructor.call(this, ...arguments);
};

const ArraytoString = Array.prototype.toString;
Array.prototype.toString = function() {
    _IOSTInstruction_counter.incr(this.length);
    return ArraytoString.call(this, ...arguments);
};

const Arrayconcat = Array.prototype.concat;
Array.prototype.concat = function() {
    _IOSTInstruction_counter.incr(this.length);
    return Arrayconcat.call(this, ...arguments);
};

const Arrayevery = Array.prototype.every;
Array.prototype.every = function() {
    _IOSTInstruction_counter.incr(this.length);
    return Arrayevery.call(this, ...arguments);
};

const Arrayfilter = Array.prototype.filter;
Array.prototype.filter = function() {
    _IOSTInstruction_counter.incr(this.length);
    return Arrayfilter.call(this, ...arguments);
};

const Arrayfind = Array.prototype.find;
Array.prototype.find = function() {
    _IOSTInstruction_counter.incr(this.length);
    return Arrayfind.call(this, ...arguments);
};

const ArrayfindIndex = Array.prototype.findIndex;
Array.prototype.findIndex = function() {
    _IOSTInstruction_counter.incr(this.length);
    return ArrayfindIndex.call(this, ...arguments);
};

const ArrayforEach = Array.prototype.forEach;
Array.prototype.forEach = function() {
    _IOSTInstruction_counter.incr(this.length);
    return ArrayforEach.call(this, ...arguments);
};

const Arrayincludes = Array.prototype.includes;
Array.prototype.includes = function() {
    _IOSTInstruction_counter.incr(this.length);
    return Arrayincludes.call(this, ...arguments);
};

const ArrayindexOf = Array.prototype.indexOf;
Array.prototype.indexOf = function() {
    _IOSTInstruction_counter.incr(this.length);
    return ArrayindexOf.call(this, ...arguments);
};

const Arrayjoin = Array.prototype.join;
Array.prototype.join = function() {
    _IOSTInstruction_counter.incr(this.length);
    return Arrayjoin.call(this, ...arguments);
};

const Arraykeys = Array.prototype.keys;
Array.prototype.keys = function() {
    _IOSTInstruction_counter.incr(this.length);
    return Arraykeys.call(this, ...arguments);
};

const ArraylastIndexOf = Array.prototype.lastIndexOf;
Array.prototype.lastIndexOf = function() {
    _IOSTInstruction_counter.incr(this.length);
    return ArraylastIndexOf.call(this, ...arguments);
};

const Arraymap = Array.prototype.map;
Array.prototype.map = function() {
    _IOSTInstruction_counter.incr(this.length);
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
    _IOSTInstruction_counter.incr(this.length);
    return Arrayreverse.call(this, ...arguments);
};

const Arrayshift = Array.prototype.shift;
Array.prototype.shift = function() {
    _IOSTInstruction_counter.incr(1);
    return Arrayshift.call(this, ...arguments);
};

const Arrayslice = Array.prototype.slice;
Array.prototype.slice = function() {
    _IOSTInstruction_counter.incr(this.length);
    return Arrayslice.call(this, ...arguments);
};

const Arraysort = Array.prototype.sort;
Array.prototype.sort = function() {
    _IOSTInstruction_counter.incr(this.length);
    return Arraysort.call(this, ...arguments);
};

const Arraysplice = Array.prototype.splice;
Array.prototype.splice = function() {
    _IOSTInstruction_counter.incr(this.length);
    return Arraysplice.call(this, ...arguments);
};

const Arrayunshift = Array.prototype.unshift;
Array.prototype.unshift = function() {
    _IOSTInstruction_counter.incr(1);
    return Arrayunshift.call(this, ...arguments);
};

// Math
const mathAllowedMethods = [
    'abs',
    'cbrt',
    'ceil',
    'floor',
    'log',
    'log10',
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
    _IOSTInstruction_counter.incr(5);
    return Mathabs.call(this, ...arguments);
};

const Mathcbrt = Math.cbrt;
Math.cbrt = function() {
    _IOSTInstruction_counter.incr(40);
    return Mathcbrt.call(this, ...arguments);
};

const Mathceil = Math.ceil;
Math.ceil = function() {
    _IOSTInstruction_counter.incr(5);
    return Mathceil.call(this, ...arguments);
};

const Mathfloor = Math.floor;
Math.floor = function() {
    _IOSTInstruction_counter.incr(5);
    return Mathfloor.call(this, ...arguments);
};

const Mathlog = Math.log;
Math.log = function() {
    _IOSTInstruction_counter.incr(40);
    return Mathlog.call(this, ...arguments);
};

const Mathlog10 = Math.log10;
Math.log10 = function() {
    _IOSTInstruction_counter.incr(40);
    return Mathlog10.call(this, ...arguments);
};

const Mathmax = Math.max;
Math.max = function() {
    _IOSTInstruction_counter.incr(20);
    return Mathmax.call(this, ...arguments);
};

const Mathmin = Math.min;
Math.min = function() {
    _IOSTInstruction_counter.incr(20);
    return Mathmin.call(this, ...arguments);
};

const Mathpow = Math.pow;
Math.pow = function() {
    _IOSTInstruction_counter.incr(40);
    return Mathpow.call(this, ...arguments);
};

const Mathround = Math.round;
Math.round = function() {
    _IOSTInstruction_counter.incr(5);
    return Mathround.call(this, ...arguments);
};

const Mathsqrt = Math.sqrt;
Math.sqrt = function() {
    _IOSTInstruction_counter.incr(40);
    return Mathsqrt.call(this, ...arguments);
};

// JSON
const JSONparse = JSON.parse;
JSON.parse = function () {
    if (arguments[0] == null) {
        _IOSTInstruction_counter.incr(10);
    } else {
        _IOSTInstruction_counter.incr(arguments[0].length * 2);
    }
    return JSONparse.call(this, ...arguments);
};

const JSONstringify = JSON.stringify;
JSON.stringify = function () {
    const rs = JSONstringify.call(this, ...arguments);
    _IOSTInstruction_counter.incr(rs.length * 2);
    return rs;
};

// Functions
parseFloat = null;
parseInt = null;
decodeURI = null;
decodeURIComponent = null;
encodeURI = null;
encodeURIComponent = null;
escape = null;
unescape = null;

// Fundamental Objects
Function = null;
Boolean = null;
EvalError = null;
RangeError = null;
ReferenceError = null;
SyntaxError = null;
TypeError = null;
URIError = null;

// Numbers and dates
Date = null;

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