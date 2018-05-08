package poly

import (
	"errors"
	"fmt"
	"github.com/LoCCS/bliss/params"
	"github.com/LoCCS/bliss/sampler"
)

type PolyArray struct {
	n     uint32
	q     uint32
	data  []int32
	param *params.BlissBParam
}

func newPolyArray(n, q uint32) (*PolyArray, error) {
	if n == 0 || q == 0 {
		return nil, errors.New("Invalid parameter: n or q cannot be zero")
	}
	data := make([]int32, n)
	array := &PolyArray{n, q, data, nil}
	return array, nil
}

func NewPolyArray(param *params.BlissBParam) (*PolyArray, error) {
	if param == nil {
		return nil, errors.New("Param is nil")
	}
	pa, err := newPolyArray(param.N, param.Q)
	if err != nil {
		return nil, err
	}
	pa.param = param
	return pa, err
}

func New(version int) (*PolyArray, error) {
	param := params.GetParam(version)
	if param == nil {
		return nil, errors.New("Failed to get parameter")
	}
	return NewPolyArray(param)
}

func (pa *PolyArray) Size() uint32 {
	return pa.n
}

func (pa *PolyArray) Param() *params.BlissBParam {
	return pa.param
}

func (pa *PolyArray) String() string {
	return fmt.Sprintf("%d", pa.data)
}

func (pa *PolyArray) SetData(data []int32) error {
	if pa.n != uint32(len(data)) {
		return errors.New("Mismatched data length!")
	}
	for i := 0; i < int(pa.n); i++ {
		pa.data[i] = data[i]
	}
	return nil
}

func (pa *PolyArray) GetData() []int32 {
	return pa.data
}

func UniformPoly(version int, entropy *sampler.Entropy) *PolyArray {
	p, err := New(version)
	if err != nil {
		return nil
	}
	n := p.param.N
	v := make([]int32, n)

	i := 0
	for i < int(p.param.Nz1) {
		x := entropy.Uint16()
		j := uint32(x>>1) % n
		mask := -(1 ^ (v[j] & 1))
		i += int(mask & 1)
		v[j] += (int32((x&1)<<1) - 1) & mask
	}

	i = 0
	for i < int(p.param.Nz2) {
		x := entropy.Uint16()
		j := uint32(x>>1) % n
		mask := -(1 ^ ((v[j] & 1) | ((v[j] & 2) >> 1)))
		i += int(mask & 1)
		v[j] += (int32((x&1)<<2) - 2) & mask
	}
	p.SetData(v)
	return p
}

func GaussPoly(version int, s *sampler.Sampler) *PolyArray {
	p, err := New(version)
	if err != nil {
		return nil
	}
	n := p.param.N
	v := make([]int32, n)
	for i := 0; i < int(n); i++ {
		v[i] = s.SampleGauss()
	}
	p.SetData(v)
	return p
}

// For splitted version
func GaussPolyAlpha(version int, s *sampler.Sampler) *PolyArray {
	p, err := New(version)
	if err != nil {
		return nil
	}
	n := p.param.N
	v := make([]int32, n)
	for i := 0; i < int(n); i++ {
		v[i] = s.SampleGaussCtAlpha()
	}
	p.SetData(v)
	return p
}

func GaussPolyBeta(version int, s *sampler.Sampler) *PolyArray {
	p, err := New(version)
	if err != nil {
		return nil
	}
	n := p.param.N
	v := make([]int32, n)
	for i := 0; i < int(n); i++ {
		v[i] = s.SampleGaussCtBeta()
	}
	p.SetData(v)
	return p
}
