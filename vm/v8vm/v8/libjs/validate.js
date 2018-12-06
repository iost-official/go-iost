'use strict';

function checkExports(ast) {
    if (ast.type !== "Program" || ast.body.length <= 1) {
        throw new Error("empty source code.");
    }
    let last = ast.body[ast.body.length - 1];
    let exp = last.expression;
    if (last.type !== "ExpressionStatement" || exp.type !== "AssignmentExpression" || exp.operator !== "="
    || exp.left.type !== "MemberExpression" || exp.left.object.type !== "Identifier" || exp.left.property.type !== "Identifier"
    || exp.left.object.name !== "module" || exp.left.property.name !== "exports" || exp.right.type !== "Identifier") {
        throw new Error("last expression must be 'module.exports = CLASS_IDENTIFIER;'");
    }
    return exp.right.name;
}

function checkOneAbi(a, methodMap) {
    const params = methodMap[a.name];
    if (params === undefined || params === null || !Array.isArray(params)) {
        throw new Error("abi not defined in source code: " + a.name);
    }
    const args = a.args || [];
    if (args.length !== params.length) {
        throw new Error("args length not match: " + a.name);
    }
    for (let i in args) {
        let arg = args[i];
        if (!["string", "bool", "number", "json"].includes(arg)) {
            throw new Error(`args should be one of ["string", "bool", "number", "json"]`)
        }
    }
}

function checkOneClass(node, abi, cls) {
    if (node.type !== "ClassDeclaration" || node.id.type !== "Identifier" || node.id.name !== cls || node.body.type !== "ClassBody") {
        return false;
    }
    let methodMap = {};
    for (m of node.body.body) {
        if (m.type !== "MethodDefinition" || m.key.type !== "Identifier" || m.value.type !== "FunctionExpression") {
            continue;
        }
        methodMap[m.key.name] = m.value.params;
    }

    for (const a of abi) {
        checkOneAbi(a, methodMap);
    }
    return true;
}

function checkAbi(body, abi, cls) {
    for (const node of body) {
        if (checkOneClass(node, abi, cls)) {
            return true;
        }
    }
    throw new Error("abi not match your source code.");
}

function validate(source, abi) {
    let ast = esprima.parseScript(source);

    try {
        let cls = checkExports(ast);
        checkAbi(ast.body, abi, cls)
    } catch (err) {
        return err.toString();
    }
    return "success";
}

module.exports = validate;