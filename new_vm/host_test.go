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
	db := database.NewMockIMultiValue(mockCtrl)
	bdb := database.NewVisitor(100, db)

	monitor := Monitor{}
	return db, Host{ctx: ctx, db: bdb, monitor: &monitor}
}

func TestHost_Put(t *testing.T) {
	mock, host := myinit(t, nil)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "commit", "abc")
	ctx = context.WithValue(ctx, "contract_name", "contractName")

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

	host.LoadContext(ctx).Put("hello", "world")
}

func TestHost_Get(t *testing.T) {
	mock, host := myinit(t, nil)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "commit", "abc")
	ctx = context.WithValue(ctx, "contract_name", "contractName")

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

	ans := host.LoadContext(ctx).Get("hello")
	if ans != "world" {
		t.Fatal(ans)
	}
}

func TestHost_MapPut(t *testing.T) {
	mock, host := myinit(t, nil)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "commit", "abc")
	ctx = context.WithValue(ctx, "contract_name", "contractName")

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

	host.LoadContext(ctx).MapPut("hello", "1", "world")
}

func TestHost_MapGet(t *testing.T) {
	mock, host := myinit(t, nil)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "commit", "abc")
	ctx = context.WithValue(ctx, "contract_name", "contractName")

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

	ans := host.LoadContext(ctx).MapGet("hello", "1")
	if ans != "world" {
		t.Fatal(ans)
	}
}

func TestHost_MapKeys(t *testing.T) {
	mock, host := myinit(t, nil)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "commit", "abc")
	ctx = context.WithValue(ctx, "contract_name", "contractName")

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

	ans := host.LoadContext(ctx).MapKeys("hello")
	if !sliceEqual(ans, []string{"a", "b", "c"}) {
		t.Fatal(ans)
	}
}

func TestHost_RequireAuth(t *testing.T) {
	_, host := myinit(t, nil)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "commit", "abc")
	ctx = context.WithValue(ctx, "contract_name", "contractName")
	ctx = context.WithValue(ctx, "auth_list", map[string]int{"a": 1, "b": 0})

	ans := host.LoadContext(ctx).RequireAuth("a")
	if !ans {
		t.Fatal(ans)
	}
	ans = host.LoadContext(ctx).RequireAuth("b")
	if ans {
		t.Fatal(ans)
	}
	ans = host.LoadContext(ctx).RequireAuth("c")
	if ans {
		t.Fatal(ans)
	}
}

func TestHost_BlockInfo(t *testing.T) {

}
