package protocol

import (
	"testing"
	"IOS/src/iosbase"
)

type MockView struct {
}

func TestOnNewView(t *testing.T) {
	prim := iosbase.Member{ID: "primary"}
	bkup0 := iosbase.Member{ID: "bkup0"}
	bkup1 := iosbase.Member{ID: "bkup1"}
	bkup2 := iosbase.Member{ID: "bkup2"}
	view := DposView{primary:prim, backup:[]iosbase.Member{bkup0, bkup1, bkup2}}

	var consensus Consensus
	consensus.onNewView(&view)

}
