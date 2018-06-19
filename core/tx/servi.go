package tx

import (
	"math"

	"encoding/binary"

	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/log"
	"github.com/iost-official/prototype/vm"
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

func (s *Servi) Clear() {
	s.v = s.v * 0.9
	s.v = math.Floor(s.v)
}

type ServiPool struct {
	btu map[string]*Servi // 贡献最大的账户集合
	hm  map[string]*Servi // 普通账户
	mu  sync.RWMutex
}

// 没有则添加该节点
func (sp *ServiPool) User(iostAccount vm.IOSTAccount) *Servi {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	var s *Servi
	if len(sp.btu) <= 7 {
		s = sp.userBtu(iostAccount)
	} else {
		s = sp.userHm(iostAccount)
	}

	return s
}

func (sp *ServiPool) userBtu(iostAccount vm.IOSTAccount) *Servi {

	if servi, ok := sp.btu[string(iostAccount)]; ok {
		return servi
	} else {
		sp.btu[string(iostAccount)] = &Servi{owner: iostAccount}
		return sp.btu[string(iostAccount)]
	}
}

//userHm 添加普通账户
func (sp *ServiPool) userHm(iostAccount vm.IOSTAccount) *Servi {

	if servi, ok := sp.hm[string(iostAccount)]; ok {
		return servi
	} else {
		s, err := sp.restoreHm(string(iostAccount))
		if err != nil {
			sp.hm[string(iostAccount)] = &Servi{owner: iostAccount}
		} else {
			sp.hm[string(iostAccount)] = s
		}

		return sp.hm[string(iostAccount)]
	}
}

func (sp *ServiPool) addBtu(iostAccount vm.IOSTAccount,s *Servi) error {

	if _, ok := sp.btu[string(iostAccount)]; ok {
		delete(sp.btu, string(iostAccount))
	}

	sp.btu[string(iostAccount)] = s
	return nil

}

//userHm 添加普通账户
func (sp *ServiPool) addHm(iostAccount vm.IOSTAccount, s *Servi) error {

	if _, ok := sp.hm[string(iostAccount)]; ok {
		delete(sp.hm, string(iostAccount))
	}

	sp.hm[string(iostAccount)] = s
	return nil
}


func (sp *ServiPool) delBtu(iostAccount vm.IOSTAccount) {

	if _, ok := sp.btu[string(iostAccount)]; ok {
		delete(sp.btu, string(iostAccount))
	}
}

func (sp *ServiPool) delHm(iostAccount vm.IOSTAccount) {

	if _, ok := sp.hm[string(iostAccount)]; ok {
		delete(sp.hm, string(iostAccount))
	}
}

func (sp *ServiPool) BestUser() []*Servi {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	slist := make([]*Servi, 1)
	for _, s := range sp.btu {
		slist = append(slist, s)
	}

	return slist
}

var StdServiPool = ServiPool{}

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

func (sp *ServiPool) updateBtu() {

	for k, v := range sp.hm {
		for k1,v1:=range sp.btu{
			if v.Total() > v1.Total(){
				sp.delBtu(vm.IOSTAccount(k1))
				sp.userBtu()
			}
		}
	}
}

func (sp *ServiPool) flushBtu() error {

	var sbuf []byte
	for _, s := range sp.btu {
		var buf = make([]byte, 8)
		binary.BigEndian.PutUint64(buf, math.Float64bits(s.v))
		var buf1 = make([]byte, 8)
		binary.BigEndian.PutUint64(buf1, math.Float64bits(s.b))
		buf = append(buf, buf1...)
		buf = append(buf, []byte(s.owner)...)
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
		buf = append(buf, []byte(s.owner)...)
		if len(buf) != 49 {
			panic("buf length error!")
		}
		sbuf = append(sbuf, buf...)

		return ldb.Put([]byte(key), sbuf)
	}

	return nil
}

func (sp *ServiPool) restoreHm(key string) (*Servi, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	bl, err := ldb.Has([]byte(key))
	if err != nil || !bl {
		return nil, err
	}

	sbuf, err := ldb.Get([]byte(key))
	if err != nil {
		return nil, err
	}

	servi := Servi{}
	bufnum := sbuf[ 0:8]
	servi.v = math.Float64frombits(binary.BigEndian.Uint64(bufnum))
	bufbln := sbuf[ 8:16]
	servi.b = math.Float64frombits(binary.BigEndian.Uint64(bufbln))
	bufkey := sbuf[16:49]
	servi.owner = vm.PubkeyToIOSTAccount(bufkey)

	return &servi, nil
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
		sp.btu[string(servi.owner)] = &servi
	}

	return nil
}

var ldb db.Database

func init() {
	var err error
	ldb, err = db.NewLDBDatabase("servi_db.ldb", 0, 0)
	if err != nil {
		panic(err)
	}
	StdServiPool.btu = make(map[string]*Servi)
	StdServiPool.hm = make(map[string]*Servi)

}
