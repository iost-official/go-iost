package blockcache

import (
	"encoding/json"
	"errors"
	"strconv"
	"sync"

	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/vm/database"
)

// WitnessList is the implementation of WitnessList
type WitnessList struct {
	activeWitnessList    []string
	pendingWitnessList   []string
	pendingWitnessNumber int64
	witnessInfo          sync.Map
}

type witInfo struct {
	Loc    string `json:"loc"`
	URL    string `json:"url"`
	NetID  string `json:"netId"`
	Online bool   `json:"online"`
	Score  int64  `json:"score"`
	Votes  int64  `json:"votes"`
}

// SetPending set pending witness list
func (wl *WitnessList) SetPending(pl []string) {
	wl.pendingWitnessList = pl
}

// SetPendingNum set block number of pending witness
func (wl *WitnessList) SetPendingNum(n int64) {
	wl.pendingWitnessNumber = n
}

// SetActive set active witness list
func (wl *WitnessList) SetActive(al []string) {
	wl.activeWitnessList = al
}

// Pending get pending witness list
func (wl *WitnessList) Pending() []string {
	return wl.pendingWitnessList
}

// Active get active witness list
func (wl *WitnessList) Active() []string {
	return wl.activeWitnessList
}

// PendingNum get block number of pending witness
func (wl *WitnessList) PendingNum() int64 {
	return wl.pendingWitnessNumber
}

// NetID get net id
func (wl *WitnessList) NetID() []string {
	r := make([]string, 0)
	wl.witnessInfo.Range(func(key, value interface{}) bool {
		if value == nil {
			return true
		}
		r = append(r, value.(*witInfo).NetID)
		return true
	})
	return r
}

// UpdatePending update pending witness list
func (wl *WitnessList) UpdatePending(mv db.MVCCDB) error {

	vi := database.NewVisitor(0, mv)
	pbn := vi.Get("vote_producer.iost-" + "pendingBlockNumber")
	spn := database.MustUnmarshal(pbn)
	if spn == nil {
		return errors.New("failed to get pending number")
	}
	pn, err := strconv.ParseInt(spn.(string), 10, 64)
	if err != nil {
		return err
	}
	wl.SetPendingNum(pn)

	jwl := database.MustUnmarshal(vi.Get("vote_producer.iost-" + "pendingProducerList"))
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

	wl.witnessInfo.Range(func(key, value interface{}) bool {
		wl.witnessInfo.Delete(key)
		return true
	})
	vi := database.NewVisitor(0, mv)
	for _, v := range wl.pendingWitnessList {
		jwl := database.MustUnmarshal(vi.MGet("vote_producer.iost-producerTable", v))
		if jwl == nil {
			continue
		}

		var str witInfo
		err := json.Unmarshal([]byte(jwl.(string)), &str)
		if err != nil {
			continue
		}
		wl.witnessInfo.Store(v, &str)
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
