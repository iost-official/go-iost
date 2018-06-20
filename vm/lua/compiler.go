package lua

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/iost-official/prototype/vm"
)

var (
	// ErrNoMain 代码中未包含main函数
	ErrNoMain = errors.New("parse failed: no main function")
	// ErrIllegalCode 代码中包含\\0字符
	ErrIllegalCode = errors.New("parse failed: non function code included")
	ErrNoGasPrice  = errors.New("parse failed: no gas price")
	ErrNoGasLimit  = errors.New("parse failed: no gas limit")
	// 代码没指定输入参数数量
	ErrNoParamCnt = errors.New("parse failed: param count not given")
	// 代码没指定返回参数数量
	ErrNoRtnCnt = errors.New("parse failed: return count not given")
)

// DocCommentParser 装入text之后调用parse即可得到contract
type DocCommentParser struct {
	Debug bool
	text  string // always ends with \0, which doesn't appear elsewhere

	index int
}

func NewDocCommentParser(text string) (*DocCommentParser, error) {
	parser := new(DocCommentParser)
	strings.Replace(text, "\\0", "", -1)
	parser.text = text
	parser.Debug = false
	return parser, nil
}

//func (p *DocCommentParser) Parse() (*Contract, error) {
//	//0. preprocess
//	//
//	//1. checking exsistence of main function
//	//2. detecting all functions and split them.
//	//3. Parse doccomment for each function.
//	//4. return contract
//
//	// 没有doc comment的代码将被忽略
//
//	content := p.text
//
//	re := regexp.MustCompile("(--- .*\n)(-- .*\n)*function(.*\n)*?end--f") //匹配代码块
//
//	hasMain := false
//	var contract Contract
//
//	var buffer bytes.Buffer
//
//	// 匹配关键字
//	gasRe := regexp.MustCompile("@gas_limit (\\d+)")
//	priceRe := regexp.MustCompile("@gas_price ([+-]?([0-9]*[.])?[0-9]+)")
//	publisherRe := regexp.MustCompile("@publisher ([a-zA-Z1-9]+)")
//
//	gas0, ok := optionalMatchOne(gasRe, content)
//	if !ok {
//		return nil, errors.New("gas undefined")
//	}
//	gas, _ := strconv.ParseInt(gas0, 10, 64)
//
//	price0, ok := optionalMatchOne(priceRe, content)
//	if !ok {
//		return nil, errors.New("price undefined")
//	}
//	price, _ := strconv.ParseFloat(price0, 64)
//	if p.Debug {
//		match, ok := optionalMatchOne(publisherRe, content)
//		if !ok {
//			return nil, errors.New("publisher undefined")
//		}
//		contract.info.Publisher = vm.IOSTAccount(match)
//	}
//
//	contract.apis = make(map[string]Method)
//
//	for _, submatches := range re.FindAllStringSubmatchIndex(content, -1) {
//
//		/*
//			--- <functionName>  summary
//			-- some description
//			-- ...
//			-- ...
//			-- @gas_limit <gasLimit>
//			-- @gas_price <gasPrice>
//			-- @param_cnt <paramCnt>
//			-- @return_cnt <returnCnt>
//		*/
//
//		funcNameRe := regexp.MustCompile("---[ \t\n]*([a-zA-Z0-9_]+)")
//		funcNameR := funcNameRe.FindStringSubmatch(content[submatches[0]:submatches[1]])
//		if len(funcNameR) < 1 {
//			return nil, errors.New("syntax error, function name not found")
//		}
//		funcName := funcNameR[1]
//
//		inputCountRe := regexp.MustCompile("@param_cnt (\\d+)")
//		rtnCountRe := regexp.MustCompile("@return_cnt (\\d+)")
//		privRe := regexp.MustCompile("@privilege ([a-z]+)")
//
//		var inputCount, rtnCount int
//		var privS string
//
//		ics := inputCountRe.FindStringSubmatch(content[submatches[0]:submatches[1]])
//		if len(ics) < 1 {
//			return nil, fmt.Errorf("function %v: input count not given", funcName)
//		}
//		rcs := rtnCountRe.FindStringSubmatch(content[submatches[0]:submatches[1]])
//		if len(rcs) < 1 {
//			return nil, fmt.Errorf("function %v: return count not given", funcName)
//		}
//		ps := privRe.FindStringSubmatch(content[submatches[0]:submatches[1]])
//		if len(ps) < 1 {
//			privS = ""
//		} else {
//			privS = ps[1]
//		}
//		inputCount, _ = strconv.Atoi(ics[1])
//		rtnCount, _ = strconv.Atoi(rcs[1])
//		var priv vm.Privilege
//		switch privS {
//		case "public":
//			priv = vm.Public
//		default:
//			priv = vm.Private
//
//		}
//
//		method := NewMethod(priv, funcName, inputCount, rtnCount)
//
//		//匹配代码部分
//
//		//endRe := regexp.MustCompile("^end--f")
//		//endPos := endRe.FindStringIndex(content[submatches[1]:])
//
//		//code part: content[submatches[1]:][:endPos[1]
//
//		contract.info.Language = "lua"
//		contract.info.GasLimit = gas
//		contract.info.Price = price
//		if funcName == "main" {
//			hasMain = true
//			contract.main = method
//			//contract.code = content[submatches[1]:][:endPos[1]]
//		} else {
//
//			contract.apis[funcName] = method
//		}
//		buffer.WriteString(content[submatches[0]:submatches[1]])
//		buffer.WriteString("\n")
//
//	}
//
//	if !hasMain {
//		return nil, ErrNoMain
//	}
//	contract.code = buffer.String()
//	return &contract, nil
//
//}

func optionalMatchOne(re *regexp.Regexp, s string) (sub string, ok bool) {
	ss := re.FindStringSubmatch(s)
	if len(ss) < 1 {
		return "", false
	} else {
		sub = ss[1]
		ok = true
		return
	}
}

func matchIntOne(code, re string) (int64, error) {
	gasRe := regexp.MustCompile(re)
	gas0, ok := optionalMatchOne(gasRe, code)
	if !ok {
		return 0, errors.New("no match")
	}
	return strconv.ParseInt(gas0, 10, 64)
}

func matchFloatOne(code, re string) (float64, error) {
	priceRe := regexp.MustCompile(re)
	price0, ok := optionalMatchOne(priceRe, code)
	if !ok {
		return 0, errors.New("no match")
	}
	return strconv.ParseFloat(price0, 64)
}

func matchGasPrice(code string) (gasPrice float64, err error) {
	return matchFloatOne(code, "@gas_price ([+]?([0-9]*[.])?[0-9]+)")
}

func matchGasLimit(code string) (gasLimit int64, err error) {
	return matchIntOne(code, "@gas_limit (\\d+)")
}

func matchPublisher(code string) (publisher string, err error) {
	publisherRe := regexp.MustCompile("@publisher ([a-zA-Z1-9]+)")
	match, ok := optionalMatchOne(publisherRe, code)
	if !ok {
		return "", errors.New("publisher undefined")
	}
	return match, nil
}

func matchPriv(code string) vm.Privilege {
	privRe := regexp.MustCompile("@privilege ([a-z]+)")
	var privS string
	ps := privRe.FindStringSubmatch(code)
	if len(ps) < 1 {
		privS = ""
	} else {
		privS = ps[1]
	}

	var priv vm.Privilege
	switch privS {
	case "public":
		priv = vm.Public
	default:
		priv = vm.Private

	}
	return priv
}

func parseAPI(code string) (method *Method, err error) {
	funcNameRe := regexp.MustCompile("---[ \t\n]*([a-zA-Z0-9_]+)")
	funcNameRe2 := regexp.MustCompile("function[ \t]+([a-zA-Z0-9_]+)")
	funcName, ok := optionalMatchOne(funcNameRe, code)
	if !ok {
		return nil, errors.New("not an api")
	}
	funcName2, ok := optionalMatchOne(funcNameRe2, code)
	if !ok {
		return nil, errors.New("function implement not found")
	}
	if funcName != funcName2 {
		return nil, errors.New("function name not match with comment")
	}

	pmc, err := matchIntOne(code, "@param_cnt (\\d+)")
	if err != nil {
		return nil, err
	}
	if pmc > 256 || pmc < 0 {
		return nil, errors.New("illegal function input count")
	}
	rtnc, err := matchIntOne(code, "@return_cnt (\\d+)")
	if err != nil {
		return nil, err
	}
	if rtnc > 256 || rtnc < 0 {
		return nil, errors.New("illegal function return count")
	}
	priv := matchPriv(code)

	methodp := NewMethod(priv, funcName, int(pmc), int(rtnc))

	return &methodp, nil

}

func checkNonFunctionCode(code string) error {
	inFunc := false
	li := 0
	reComment := regexp.MustCompile("^[ \t]*--")
	reEmpty := regexp.MustCompile("^[ \t]*$")
	for _, line := range strings.Split(code, "\n") {
		//fmt.Println(line)
		li++
		if inFunc {
			if strings.HasPrefix(line, "end--f") {
				inFunc = false
				continue
			}
		} else {
			switch {
			case strings.HasPrefix(line, "function"):
				inFunc = true
				continue
			case reComment.MatchString(line):
				continue
			case reEmpty.MatchString(line):
				continue
			case line == "\\0":
				continue
			default:
				return errors.New("parse failed: non function code included. line: " + strconv.Itoa(li))
			}
		}
	}
	return nil
}

func (p *DocCommentParser) Parse() (*Contract, error) {
	content := p.text

	err := checkNonFunctionCode(content)
	if err != nil {
		return nil, err
	}

	var contract Contract
	contract.code = p.text

	hasMain := false

	gasPrice, err := matchGasPrice(content)
	if err != nil {
		return nil, err
	}

	gasLimit, err := matchGasLimit(content)
	if err != nil {
		return nil, err
	}
	var publisher string
	if p.Debug {
		publisher, err = matchPublisher(content)
		if err != nil {
			return nil, err
		}
		contract.info.Publisher = vm.IOSTAccount(publisher)
	}
	contract.info.Language = "lua"
	contract.info.GasLimit = gasLimit
	contract.info.Price = gasPrice
	contract.apis = make(map[string]Method)

	re := regexp.MustCompile("(--- .*\n)(-- .*\n)*function(.*\n)*?end--f")
	for _, sub := range re.FindAllStringSubmatchIndex(content, -1) {
		mtd, err := parseAPI(content[sub[0]:sub[1]])
		if err != nil {
			return nil, err
		}
		if mtd.name == "main" {
			contract.main = *mtd
			hasMain = true
		} else {
			contract.apis[mtd.name] = *mtd
		}
	}
	if !hasMain {
		return nil, ErrNoMain
	}
	return &contract, nil
}
