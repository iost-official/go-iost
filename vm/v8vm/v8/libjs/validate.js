'use strict';

function checkAbiNameValid(name) {
    if (name.length <= 0 || name.length > 32) {
        throw new Error("abi name invalid. abi name length should be between 1,32  got " + name)
    }
    if (name[0] === "_") {
        throw new Error("abi name invalid. abi name shouldn't start with _");
    }
    for (let i in name) {
        let ch = name.charAt(i);
        if (!(ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '_')) {
            throw new Error("abi name invalid. abi name contains invalid character " + ch);
        }
    }
}

function checkAmountLimitValid(amountLimit) {
    for (let i in amountLimit) {
        let limit = amountLimit[i];
        if (typeof limit.token !== "string") {
            throw new Error("amountLimit invalid. token in amountLimit should be string type. limit = " + limit);
        }
        if (typeof limit.val!== "string") {
            throw new Error("amountLimit invalid. val in amountLimit should be string type. limit = " + limit);
        }
    }
}

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
    checkAbiNameValid(a.name);
    if (a.amountLimit !== undefined && a.amountLimit !== null) {
        checkAmountLimitValid(a.amountLimit);
    }
    if (a.name === "init") {
        throw new Error("abi shouldn't contain internal function: init");
    }
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
        if (m.key.name === "constructor") {
            throw new Error("smart contract class shouldn't contain constructor method!!");
        }
        methodMap[m.key.name] = m.value.params;
    }
    if (methodMap["init"] === undefined || methodMap["init"] === null) {
        throw new Error("init not found!");
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