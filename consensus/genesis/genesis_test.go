package genesis

import (
	"fmt"
	"os"
	"testing"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
)

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

	blk, err := GenGenesis(d, &common.GenesisConfig{
		WitnessInfo: []string{"IOSTjBxx7sUJvmxrMiyjEQnz9h5bfNrXwLinkoL9YvWjnrGdbKnBP",
			"13100000000",
			"IOSTgw6cmmWyiW25TMAK44N9coLCMaygx5eTfGVwjCcriEWEEjK2H",
			"13200000000",
			"IOSTxHn7wtQMpgvDbiypByZVNHrE6ELdXFbL1Vic8B23EgRNjQGbs",
			"13300000000",
			"IOST2gxCPceKrWauFTqMCjMgZKRykp4Gt2Nd1H1XGRP1saYFXGqH4Y",
			"13400000000",
			"IOST24jsSGj2WxSRtgZkCDng19LPbT48HMsv2Nz13NXEYoqR1aYyvS",
			"13500000000",
			"IOST2v2ZumgyNXtpf1MEbkbbAK3tFfC856oMoVUYfYDvC1mpX14AvA",
			"13600000000",
			"IOSTCJqjtLBntuWRGaZumevYgBEZsU8AaAdUpEMnpGieKV676B9St",
			"13700000000",
		},
		InitialTimestamp: "2006-01-02T15:04:05Z",
		VoteContractPath: os.Getenv("GOPATH") + "/src/github.com/iost-official/go-iost/contract/",
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(blk)
}
