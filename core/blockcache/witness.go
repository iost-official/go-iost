package blockcache

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/iost-official/go-iost/core/version"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/vm/database"
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

// UpdatePending update pending witness list
func (wl *WitnessList) UpdatePending(mv db.MVCCDB, rules *version.Rules) error {

	vi := database.NewVisitor(0, mv, rules)

	jwl := database.MustUnmarshal(vi.Get("vote_producer.iost-" + "pendingProducerList"))
	if jwl == nil {
		return errors.New("failed to get pending list")
	}
	str := make([]string, 0)
	err := json.Unmarshal([]byte(jwl.(string)), &str)
	if err != nil {
		return err
	}
	wl.SetPending(str)

	return nil
}

// UpdateInfo update pending witness list
func (wl *WitnessList) UpdateInfo(mv db.MVCCDB, rules *version.Rules) error {

	wl.WitnessInfo = make([]string, 0, 0)
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
