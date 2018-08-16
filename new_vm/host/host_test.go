package host

import (
	"testing"

	. "github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
)

func sliceEqual(a, b []string) bool {
	if len(a) == len(b) {
		for i, s := range a {
			if s != b[i] {
				return false
			}
		}
		return true
	}
	return false
}

func myinit(t *testing.T, ctx *Context) (*database.MockIMultiValue, Host) {
	mockCtrl := NewController(t)
	defer mockCtrl.Finish()
	db := database.NewMockIMultiValue(mockCtrl)
	bdb := database.NewVisitor(100, db)

	//monitor := Monitor{}

	host := NewHost(ctx, bdb, nil, nil)
	return db, *host
}

func TestHost_Put(t *testing.T) {

	ctx := NewContext(nil)
	ctx.Set("commit", "abc")
	ctx.Set("contract_name", "contractName")

	mock, host := myinit(t, ctx)

	mock.EXPECT().Put(Any(), Any(), Any()).Do(func(a, b, c string) {
		if a != "state" || b != "b-contractName-hello" || c != "sworld" {
			t.Fatal(a, b, c)
		}
	})

	host.Put("hello", "world")
}

func TestHost_Get(t *testing.T) {
	ctx := NewContext(nil)
	ctx.Set("commit", "abc")
	ctx.Set("contract_name", "contractName")

	mock, host := myinit(t, ctx)

	mock.EXPECT().Get(Any(), Any()).DoAndReturn(func(a, b string) (string, error) {
		if a != "state" || b != "b-contractName-hello" {
			t.Fatal(a, b)
		}
		return "sworld", nil
	})

	ans, _ := host.Get("hello")
	if ans != "world" {
		t.Fatal(ans)
	}
}

func TestHost_MapPut(t *testing.T) {

	ctx := NewContext(nil)
	ctx.Set("commit", "abc")
	ctx.Set("contract_name", "contractName")

	mock, host := myinit(t, ctx)

	mock.EXPECT().Put(Any(), Any(), Any()).Do(func(a, b, c string) {
		if a != "state" || b != "m-contractName-hello-1" || c != "sworld" {
			t.Fatal(a, b, c)
		}
	})

	host.MapPut("hello", "1", "world")
}

func TestHost_MapGet(t *testing.T) {

	ctx := NewContext(nil)
	ctx.Set("commit", "abc")
	ctx.Set("contract_name", "contractName")

	mock, host := myinit(t, ctx)

	mock.EXPECT().Get(Any(), Any()).DoAndReturn(func(a, b string) (string, error) {
		if a != "state" || b != "m-contractName-hello-1" {
			t.Fatal(a, b)
		}
		return "sworld", nil
	})

	ans, _ := host.MapGet("hello", "1")
	if ans != "world" {
		t.Fatal(ans)
	}
}

func TestHost_MapKeys(t *testing.T) {

	ctx := NewContext(nil)
	ctx.Set("commit", "abc")
	ctx.Set("contract_name", "contractName")

	mock, host := myinit(t, ctx)

	mock.EXPECT().Keys(Any(), Any()).DoAndReturn(func(a, b string) ([]string, error) {
		if a != "state" || b != "m-contractName-hello-" {
			t.Fatal(a, b)
		}
		return []string{"m-contractName-hello-a", "m-contractName-hello-b", "m-contractName-hello-c"}, nil
	})

	ans, _ := host.MapKeys("hello")
	if !sliceEqual(ans, []string{"a", "b", "c"}) {
		t.Fatal(ans)
	}
}

func TestHost_RequireAuth(t *testing.T) {

	ctx := NewContext(nil)
	ctx.Set("commit", "abc")
	ctx.Set("contract_name", "contractName")
	ctx.Set("auth_list", map[string]int{"a": 1, "b": 0})

	_, host := myinit(t, ctx)

	ans, _ := host.RequireAuth("a")
	if !ans {
		t.Fatal(ans)
	}
	ans, _ = host.RequireAuth("b")
	if ans {
		t.Fatal(ans)
	}
	ans, _ = host.RequireAuth("c")
	if ans {
		t.Fatal(ans)
	}
}

func TestHost_BlockInfo(t *testing.T) {

}
