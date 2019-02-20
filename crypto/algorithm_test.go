package crypto

import (
	"crypto/rand"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var algos = []Algorithm{
	Secp256k1,
	Ed25519,
}

func TestVerify(t *testing.T) {
	for _, algo := range algos {
		seckey := algo.GenSeckey()
		pubkey := algo.GetPubkey(seckey)
		msg := make([]byte, 32)
		rand.Read(msg)
		sig := algo.Sign(msg, seckey)
		assert.True(t, algo.Verify(msg, pubkey, sig))
		assert.False(t, algo.Verify(msg, pubkey[:31], sig))
		sig[0]++
		assert.False(t, algo.Verify(msg, pubkey, sig))
	}
}

func BenchmarkSign(b *testing.B) {
	for _, algo := range algos {
		b.Run(reflect.TypeOf(algo.getBackend()).String(), func(b *testing.B) {
			seckeys := make([][]byte, 0)
			msgs := make([][]byte, 0)
			for i := 0; i < b.N; i++ {
				msg := make([]byte, 32)
				rand.Read(msg)
				msgs = append(msgs, msg)
				seckeys = append(seckeys, algo.GenSeckey())
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				algo.Sign(msgs[i], seckeys[i])
			}
		})
	}
}

func BenchmarkVerify(b *testing.B) {
	for _, algo := range algos {
		b.Run(reflect.TypeOf(algo.getBackend()).String(), func(b *testing.B) {
			seckeys := make([][]byte, 0)
			pubkeys := make([][]byte, 0)
			sigs := make([][]byte, 0)
			msgs := make([][]byte, 0)
			for i := 0; i < b.N; i++ {
				msg := make([]byte, 32)
				rand.Read(msg)
				msgs = append(msgs, msg)
				seckeys = append(seckeys, algo.GenSeckey())
				pubkeys = append(pubkeys, algo.GetPubkey(seckeys[i]))
				sigs = append(sigs, algo.Sign(msgs[i], seckeys[i]))
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				algo.Verify(msgs[i], pubkeys[i], sigs[i])
			}
		})
	}
}

func BenchmarkGetPubkey(b *testing.B) {
	for _, algo := range algos {
		b.Run(reflect.TypeOf(algo.getBackend()).String(), func(b *testing.B) {
			seckey := algo.GenSeckey()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				algo.GetPubkey(seckey)
			}
		})
	}
}

func BenchmarkGenSeckey(b *testing.B) {
	for _, algo := range algos {
		b.Run(reflect.TypeOf(algo.getBackend()).String(), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				algo.GenSeckey()
			}
		})
	}
}
