'use strict';
class ArrayFunc{
    from() {
        return Array.from('foo');
    }
    of() {
        return Array.of(1, 2, 3);
    }
}

module.exports = ArrayFunc;