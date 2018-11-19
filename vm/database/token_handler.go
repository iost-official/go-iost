package database

import (
	"encoding/json"
	"errors"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/ilog"
)

// TokenContractName name of basic token contract
const TokenContractName = "token.iost"

// TokenHandler easy to get balance of token.iost
type TokenHandler struct {
	db database
}

// FreezeItem represents freezed balance, will unfreeze after Ftime
type FreezeItem struct {
	Amount int64
	Ftime  int64
}

// FreezeItemFixed ...
type FreezeItemFixed struct {
	Amount common.Fixed
	Ftime  int64
}

func (m *TokenHandler) balanceKey(tokenName, acc string) string {
	return "m-" + TokenContractName + "@" + acc + "-" + "TB" + "-" + tokenName
}

func (m *TokenHandler) freezedBalanceKey(tokenName, acc string) string {
	return "m-" + TokenContractName + "@" + acc + "-" + "TF" + "-" + tokenName
}

func (m *TokenHandler) decimalKey(tokenName string) string {
	issuerKey := "m-" + TokenContractName + "-" + "TI" + tokenName + "-" + "issuer"
	issuer := Unmarshal(m.db.Get(issuerKey))
	if issuer == nil {
		return ""
	}
	key := "m-" + TokenContractName + "@" + issuer.(string) + "-" + "TI" + tokenName + "-" + "decimal"
	return key
}

// TokenBalance get token balance of acc
func (m *TokenHandler) TokenBalance(tokenName, acc string) int64 {
	currentRaw := m.db.Get(m.balanceKey(tokenName, acc))
	balance := Unmarshal(currentRaw)
	ib, ok := balance.(int64)
	if !ok {
		ib = 0
	}
	return ib
}

// TokenBalanceFixed get token balance of acc
func (m *TokenHandler) TokenBalanceFixed(tokenName, acc string) *common.Fixed {
	ib := m.TokenBalance(tokenName, acc)
	return &common.Fixed{Value: ib, Decimal: m.Decimal(tokenName)}
}

// FreezedTokenBalance get freezed token balance of acc
func (m *TokenHandler) FreezedTokenBalance(tokenName, acc string) int64 {
	freezeJSON := Unmarshal(m.db.Get(m.freezedBalanceKey(tokenName, acc)))
	if freezeJSON == nil {
		return 0
	}
	freezeList := []FreezeItem{}
	//fmt.Println(string(freezeJSON.(SerializedJSON)))
	err := json.Unmarshal([]byte(freezeJSON.(SerializedJSON)), &freezeList)
	if err != nil {
		return 0
	}
	ilog.Debugf("FreezedTokenBalance is %v %v %v", tokenName, acc, []byte(freezeJSON.(SerializedJSON)))

	ib := int64(0)
	for _, item := range freezeList {
		ib += item.Amount
	}
	return ib
}

// AllFreezedTokenBalance get freezed token balance of acc
func (m *TokenHandler) AllFreezedTokenBalance(tokenName, acc string) []FreezeItem {
	freezeList := make([]FreezeItem, 0)
	freezeJSON := Unmarshal(m.db.Get(m.freezedBalanceKey(tokenName, acc)))
	if freezeJSON == nil {
		return freezeList
	}
	//fmt.Println(string(freezeJSON.(SerializedJSON)))
	err := json.Unmarshal([]byte(freezeJSON.(SerializedJSON)), &freezeList)
	if err != nil {
		ilog.Errorf("frozen token balance is invalid json %v %v", string(freezeJSON.(SerializedJSON)), err)
		return freezeList
	}
	ilog.Debugf("FreezedTokenBalance is %v %v %v", tokenName, acc, string(freezeJSON.(SerializedJSON)))
	return freezeList
}

// AllFreezedTokenBalanceFixed get freezed token balance of acc
func (m *TokenHandler) AllFreezedTokenBalanceFixed(tokenName, acc string) []FreezeItemFixed {
	freezeList := m.AllFreezedTokenBalance(tokenName, acc)
	result := make([]FreezeItemFixed, 0)
	for _, item := range freezeList {
		result = append(result, FreezeItemFixed{Amount: common.Fixed{Value: item.Amount, Decimal: m.Decimal(tokenName)}, Ftime: item.Ftime})
	}
	return result
}

// FreezedTokenBalanceFixed get token balance of acc
func (m *TokenHandler) FreezedTokenBalanceFixed(tokenName, acc string) *common.Fixed {
	ib := m.FreezedTokenBalance(tokenName, acc)
	return &common.Fixed{Value: ib, Decimal: m.Decimal(tokenName)}
}

// SetTokenBalance set token balance of acc, used for test
func (m *TokenHandler) SetTokenBalance(tokenName, acc string, amount int64) {
	m.db.Put(m.balanceKey(tokenName, acc), MustMarshal(amount))
}

// SetTokenBalanceFixed set token balance of acc, used for test
func (m *TokenHandler) SetTokenBalanceFixed(tokenName, acc string, amountStr string) {
	amountNumber, err := common.NewFixed(amountStr, m.Decimal(tokenName))
	if err != nil {
		panic(errors.New("construct Fixed number failed. str = " + amountStr + ", decimal = " + string(m.Decimal(tokenName))))
	}
	m.db.Put(m.balanceKey(tokenName, acc), MustMarshal(amountNumber.Value))
}

// Decimal get decimal in token info
func (m *TokenHandler) Decimal(tokenName string) int {
	decimalRaw := m.db.Get(m.decimalKey(tokenName))
	decimal, ok := Unmarshal(decimalRaw).(int64)
	if !ok {
		decimal = -1
	}
	return int(decimal)
}
