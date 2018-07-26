package log

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var Server = "http://18.191.77.190:30333"
var LocalID = "default"

func toBase64(hash []byte) string {
	return base64.StdEncoding.EncodeToString(hash)
}

func fromBase64(str string) []byte {
	ret, _ := base64.StdEncoding.DecodeString(str)
	return ret
}

func Report(msg Msg) error {
	return nil
}

func report(msg Msg) error {
	resp, err := http.PostForm(Server+"/report", msg.Form())
	defer func() {
		if err != nil {
			log, err2 := NewLogger("NET_LOG")
			if err2 != nil {
				fmt.Println("error :", err2)
			}
			log.E(err.Error())
		}
	}()

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		err = errors.New(strconv.Itoa(resp.StatusCode) + " " + string(body))
		return err
	}
	return nil
}

type Timestamp struct {
	second int64
	nano   int
}

func (t Timestamp) String() string {
	return strconv.FormatInt(t.second, 10) + "+" + strconv.Itoa(t.nano)
}

func Now() Timestamp {
	now := time.Now()
	return Timestamp{
		second: now.Unix(),
		nano:   now.Nanosecond(),
	}
}

type Msg interface {
	Form() url.Values
}

var Subtypes map[string][]string = map[string][]string{
	"MsgBlock": []string{"confirm", "verify.pass", "verify.fail", "receive", "send"},
	"MsgTx":    []string{"confirm", "verify.pass", "verify.fail"},
	"MsgNode":  []string{"online", "offline", "conn.fail", "conn.succ", "conn.count"},
}

type MsgBlock struct {
	SubType       string
	BlockHeadHash []byte
	BlockNum      int64
}

func (m *MsgBlock) Form() url.Values {
	return url.Values{
		"from":            {LocalID},
		"time":            {Now().String()},
		"type":            {"Block", m.SubType},
		"block_head_hash": {toBase64(m.BlockHeadHash)},
		"block_number":    {strconv.FormatInt(m.BlockNum, 10)},
	}
}

type MsgTx struct {
	SubType   string
	TxHash    []byte
	Publisher []byte
	Nonce     int64
}

func (m *MsgTx) Form() url.Values {
	return url.Values{
		"from":      {LocalID},
		"time":      {Now().String()},
		"type":      {"Tx", m.SubType},
		"hash":      {toBase64(m.TxHash)},
		"publisher": {toBase64(m.Publisher)},
		"nonce":     {strconv.FormatInt(m.Nonce, 10)},
	}
}

type MsgNode struct {
	SubType string
	Log     string
}

func (m *MsgNode) Form() url.Values {
	return url.Values{
		"from": {LocalID},
		"time": {Now().String()},
		"type": {"Node", m.SubType},
		"log":  {m.Log},
	}
}

var BlockSubType = []string{"receive"}

func ParseMsg(v url.Values) Msg {
	switch v["type"][0] {
	case "Block":
		num, _ := strconv.ParseInt(v["block_number"][0], 10, 64)
		return &MsgBlock{
			SubType:       v["type"][1],
			BlockHeadHash: fromBase64(v["block_head_hash"][0]),
			BlockNum:      num,
		}
	case "Tx":
		num, _ := strconv.ParseInt(v["nonce"][0], 10, 64)
		return &MsgTx{
			SubType:   v["type"][1],
			TxHash:    fromBase64(v["hash"][0]),
			Publisher: fromBase64(v["publisher"][0]),
			Nonce:     num,
		}
	case "Node":
		return &MsgNode{
			SubType: v["type"][1],
			Log:     v["log"][0],
		}
	}
	return nil
}
