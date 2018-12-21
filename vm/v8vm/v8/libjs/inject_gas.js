'use strict';

// expression to be charged for some instruction, used to calculate inject value
const chargedExpression = {
    ThrowStatement: 50,
    // expression
    CallExpression: 4,
    TaggedTemplateExpression: 4,
    NewExpression: 8,
    YieldExpression: 8,
    MemberExpression: 4,
    MetaProperty: 4,
    AssignmentExpression: 3,
    UpdateExpression: 3,
    BinaryExpression: 3,
    UnaryExpression: 3,
    LogicalExpression: 3,
    ConditionalExpression: 3,

    ObjectExpression: 0.1,
    ArrayExpression: 0.1,
    FunctionExpression: 1,
    ArrowFunctionExpression: 3,
    // declaration
    ClassDeclaration: 150,
    FunctionDeclaration: 3,
    VariableDeclarator: 3,
    VariableDeclaratorWithoutInit: 3,
    MethodDefinition: 2,
    // literal
    StringLiteral: 0.1,
    // statement
    ForStatement: 1,
    ForInStatement: 4,
    ForOfStatement: 2,
    WhileStatement: 1,
    DoWhileStatement: 1
};
// statement before which can inject gas, used to find inject location
const InjectableStatement = {
    ExpressionStatement: 1,
    ReturnStatement: 1,
    ThrowStatement: 1,
    SwitchStatement: 1,
    DoWhileStatement: 1,
    ForStatement: 1,
    ForInStatement: 1,
    ForOfStatement: 1,
    IfStatement: 1,
    WhileStatement: 1
};

const InjectType = {
    gasIncr: 1,
    gasIncrWithComma: 2,
    gasIncrWithSemicolon: 3,
    leftBracket: 4,
    rightBracket: 5,
    leftBrace: 6,
    rightBrace: 7,
    str: 8
};

function genInjectionStr(injectionPoint) {
    switch (injectionPoint["type"]) {
        case InjectType.gasIncr:
            return "_IOSTInstruction_counter.incr(" + injectionPoint.value + ")";
        case InjectType.gasIncrWithComma:
            return "_IOSTInstruction_counter.incr(" + injectionPoint.value + "),";
        case InjectType.gasIncrWithSemicolon:
            return "_IOSTInstruction_counter.incr(" + injectionPoint.value + ");";
        case InjectType.leftBracket:
            return "(";
        case InjectType.rightBracket:
            return ")";
        case InjectType.leftBrace:
            return "{";
        case InjectType.rightBrace:
            return "}";
        case InjectType.str:
            return injectionPoint.value;
    }
}


function checkInvalidKeyword(tokens) {
    for (let i = 0; i < tokens.length; i++) {
        if ((tokens[i].type === "Identifier" || tokens[i].type === "Literal") &&
            (tokens[i].value === "_IOSTInstruction_counter" || tokens[i].value === "_IOSTBinaryOp" || tokens[i].value === "IOSTInstruction" ||
             tokens[i].value === "_IOSTTemplateTag" || tokens[i].value === "_IOSTSpreadElement")) {
            throw new Error("use of _IOSTInstruction_counter or _IOSTBinaryOp keyword is not allowed");
        }
        if (tokens[i].type === "RegularExpression") {
            throw new Error("use of RegularExpression is not allowed." + tokens[i].val)
        }
    }
}

function checkOperator(tokens) {
    for (let i = 0; i < tokens.length; i++) {
        if (tokens[i].type === "Punctuator" &&
            (tokens[i].value === "*" || tokens[i].value === "*" || tokens[i].value === "*" || tokens[i].value === "/" || tokens[i].value === "%" ||
                tokens[i].value === "+=" || tokens[i].value === "-=" || tokens[i].value === "*=" || tokens[i].value === "/=" || tokens[i].value === "%=" )) {
            throw new Error("use of +-*/% operators is not allowed");
        }
    }
}

let injectionMap = new Map();

function addInjection(pos, type, value, before = false) {
    if (!injectionMap.has(pos)) {
        injectionMap.set(pos, []);
    }
    let arr = injectionMap.get(pos);
    let index;
    if (before) {
        index = arr.length > 0 ? 0 : -1;
    }
    else {
        index = arr.length > 0 ? arr.length - 1 : -1;
    }

    if (type > 3) {
        if (before) {
            arr.unshift({
                "type": type,
                "value": value
            });
            index = 0;
        } else {
            arr.push({
                "type": type,
                "value": value
            });
            index = arr.length - 1;
        }
    }
    else {
        if (index >= 0 && (arr[index].type === InjectType.gasIncr || arr[index].type === InjectType.gasIncrWithSemicolon || arr[index].type === InjectType.gasIncrWithComm)) {
            arr[index].value += value;
        }
        else {
            if (before) {
                arr.unshift({
                    "type": type,
                    "value": 0
                });
                index = 0;
            }
            else {
                arr.push({
                    "type": type,
                    "value": 0
                });
                index = arr.length - 1;
            }
            arr[index].value += value;
        }

    }
    injectionMap.set(pos, arr);
    return {
        "pos": pos,
        "index": index
    }
}

function addInjectionPoint(node, type, value) {
    if (!node || node === null) {
        return {};
    }
    if (node.type !== "EmptyExpression") {
        return addInjection(node.range[0], type, value);
    }
    else {
        return addInjection(node.range[0], InjectType.gasIncr, value);
    }
}

function ensure_block(node) {
    if (!node || node === null) {
        return;
    }
    if (node.type !== "BlockStatement") {
        addInjection(node.range[0], InjectType.leftBrace, 0);
        addInjection(node.range[1], InjectType.rightBrace, 0, true);
    }
}

function ensure_bracket(node) {
    if (!node || node === null) {
        return;
    }
    addInjection(node.range[0], InjectType.leftBracket, 0);
    addInjection(node.range[1], InjectType.rightBracket, 0, true);
}

function processNode(node, parentNode, lastInjection) {
    let newLastInjection = lastInjection;

    if (node.type in InjectableStatement) {
        newLastInjection = addInjection(node.range[0], InjectType.gasIncrWithSemicolon, 0);
    }
    if (node.type === "VariableDeclaration" && (parentNode === null ||  parentNode.type !== "ForStatement" &&
            parentNode.type !== "ForInStatement" && parentNode.type !== "ForOfStatement")) {
        newLastInjection = addInjection(node.range[0], InjectType.gasIncrWithSemicolon, 0);
    }

    if (node.type === "IfStatement") {
        ensure_block(node.consequent);
        ensure_block(node.alternate);
        let injectionPoint = addInjectionPoint(node.test, InjectType.gasIncrWithComma, 0);
        return [newLastInjection, {
            "test": injectionPoint
        }];

    } else if (node.type === "ForStatement") {
        ensure_block(node.body);

        let body = node.body;
        let pos = body.range[0];
        if (body.type === 'BlockStatement') {
            pos = body.range[0] + 1;
        }
        let ip0 = addInjection(pos, InjectType.gasIncrWithSemicolon, chargedExpression[node.type]);

        let injectionPoint2 = addInjectionPoint(node.test, InjectType.gasIncrWithComma, 0);
        let injectionPoint3 = addInjectionPoint(node.update, InjectType.gasIncrWithComma, 0);
        return [newLastInjection, {
            "test": injectionPoint2,
            "update": injectionPoint3,
            "body": ip0
        }];

    } else if (node.type === "ForInStatement" || node.type === "ForOfStatement") {
        ensure_block(node.body);

        let body = node.body;
        let pos = body.range[0];
        if (body.type === 'BlockStatement') {
            pos = body.range[0] + 1;
        }
        let ip0 = addInjection(pos, InjectType.gasIncrWithSemicolon, chargedExpression[node.type]);

        return [newLastInjection, {
            "body": ip0
        }];

    } else if (node.type === "WhileStatement" || node.type === "DoWhileStatement") {
        ensure_block(node.body);
        let body = node.body;
        let pos = body.range[0];
        if (body.type === 'BlockStatement') {
            pos = body.range[0] + 1;
        }
        let ip0 = addInjection(pos, InjectType.gasIncrWithSemicolon, 1);

        let injectionPoint = addInjectionPoint(node.test, InjectType.gasIncrWithComma, chargedExpression[node.type]);
        return [newLastInjection, {
            "test": injectionPoint,
            "body": ip0
        }];

    } else if (node.type === "WithStatement") {
        ensure_block(node.body);
        return [newLastInjection, {}];

    } else if (node.type === "SwitchStatement") {
        return [newLastInjection, {}];

    } else if (node.type === "SwitchCase") {
        let injectionPoint = addInjectionPoint(node.test, InjectType.gasIncrWithComma, 0);
        return [newLastInjection, {
            "test": injectionPoint
        }];

    } else if (node.type === "ArrowFunctionExpression") {
        let value = chargedExpression[node.type];
        console.log("arrow function value, ", value);
        if (newLastInjection === null) {
            newLastInjection = addInjection(node.range[0], InjectType.gasIncrWithSemicolon, 0);
        }
        injectionMap.get(newLastInjection.pos)[newLastInjection.index].value += value;

        if (node.body.type !== 'BlockStatement') {
            addInjection(node.body.range[0], InjectType.str, "function(){");
            let injectionPoint = addInjectionPoint(node.body, InjectType.gasIncrWithSemicolon, 0);
            addInjection(node.body.range[0], InjectType.str, "return ");
            addInjection(node.body.range[1], InjectType.str, "}()", true);

            return [newLastInjection, {
                "body": injectionPoint
            }];
        }
        return [newLastInjection, {}];

    } else if (node.type === "ConditionalExpression") {
        ensure_bracket(node.test);
        ensure_bracket(node.consequent);
        ensure_bracket(node.alternate);
        let injectionPoint1 = addInjectionPoint(node.test, InjectType.gasIncrWithComma, 0);
        let injectionPoint2 = addInjectionPoint(node.consequent, InjectType.gasIncrWithComma, 0);
        let injectionPoint3 = addInjectionPoint(node.alternate, InjectType.gasIncrWithComma, 0);

        let value = chargedExpression[node.type];
        if (newLastInjection === null) {
            newLastInjection = addInjection(node.range[0], InjectType.gasIncrWithSemicolon, 0);
        }
        injectionMap.get(newLastInjection.pos)[newLastInjection.index].value += value;

        return [newLastInjection, {
            "test": injectionPoint1,
            "consequent": injectionPoint2,
            "alternate": injectionPoint3
        }];

    } else {

        let value = chargedExpression[node.type];
        if (value === null || value === undefined) {
            if (node.type === "Literal" && typeof node.value === "string") {
                value = 2 + chargedExpression["StringLiteral"] * node.value.length;
            } else {
                return [newLastInjection, {}];
            }
        }
        if (node.type === "VariableDeclarator" && (node.init === null || node.init === undefined)) {
            value = chargedExpression["VariableDeclaratorWithoutInit"]
        }
        if (node.type === "ObjectExpression" && node.properties !== undefined && node.properties.length > 0) {
            value = 2 + value * node.properties.length;
        }
        if (node.type === "ArrayExpression" && node.elements !== undefined && node.elements.length > 0) {
            value = 2 + value * node.elements.length;
        }
        if (newLastInjection === null) {
            newLastInjection = addInjection(node.range[0], InjectType.gasIncrWithSemicolon, 0);
        }
        injectionMap.get(newLastInjection.pos)[newLastInjection.index].value += value;
        return [newLastInjection, {}];
    }
}

function traverse(node, parentNode, lastInjection) {
    let newLastInjection;
    let childInjection;
    [newLastInjection, childInjection] = processNode(node, parentNode, lastInjection);
    if (childInjection === false || childInjection === null) {
        childInjection = {};
    }
    for (let key in node) {
        if (node.hasOwnProperty(key)) {
            let child = node[key];
            if (typeof child === 'object' && child !== null) {

                let keyInjection = newLastInjection;
                if (childInjection.hasOwnProperty(key)) {
                    keyInjection = childInjection[key];
                }

                traverse(child, node, keyInjection);
            }
        }
    }
}

function genNewScript(source) {
    let arr = Array.from(injectionMap.keys());
    arr.sort(function (a, b) {
        return a - b;
    });

    let offset = 0,
    newSource = "";
    arr.forEach(function (pos) {
        let injectionArr = injectionMap.get(pos);

        newSource += source.slice(offset, pos);
        injectionArr.forEach(function (injectionPoint) {
            if (injectionPoint.type > 3 || injectionPoint.value > 0) {
                newSource += genInjectionStr(injectionPoint);
            }
        });
        offset = pos;
    });
    newSource += source.slice(offset);

    return newSource;
}

function processOperator(node, pnode) {
    if (node.type === "ArrayPattern" || node.type === "ObjectPattern") {
        throw new Error("use of ArrayPattern or ObjectPattern is not allowed." + JSON.stringify(node));
    }
    let ops = ['+', '-', '*', '/', '%', '**', '|', '&', '^', '>>', '>>>', '<<', '==', '!=', '===', '!==', '>', '>=', '<', '<='];

    if (node.type === "AssignmentExpression" && node.operator !== '=') {
        let subnode = {};
        subnode.operator = node.operator.substr(0, node.operator.length - 1);
        subnode.type = 'BinaryExpression';
        subnode.left = Object.assign({}, node.left);
        subnode.right = node.right;
        node.operator = '=';
        node.right = subnode;

    } else if (node.type === "BinaryExpression" && ops.includes(node.operator)) {
        let newnode = {};
        newnode.type = "CallExpression";
        let calleeNode = {};
        calleeNode.type = 'Identifier';
        calleeNode.name = '_IOSTBinaryOp';
        newnode.callee = calleeNode;
        let opNode = {};
        opNode.type = 'Literal';
        opNode.value = node.operator;
        opNode.raw = '\'' + node.operator + '\'';
        newnode.arguments = [node.left, node.right, opNode];
        node = newnode;
    } else if (node.type === "TemplateLiteral" && (pnode === undefined || pnode.type !== "TaggedTemplateExpression")) {
        let newnode = {};
        newnode.type = "TaggedTemplateExpression";
        let tagNode = {};
        tagNode.type = 'Identifier';
        tagNode.name = '_IOSTTemplateTag';
        newnode.tag = tagNode;
        newnode.quasi = node;
        node = newnode;
    } else if (node.type === "SpreadElement") {
        let newnode = {};
        newnode.type = "CallExpression";
        let calleeNode = {};
        calleeNode.type = 'Identifier';
        calleeNode.name = '_IOSTSpreadElement';
        newnode.callee = calleeNode;
        newnode.arguments = [node.argument];
        node.argument = newnode;
    }
    return node;
}

function traverseOperator(node, pnode) {
    node = processOperator(node, pnode);
    for (let key in node) {
        if (node.hasOwnProperty(key)) {
            let child = node[key];
            if (typeof child === 'object' && child !== null) {
                node[key] = traverseOperator(child, node);
            }
        }
    }
    return node;
}

function handleOperator(ast) {
    ast = traverseOperator(ast);
    // generate source from ast
    return escodegen.generate(ast);
}

function injectGas(source) {
    let ast = esprima.parseScript(source, {
        comment: true,
        tokens: true,
        loc: true
    });

    checkInvalidKeyword(ast.tokens);
    // checkOperator(ast.tokens);

    // replace operator with function
    source = handleOperator(ast);
    ast = esprima.parseScript(source, {
        range: true,
        comment: true,
    });
    traverse(ast, null, null);

    return genNewScript(source);
}

module.exports = injectGas;
