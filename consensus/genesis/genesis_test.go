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
		WitnessInfo: []*common.Witness{
			{"a1", "IOSTjBxx7sUJvmxrMiyjEQnz9h5bfNrXwLinkoL9YvWjnrGdbKnBP", "IOSTjBxx7sUJvmxrMiyjEQnz9h5bfNrXwLinkoL9YvWjnrGdbKnBP", 3 * 1e8},
			{"a2", "IOSTgw6cmmWyiW25TMAK44N9coLCMaygx5eTfGVwjCcriEWEEjK2H", "IOSTgw6cmmWyiW25TMAK44N9coLCMaygx5eTfGVwjCcriEWEEjK2H", 3 * 1e8},
			{"a3", "IOSTxHn7wtQMpgvDbiypByZVNHrE6ELdXFbL1Vic8B23EgRNjQGbs", "IOSTxHn7wtQMpgvDbiypByZVNHrE6ELdXFbL1Vic8B23EgRNjQGbs", 3 * 1e8},
			{"a4", "IOST2gxCPceKrWauFTqMCjMgZKRykp4Gt2Nd1H1XGRP1saYFXGqH4Y", "IOST2gxCPceKrWauFTqMCjMgZKRykp4Gt2Nd1H1XGRP1saYFXGqH4Y", 3 * 1e8},
			{"a5", "IOST24jsSGj2WxSRtgZkCDng19LPbT48HMsv2Nz13NXEYoqR1aYyvS", "IOST24jsSGj2WxSRtgZkCDng19LPbT48HMsv2Nz13NXEYoqR1aYyvS", 3 * 1e8},
			{"a6", "IOST2v2ZumgyNXtpf1MEbkbbAK3tFfC856oMoVUYfYDvC1mpX14AvA", "IOST2v2ZumgyNXtpf1MEbkbbAK3tFfC856oMoVUYfYDvC1mpX14AvA", 3 * 1e8},
			{"a7", "IOSTCJqjtLBntuWRGaZumevYgBEZsU8AaAdUpEMnpGieKV676B9St", "IOSTCJqjtLBntuWRGaZumevYgBEZsU8AaAdUpEMnpGieKV676B9St", 3 * 1e8},
		},
		TokenInfo: &common.TokenInfo{
			FoundationAccount: "f8",
			IOSTTotalSupply:   90000000000,
			IOSTDecimal:       8,
			RAMTotalSupply:    9000000000000000000,
			RAMGenesisAmount:  137438953472,
		},
		InitialTimestamp: "2006-01-02T15:04:05Z",
		ContractPath:     os.Getenv("GOPATH") + "/src/github.com/iost-official/go-iost/contract/",
		AdminInfo:        &common.Witness{"a8", "IOSTjBxx7sUJvmxrMiyjEQnz9h5bfNrXwLinkoL9YvWjnrGdbKnBP", "IOSTjBxx7sUJvmxrMiyjEQnz9h5bfNrXwLinkoL9YvWjnrGdbKnBP", 3 * 1e8},
		FoundationInfo:   &common.Witness{"f8", "IOSTjBxx7sUJvmxrMiyjEQnz9h5bfNrXwLinkoL9YvWjnrGdbKnBP", "IOSTjBxx7sUJvmxrMiyjEQnz9h5bfNrXwLinkoL9YvWjnrGdbKnBP", 0},
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(blk)
	return
}
