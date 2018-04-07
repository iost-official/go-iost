package drp

import (
	"fmt"
	"math/big"
)

type ShareId int

type EncryptedShare struct {
	sid          ShareId
	encryptedVal Point
}

type DecryptedShare struct {
	sid          ShareId
	decryptedVal Point
}

func escrow(threshold int, participants []Point) (DhSecret, []EncryptedShare) {
	poly := polynomialGen(threshold)
	dh := poly.elements[0].toPoint()
	dhsec := dh.ToDhSecret()
	fmt.Println(poly)
	shares := make([]EncryptedShare, len(participants))
	for p := 0; p < len(participants); p++ {
		evalVal := poly.evaluate(big.NewInt(int64(p + 2)))
		key := participants[p]
		yi := PointMul(&key, &evalVal)
		eshare := EncryptedShare{ShareId(p + 2), *yi}
		shares[p] = eshare
	}
	return dhsec, shares
}

func (share *EncryptedShare) decryptShare(priv *Scalar) DecryptedShare {
	decryptedVal := PointDiv(&share.encryptedVal, priv)
	return DecryptedShare{share.sid, *decryptedVal}
}

func pool(shares []DecryptedShare) DhSecret {
	var v *Point
	fmt.Println("pool with", len(shares))
	for j := 0; j < len(shares); j++ {
		r := new(Scalar).fromSmallInt(1)
		for m := 0; m < len(shares); m++ {
			if j != m {
				num := new(Scalar).fromSmallInt(int(shares[m].sid))
				denum := new(Scalar).fromSmallInt(int(shares[m].sid) - int(shares[j].sid))
				dinv := denum.Inverse(denum)
				t := new(Scalar).Mul(num, dinv)
				r = r.Mul(r, t)
			}
		}
		p := PointMul(&shares[j].decryptedVal, r)
		if v == nil {
			v = p
		} else {
			v = v.Add(v, p)
		}
	}
	return v.ToDhSecret()
}
