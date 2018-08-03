package new_vm

import (
	"context"
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

func myinit(t *testing.T, ctx context.Context) (*database.MockIMultiValue, Host) {
	mockCtrl := NewController(t)
	defer mockCtrl.Finish()
	db := database.NewMockIMultiValue(mockCtrl)
	bdb := database.NewVisitor(100, db)

	//monitor := Monitor{}
	return db, Host{ctx: ctx, db: bdb}
}

func TestHost_Put(t *testing.T) {

	ctx := context.Background()
	ctx = context.WithValue(ctx, "commit", "abc")
	ctx = context.WithValue(ctx, "contract_name", "contractName")

	mock, host := myinit(t, ctx)

	mock.EXPECT().Checkout(Any()).Do(func(commit string) {
		if commit != "abc" {
			t.Fatal(commit)
		}
	})
	mock.EXPECT().Put(Any(), Any(), Any()).Do(func(a, b, c string) {
		if a != "state" || b != "b-contractName-hello" || c != "world" {
			t.Fatal(a, b, c)
		}
	})

	host.Put("hello", "world")
}

func TestHost_Get(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "commit", "abc")
	ctx = context.WithValue(ctx, "contract_name", "contractName")

	mock, host := myinit(t, ctx)

	mock.EXPECT().Checkout(Any()).Do(func(commit string) {
		if commit != "abc" {
			t.Fatal(commit)
		}
	})
	mock.EXPECT().Get(Any(), Any()).DoAndReturn(func(a, b string) (string, error) {
		if a != "state" || b != "b-contractName-hello" {
			t.Fatal(a, b)
		}
		return "world", nil
	})

	ans := host.Get("hello")
	if ans != "world" {
		t.Fatal(ans)
	}
}

func TestHost_MapPut(t *testing.T) {

	ctx := context.Background()
	ctx = context.WithValue(ctx, "commit", "abc")
	ctx = context.WithValue(ctx, "contract_name", "contractName")

	mock, host := myinit(t, ctx)

	mock.EXPECT().Checkout(Any()).Do(func(commit string) {
		if commit != "abc" {
			t.Fatal(commit)
		}
	})
	mock.EXPECT().Put(Any(), Any(), Any()).Do(func(a, b, c string) {
		if a != "state" || b != "m-contractName-hello-1" || c != "world" {
			t.Fatal(a, b, c)
		}
	})

	host.MapPut("hello", "1", "world")
}

func TestHost_MapGet(t *testing.T) {

	ctx := context.Background()
	ctx = context.WithValue(ctx, "commit", "abc")
	ctx = context.WithValue(ctx, "contract_name", "contractName")

	mock, host := myinit(t, ctx)

	mock.EXPECT().Checkout(Any()).Do(func(commit string) {
		if commit != "abc" {
			t.Fatal(commit)
		}
	})
	mock.EXPECT().Get(Any(), Any()).DoAndReturn(func(a, b string) (string, error) {
		if a != "state" || b != "m-contractName-hello-1" {
			t.Fatal(a, b)
		}
		return "world", nil
	})

	ans := host.MapGet("hello", "1")
	if ans != "world" {
		t.Fatal(ans)
	}
}

func TestHost_MapKeys(t *testing.T) {

	ctx := context.Background()
	ctx = context.WithValue(ctx, "commit", "abc")
	ctx = context.WithValue(ctx, "contract_name", "contractName")

	mock, host := myinit(t, ctx)

	mock.EXPECT().Checkout(Any()).Do(func(commit string) {
		if commit != "abc" {
			t.Fatal(commit)
		}
	})
	mock.EXPECT().Keys(Any(), Any()).DoAndReturn(func(a, b string) ([]string, error) {
		if a != "state" || b != "m-contractName-hello-" {
			t.Fatal(a, b)
		}
		return []string{"m-contractName-hello-a", "m-contractName-hello-b", "m-contractName-hello-c"}, nil
	})

	ans := host.MapKeys("hello")
	if !sliceEqual(ans, []string{"a", "b", "c"}) {
		t.Fatal(ans)
	}
}

func TestHost_RequireAuth(t *testing.T) {

	ctx := context.Background()
	ctx = context.WithValue(ctx, "commit", "abc")
	ctx = context.WithValue(ctx, "contract_name", "contractName")
	ctx = context.WithValue(ctx, "auth_list", map[string]int{"a": 1, "b": 0})

	_, host := myinit(t, ctx)

	ans := host.RequireAuth("a")
	if !ans {
		t.Fatal(ans)
	}
	ans = host.RequireAuth("b")
	if ans {
		t.Fatal(ans)
	}
	ans = host.RequireAuth("c")
	if ans {
		t.Fatal(ans)
	}
}

func TestHost_BlockInfo(t *testing.T) {

}
