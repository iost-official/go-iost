package lua

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/iost-official/prototype/vm"
)

var (
	// ErrNoMain 代码中未包含main函数
	ErrNoMain = errors.New("parse failed: no main function")
	// ErrIllegalCode 代码中包含\\0字符
	ErrIllegalCode = errors.New("parse failed: Text contains character \\0")
	// 代码没指定输入参数数量
	ErrNoParamCnt = errors.New("parse failed: param count not given \\0")
	// 代码没指定返回参数数量
	ErrNoRtnCnt = errors.New("parse failed: return count not given \\0")
)

// DocCommentParser 装入text之后调用parse即可得到contract
type DocCommentParser struct {
	Debug bool
	text  string // always ends with \0, which doesn't appear elsewhere

	index int
}

func NewDocCommentParser(text string) (*DocCommentParser, error) {
	parser := new(DocCommentParser)
	if strings.Contains(text, "\\0") {
		return nil, ErrIllegalCode
	}
	parser.text = text + "\\0"
	parser.Debug = false
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
		privRe := regexp.MustCompile("@privilege ([a-z]+)")

		var inputCount, rtnCount int
		var privS string

		ics := inputCountRe.FindStringSubmatch(content[submatches[0]:submatches[1]])
		if len(ics) < 1 {
			return nil, ErrNoParamCnt
		}
		rcs := rtnCountRe.FindStringSubmatch(content[submatches[0]:submatches[1]])
		if len(rcs) < 1 {
			return nil, ErrNoRtnCnt
		}
		ps := privRe.FindStringSubmatch(content[submatches[0]:submatches[1]])
		if len(ps) < 1 {
			privS = ""
		} else {
			privS = ps[1]
		}
		inputCount, _ = strconv.Atoi(ics[1])
		rtnCount, _ = strconv.Atoi(rcs[1])
		var priv vm.Privilege
		switch privS {
		case "public":
			priv = vm.Public
		default:
			priv = vm.Private

		}

		method := NewMethod(priv, funcName, inputCount, rtnCount) // TODO 从代码中获取权限信息

		//匹配代码部分

		endRe := regexp.MustCompile("end")
		endPos := endRe.FindStringIndex(content[submatches[1]:])

		//code part: content[submatches[1]:][:endPos[1]
		contract.apis = make(map[string]Method)
		if funcName == "main" {
			hasMain = true
			// 匹配关键字
			gasRe := regexp.MustCompile("@gas_limit (\\d+)")
			priceRe := regexp.MustCompile("@gas_price ([+-]?([0-9]*[.])?[0-9]+)")
			publisherRe := regexp.MustCompile("@publisher ([a-zA-Z1-9]+)")

			gas, _ := strconv.ParseInt(gasRe.FindStringSubmatch(content[submatches[0]:submatches[1]])[1], 10, 64)
			price, _ := strconv.ParseFloat(priceRe.FindStringSubmatch(content[submatches[0]:submatches[1]])[1], 64)
			if p.Debug {
				match := publisherRe.FindStringSubmatch(content[submatches[0]:submatches[1]])
				fmt.Println("compile publisher:", match[1])
				contract.info.Publisher = vm.IOSTAccount(match[1])
			}
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
		return nil, ErrNoMain
	}
	contract.code = buffer.String()
	return &contract, nil

}
