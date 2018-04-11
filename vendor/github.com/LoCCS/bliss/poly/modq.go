package poly

func (lh *PolyArray) IncModQ(rh *PolyArray) *PolyArray {
	if lh.n != rh.n || lh.q != rh.q {
		return nil
	}
	n, q := lh.n, lh.q
	for i := 0; i < int(n); i++ {
		lh.data[i] = addMod(lh.data[i], rh.data[i], q)
	}
	return lh
}

func (lh *PolyArray) ScalarIncModQ(rh int32) *PolyArray {
	lh.data[0] = addMod(lh.data[0], rh, lh.q)
	return lh
}

func (lh *PolyArray) AddModQ(rh *PolyArray) *PolyArray {
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
		ret.data[i] = addMod(lh.data[i], rh.data[i], q)
	}
	return ret
}

func (lh *PolyArray) DecModQ(rh *PolyArray) *PolyArray {
	if lh.n != rh.n || lh.q != rh.q {
		return nil
	}
	n, q := lh.n, lh.q
	for i := 0; i < int(n); i++ {
		lh.data[i] = subMod(lh.data[i], rh.data[i], q)
	}
	return lh
}

func (lh *PolyArray) SubModQ(rh *PolyArray) *PolyArray {
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
		ret.data[i] = subMod(lh.data[i], rh.data[i], q)
	}
	return ret
}

func (lh *PolyArray) MulModQ(rh *PolyArray) *PolyArray {
	if lh.n != rh.n || lh.q != rh.q {
		return nil
	}
	n, q := lh.n, lh.q
	for i := 0; i < int(n); i++ {
		lh.data[i] = mulMod(lh.data[i], rh.data[i], q)
	}
	return lh
}

func (lh *PolyArray) TimesModQ(rh *PolyArray) *PolyArray {
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
		ret.data[i] = mulMod(lh.data[i], rh.data[i], q)
	}
	return ret
}

func (lh *PolyArray) ScalarMulModQ(rh int32) *PolyArray {
	n, q := lh.n, lh.q
	for i := 0; i < int(n); i++ {
		lh.data[i] = mulMod(lh.data[i], rh, q)
	}
	return lh
}

func (lh *PolyArray) ScalarTimesModQ(rh int32) *PolyArray {
	n, q := lh.n, lh.q
	var ret *PolyArray
	if lh.param != nil {
		ret, _ = NewPolyArray(lh.param)
	} else {
		ret, _ = newPolyArray(n, q)
	}
	for i := 0; i < int(n); i++ {
		ret.data[i] = mulMod(lh.data[i], rh, q)
	}
	return ret
}

func (lh *PolyArray) ExpModQ(e uint32) *PolyArray {
	n, q := lh.n, lh.q
	var ret *PolyArray
	if lh.param != nil {
		ret, _ = NewPolyArray(lh.param)
	} else {
		ret, _ = newPolyArray(n, q)
	}
	for i := 0; i < int(n); i++ {
		ret.data[i] = expMod(lh.data[i], e, q)
	}
	return ret
}

func (lh *PolyArray) ModQ() *PolyArray {
	n, q := lh.n, lh.q
	for i := 0; i < int(n); i++ {
		lh.data[i] = bound(lh.data[i], q)
	}
	return lh
}

func (lh *PolyArray) NumModQ(a int32) int32 {
	return bound(a, lh.q)
}

func (lh *PolyArray) NumMod2Q(a int32) int32 {
	return bound(a%int32(lh.q*2), lh.q*2)
}

func (lh *PolyArray) Mod2Q() *PolyArray {
	n := lh.n
	for i := 0; i < int(n); i++ {
		lh.data[i] = lh.NumMod2Q(lh.data[i])
	}
	return lh
}

func (lh *PolyArray) ModP() *PolyArray {
	n := lh.n
	for i := 0; i < int(n); i++ {
		lh.data[i] = bound(lh.data[i]%int32(lh.param.Modp), lh.param.Modp)
	}
	return lh
}
