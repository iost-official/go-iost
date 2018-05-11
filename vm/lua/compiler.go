package lua

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

type DocCommentParser struct {
	text string // always ends with \0, which doesn't appear elsewhere

	index int
}

func NewDocCommentParser(text string) (*DocCommentParser, error) {
	parser := new(DocCommentParser)
	if strings.Contains(text, "\\0") {
		return nil, errors.New("Text contains character \\0, parse failed")
	}
	parser.text = text + "\\0"
	return parser, nil
}

func (p *DocCommentParser) Parse() (*Contract, error) {
	//0. preprocess
	//
	//1. checking exsistence of main function
	//2. detecting all functions and split them.
	//3. Parse doccomment for each function.
	//4. return contract

	// 没有doc comment的代码将被忽略

	content := p.text

	re := regexp.MustCompile("--- .*\n(-- .*\n)*") //匹配全部注释代码

	hasMain := false
	var contract Contract

	var buffer bytes.Buffer

	for _, submatches := range re.FindAllStringSubmatchIndex(content, -1) {
		/*
			--- <functionName>  summary
			-- some description
			-- ...
			-- ...
			-- @gas_limit <gasLimit>
			-- @gas_price <gasPrice>
			-- @param_cnt <paramCnt>
			-- @return_cnt <returnCnt>
		*/
		funcName := strings.Split(content[submatches[0]:submatches[1]], " ")[1]

		inputCountRe := regexp.MustCompile("@param_cnt (\\d+)")
		rtnCountRe := regexp.MustCompile("@return_cnt (\\d+)")

		inputCount, _ := strconv.Atoi(inputCountRe.FindStringSubmatch(content[submatches[0]:submatches[1]])[1])
		rtnCount, _ := strconv.Atoi(rtnCountRe.FindStringSubmatch(content[submatches[0]:submatches[1]])[1])
		method := NewMethod(funcName, inputCount, rtnCount)

		//匹配代码部分

		endRe := regexp.MustCompile("end")
		endPos := endRe.FindStringIndex(content[submatches[1]:])

		//code part: content[submatches[1]:][:endPos[1]
		contract.apis = make(map[string]Method)
		if funcName == "main" {
			hasMain = true
			gasRe := regexp.MustCompile("@gas_limit (\\d+)")
			priceRe := regexp.MustCompile("@gas_price ([+-]?([0-9]*[.])?[0-9]+)")

			gas, _ := strconv.ParseInt(gasRe.FindStringSubmatch(content[submatches[0]:submatches[1]])[1], 10, 64)
			price, _ := strconv.ParseFloat(priceRe.FindStringSubmatch(content[submatches[0]:submatches[1]])[1], 64)
			contract.info.Language = "lua"
			contract.info.GasLimit = gas
			contract.info.Price = price
			contract.main = method
			//contract.code = content[submatches[1]:][:endPos[1]]
		} else {

			contract.apis[funcName] = method
		}
		buffer.WriteString(content[submatches[1]:][:endPos[1]])
		buffer.WriteString("\n")

	}

	if !hasMain {
		return nil, errors.New("No main function!, parse failed")
	}
	contract.code = buffer.String()
	return &contract, nil

}
