"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
/**
 * Provides simple way to "proxify" nested objects and validate the changes.
 */
module.exports = (function () {
    let _updatedMap = {};

    var _create = function(target, path) {
        var proxies = {};
        var getPath = function getPath(path, prop) {
            if (path.length !== 0)
                return path + "." + prop;
            else
                return prop;
        };

        var handler = {
            get: function(target, property, receiver) {
                // var aa = {
                //     target: target,
                //     prop: property,
                //     path: path,
                //     // path: getPath(path, property),
                //     type: typeof target[property]
                // };
                // _native_log('observer get: ' + JSON.stringify(aa));

                var totalPath = getPath(path, property);
                var dotIndex = totalPath.indexOf('.');

                if (dotIndex === -1) {
                    if (typeof target[property] !== 'function' && typeof target[property] !== 'object') {
                        let objectStorage = IOSTContractStorage.get(property, target[property]);
                        // _native_log("observer get return: " + JSON.stringify(objectStorage));
                        return objectStorage;
                    }
                } else {
                    if (typeof target[property] !== 'function' && typeof target[property] !== 'object') {
                        return target[property];
                    }
                }

                // if (target[property] instanceof BigNumber || target[property] instanceof Int64 || typeof target[property] === 'string' || typeof target[property] === 'number') {
                //     if (dotIndex === -1) {
                //         var objectStorage = IOSTContractStorage.get(property, target[property]);
                //         return objectStorage;
                //     } else {
                //         return target[property];
                //     }
                // }

                var value = Reflect.get(target, property, receiver);
                if (typeof target[property] === 'object' && target[property] !== null) {
                    if (dotIndex === -1) {
                        let objectStorage = IOSTContractStorage.get(property, target[property]);

                        // _native_log("observer get return: " + JSON.stringify(objectStorage));
                        let proxy = _create(objectStorage, getPath(path, property));
                        proxies[property] = proxy;
                        return proxy;
                    } else {
                        let objInMemory = target[property];
                        // _native_log("observer get return: " + JSON.stringify(objInMemory));
                        let proxy = _create(objInMemory, getPath(path, property));
                        proxies[property] = proxy;
                        return proxy;
                    }
                }

                if ((property === "set" || property === "get") && typeof value === "function") {
                    const origSet = value;
                    var args = [];
                    for (var _i = 0; _i < arguments.length; _i++) {
                        args[_i] = arguments[_i];
                    }
                    value = function(...args) {
                        return origSet.apply(target, arguments);
                    };
                }
                return value;
            },
            set: function(target, prop, value, receiver) {
                // var aa = {
                //     target: target,
                //     prop: prop,
                //     path: getPath(path, prop),
                //     value: value,
                //     type: typeof target[prop]
                // };
                // _native_log('observer set: ' + JSON.stringify(aa));
                target[prop] = value;

                var totalPath = getPath(path, prop);
                var dotIndex = totalPath.indexOf('.');

                var pathList = totalPath.split('.');
                if (pathList.length === 2) {
                    _updatedMap[pathList[0]] = target;
                }

                // _native_log("setttt: " + JSON.stringify(target) + ' : ' + pathList.length + " : " + path + " : " + prop);

                if (pathList.length === 1) {
                    IOSTContractStorage.put(prop, value)
                } else if (pathList.length === 2) {
                    IOSTContractStorage.put(path, target);
                } else {
                    if (pathList.length === 3) {
                        let obj = IOSTContractStorage.get(pathList[0]);
                        // pathList.forEach(function (value, index) {
                        //     _native_log("xxzzxx: " + value + " " + index)
                        // });
                        // let obj = _updatedMap[pathList[0]];
                        // _native_log("obj: " + JSON.stringify(obj));
                        obj[pathList[1]][pathList[2]] = value;
                        IOSTContractStorage.put(pathList[0], obj);
                    }
                    if (pathList.length === 4) {
                        let obj = IOSTContractStorage.get(pathList[0]);
                        // let obj = _updatedMap[pathList[0]];
                        obj[pathList[1]][pathList[2]][pathList[3]] = value;
                        IOSTContractStorage.put(pathList[0], obj);
                    }
                }

                // if (dotIndex !== -1) {
                //     IOSTContractStorage.put(totalPath.substr(0, totalPath.indexOf('.')), target);
                // } else {
                //     IOSTContractStorage.put(prop, value);
                // }
                return true;
            },
        };

        return new Proxy(target, handler);
    };

    return {
        create: function(target) {
            return _create(target, '');
        }
    }
})();