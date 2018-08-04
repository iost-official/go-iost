const esprima = require('esprima-4.0.1.js');
const estraverse = require('estraverse.js');

var arr = new Array();

function InjectInstructions(script) {
    var ast = esprima.parseScript(script, {
        range: true,
    });

    estraverse(ast, {
        enter: function (node, parent) {
        },
        leave: function(node, parent) {
            if (node.type == 'IfStatement') {
                arr.push({
                    range: node.test.range,
                    type: 'simpleInsert',
                })
            }
            if (node.type == 'ForStatement') {
                arr.push({
                    range: node.test.range,
                    type: 'simpleInsert',
                })
            }
            if (node.type == 'WhileStatement') {
                arr.push({
                    range: node.test.range,
                    type: 'simpleInsert',
                })
            }
        },

        fallback: "iteration"
    });

    var injectedScript = '';
    var offset = 0;
    arr.forEach(function (segment) {
        injectedScript += script.substring(offset, offset+segment.range[0]);
        offset += segment.range[0]
        if (segment.type == 'simpleInsert') {
            injectedScript += '_instruction.add(1) && (' + script.substring(offset, segment.range[1]) + ')'
        }
        injectedScript += script.substring(segment.range[1])
    })

    return script;
}

module.exports = InjectInstructions;