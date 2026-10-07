// v8sort.js — V8 6.3 时代 array.js 中 InnerArraySort 的忠实移植
//（QuickSort + InsertionSort + GetThirdIndex）。
//
// 主网节点的 C++ V8(6.3.292.48) 用该算法执行 Array.prototype.sort，
// 其调用用户比较器的次数/顺序是确定性的，并计入合约 gas。
// goja 原生 sort 的比较次数不同会导致 gas 分叉，因此这里用纯 JS 复刻
// V8 的比较器调用序列。本文件代码不插桩计费（与 V8 原生 sort 内部免费一致），
// 只有比较器（合约代码，已插桩）产生 gas。
//
// 已知简化：只完整支持稠密数组；空洞/原型链元素按 V8 %RemoveArrayHoles
// 的语义压缩到末尾再排序前缀。稀疏数组排序在合约中几乎不出现。
(function () {
    "use strict";

    function defaultCompare(x, y) {
        if (x === y) return 0;
        // V8: %_IsSmi 双方为 32 位整数时按数值比较，否则转字符串比较
        var smiX = typeof x === 'number' && isFinite(x) && Math.floor(x) === x && Math.abs(x) < 1073741824;
        var smiY = typeof y === 'number' && isFinite(y) && Math.floor(y) === y && Math.abs(y) < 1073741824;
        if (smiX && smiY) {
            return x < y ? -1 : 1;
        }
        x = String(x);
        y = String(y);
        if (x == y) return 0;
        return x < y ? -1 : 1;
    }

    function innerSort(array, length, comparefn) {
        if (typeof comparefn !== 'function') {
            comparefn = defaultCompare;
        }

        function InsertionSort(a, from, to) {
            for (var i = from + 1; i < to; i++) {
                var element = a[i];
                for (var j = i - 1; j >= from; j--) {
                    var tmp = a[j];
                    var order = comparefn(tmp, element);
                    if (order > 0) {
                        a[j + 1] = tmp;
                    } else {
                        break;
                    }
                }
                a[j + 1] = element;
            }
        }

        function GetThirdIndex(a, from, to) {
            var t_array = [];
            var increment = 200 + ((to - from) & 15);
            var j = 0;
            from += 1;
            to -= 1;
            for (var i = from; i < to; i += increment) {
                t_array[j] = [i, a[i]];
                j++;
            }
            // 用原始 innerSort 而不是 Array.prototype.sort（后者被
            // environment.js 包装会计费，而 V8 的 InternalArray sort 不走包装）
            innerSort(t_array, t_array.length, function (a, b) {
                return comparefn(a[1], b[1]);
            });
            var third_index = t_array[t_array.length >> 1][0];
            return third_index;
        }

        function QuickSort(a, from, to) {
            var third_index = 0;
            while (true) {
                if (to - from <= 10) {
                    InsertionSort(a, from, to);
                    return;
                }
                if (to - from > 1000) {
                    third_index = GetThirdIndex(a, from, to);
                } else {
                    third_index = from + ((to - from) >> 1);
                }
                var v0 = a[from];
                var v1 = a[to - 1];
                var v2 = a[third_index];
                var c01 = comparefn(v0, v1);
                if (c01 > 0) {
                    var tmp = v0;
                    v0 = v1;
                    v1 = tmp;
                }
                var c02 = comparefn(v0, v2);
                if (c02 >= 0) {
                    var tmp = v0;
                    v0 = v2;
                    v2 = v1;
                    v1 = tmp;
                } else {
                    var c12 = comparefn(v1, v2);
                    if (c12 > 0) {
                        var tmp = v1;
                        v1 = v2;
                        v2 = tmp;
                    }
                }
                a[from] = v0;
                a[to - 1] = v2;
                var pivot = v1;
                var low_end = from + 1;
                var high_start = to - 1;
                a[third_index] = a[low_end];
                a[low_end] = pivot;

                partition: for (var i = low_end + 1; i < high_start; i++) {
                    var element = a[i];
                    var order = comparefn(element, pivot);
                    if (order < 0) {
                        a[i] = a[low_end];
                        a[low_end] = element;
                        low_end++;
                    } else if (order > 0) {
                        do {
                            high_start--;
                            if (high_start == i) break partition;
                            var top_elem = a[high_start];
                            order = comparefn(top_elem, pivot);
                        } while (order > 0);
                        a[i] = a[high_start];
                        a[high_start] = element;
                        if (order < 0) {
                            element = a[i];
                            a[i] = a[low_end];
                            a[low_end] = element;
                            low_end++;
                        }
                    }
                }
                if (to - high_start < low_end - from) {
                    QuickSort(a, high_start, to);
                    to = low_end;
                } else {
                    QuickSort(a, from, low_end);
                    from = high_start;
                }
            }
        }

        // 等价于 V8 的 %RemoveArrayHoles：把已定义元素按原序压实到前面，
        // undefined/空洞挪到末尾，返回已定义元素个数。
        // 注意不能用 Array.prototype.push 等（environment.js 给它们加了计费包装），
        // 只能用索引读写。
        var definedLen = 0;
        for (var i = 0; i < length; i++) {
            if (i in array && array[i] !== undefined) {
                if (definedLen !== i) {
                    array[definedLen] = array[i];
                }
                definedLen++;
            }
        }
        for (var i = definedLen; i < length; i++) {
            array[i] = undefined;
        }

        QuickSort(array, 0, definedLen);
        return array;
    }

    Array.prototype.sort = function (comparefn) {
        if (this === null || this === undefined) {
            throw new TypeError("Array.prototype.sort called on null or undefined");
        }
        if (comparefn !== undefined && typeof comparefn !== 'function') {
            throw new TypeError("The comparison function must be either a function or undefined");
        }
        var array = Object(this);
        var length = array.length >>> 0;
        if (length < 2) return array;
        return innerSort(array, length, comparefn);
    };
})();
