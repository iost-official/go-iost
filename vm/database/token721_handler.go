package database

import (
	"fmt"
	"strings"
)

// Token721ContractName name of basic token contract
const Token721ContractName = "token721.iost"

// Token721Handler easy to get balance of token.iost
type Token721Handler struct {
	db database
}

func (m *Token721Handler) balanceKey(tokenName, acc string) string {
	return "m-" + Token721ContractName + "-" + "T721B" + acc + "-" + tokenName
}
func (m *Token721Handler) idKey(tokenName, acc string) string {
	return "m-" + Token721ContractName + "-" + "T721M" + tokenName + "#" + acc
}
func (m *Token721Handler) metadataKey(tokenName, acc, tokenID string) string {
	return "m-" + Token721ContractName + "-" + "T721M" + tokenName + "#" + acc + "-" + tokenID
}
func (m *Token721Handler) ownerKey(tokenName, tokenID string) string {
	return "m-" + Token721ContractName + "-" + "T721I" + tokenName + "-" + tokenID
}

// Token721Balance get token balance of acc
func (m *Token721Handler) Token721Balance(tokenName, acc string) int64 {
	currentRaw := m.db.Get(m.balanceKey(tokenName, acc))
	balance := Unmarshal(currentRaw)
	ib, ok := balance.(int64)
	if !ok {
		ib = 0
	}
	return ib
}

// Token721IDList get token balance of acc
func (m *Token721Handler) Token721IDList(tokenName, acc string) []string {
	ids := m.db.Get(m.idKey(tokenName, acc))
	if len(ids) == 0 {
		return []string{}
	}
	// adapt new format of map keys
	if strings.HasPrefix(ids, "@@") {
		return strings.Split(ids, "@")[2:]
	}
	return strings.Split(ids, "@")[1:]
}

// Token721Metadata get token balance of acc
func (m *Token721Handler) Token721Metadata(tokenName, tokenID string) (string, error) {
	owner, err := m.Token721Owner(tokenName, tokenID)
	if err != nil {
		return "", err
	}
	currentRaw := m.db.Get(m.metadataKey(tokenName, owner, tokenID))
	metadata, ok := Unmarshal(currentRaw).(string)
	if !ok {
		return "", nil
	}
	return metadata, nil
}

// Token721Owner get token owner of tokenID
func (m *Token721Handler) Token721Owner(tokenName, tokenID string) (string, error) {
	currentRaw := m.db.Get(m.ownerKey(tokenName, tokenID))
	owner, ok := Unmarshal(currentRaw).(string)
	if !ok || owner == "" {
		return "", fmt.Errorf("token %v %v not found", tokenName, tokenID)
	}
	return owner, nil
}
