package blockcache

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/core/block"
	"github.com/iost-official/go-iost/v3/ilog"

	"github.com/iost-official/go-iost/v3/core/version"
	"github.com/iost-official/go-iost/v3/vm/database"
)

// SetPending set pending witness list
func (wl *WitnessList) SetPending(pl []string) {
	wl.PendingWitnessList = pl
}

// SetActive set active witness list
func (wl *WitnessList) SetActive(al []string) {
	wl.ActiveWitnessList = al
}

// Pending get pending witness list
func (wl *WitnessList) Pending() []string {
	return wl.PendingWitnessList
}

// Active get active witness list
func (wl *WitnessList) Active() []string {
	return wl.ActiveWitnessList
}

// NetID get net id
func (wl *WitnessList) NetID() []string {
	return wl.WitnessInfo
}

// getPendingFromDB get pending witness list
func (wl *WitnessList) getPendingFromDB(mv database.IMultiValue, rules *version.Rules) ([]string, error) {
	result := make([]string, 0)
	vi := database.NewVisitor(0, mv, rules)
	jwl := database.MustUnmarshal(vi.Get("vote_producer.iost-" + "pendingProducerList"))
	if jwl == nil {
		return result, errors.New("failed to get pending list")
	}
	err := json.Unmarshal([]byte(jwl.(string)), &result)
	return result, err
}

// UpdateInfo update pending witness list
func (wl *WitnessList) UpdateInfo(mv database.IMultiValue, rules *version.Rules) error {
	wl.WitnessInfo = make([]string, 0)
	vi := database.NewVisitor(0, mv, rules)
	for _, v := range wl.PendingWitnessList {
		iAcc := database.MustUnmarshal(vi.MGet("vote_producer.iost-producerKeyToId", v))
		if iAcc == nil {
			continue
		}
		acc := strings.Trim(iAcc.(string), "\"")
		jwl := database.MustUnmarshal(vi.MGet("vote_producer.iost-producerTable", acc))
		if jwl == nil {
			continue
		}

		var str WitnessInfo
		err := json.Unmarshal([]byte(jwl.(string)), &str)
		if err != nil {
			continue
		}
		wl.WitnessInfo = append(wl.WitnessInfo, str.NetID)
	}
	return nil
}

// CopyWitness is copy witness
func (wl *WitnessList) CopyWitness(n *BlockCacheNode) {
	if n == nil {
		return
	}
	wl.SetActive(n.Active())
	wl.SetPending(n.Pending())
}

type WitnessStatus struct {
	PendingList []string `json:"pendingList"`
	CurrentList []string `json:"currentList"`
}

func GetWitnessStatusFromBlock(b *block.Block) (*WitnessStatus, error) {
	result := &WitnessStatus{}
	for _, r := range b.Receipts {
		for _, rr := range r.Receipts {
			if rr.FuncName == "vote_producer.iost/stat" {
				err := json.Unmarshal([]byte(rr.Content), result)
				if err != nil {
					ilog.Warn("invalid vote_producer.iost/stat receipt", rr.Content, err)
					continue
				}
				return result, nil
			}
		}
	}
	ilog.Warn("vote_producer.iost/stat receipt not found at block ", b.Head.Number, ",hash:", common.Base58Encode(b.HeadHash()))
	return result, nil
}
