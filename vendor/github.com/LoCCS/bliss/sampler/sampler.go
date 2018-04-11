package sampler

import (
	"fmt"
	"github.com/LoCCS/bliss/params"
)

type Sampler struct {
	sigma      uint32
	ell        uint32
	prec       uint32
	columns    uint32
	kSigma     uint16
	kSigmaBits uint16

	ctable []uint8

	/* For splitting */
	ell1        uint32
	ell2        uint32
	kSigma1     uint16
	kSigma2     uint16
	kSigmaBits1 uint16
	kSigmaBits2 uint16
	ctable1     []uint8
	ctable2     []uint8

	random *Entropy
}

func invalidSampler() *Sampler {
	return &Sampler{0, 0, 0, 0, 0, 0, []uint8{}, 0, 0, 0, 0, 0, 0,
		[]uint8{}, []uint8{}, nil}
}

func NewSampler(sigma, ell, prec uint32, entropy *Entropy) (*Sampler, error) {
	columns := prec / 8
	ctable, err := getTable(sigma, ell, prec)
	if err != nil {
		return invalidSampler(), err
	}
	ksigma := getKSigma(sigma, prec)
	if ksigma == 0 {
		return invalidSampler(), fmt.Errorf("Failed to get kSigma")
	}
	ksigmabits := getKSigmaBits(sigma, prec)
	if ksigmabits == 0 {
		return invalidSampler(), fmt.Errorf("Failed to get kSigmaBits")
	}
	sigma1, sigma2, ell1, ell2 := splitSigma(sigma)
	if sigma1 == 0 || sigma2 == 0 || ell1 == 0 || ell2 == 0 {
		return invalidSampler(), fmt.Errorf("Failed to split sigma")
	}
	ctable1, err := getTable(sigma1, ell1, prec)
	if err != nil {
		return invalidSampler(), err
	}
	ksigma1 := getKSigma(sigma1, prec)
	if ksigma1 == 0 {
		return invalidSampler(), fmt.Errorf("Failed to get kSigma1")
	}
	ksigmabits1 := getKSigmaBits(sigma1, prec)
	if ksigmabits1 == 0 {
		return invalidSampler(), fmt.Errorf("Failed to get kSigmaBits1")
	}
	ctable2, err := getTable(sigma2, ell2, prec)
	if err != nil {
		return invalidSampler(), err
	}
	ksigma2 := getKSigma(sigma2, prec)
	if ksigma2 == 0 {
		return invalidSampler(), fmt.Errorf("Failed to get kSigma2")
	}
	ksigmabits2 := getKSigmaBits(sigma2, prec)
	if ksigmabits2 == 0 {
		return invalidSampler(), fmt.Errorf("Failed to get kSigmaBits2")
	}
	return &Sampler{sigma, ell, prec, columns, ksigma, ksigmabits, ctable,
		ell1, ell2, ksigma1, ksigma2, ksigmabits1, ksigmabits2, ctable1, ctable2,
		entropy}, nil
}

func New(version int, entropy *Entropy) (*Sampler, error) {
	param := params.GetParam(version)
	if param == nil {
		return nil, fmt.Errorf("Failed to get parameter")
	}
	return NewSampler(param.Sigma, param.Ell, param.Prec, entropy)
}

// Sample Bernoulli distribution with probability p
// p is stored as a large big-endian integer in an array
// the real probability is p/2^d, where d is the number of
// bits of p
func (sampler *Sampler) sampleBer(p []uint8) bool {
	for _, pi := range p {
		uc := sampler.random.Char()
		if uc < pi {
			return true
		}
		if uc > pi {
			return false
		}
	}
	return true
}

// Sample Bernoulli distribution with probability p = exp(-x/(2*sigma^2))
func (sampler *Sampler) sampleBerExp(x uint32, table []uint8, ell uint32) bool {
	ri := ell - 1
	mask := uint32(1) << ri
	start := ri * sampler.columns
	for mask > 0 {
		if x&mask != 0 {
			if !sampler.sampleBer(table[start : start+sampler.columns]) {
				return false
			}
		}
		mask >>= 1
		start -= sampler.columns
	}
	return true
}

// Sample Bernoulli distribution with probability p = exp(-x/(2*sigma^2))
func (sampler *Sampler) sampleBerExpCt(x uint32, table []uint8, ell uint32) bool {
	var xi, i, ret, start, bit uint32
	start = 0
	ret = 1

	xi = x
	for i = ell - 1; i != 0; i-- {
		if sampler.sampleBer(table[start : start+sampler.columns]) {
			bit = 1
		} else {
			bit = 0
		}
		ret = ret * (1 - (xi & 1) + uint32(bit)*(xi&1))
		xi >>= 1
		start += sampler.columns
	}

	return ret != 0
}

// Sample Bernoulli distribution with probability p = 1/cosh(-x/(2*sigma^2))
func (sampler *Sampler) sampleBerCosh(x int32, table []uint8, ell uint32) bool {
	if x < 0 {
		x = -x
	}
	x <<= 1
	for {
		bit := sampler.sampleBerExp(uint32(x), table, ell)
		if bit {
			return true
		}
		bit = sampler.random.Bit()
		if !bit {
			bit = sampler.sampleBerExp(uint32(x), table, ell)
			if !bit {
				return false
			}
		}
	}
}

func (sampler *Sampler) SampleBerExp(x uint32) bool {
	return sampler.sampleBerExp(x, sampler.ctable, sampler.ell)
}

func (sampler *Sampler) SampleBerExpCt(x uint32) bool {
	return sampler.sampleBerExpCt(x, sampler.ctable, sampler.ell)
}

func (sampler *Sampler) SampleBerCosh(x int32) bool {
	return sampler.sampleBerCosh(x, sampler.ctable, sampler.ell)
}

// Discrete Binary Gauss distribution is Discrete Gauss Distribution with
// a specific variance sigma = sqrt(1/(2 ln 2)) = 0.849...
// This is used as foundation of SampleGauss.
func (sampler *Sampler) SampleBinaryGauss() uint32 {
restart:
	if sampler.random.Bit() {
		return 0
	}
	for i := 1; i <= 16; i++ {
		u := sampler.random.Bits(2*i - 1)
		if u == 0 {
			return uint32(i)
		}
		if u != 1 {
			goto restart
		}
	}
	return 0
}

// Sample according to Discrete Gauss Distribution
// exp(-x^2/(2*sigma*sigma))
func (sampler *Sampler) sampleGauss(ksigma uint16, ksigmabits uint16, table []uint8, ell uint32) int32 {
	var x, y uint32
	var u bool
	for {
		x = sampler.SampleBinaryGauss()
		for {
			y = sampler.random.Bits(int(ksigmabits))
			if y < uint32(ksigma) {
				break
			}
		}
		e := y * (y + 2*uint32(ksigma)*x)
		u = sampler.random.Bit()
		if (x|y) != 0 || u {
			if sampler.sampleBerExp(e, table, ell) {
				break
			}
		}
	}

	valPos := int32(uint32(ksigma)*x + y)
	if u {
		return valPos
	} else {
		return -valPos
	}
}

func (sampler *Sampler) SampleGauss() int32 {
	return sampler.sampleGauss(sampler.kSigma, sampler.kSigmaBits, sampler.ctable, sampler.ell)
}

// Sample according to Discrete Gauss Distribution, constant time
// exp(-x^2/(2*sigma*sigma))
func (sampler *Sampler) sampleGaussCt(ksigma uint16, ksigmabits uint16, table []uint8, ell uint32) int32 {
	var x, y uint32
	var u bool
	for {
		x = sampler.SampleBinaryGauss()
		for {
			y = sampler.random.Bits(int(ksigmabits))
			if y < uint32(ksigma) {
				break
			}
		}
		e := y * (y + 2*uint32(ksigma)*x)
		u = sampler.random.Bit()
		if (x|y) != 0 || u {
			if sampler.sampleBerExpCt(e, table, ell) {
				break
			}
		}
	}

	valPos := int32(uint32(ksigma)*x + y)
	if u {
		return valPos
	} else {
		return -valPos
	}
}

func (sampler *Sampler) SampleGaussCt() int32 {
	return sampler.sampleGaussCt(sampler.kSigma, sampler.kSigmaBits, sampler.ctable, sampler.ell)
}

// Sample according to Discrete Gauss Distribution, by sigma1
func (sampler *Sampler) SampleGaussCtAlpha() int32 {
	return sampler.sampleGaussCt(sampler.kSigma1, sampler.kSigmaBits1, sampler.ctable1, sampler.ell1)
}

// Sample according to Discrete Gauss Distribution, by sigma2
func (sampler *Sampler) SampleGaussCtBeta() int32 {
	return sampler.sampleGaussCt(sampler.kSigma2, sampler.kSigmaBits2, sampler.ctable2, sampler.ell2)
}
