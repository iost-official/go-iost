package tx

import (
	"math"

	"encoding/binary"

	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/vm"
)

const base float64 = 1.0

type Servi struct {
	v     float64
	owner vm.IOSTAccount
}

func (s *Servi) Incr(time int) {
	s.v += base * float64(time)
}

func (s *Servi) Clear() {
	s.v = s.v * 0.9
	s.v = math.Floor(s.v)
}

type ServiPool struct {
	hm map[string]*Servi
}

func (sp *ServiPool) User(iostAccount vm.IOSTAccount) *Servi {
	if servi, ok := sp.hm[string(iostAccount)]; ok {
		return servi
	} else {
		sp.hm[string(iostAccount)] = &Servi{owner: iostAccount}
		return sp.hm[string(iostAccount)]
	}
}

func (sp *ServiPool) BestUser() *Servi {
	var bestk string
	var best float64
	for k, v := range sp.hm {
		if v.v > best {
			best = v.v
			bestk = k
		}
	}
	return sp.hm[bestk]
}

var StdServiPool = ServiPool{}

func (sp *ServiPool) Flush(headhash []byte) {
	var sbuf []byte
	for _, s := range sp.hm {
		var buf = make([]byte, 8)
		binary.BigEndian.PutUint64(buf, math.Float64bits(s.v))
		buf = append(buf, vm.IOSTAccountToPubkey(s.owner)...)
		if len(buf) != 41 {
			panic("buf length error!")
		}
		sbuf = append(sbuf, buf...)
	}
	ldb.Put(headhash, sbuf)
}

func (sp *ServiPool) Restore(headhash []byte) error {
	sbuf, err := ldb.Get(headhash)
	if err != nil {
		return err
	}

	for i := 0; i < len(sbuf); i += 41 {
		servi := Servi{}
		bufnum := sbuf[i : i+8]
		servi.v = math.Float64frombits(binary.BigEndian.Uint64(bufnum))
		bufkey := sbuf[i+8 : i+41]
		servi.owner = vm.PubkeyToIOSTAccount(bufkey)
		sp.hm[string(servi.owner)] = &servi
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
	StdServiPool.hm = make(map[string]*Servi)
}
