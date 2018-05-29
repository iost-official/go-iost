package pob

import (
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
	s.v = 0
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
