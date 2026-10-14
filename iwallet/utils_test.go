package iwallet

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/iost-official/go-iost/v3/account"
	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/crypto"
	rpcpb "github.com/iost-official/go-iost/v3/rpc/pb"
	"github.com/iost-official/go-iost/v3/sdk"
)

func TestHandleMultiSigSignKeyFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	files := make([]string, 2)
	publicKeys := make([][]byte, len(files))
	for i := range files {
		key, err := account.NewKeyPair(nil, crypto.Ed25519)
		if err != nil {
			t.Fatal(err)
		}
		info := sdk.NewAccountInfo()
		info.Name = fmt.Sprintf("signer%d", i)
		info.Keypairs["active"] = &sdk.KeyPairInfo{
			RawKey:  common.Base58Encode(key.Seckey),
			KeyType: "ed25519",
			PubKey:  key.ReadablePubkey(),
		}
		files[i] = filepath.Join(dir, info.Name+".json")
		if err := info.SaveTo(files[i]); err != nil {
			t.Fatal(err)
		}
		publicKeys[i] = key.Pubkey
	}

	for _, tt := range []struct {
		name          string
		publisherSign bool
	}{
		{name: "transaction signatures"},
		{name: "publisher signatures", publisherSign: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tx := &rpcpb.TransactionRequest{
				Time:       1,
				Expiration: 2,
				Publisher:  "publisher",
				Actions:    []*rpcpb.Action{sdk.NewAction("test", "call", "[]")},
			}
			if err := handleMultiSig(tx, nil, files, tt.publisherSign); err != nil {
				t.Fatal(err)
			}
			signatures, otherSignatures := tx.Signatures, tx.PublisherSigs
			if tt.publisherSign {
				signatures, otherSignatures = tx.PublisherSigs, tx.Signatures
			}
			if len(signatures) != len(files) {
				t.Fatalf("got %d signatures, want %d", len(signatures), len(files))
			}
			if len(otherSignatures) != 0 {
				t.Fatal("unexpected signatures in the other signature field")
			}
			for i, sig := range signatures {
				if !bytes.Equal(sig.PublicKey, publicKeys[i]) {
					t.Errorf("signature %d does not match the active key in its input file", i)
				}
				if !sdk.VerifySigForTx(tx, sig, tt.publisherSign) {
					t.Errorf("signature %d failed verification", i)
				}
			}
		})
	}
}
