package blockcache

import (
	"strconv"
	"errors"
	"encoding/json"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/vm/database"
)

// SetPending set pending witness list
func (wl *WitnessList) SetPending(pl []string) {
	wl.PendingWitnessList = pl
}

// SetPendingNum set block number of pending witness
func (wl *WitnessList) SetPendingNum(n int64) {
	wl.PendingWitnessNumber = n
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

// PendingNum get block number of pending witness
func (wl *WitnessList) PendingNum() int64 {
	return wl.PendingWitnessNumber
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
	pbn := vi.Get("iost.vote_producer-" + "pendingBlockNumber")
	spn := database.MustUnmarshal(pbn)
	if spn == nil {
		return errors.New("failed to get pending number")
	}
	pn, err := strconv.ParseInt(spn.(string), 10, 64)
	if err != nil {
		return err
	}
	wl.SetPendingNum(pn)

	jwl := database.MustUnmarshal(vi.Get("iost.vote_producer-" + "pendingProducerList"))
	if jwl == nil {
		return errors.New("failed to get pending list")
	}
	str := make([]string, 0)
	err = json.Unmarshal([]byte(jwl.(string)), &str)
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
		jwl := database.MustUnmarshal(vi.MGet("iost.vote_producer-producerTable", v))
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

// LibWitnessHandle is set active list
func (wl *WitnessList) LibWitnessHandle() {
	wl.SetActive(wl.Pending())
}

// CopyWitness is copy witness
func (wl *WitnessList) CopyWitness(n *BlockCacheNode) {
	if n == nil {
		return
	}
	wl.SetActive(n.Active())
	wl.SetPending(n.Pending())
	wl.SetPendingNum(n.PendingNum())
}
