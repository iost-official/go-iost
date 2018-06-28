package tx

import (
	"math"

	"encoding/binary"

	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/log"
	"github.com/iost-official/prototype/vm"
	"sort"
	"sync"
)

const base float64 = 1.0

var (
	bestUser = []byte("bestUser")
)

type Servi struct {
	v     float64 // behavior
	b     float64 // balance
	owner vm.IOSTAccount
}

func (s *Servi) IncrBehavior(time int) {
	s.v += base * float64(time)
}

func (s *Servi) SetBalance(b float64) {
	s.b = b
}

func (s *Servi) Total() float64 {
	return s.b + s.v
}

func (s *Servi) Owner() vm.IOSTAccount {
	return s.owner
}

func (s *Servi) Clear() {
	s.v = s.v * 0.9
	s.v = math.Floor(s.v)
}

type ServiPool struct {
	btu       map[vm.IOSTAccount]*Servi
	hm        map[vm.IOSTAccount]*Servi
	btuCnt    int
	cacheSize int
	mu        sync.RWMutex
}

var ldb db.Database

var StdServiPool *ServiPool
var sonce sync.Once

func NewServiPool(num int, cacheSize int) (*ServiPool, error) {

	var err error
	sonce.Do(func() {
		ldb, err = db.NewLDBDatabase(LdbPath+"serviDb", 0, 0)
		if err != nil {
			panic(err)
		}

		StdServiPool = &ServiPool{
			btu:       make(map[vm.IOSTAccount]*Servi, 0),
			hm:        make(map[vm.IOSTAccount]*Servi, 0),
			btuCnt:    num,
			cacheSize: cacheSize,
		}
	})
	return StdServiPool, nil
}

func (sp *ServiPool) User(iostAccount vm.IOSTAccount) (*Servi, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	var s *Servi

	if servi, ok := sp.btu[iostAccount]; ok {
		return servi, nil
	}

	if servi, ok := sp.hm[iostAccount]; ok {
		return servi, nil
	}

	if len(sp.btu) < sp.btuCnt {
		s = sp.userBtu(iostAccount)
	} else {

		s = sp.userHm(iostAccount)
	}

	return s, nil
}

func (sp *ServiPool) BestUser() ([]*Servi, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	var slist BestUserList
	for i, _ := range sp.btu {
		slist = append(slist, sp.btu[i])
	}

	sort.Sort(slist)

	return slist, nil
}

func (sp *ServiPool) ClearBtu() {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	for k, _ := range sp.btu {
		sp.delBtu(k)
	}

}

func (sp *ServiPool) Flush() {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	err := sp.flushHm()
	if err != nil {
		log.Log.D("Failed to ServiPool flushHm")
	}

	err = sp.flushBtu()
	if err != nil {
		log.Log.D("Failed to ServiPool flushBtu")
	}

}

func (sp *ServiPool) UpdateBtu() {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	for k, v := range sp.hm {
		for k1, v1 := range sp.btu {
			if v.Total() > v1.Total() {
				sp.delBtu(k1)
				sp.delHm(k)
				sp.addBtu(k, v)
				break
			}
		}
	}
}

func (sp *ServiPool) Restore() error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	sbuf, err := ldb.Get(bestUser)

	if err != nil {
		return err
	}

	for i := 0; i < len(sbuf); i += 49 {
		servi := Servi{}
		bufnum := sbuf[i : i+8]
		servi.v = math.Float64frombits(binary.BigEndian.Uint64(bufnum))
		bufbln := sbuf[i+8 : i+16]
		servi.b = math.Float64frombits(binary.BigEndian.Uint64(bufbln))
		bufkey := sbuf[i+16 : i+49]
		servi.owner = vm.PubkeyToIOSTAccount(bufkey)
		sp.btu[servi.owner] = &servi
	}

	return nil
}

func (sp *ServiPool) userBtu(iostAccount vm.IOSTAccount) *Servi {

	if servi, ok := sp.btu[iostAccount]; ok {
		return servi
	} else {
		sp.btu[iostAccount] = &Servi{owner: iostAccount}
		return sp.btu[iostAccount]
	}
}

func (sp *ServiPool) userHm(iostAccount vm.IOSTAccount) *Servi {

	if servi, ok := sp.hm[iostAccount]; ok {
		return servi
	} else {
		s, _ := sp.restoreHm(vm.IOSTAccount(iostAccount))
		if s == nil {
			sp.hm[iostAccount] = &Servi{owner: iostAccount}
		} else {
			sp.hm[iostAccount] = &Servi{owner: s.owner, b: s.b, v: s.v}
		}

		return sp.hm[iostAccount]
	}
}

func (sp *ServiPool) addBtu(iostAccount vm.IOSTAccount, s *Servi) error {

	if _, ok := sp.btu[iostAccount]; ok {
		delete(sp.btu, iostAccount)
	}

	sp.btu[iostAccount] = &Servi{owner: s.owner, b: s.b, v: s.v}
	return nil

}

func (sp *ServiPool) addHm(iostAccount vm.IOSTAccount, s *Servi) error {

	if _, ok := sp.hm[iostAccount]; ok {
		delete(sp.hm, iostAccount)
	}

	sp.hm[iostAccount] = &Servi{owner: s.owner, b: s.b, v: s.v}
	return nil
}

func (sp *ServiPool) delBtu(iostAccount vm.IOSTAccount) error {

	if _, ok := sp.btu[iostAccount]; ok {
		delete(sp.btu, iostAccount)
	}

	return nil
}

func (sp *ServiPool) delHm(iostAccount vm.IOSTAccount) error {

	if _, ok := sp.hm[iostAccount]; ok {
		delete(sp.hm, iostAccount)
		err := ldb.Delete(vm.IOSTAccountToPubkey(iostAccount))
		if err != nil {
			return err
		}
	}

	return nil
}

func (sp *ServiPool) flushBtu() error {

	var sbuf []byte
	for _, s := range sp.btu {
		var buf = make([]byte, 8)
		binary.BigEndian.PutUint64(buf, math.Float64bits(s.v))
		var buf1 = make([]byte, 8)
		binary.BigEndian.PutUint64(buf1, math.Float64bits(s.b))
		buf = append(buf, buf1...)
		buf = append(buf, vm.IOSTAccountToPubkey(s.owner)...)
		if len(buf) != 49 {
			panic("buf length error!")
		}
		sbuf = append(sbuf, buf...)
	}
	return ldb.Put(bestUser, sbuf)
}

func (sp *ServiPool) flushHm() error {

	var sbuf []byte
	for key, s := range sp.hm {
		var buf = make([]byte, 8)
		binary.BigEndian.PutUint64(buf, math.Float64bits(s.v))
		var buf1 = make([]byte, 8)
		binary.BigEndian.PutUint64(buf1, math.Float64bits(s.b))
		buf = append(buf, buf1...)
		buf = append(buf, vm.IOSTAccountToPubkey(s.owner)...)
		if len(buf) != 49 {
			panic("buf length error!")
		}
		sbuf = append(sbuf, buf...)

		err := ldb.Put(vm.IOSTAccountToPubkey(key), sbuf)
		if err != nil {
			log.Log.I("Failed to flushHm db", err)
		}
	}

	if len(sp.hm) > sp.cacheSize {
		sp.hm = make(map[vm.IOSTAccount]*Servi, 0)
	}

	return nil
}

func (sp *ServiPool) restoreHm(key vm.IOSTAccount) (*Servi, error) {

	bl, err := ldb.Has(vm.IOSTAccountToPubkey(key))
	if err != nil || !bl {
		return nil, err
	}

	sbuf, err := ldb.Get(vm.IOSTAccountToPubkey(key))
	if err != nil {
		return nil, err
	}

	servi := Servi{}
	bufnum := sbuf[0:8]
	servi.v = math.Float64frombits(binary.BigEndian.Uint64(bufnum))
	bufbln := sbuf[8:16]
	servi.b = math.Float64frombits(binary.BigEndian.Uint64(bufbln))
	bufkey := sbuf[16:49]
	servi.owner = vm.PubkeyToIOSTAccount(bufkey)

	return &servi, nil
}

type BestUserList []*Servi

func (s BestUserList) Len() int { return len(s) }
func (s BestUserList) Less(i, j int) bool {
	return s[i].Total() > s[j].Total()
}
func (s BestUserList) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
