package genesis

import (
	"fmt"
	"os"
	"testing"

	"github.com/iost-official/go-iost/v3/account"
	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/crypto"
	"github.com/iost-official/go-iost/v3/db"
	"github.com/iost-official/go-iost/v3/ilog"
)

func randWitness(idx int) *common.Witness {
	k := account.EncodePubkey(crypto.Ed25519.GetPubkey(crypto.Ed25519.GenSeckey()))
	return &common.Witness{ID: fmt.Sprintf("a%d", idx), Owner: k, Active: k, SignatureBlock: k, Balance: 3 * 1e8}
}

func TestGenGenesis(t *testing.T) {
	ilog.Stop()

	d, err := db.NewMVCCDB("mvcc")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		d.Close()
		os.RemoveAll("mvcc")
	}()
	k := account.EncodePubkey(crypto.Ed25519.GetPubkey(crypto.Ed25519.GenSeckey()))
	//fmt.Println("path", os.Getenv("GOBASE") + "//config/genesis/contract/")
	blk, err := GenGenesis(d, &common.GenesisConfig{
		WitnessInfo: []*common.Witness{
			randWitness(1),
			randWitness(2),
			randWitness(3),
			randWitness(4),
			randWitness(5),
			randWitness(6),
			randWitness(7),
		},
		TokenInfo: &common.TokenInfo{
			FoundationAccount: "f8",
			IOSTTotalSupply:   90000000000,
			IOSTDecimal:       8,
		},
		InitialTimestamp: "2006-01-02T15:04:05Z",
		ContractPath:     os.Getenv("GOBASE") + "//config/genesis/contract/",
		AdminInfo:        randWitness(8),
		FoundationInfo:   &common.Witness{ID: "f8", Owner: k, Active: k, Balance: 0},
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(blk.Head)
	return
}
