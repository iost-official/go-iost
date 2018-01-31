package iost

import (
	"github.com/golang-collections/collections/stack"
	"strings"
	"fmt"
)

type BtcScript struct {
	myStack * stack.Stack
}

func NewBtcScriptor () BtcScript {
	var bb BtcScript
	bb.myStack = stack.New()
	return bb
}

func (bs *BtcScript) Run(s string) (result string, err error) {
	for _, str := range strings.Fields(s) {
		err = bs.cmd(str)
		if err != nil {
			result = ""
			return
		}
	}
	result, _ = bs.myStack.Pop().(string)
	return
}

func (bs *BtcScript) cmd(s string) error {
	switch s {
	case "DUP":
		if bs.myStack.Len() < 1 {
			return fmt.Errorf("SyntaxError")
		}
		bs.myStack.Push(bs.myStack.Peek())
	case "HASH160":
		if bs.myStack.Len() < 1 {
			return fmt.Errorf("SyntaxError")
		}
		word, _ := bs.myStack.Pop().(string)
		cmdBinary := BinaryFromBase58(word).Hash160()
		bs.myStack.Push(cmdBinary.ToBase58())
	case "EQUALVERIFY":
		if bs.myStack.Len() < 2 {
			return fmt.Errorf("SyntaxError")
		}
		word1, _ := bs.myStack.Pop().(string)
		word2, _ := bs.myStack.Pop().(string)
		if word1 != word2 {
			return fmt.Errorf("EQUALVERIFY_FAILED")
		}
	case "CHECKSIG":
		if bs.myStack.Len() < 2 {
			return fmt.Errorf("SyntaxError")
		}
		pubkey, _ := bs.myStack.Pop().(string)
		sig, _ := bs.myStack.Pop().(string)
		pubkeyBinary := BinaryFromBase58(pubkey)
		sigBinary := BinaryFromBase58(sig)
		if VerifySignature(pubkeyBinary.Bytes(), pubkeyBinary.Bytes(), sigBinary.Bytes()) {
			bs.myStack.Push("TRUE")
		} else {
			bs.myStack.Push("FALSE")
		}

	default:
		bs.myStack.Push(s)
	}

	return nil
}

func BtcLockScript(format string, addr string) (script string, err error) {
	err = nil
	switch format {
	case "P2PKH":
		script = fmt.Sprintf("DUP HASH160 %v EQUALVERIFY CHECKSIG", addr)
		return
	default:
		script = ""
		err = fmt.Errorf("FormatNotExist")
		return
	}

}