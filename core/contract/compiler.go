package contract

import (
	"encoding/json"
)

// Compiler parse contract from json string
type Compiler struct {
}

func (c *Compiler) parseInfo(infoStr string) (*Info, error) {
	var info *Info
	err := json.Unmarshal([]byte(infoStr), &info)
	return info, err
}

// Parse parse contract from json abi string, set code and id
func (c *Compiler) Parse(id, code, abi string) (*Contract, error) {
	info, err := c.parseInfo(abi)
	if err != nil {
		return nil, err
	}

	return &Contract{
		Info: info,
		Code: code,
		ID:   id,
	}, nil
}
