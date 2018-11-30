'use strict';

let esprima = require('esprima/dist/esprima.js');

let lang = "javascript";
let version = "1.0.0";

function isClassDecl(stat) {
	return !!(stat && stat.type === "ClassDeclaration");
}

function isExport(stat) {
	return !!(stat && stat.type === "AssignmentExpression" && stat.left && stat.left.type === "MemberExpression"
    && stat.left.object && stat.left.object.type === "Identifier" && stat.left.object.name === "module"
    && stat.left.property && stat.left.property.type === "Identifier" && stat.left.property.name === "exports");
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

function genAbi(def) {
	return {
		"name": def.key.name,
		"args": new Array(def.value.params.length).fill("string"),
		"amountLimit": [{
            "token": "iost",
            "val": "0"
        }]
	};
}

function genAbiArr(stat) {
	let abiArr = [];
	if (!isClassDecl(stat) || stat.body.type !== "ClassBody") {
		console.error("invalid statement for generate abi. stat = " + stat);
		return null;
	}
	let initFound = false;
	for (let def of stat.body.body) {
		if (def.type === "MethodDefinition" && isPublicMethod(def)) {
			if (def.key.name === "constructor") {
			} else if (def.key.name === "init") {
				initFound = true;
			} else {
				abiArr.push(genAbi(def));
			}
		}
	}
	if (!initFound) {
		console.error("init not found!");
		return null;
	}
	return abiArr;
}

function checkInvalidKeyword(tokens) {
    for (let i = 0; i < tokens.length; i++) {
        if ((tokens[i].type === "Identifier" || tokens[i].type === "Literal") &&
            (tokens[i].value === "_IOSTInstruction_counter" || tokens[i].value === "_IOSTBinaryOp")) {
            throw new Error("use of _IOSTInstruction_counter or _IOSTBinaryOp keyword is not allowed");
        }
        if (tokens[i].type === "RegularExpression") {
            throw new Error("use of RegularExpression is not allowed. " + tokens[i].value)
        }
    }
}

function checkOperator(tokens) {
    for (let i = 0; i < tokens.length; i++) {
        if (tokens[i].type === "Punctuator" &&
            (tokens[i].value === "+" || tokens[i].value === "-" || tokens[i].value === "*" || tokens[i].value === "/" || tokens[i].value === "%" ||
                tokens[i].value === "+=" || tokens[i].value === "-=" || tokens[i].value === "*=" || tokens[i].value === "/=" || tokens[i].value === "%=" ||
                tokens[i].value === "++" || tokens[i].value === "--")) {
            throw new Error("use of +-*/% operators is not allowed");
        }
    }
}

function processContract(source) {
  let ast = esprima.parseModule(source, {
		range: true,
		loc: true,
		tokens: true
	});

	let abiArr = [];
	if (!ast || ast === null || !ast.body || ast.body === null || ast.body.length === 0) {
		console.error("invalid source! ast = " + ast);
		return ["", ""];
	}

    checkInvalidKeyword(ast.tokens);
	// checkOperator(ast.tokens);

	let validRange = [];
	let className;
	for (let stat of ast.body) {
		if (isClassDecl(stat)) {
			validRange.push(stat.range);
		}
		else if (stat.type === "ExpressionStatement" && isExport(stat.expression)) {
			validRange.push(stat.range);
			className = getExportName(stat.expression);
		}
	}

	for (let stat of ast.body) {
		if (isClassDecl(stat) && stat.id.type === "Identifier" && stat.id.name === className) {
			abiArr = genAbiArr(stat);
		}
	}

	let newSource = 'use strict;\n';
	for (let r of validRange) {
		newSource += source.slice(r[0], r[1]) + "\n";
	}

	let abi = {};
	abi["lang"] = lang;
	abi["version"] = version;
	abi["abi"] = abiArr;
	let abiStr = JSON.stringify(abi, null, 4);

	return [newSource, abiStr]
}
module.exports = processContract;


let fs = require('fs');

let file = process.argv[2];
fs.readFile(file, 'utf8', function(err, contents) {
	console.log('before calling process, len = ' + contents.length);
	let [newSource, abi] = processContract(contents);
	console.log('after calling process, newSource len = ' + newSource.length + ", abi len = " + abi.length);

	fs.writeFile(file + ".after", newSource, function(err) {
    	if(err) {
    	    return console.log(err);
    	}
    	console.log("The new contract file was saved as " + file + ".after");
	});

	fs.writeFile(file + ".abi", abi, function(err) {
    	if(err) {
    	    return console.log(err);
    	}
    	console.log("The new abi file was saved as " + file + ".abi");
	});
});
