package poly

func addMod(a,b int32, q uint32) int32 {
	// inputs a,b must be in [0,q)
	// this algorithm ensures the result is also in [0,q)
	a += b-int32(q)
	return a + (int32)((uint32)(a>>31) & q)
}

func subMod(a,b int32, q uint32) int32 {
	// inputs a,b must be in [0,q)
	// this algorithm ensures the result is also in [0,q)
	a -= b
	return a + (int32)((uint32)(a>>31) & q)
}

func mulMod(a,b int32, q uint32) int32 {
	a = (a * b) % int32(q)
	return a + (int32)((uint32)(a>>31) & q)
}

func bound(a int32, q uint32) int32 {
	return a + (int32)((uint32)(a>>31) & q)
}

func expMod(a int32, e,q uint32) int32 {
	var y int32
	y = 1
	if e & 1 == 1 {
		y = a
	}
	e >>= 1
	for e > 0 {
		a = (a * a) % int32(q)
		if e & 1 == 1 {
			y = (a * y) % int32(q)
		}
		e >>= 1
	}
	return y
}

