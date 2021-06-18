"use strict";

const fs = require("fs");
const esprima = require("esprima/dist/esprima.js");

const lang = "javascript";
const version = "1.0.0";

function isClassDecl(stat) {
    return !!(stat && stat.type === "ClassDeclaration");
}

function isExport(stat) {
    return !!(
        stat &&
        stat.type === "AssignmentExpression" &&
        stat.left &&
        stat.left.type === "MemberExpression" &&
        stat.left.object &&
        stat.left.object.type === "Identifier" &&
        stat.left.object.name === "module" &&
        stat.left.property &&
        stat.left.property.type === "Identifier" &&
        stat.left.property.name === "exports"
    );
}

function getExportName(stat) {
    if (stat.right.type !== "Identifier") {
        throw new Error("module.exports should be assigned to an identifier");
    }
    return stat.right.name;
}

function isPublicMethod(def) {
    return def.key.type === "Identifier" && def.value.type === "FunctionExpression" && !def.key.name.startsWith("_");
}

function genAbi(def, lastPos, comments) {
    for (let param of def.value.params) {
        if (param.type !== "Identifier") {
            throw new Error("invalid method parameter type. must be Identifier, got " + param.type);
        }
    }
    let abi = {
        name: def.key.name,
        args: new Array(def.value.params.length).fill("string"),
        amountLimit: []
        //"description": ""
    };
    for (let i = comments.length - 1; i >= 0; i--) {
        let comment = comments[i];
        if (comment.range[0] > lastPos && comment.range[1] < def.range[0]) {
            for (let i in def.value.params) {
                let name = def.value.params[i].name;
                let reg = new RegExp("@param\\s*{([a-zA-Z]+)}\\s*" + name);
                let reg1 = new RegExp("@param\\s*" + name + "\\s*{([a-zA-Z]+)}");
                let res = null;
                if (((res = comment.value.match(reg)), res !== null)) {
                    abi.args[i] = res[1];
                } else if (((res = comment.value.match(reg1)), res !== null)) {
                    abi.args[i] = res[1];
                }
            }
            break;
        }
    }
    return abi;
}

function genAbiArr(stat, comments) {
    let abiArr = [];
    if (!isClassDecl(stat) || stat.body.type !== "ClassBody") {
        throw new Error("invalid statement for generate abi. stat = " + stat);
        return null;
    }
    let initFound = false;
    let lastPos = stat.body.range[0];
    for (let def of stat.body.body) {
        if (def.type === "MethodDefinition" && isPublicMethod(def)) {
            if (def.key.name === "constructor") {
                throw new Error("smart contract class shouldn't contain constructor method!");
            } else if (def.key.name === "init") {
                initFound = true;
            } else {
                abiArr.push(genAbi(def, lastPos, comments));
                lastPos = def.range[1];
            }
        }
    }
    if (!initFound) {
        throw new Error("init not found!");
        return null;
    }
    return abiArr;
}

function checkInvalidKeyword(tokens) {
    for (let i = 0; i < tokens.length; i++) {
        if (
            (tokens[i].type === "Identifier" || tokens[i].type === "Literal") &&
            (tokens[i].value === "_IOSTInstruction_counter" ||
                tokens[i].value === "_IOSTBinaryOp" ||
                tokens[i].value === "IOSTInstruction" ||
                tokens[i].value === "_IOSTTemplateTag" ||
                tokens[i].value === "_IOSTSpreadElement")
        ) {
            throw new Error("use of _IOSTInstruction_counter or _IOSTBinaryOp keyword is not allowed");
        }
        if (tokens[i].type === "RegularExpression") {
            throw new Error("use of RegularExpression is not allowed." + tokens[i].val);
        }
        if (tokens[i].type === "Keyword" && (tokens[i].value === "try" || tokens[i].value === "catch")) {
            throw new Error("use of try catch is not supported");
        }
    }
}

function processContractSrc(source) {
    let ast = esprima.parseModule(source, {
        range: true,
        loc: false,
        comment: true,
        tokens: true
    });

    let abiArr = [];
    if (!ast || ast === null || !ast.body || ast.body === null || ast.body.length === 0) {
        throw new Error("invalid source! ast = " + ast);
    }

    checkInvalidKeyword(ast.tokens);

    let className;
    for (let stat of ast.body) {
        if (isClassDecl(stat)) {
        } else if (stat.type === "ExpressionStatement" && isExport(stat.expression)) {
            className = getExportName(stat.expression);
        }
    }
    for (let stat of ast.body) {
        if (isClassDecl(stat) && stat.id.type === "Identifier" && stat.id.name === className) {
            abiArr = genAbiArr(stat, ast.comments);
        }
    }

    let abi = {};
    abi["lang"] = lang;
    abi["version"] = version;
    abi["abi"] = abiArr;
    let abiStr = JSON.stringify(abi, null, 4);
    return abiStr;
}

function processContract(file) {
    const source = fs.readFileSync(file);

    if (source === undefined) {
        throw new Error("invalid file content. Is " + file + " exists?");
    }
    const abiStr = processContractSrc(source.toString());

    fs.writeFile(file + ".abi", abiStr, function(err) {
        if (err) {
            return console.log(err);
        }
        console.log("The new abi file was saved as " + file + ".abi");
    });
}

module.exports = processContract;
