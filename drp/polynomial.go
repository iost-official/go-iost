package drp

import (
	"fmt"
	"math/big"
)

type Polynomial struct {
	elements []Scalar
}

func (poly Polynomial) String() string {
	s := poly.elements[0].toInt().String()
	for d := 1; d < poly.degree(); d++ {
		c := poly.elements[d].toInt()
		if c.Cmp(big.NewInt(0)) != 0 {
			s = fmt.Sprintf("%s * x^%d", c.String(), d) + " + " + s
		}
	}
	return s
}

func polynomialGen(threshold int) Polynomial {
	elements := make([]Scalar, threshold)
	for i := 0; i < threshold; i++ {
		elements[i] = keypairGen().private
	}
	return Polynomial{elements}
}

func (p *Polynomial) degree() int {
	return cap(p.elements)
}

func (p *Polynomial) evaluate(x *big.Int) Scalar {
	order := getCurveParams().N
	xN := new(big.Int)
	v := new(Scalar)
	*v = p.elements[0]
	*xN = *x
	for power := 1; power < p.degree(); power++ {
		coeff := p.elements[power]
		d := new(Scalar).Mul(&coeff, new(Scalar).fromInt(xN))
		v = v.Add(v, d)
		if power <= p.degree()-1 {
			xN = xN.Mod(xN.Mul(xN, x), order)
		}
	}
	return *v
}
