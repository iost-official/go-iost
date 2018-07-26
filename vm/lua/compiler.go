package lua

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/iost-official/prototype/vm"
)

var (
	// compile errors
	ErrNoMain      = errors.New("parse failed: no main function")
	ErrIllegalCode = errors.New("parse failed: non function code included")
	ErrNoGasPrice  = errors.New("parse failed: no gas price")
	ErrNoGasLimit  = errors.New("parse failed: no gas limit")
	ErrNoParamCnt  = errors.New("parse failed: param count not given")
	ErrNoRtnCnt    = errors.New("parse failed: return count not given")
)

// DocCommentParser parser of codes
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
