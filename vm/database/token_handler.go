package database

import (
	"errors"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/ilog"
)

// TokenContractName name of basic token contract
const TokenContractName = "iost.token"

// TokenHandler easy to get balance of iost.token
type TokenHandler struct {
	db database
}

func (m *TokenHandler) balanceKey(tokenName, acc string) string {
	return "m-" + TokenContractName + "-" + "TB" + acc + "-" + tokenName
}

func (m *TokenHandler) decimalKey(tokenName string) string {
	key := "m-" + TokenContractName + "-" + "TI" + tokenName + "-" + "decimal"
	return key
}

// TokenBalance get token balance of acc
func (m *TokenHandler) TokenBalance(tokenName, acc string) int64 {
	currentRaw := m.db.Get(m.balanceKey(tokenName, acc))
	ilog.Errorf("TokenBalance is %v %v %v", tokenName, acc, currentRaw)
	balance := Unmarshal(currentRaw)
	ib, ok := balance.(int64)
	if !ok {
		ib = 0
	}
	return ib
}

// TokenBalanceFixed get token balance of acc
func (m *TokenHandler) TokenBalanceFixed(tokenName, acc string) common.Fixed {
	currentRaw := m.db.Get(m.balanceKey(tokenName, acc))
	balance := Unmarshal(currentRaw)
	ib, ok := balance.(int64)
	if !ok {
		ib = 0
	}
	return common.Fixed{Value: ib, Decimal: m.Decimal(tokenName)}
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
		decimal = 0
	}
	return int(decimal)
}
