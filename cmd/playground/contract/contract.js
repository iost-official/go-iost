'use strict';

var esprima = require('esprima/dist/esprima.js');

var lang = "js";
var version = "1.0.0";

function isFunctionDecl(stat) {
	if (!stat || stat === null) {
		return false;
	}
	if (stat.type === "FunctionDeclaration") {
		return true;
	}
	return false;
}

function isValid(stat) {
	console.log("is valid stat = " + stat);
	if (isFunctionDecl(stat)) {
		return true;
	}
	if (stat.type === "ExpressionStatement" && stat.expression.type === "CallExpression" &&
			stat.expression.callee.type === "Identifier" && stat.expression.callee.name === "require") {
		return true;
	}
	return false;
}

function genAbi(stat) {
	if (!isFunctionDecl(stat)) {
		console.error("invalid statment for generate abi. stat = " + stat);
		return null;
	}
	var abi = {
		"name": stat.id.name,
		"args": new Array(stat.params.length).fill("string"),
		"payment": 0,
		"cost_limit": new Array(stat.params.length).fill(1),
		"price_limit": 1
	}
	return abi;
}

function processContract(source) {
	var newSource, abi;
    var ast = esprima.parseScript(source, {
		range: true,
		loc: true,
		tokens: true
	});
	console.log(ast);

	var abiArr = [];
	if (!ast || ast === null || !ast.body || ast.body === null || ast.body.length === 0) {
		console.error("invalid source! ast = " + ast);
	}
	var validRange = [];
	for (var i in ast.body) {
		var stat = ast.body[i];
		if (isValid(stat)) {
			validRange.push(stat.range);
			if (isFunctionDecl(stat)) {
				abiArr.push(genAbi(stat));
			}
		}
	}

	newSource = 'use strict;\n';
	for (var i in validRange) {
		var r = validRange[i];
    	newSource += source.slice(r[0], r[1]) + "\n";
	}

	var abi = {};
	abi["lang"] = lang;
	abi["version"] = version;
	abi["abi"] = abiArr;
	var abiStr = JSON.stringify(abi, null, 4); 

	return [newSource, abiStr]
}


var fs = require('fs');

var file = process.argv[2]
fs.readFile(file, 'utf8', function(err, contents) {
	console.log('before calling process, len = ' + contents.length);
	var [newSource, abi] = processContract(contents)
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
