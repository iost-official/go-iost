package poly

func (lh *PolyArray) Inc(rh *PolyArray) *PolyArray {
	if lh.n != rh.n || lh.q != rh.q {
		return nil
	}
	n := lh.n
	for i := 0; i < int(n); i++ {
		lh.data[i] = lh.data[i] + rh.data[i]
	}
	return lh
}

func (lh *PolyArray) ScalarInc(rh int32) *PolyArray {
	lh.data[0] = lh.data[0] + rh
	return lh
}

func (lh *PolyArray) Add(rh *PolyArray) *PolyArray {
	if lh.n != rh.n || lh.q != rh.q {
		return nil
	}
	n, q := lh.n, lh.q
	var ret *PolyArray
	if lh.param != nil {
		ret, _ = NewPolyArray(lh.param)
	} else {
		ret, _ = newPolyArray(n, q)
	}
	for i := 0; i < int(n); i++ {
		ret.data[i] = lh.data[i] + rh.data[i]
	}
	return ret
}

func (lh *PolyArray) Dec(rh *PolyArray) *PolyArray {
	if lh.n != rh.n || lh.q != rh.q {
		return nil
	}
	n := lh.n
	for i := 0; i < int(n); i++ {
		lh.data[i] = lh.data[i] - rh.data[i]
	}
	return lh
}

func (lh *PolyArray) Sub(rh *PolyArray) *PolyArray {
	if lh.n != rh.n || lh.q != rh.q {
		return nil
	}
	n, q := lh.n, lh.q
	var ret *PolyArray
	if lh.param != nil {
		ret, _ = NewPolyArray(lh.param)
	} else {
		ret, _ = newPolyArray(n, q)
	}
	for i := 0; i < int(n); i++ {
		ret.data[i] = lh.data[i] - rh.data[i]
	}
	return ret
}

func (lh *PolyArray) Mul(rh *PolyArray) *PolyArray {
	if lh.n != rh.n || lh.q != rh.q {
		return nil
	}
	n := lh.n
	for i := 0; i < int(n); i++ {
		lh.data[i] = lh.data[i] * rh.data[i]
	}
	return lh
}

func (lh *PolyArray) Times(rh *PolyArray) *PolyArray {
	if lh.n != rh.n || lh.q != rh.q {
		return nil
	}
	n, q := lh.n, lh.q
	var ret *PolyArray
	if lh.param != nil {
		ret, _ = NewPolyArray(lh.param)
	} else {
		ret, _ = newPolyArray(n, q)
	}
	for i := 0; i < int(n); i++ {
		ret.data[i] = lh.data[i] * rh.data[i]
	}
	return ret
}

func (lh *PolyArray) ScalarMul(rh int32) *PolyArray {
	n := lh.n
	for i := 0; i < int(n); i++ {
		lh.data[i] = lh.data[i] * rh
	}
	return lh
}

func (lh *PolyArray) ScalarTimes(rh int32) *PolyArray {
	n, q := lh.n, lh.q
	var ret *PolyArray
	if lh.param != nil {
		ret, _ = NewPolyArray(lh.param)
	} else {
		ret, _ = newPolyArray(n, q)
	}
	for i := 0; i < int(n); i++ {
		ret.data[i] = lh.data[i] * rh
	}
	return ret
}

func (pa *PolyArray) Norm2() int32 {
	sum := int32(0)
	for i := 0; i < len(pa.data); i++ {
		sum += pa.data[i] * pa.data[i]
	}
	return sum
}

func (pa *PolyArray) MaxNorm() int32 {
	max := int32(0)
	for i := 0; i < len(pa.data); i++ {
		if pa.data[i] > max {
			max = pa.data[i]
		} else if -pa.data[i] > max {
			max = -pa.data[i]
		}
	}
	return max
}

func (lh *PolyArray) InnerProduct(rh *PolyArray) int32 {
	if lh.n != rh.n || lh.q != rh.q {
		return 0
	}
	n := lh.n
	sum := int32(0)
	for i := 0; i < int(n); i++ {
		sum += lh.data[i] * rh.data[i]
	}
	return sum
}

func (pa *PolyArray) DropBits() *PolyArray {
	var ret *PolyArray
	if pa.param != nil {
		ret, _ = NewPolyArray(pa.param)
	} else {
		return nil
	}
	delta := int32(1) << pa.param.D
	halfDelta := delta >> 1
	for i := 0; i < len(pa.data); i++ {
		ret.data[i] = (pa.data[i] + halfDelta) / delta
	}
	return ret
}

func (pa *PolyArray) Mul2d() *PolyArray {
	var ret *PolyArray
	if pa.param != nil {
		ret, _ = NewPolyArray(pa.param)
	} else {
		return nil
	}
	D := pa.param.D
	for i := 0; i < len(pa.data); i++ {
		ret.data[i] = pa.data[i] << D
	}
	return ret
}

func (pa *PolyArray) BoundByP() *PolyArray {
	var ret *PolyArray
	if pa.param != nil {
		ret, _ = NewPolyArray(pa.param)
	} else {
		return nil
	}
	p := int32(pa.param.Modp)
	for i := 0; i < len(pa.data); i++ {
		if pa.data[i] < -p/2 {
			ret.data[i] = pa.data[i] + p
		} else if pa.data[i] >= p/2 {
			ret.data[i] = pa.data[i] - p
		} else {
			ret.data[i] = pa.data[i]
		}
	}
	return ret
}
