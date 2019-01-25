package blockcache

import (
	"encoding/json"
	"errors"

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
	r := make([]string, 0)
	for _, value := range wl.WitnessInfo {
		if value != nil {
			r = append(r, value.NetID)
		}
	}
	return r
}

// UpdatePending update pending witness list
func (wl *WitnessList) UpdatePending(mv db.MVCCDB) error {

	vi := database.NewVisitor(0, mv)

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
func (wl *WitnessList) UpdateInfo(mv db.MVCCDB) error {

	for key := range wl.WitnessInfo {
		delete(wl.WitnessInfo, key)
	}
	vi := database.NewVisitor(0, mv)
	for _, v := range wl.PendingWitnessList {
		jwl := database.MustUnmarshal(vi.MGet("vote_producer.iost-producerTable", v))
		if jwl == nil {
			continue
		}

		var str WitnessInfo
		err := json.Unmarshal([]byte(jwl.(string)), &str)
		if err != nil {
			continue
		}
		wl.WitnessInfo[v] = &str
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
