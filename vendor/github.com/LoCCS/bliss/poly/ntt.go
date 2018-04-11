package poly

import (
	"errors"
)

func (ma *PolyArray) FFT() (*PolyArray, error) {
	var i, j, k uint32
	n := ma.param.N
	q := ma.param.Q
	psi := ma.param.Psi
	array, err := NewPolyArray(ma.param)
	if err != nil {
		return nil, err
	}
	array.SetData(ma.data)
	v := array.data

	// Bit-Inverse Shuffle
	j = n >> 1
	for i = 1; i < n-1; i++ {
		if i < j {
			tmp := v[i]
			v[i] = v[j]
			v[j] = tmp
		}
		k := n
		for {
			k >>= 1
			j ^= k
			if (j & k) != 0 {
				break
			}
		}
	}

	// Main loop
	l := n
	for i = 1; i < n; i <<= 1 {
		i2 := i + i
		for k = 0; k < n; k += i2 {
			tmp := v[k+i]
			v[k+i] = subMod(v[k], tmp, q)
			v[k] = addMod(v[k], tmp, q)
		}
		for j = 1; j < i; j++ {
			y := psi[j*l]
			for k = j; k < n; k += i2 {
				tmp := (v[k+i] * y) % int32(q)
				v[k+i] = subMod(v[k], tmp, q)
				v[k] = addMod(v[k], tmp, q)
			}
		}
		l >>= 1
	}

	return array, nil
}

func (p *PolyArray) NTT() (*PolyArray, error) {
	psi, err := NewPolyArray(p.param)
	if err != nil {
		return nil, err
	}
	psi.SetData(p.param.Psi)
	f := p.TimesModQ(psi)
	g, err := f.FFT()
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (ntt *PolyArray) INTT() (*PolyArray, error) {
	rpsi, err := NewPolyArray(ntt.param)
	rpsi.SetData(ntt.param.RPsi)
	if err != nil {
		return nil, err
	}
	f, err := ntt.FFT()
	if err != nil {
		return nil, err
	}
	f.MulModQ(rpsi)
	f.flip()
	return f, nil
}

func (ntt *PolyArray) InvertAsNTT() (*PolyArray, error) {
	for i := 0; i < int(ntt.n); i++ {
		if ntt.data[i] == 0 {
			return nil, errors.New("PolyArray not invertible")
		}
	}
	ret := ntt.ExpModQ(ntt.q - 2)
	return ret, nil
}

func (p *PolyArray) MultiplyNTT(ntt *PolyArray) (*PolyArray, error) {
	lh, err := p.NTT()
	if err != nil {
		return nil, err
	}
	lh.MulModQ(ntt)
	return lh.INTT()
}

func (lh *PolyArray) flip() *PolyArray {
	n, q := lh.n, lh.q
	for i, j := 1, n-1; i < int(j); i, j = i+1, j-1 {
		tmp := lh.data[i]
		lh.data[i] = lh.data[j]
		lh.data[j] = tmp
	}
	tmp := int32(q) & ((-lh.data[0]) >> 31)
	lh.data[0] = tmp - lh.data[0]
	return lh
}
