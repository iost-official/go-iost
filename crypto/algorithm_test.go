package crypto

import (
	"crypto/rand"
	"testing"
)

var algos = []struct {
	name string
	data Algorithm
}{
	{"Secp256k1", Secp256k1},
	{"Ed25519", Ed25519},
}

func BenchmarkSign(b *testing.B) {
	for _, algo := range algos {
		b.Run(algo.name, func(b *testing.B) {
			seckeys := make([][]byte, 0)
			msgs := make([][]byte, 0)
			for i := 0; i < b.N; i++ {
				msg := make([]byte, 32)
				rand.Read(msg)
				msgs = append(msgs, msg)
				seckeys = append(seckeys, algo.data.GenSeckey())
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				algo.data.Sign(msgs[i], seckeys[i])
			}
		})
	}
}

func BenchmarkVerify(b *testing.B) {
	for _, algo := range algos {
		b.Run(algo.name, func(b *testing.B) {
			seckeys := make([][]byte, 0)
			pubkeys := make([][]byte, 0)
			sigs := make([][]byte, 0)
			msgs := make([][]byte, 0)
			for i := 0; i < b.N; i++ {
				msg := make([]byte, 32)
				rand.Read(msg)
				msgs = append(msgs, msg)
				seckeys = append(seckeys, algo.data.GenSeckey())
				pubkeys = append(pubkeys, algo.data.GetPubkey(seckeys[i]))
				sigs = append(sigs, algo.data.Sign(msgs[i], seckeys[i]))
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				algo.data.Verify(msgs[i], pubkeys[i], sigs[i])
			}
		})
	}
}

func BenchmarkGetPubkey(b *testing.B) {
	for _, algo := range algos {
		b.Run(algo.name, func(b *testing.B) {
			seckey := algo.data.GenSeckey()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				algo.data.GetPubkey(seckey)
			}
		})
	}
}

func BenchmarkGenSeckey(b *testing.B) {
	for _, algo := range algos {
		b.Run(algo.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				algo.data.GenSeckey()
			}
		})
	}
}
