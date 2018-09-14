package blockcache

import (
	"encoding/json"
	"errors"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
	"strconv"
	"sync"
)

type WitnessList struct {
	activeWitnessList    []string
	pendingWitnessList   []string
	pendingWitnessNumber int64
	witnessInfo          sync.Map
}

type witInfo struct {
	Loc    string `json:"loc"`
	Url    string `json:"url"`
	NetId  string `json:"netId"`
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

// SetPendingNum get block number of pending witness
func (wl *WitnessList) PendingNum() int64 {
	return wl.pendingWitnessNumber
}

// NetId get net id
func (wl *WitnessList) NetId() []string {
	r := make([]string, 0)
	wl.witnessInfo.Range(func(key, value interface{}) bool {
		if value == nil {
			return true
		}
		r = append(r, value.(*witInfo).NetId)
		return true
	})
	return r
}

// UpdatePending update pending witness list
func (wl *WitnessList) UpdatePending(mv db.MVCCDB) error {

	vi := database.NewVisitor(0, mv)
	spn := database.MustUnmarshal(vi.Get("iost.vote-" + "pendingBlockNumber"))
	if spn == nil {
		return errors.New("failed to get pending number")
	}
	pn, err := strconv.ParseInt(spn.(string), 10, 64)
	if err != nil {
		return err
	}
	wl.SetPendingNum(pn)

	jwl := database.MustUnmarshal(vi.Get("iost.vote-" + "pendingProducerList"))
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

// UpdatePending update pending witness list
func (wl *WitnessList) UpdateInfo(mv db.MVCCDB) error {

	wl.witnessInfo.Range(func(key, value interface{}) bool {
		wl.witnessInfo.Delete(key)
		return true
	})
	vi := database.NewVisitor(0, mv)
	for _, v := range wl.pendingWitnessList {
		jwl := database.MustUnmarshal(vi.MGet("iost.vote-producerTable", v))
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

func (wl *WitnessList) LibWitnessHandle() {
	wl.SetActive(wl.Pending())
}

func (wl *WitnessList) CopyWitness(n *BlockCacheNode) {
	if n == nil {
		return
	}
	wl.SetActive(n.Active())
	wl.SetPending(n.Pending())
	wl.SetPendingNum(n.PendingNum())
}
