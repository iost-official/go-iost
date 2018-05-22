package log

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var Server = "127.0.0.1:1001"
var LocalID = "default"

func Report(msg Msg) error {
	resp, err := http.PostForm(Server+"/report",
		msg.Form())

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(strconv.Itoa(resp.StatusCode) + " " + string(body))
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

type MsgBlock struct {
	SubType       string
	BlockHeadHash string // base64
	BlockNum      uint64
}

func (m *MsgBlock) Form() url.Values {
	return url.Values{
		"from":            {LocalID},
		"time":            {Now().String()},
		"type":            {"Block", m.SubType},
		"block-head-hash": {m.BlockHeadHash},
		"block-number":    {strconv.FormatUint(m.BlockNum, 10)},
	}
}

type MsgTx struct {
	SubType   string
	TxHash    string
	Publisher string
	Nonce     int64
}

func (m *MsgTx) Form() url.Values {
	return url.Values{
		"from":      {LocalID},
		"time":      {Now().String()},
		"type":      {"Tx", m.SubType},
		"hash":      {m.TxHash},
		"publisher": {m.Publisher},
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
		num, _ := strconv.ParseUint(v["block-number"][0], 10, 64)
		return &MsgBlock{
			SubType:       v["type"][1],
			BlockHeadHash: v["block-head-hash"][0],
			BlockNum:      num,
		}
	case "Tx":
		num, _ := strconv.ParseInt(v["nonce"][0], 10, 64)
		return &MsgTx{
			SubType:   v["type"][1],
			TxHash:    v["hash"][0],
			Publisher: v["publisher"][0],
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
